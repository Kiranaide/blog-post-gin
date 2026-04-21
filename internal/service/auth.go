package service

import (
	"blog-post-gin/internal/dto"
	"blog-post-gin/internal/entity"
	"blog-post-gin/internal/helper"
	"blog-post-gin/internal/response"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AuthService struct {
	db *gorm.DB
	jc entity.JWTConfig
	ac helper.ArgonConfig
	sc entity.SessionCookieConfig
}

func NewAuthService(db *gorm.DB, jc entity.JWTConfig, ac helper.ArgonConfig, sc entity.SessionCookieConfig) *AuthService {
	return &AuthService{db: db, jc: jc, ac: ac, sc: sc}
}

func (s *AuthService) Register(req dto.RegisterRequest) (*response.RegisterResponse, string, error) {
	var existingUser entity.User

	err := s.db.Where("username = ?", req.Username).First(&existingUser).Error
	if err == nil {
		return nil, "", helper.ErrUsernameAlreadyTaken.WithMessage("Username is already taken")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", helper.ErrInternal.WithCause(err)
	}

	hashed, err := helper.HashPassword(req.Password, s.ac)
	if err != nil {
		return nil, "", helper.ErrInternal.WithCause(err)
	}

	user := entity.User{
		ID:       uuid.New(),
		Name:     req.Name,
		Username: req.Username,
		Password: hashed,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, "", helper.ErrInternal.WithCause(err)
	}

	data := response.RegisterResponse{
		User: response.ToUserResponse(user),
	}

	return &data, "User registered successfully", nil
}

func (s *AuthService) Login(req dto.LoginRequest, userAgent, ip string) (*response.LoginResponse, string, string, error) {
	var user entity.User

	if err := s.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		return nil, "", "", helper.ErrUserNotFound.WithMessage("Username not found")
	}

	verify, err := helper.VerifyPassword(req.Password, user.Password)
	if err != nil {
		return nil, "", "", helper.ErrInternal.WithCause(err)
	}
	if !verify {
		return nil, "", "", helper.ErrInvalidCredentials.WithMessage("Invalid username or password")
	}

	accessToken, err := s.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, "", "", helper.ErrInternal.WithCause(err)
	}

	refreshToken, err := s.createSession(user.ID, uuid.Nil, userAgent, ip, time.Now())
	if err != nil {
		return nil, "", "", helper.ErrInternal.WithCause(err)
	}

	data := response.ToLoginResponse(accessToken, user)

	return &data, refreshToken, "Login successful", nil
}

func (s *AuthService) Refresh(refreshToken, userAgent, ip string) (*response.RefreshResponse, string, string, error) {
	now := time.Now()
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, "", "", helper.ErrRefreshTokenInvalid.WithMessage("Refresh token is required")
	}

	tokenHash := s.hashRefreshToken(refreshToken)

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, "", "", helper.ErrInternal.WithCause(tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var session entity.Session
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("token_hash = ?", tokenHash).First(&session).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", "", helper.ErrRefreshTokenInvalid.WithMessage("Refresh token is invalid")
		}
		return nil, "", "", helper.ErrInternal.WithCause(err)
	}

	if session.RevokedAt != nil {
		tx.Rollback()
		return nil, "", "", helper.ErrRefreshTokenInvalid.WithMessage("Refresh token has been revoked")
	}

	if now.After(session.ExpiresAt) || now.After(session.FinalExpiresAt) {
		tx.Rollback()
		return nil, "", "", helper.ErrRefreshTokenExpired.WithMessage("Refresh token has expired")
	}

	if session.UsedAt != nil {
		if err := s.revokeFamilyTx(tx, session.FamilyID, now); err != nil {
			tx.Rollback()
			return nil, "", "", helper.ErrInternal.WithCause(err)
		}
		if err := tx.Commit().Error; err != nil {
			return nil, "", "", helper.ErrInternal.WithCause(err)
		}
		return nil, "", "", helper.ErrRefreshTokenInvalid.WithMessage("Refresh token has already been used")
	}

	newToken, err := s.generateOpaqueToken()
	if err != nil {
		tx.Rollback()
		return nil, "", "", helper.ErrInternal.WithCause(err)
	}

	newSession := entity.Session{
		ID:             uuid.New(),
		UserID:         session.UserID,
		FamilyID:       session.FamilyID,
		TokenHash:      s.hashRefreshToken(newToken),
		ExpiresAt:      now.Add(s.jc.RefreshExpiresAt),
		FinalExpiresAt: session.FinalExpiresAt,
		UserAgent:      userAgent,
		IPAddress:      ip,
	}

	if err := tx.Create(&newSession).Error; err != nil {
		tx.Rollback()
		return nil, "", "", helper.ErrInternal.WithCause(err)
	}

	update := tx.Model(&entity.Session{}).
		Where("id = ? and used_at IS NULL", session.ID).
		Updates(map[string]any{
			"used_at":        now,
			"rotated_at":     now,
			"replaced_by_id": newSession.ID,
		})
	if update.Error != nil {
		tx.Rollback()
		return nil, "", "", helper.ErrInternal.WithCause(update.Error)
	}
	if update.RowsAffected == 0 {
		if err := s.revokeFamilyTx(tx, session.FamilyID, now); err != nil {
			tx.Rollback()
			return nil, "", "", helper.ErrInternal.WithCause(err)
		}
		if err := tx.Commit().Error; err != nil {
			return nil, "", "", helper.ErrInternal.WithCause(err)
		}
		return nil, "", "", helper.ErrRefreshTokenInvalid.WithMessage("Refresh token has already been used")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, "", "", helper.ErrInternal.WithCause(err)
	}

	var user entity.User
	if err := s.db.Where("id = ?", session.UserID).First(&user).Error; err != nil {
		return nil, "", "", helper.ErrInternal.WithCause(err)
	}

	accessToken, err := s.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, "", "", helper.ErrInternal.WithCause(err)
	}

	data := response.ToRefreshResponse(accessToken)

	return &data, newToken, "Token refreshed successfully", nil
}

func (s *AuthService) Logout(refreshToken string) error {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return helper.ErrRefreshTokenInvalid.WithMessage("Refresh token is required")
	}

	now := time.Now()
	tokenHash := s.hashRefreshToken(refreshToken)

	if err := s.db.Model(&entity.Session{}).Where("token_hash = ?", tokenHash).Update("revoked_at", now).Error; err != nil {
		return helper.ErrInternal.WithCause(err)
	}

	return nil
}

func (s *AuthService) GenerateToken(userID uuid.UUID, username string) (string, error) {
	now := time.Now()

	claims := entity.JWTClaims{
		UserID:    userID.String(),
		Username:  username,
		TokenType: entity.AccessTokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			Issuer:    s.jc.Issuer,
			Audience:  jwt.ClaimStrings{s.jc.Audience},
			ID:        uuid.NewString(),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.jc.AccessExpiresAt)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(s.jc.AccessSecret))
	if err != nil {
		return "", helper.ErrInternal.WithCause(err)
	}

	return signedToken, nil
}

func (s *AuthService) ValidateToken(signedToken string) (*entity.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(signedToken, &entity.JWTClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, helper.ErrTokenInvalid.WithMessage("Unexpected signing method")
		}
		return []byte(s.jc.AccessSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, helper.ErrTokenExpired
		}
		return nil, helper.ErrTokenInvalid.WithCause(err)
	}

	claims, ok := token.Claims.(*entity.JWTClaims)
	if !ok || !token.Valid {
		return nil, helper.ErrTokenInvalid.WithMessage("Invalid token claims")
	}
	if claims.TokenType != entity.AccessTokenType {
		return nil, helper.ErrTokenInvalid.WithMessage("Invalid token type")
	}
	if claims.Issuer != s.jc.Issuer {
		return nil, helper.ErrTokenInvalid.WithMessage("Invalid token issuer")
	}

	return claims, nil
}

func (s *AuthService) SessionCookieConfig() entity.SessionCookieConfig {
	return s.sc
}

func (s *AuthService) SessionCookieMaxAgeSeconds() int {
	return int(s.jc.RefreshExpiresAt.Seconds())
}

func (s *AuthService) revokeFamilyTx(tx *gorm.DB, familyID uuid.UUID, now time.Time) error {
	return tx.Model(&entity.Session{}).Where("family_id = ? and revoked_at IS NULL", familyID).Update("revoked_at", now).Error
}

func (s *AuthService) createSession(userID, familyID uuid.UUID, userAgent, ip string, now time.Time) (string, error) {
	rawToken, err := s.generateOpaqueToken()
	if err != nil {
		return "", err
	}
	if familyID == uuid.Nil {
		familyID = uuid.New()
	}

	session := entity.Session{
		ID:             uuid.New(),
		UserID:         userID,
		FamilyID:       familyID,
		TokenHash:      s.hashRefreshToken(rawToken),
		ExpiresAt:      now.Add(s.jc.RefreshExpiresAt),
		FinalExpiresAt: now.Add(s.jc.FinalExpiresAt),
		UserAgent:      userAgent,
		IPAddress:      ip,
	}

	if err := s.db.Create(&session).Error; err != nil {
		return "", err
	}

	return rawToken, nil
}

func (s *AuthService) generateOpaqueToken() (string, error) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func (s *AuthService) hashRefreshToken(token string) string {
	mac := hmac.New(sha256.New, []byte(s.jc.RefreshSecret))
	mac.Write([]byte(token))
	return hex.EncodeToString(mac.Sum(nil))
}

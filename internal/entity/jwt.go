package entity

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTConfig struct {
	AccessSecret     string
	RefreshSecret    string
	AccessExpiresAt  time.Duration
	RefreshExpiresAt time.Duration
	FinalExpiresAt   time.Duration
	Issuer           string
	Audience         string
}

type SessionCookieConfig struct {
	Name     string
	Domain   string
	Path     string
	Secure   bool
	HTTPOnly bool
	SameSite string
}

type JWTClaims struct {
	UserID    string `json:"userId" type:"uuid"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	TokenType string `json:"tokenType"`
	jwt.RegisteredClaims
}

const (
	AccessTokenType  = "access"
	RefreshTokenType = "refresh"
)

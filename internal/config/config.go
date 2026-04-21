package config

import (
	"blog-post-gin/internal/entity"
	"blog-post-gin/internal/helper"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost                   string
	DBPort                   string
	DBUser                   string
	DBPassword               string
	DBName                   string
	DBSSLMode                string
	AppPort                  string
	JWTSecret                string
	JWTExpiresAt             string
	JWTRefreshSecret         string
	JWTRefreshExpiresAt      string
	JWTRefreshFinalExpiresAt string
	JWTIssuer                string
	JWTAudience              string
	ArgonTime                string
	ArgonMemory              string
	ArgonThreads             string
	ArgonKeyLength           string
	ArgonSaltLength          string
	SessionCookieName        string
	SessionCookieDomain      string
	SessionCookiePath        string
	SessionCookieHTTPOnly    string
	SessionCookieSecure      string
	SessionCookieSameSite    string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using systme env")
	}

	return &Config{
		DBHost:                   os.Getenv("DB_HOST"),
		DBPort:                   os.Getenv("DB_PORT"),
		DBUser:                   os.Getenv("DB_USER"),
		DBPassword:               os.Getenv("DB_PASSWORD"),
		DBName:                   os.Getenv("DB_NAME"),
		DBSSLMode:                os.Getenv("DB_SSLMODE"),
		AppPort:                  os.Getenv("APP_PORT"),
		JWTSecret:                os.Getenv("JWT_SECRET"),
		JWTExpiresAt:             os.Getenv("JWT_EXPIRES_AT"),
		JWTRefreshSecret:         os.Getenv("JWT_REFRESH_SECRET"),
		JWTRefreshExpiresAt:      os.Getenv("JWT_REFRESH_EXPIRES_AT"),
		JWTRefreshFinalExpiresAt: os.Getenv("JWT_REFRESH_FINAL_EXPIRES_AT"),
		JWTIssuer:                os.Getenv("JWT_ISSUER"),
		JWTAudience:              os.Getenv("JWT_AUDIENCE"),
		ArgonTime:                os.Getenv("ARGON_TIME"),
		ArgonMemory:              os.Getenv("ARGON_MEMORY"),
		ArgonThreads:             os.Getenv("ARGON_THREADS"),
		ArgonKeyLength:           os.Getenv("ARGON_KEYLENGTH"),
		ArgonSaltLength:          os.Getenv("ARGON_SALTLENGTH"),
		SessionCookieName:        os.Getenv("SESSION_COOKIE_NAME"),
		SessionCookieDomain:      os.Getenv("SESSION_COOKIE_DOMAIN"),
		SessionCookiePath:        os.Getenv("SESSION_COOKIE_PATH"),
		SessionCookieSecure:      os.Getenv("SESSION_COOKIE_SECURE"),
		SessionCookieHTTPOnly:    os.Getenv("SESSION_COOKIE_HTTP_ONLY"),
		SessionCookieSameSite:    os.Getenv("SESSION_COOKIE_SAME_SITE"),
	}
}

func ParseArgonConfig(cfg *Config) (helper.ArgonConfig, error) {
	time, err := strconv.ParseUint(cfg.ArgonTime, 10, 32)
	if err != nil {
		return helper.ArgonConfig{}, fmt.Errorf("invalid ArgonTime: %w", err)
	}

	memory, err := strconv.ParseUint(cfg.ArgonMemory, 10, 32)
	if err != nil {
		return helper.ArgonConfig{}, fmt.Errorf("invalid ArgonMemory: %w", err)
	}

	threads, err := strconv.ParseUint(cfg.ArgonThreads, 10, 8)
	if err != nil {
		return helper.ArgonConfig{}, fmt.Errorf("invalid ArgonThreads: %w", err)
	}

	keyLength, err := strconv.ParseUint(cfg.ArgonKeyLength, 10, 32)
	if err != nil {
		return helper.ArgonConfig{}, fmt.Errorf("invalid ArgonKeyLength: %w", err)
	}

	saltLength, err := strconv.ParseUint(cfg.ArgonSaltLength, 10, 32)
	if err != nil {
		return helper.ArgonConfig{}, fmt.Errorf("invalid ArgonSaltLength: %w", err)
	}

	return helper.ArgonConfig{
		Time:       uint32(time),
		Memory:     uint32(memory),
		Threads:    uint8(threads),
		KeyLength:  uint32(keyLength),
		SaltLength: uint32(saltLength),
	}, nil
}

func ParseJWTConfig(cfg *Config) (entity.JWTConfig, entity.SessionCookieConfig, error) {
	accessTTL, err := time.ParseDuration(cfg.JWTExpiresAt)
	if err != nil {
		return entity.JWTConfig{}, entity.SessionCookieConfig{}, fmt.Errorf("invalid JWTExpiresAt: %w", err)
	}

	refreshTTL, err := time.ParseDuration(cfg.JWTRefreshExpiresAt)
	if err != nil {
		return entity.JWTConfig{}, entity.SessionCookieConfig{}, fmt.Errorf("invalid JWTRefreshExpiresAt: %w", err)
	}

	refreshAbsoluteTTL, err := time.ParseDuration(cfg.JWTRefreshFinalExpiresAt)
	if err != nil {
		return entity.JWTConfig{}, entity.SessionCookieConfig{}, fmt.Errorf("invalid JWTRefreshFinalExpiresAt: %w", err)
	}

	secure, err := strconv.ParseBool(cfg.SessionCookieSecure)
	if err != nil {
		return entity.JWTConfig{}, entity.SessionCookieConfig{}, fmt.Errorf("invalid SessionCookieSecure: %w", err)
	}

	httpOnly, err := strconv.ParseBool(cfg.SessionCookieHTTPOnly)
	if err != nil {
		return entity.JWTConfig{}, entity.SessionCookieConfig{}, fmt.Errorf("invalid SessionCookieHTTPOnly: %w", err)
	}

	jc := entity.JWTConfig{
		AccessSecret:     cfg.JWTSecret,
		RefreshSecret:    cfg.JWTRefreshSecret,
		AccessExpiresAt:  accessTTL,
		RefreshExpiresAt: refreshTTL,
		FinalExpiresAt:   refreshAbsoluteTTL,
		Issuer:           cfg.JWTIssuer,
		Audience:         cfg.JWTAudience,
	}

	sc := entity.SessionCookieConfig{
		Name:     cfg.SessionCookieName,
		Domain:   cfg.SessionCookieDomain,
		Path:     cfg.SessionCookiePath,
		Secure:   secure,
		HTTPOnly: httpOnly,
		SameSite: cfg.SessionCookieSameSite,
	}

	if strings.TrimSpace(jc.AccessSecret) == "" {
		return entity.JWTConfig{}, entity.SessionCookieConfig{}, fmt.Errorf("JWT_SECRET cannot be empty")
	}

	if strings.TrimSpace(jc.RefreshSecret) == "" {
		return entity.JWTConfig{}, entity.SessionCookieConfig{}, fmt.Errorf("JWT_REFRESH_SECRET cannot be empty")
	}

	return jc, sc, nil
}

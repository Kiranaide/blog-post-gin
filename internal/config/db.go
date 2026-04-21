package config

import (
	"blog-post-gin/internal/entity"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB(cfg *Config) (*gorm.DB, error) {
	params := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBPort,
		cfg.DBSSLMode)

	db, err := gorm.Open(postgres.Open(params), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sql, err := db.DB()
	if err != nil {
		return nil, err
	}

	sql.SetMaxOpenConns(10)
	sql.SetMaxIdleConns(5)
	sql.SetConnMaxLifetime(time.Hour)
	sql.SetConnMaxIdleTime(30 * time.Minute)

	return db, nil
}

func RunMigrations(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&entity.Role{},
		&entity.User{},
		&entity.Session{},
		&entity.Post{},
		&entity.Comment{},
		&entity.Tag{},
		&entity.PostTag{},
	); err != nil {
		return err
	}

	for _, roleName := range []string{"admin", "author", "reader"} {
		role := entity.Role{Name: roleName}
		if err := db.Where("name = ?", roleName).FirstOrCreate(&role).Error; err != nil {
			return err
		}
	}

	return nil
}

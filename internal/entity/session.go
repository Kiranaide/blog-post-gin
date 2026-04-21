package entity

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID             uuid.UUID  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID         uuid.UUID  `gorm:"type:uuid;index;not null"`
	FamilyID       uuid.UUID  `gorm:"type:uuid;index;not null"`
	TokenHash      string     `gorm:"uniqueIndex;not null"`
	ExpiresAt      time.Time  `gorm:"index;not null"`
	FinalExpiresAt time.Time  `gorm:"index;not null"`
	UsedAt         *time.Time `gorm:"index"`
	RotatedAt      *time.Time `gorm:"index"`
	RevokedAt      *time.Time `gorm:"index"`
	ReplacedByID   *uuid.UUID `gorm:"type:uuid;index"`
	UserAgent      string     `gorm:"text"`
	IPAddress      string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (Session) TableName() string {
	return "sessions"
}

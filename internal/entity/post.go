package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Post struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Title     string
	Slug      string    `gorm:"unique"`
	AuthorID  uuid.UUID `gorm:"type:uuid"`
	Author    User      `gorm:"foreignKey:AuthorID;references:ID;constraint:OnUpdate:CASCADE;OnDelete:CASCADE"`
	Tags      []*Tag    `gorm:"many2many:post_tags;constraint:OnUpdate:CASCADE;OnDelete:CASCADE"`
	Comments  []Comment `gorm:"foreignKey:PostID"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

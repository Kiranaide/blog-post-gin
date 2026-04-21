package entity

import (
	"github.com/google/uuid"
)

type Comment struct {
	ID      uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID  uuid.UUID `gorm:"primaryKey;type:uuid"`
	User    User      `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	PostID  uuid.UUID `gorm:"primaryKey;type:uuid"`
	Post    Post      `gorm:"foreignKey:PostID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Comment string
}

package entity

import "github.com/google/uuid"

type Tag struct {
	ID    uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name  string    `gorm:"unique"`
	Posts []*Post   `gorm:"many2many:post_tags"`
}

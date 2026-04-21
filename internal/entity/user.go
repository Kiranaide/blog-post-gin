package entity

import (
	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name     string
	RoleID   uuid.UUID `gorm:"type:uuid;not null;index"`
	Role     Role      `gorm:"foreignKey:RoleID;type:uuid"`
	Username string    `gorm:"unique"`
	Password string
	Posts    []Post `gorm:"foreignKey:AuthorID"`
}

type Role struct {
	ID   uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name string    `gorm:"unique"`
}

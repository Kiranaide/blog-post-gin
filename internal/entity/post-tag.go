package entity

import (
	"time"

	"github.com/google/uuid"
)

type PostTag struct {
	PostID    uuid.UUID `gorm:"primaryKey;type:uuid"`
	TagID     uuid.UUID `gorm:"primaryKey;type:uuid"`
	CreatedAt time.Time
}

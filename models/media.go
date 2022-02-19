package models

import (
	"database/sql"

	"gorm.io/gorm"
)

type MediaItem struct{
	gorm.Model
	MediaID string `gorm:"not null uniqueIndex"`
	ThumbnailURL sql.NullString
	DisplayName sql.NullString
	CanonicalURL sql.NullString
	Data sql.NullString
}
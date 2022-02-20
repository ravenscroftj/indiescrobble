package models

import (
	"database/sql"

	"gorm.io/gorm"
)

type Post struct {
	gorm.Model
	URL string `gorm:"uniqueIndex"`
	PostType string `gorm:"index"`
	UserID int `gorm:"foreignKey"`
	User User
	MediaItemID int `gorm:"index"`
	MediaItem MediaItem
	ScrobbledAt sql.NullTime
	Content sql.NullString
	Rating sql.NullString
}
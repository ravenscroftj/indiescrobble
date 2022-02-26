package models

import (
	"math/rand"

	"gorm.io/gorm"
)

var alphanum = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = alphanum[rand.Intn(len(alphanum))]
	}
	return string(b)
}

type BaseUser struct {
	Me         string
	Token      string
	UserRecord *User
}

type User struct {
	gorm.Model
	Me     string `gorm:"uniqueIndex"`
	APIKey string
	DefaultSharePost bool `gorm:"default:true"`
	DefaultEnableWatchOf bool `gorm:"default:true"`
}

func (u *User) GenerateRandomKey() {
	u.APIKey = randSeq(32)
}

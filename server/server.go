package server

import (
	"fmt"
	"log"

	"git.jamesravey.me/ravenscroftj/indiescrobble/config"
	"git.jamesravey.me/ravenscroftj/indiescrobble/models"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Init() {
	config := config.GetConfig()

	var dialect gorm.Dialector

	if config.GetString("server.database.driver") == "sqlite" {
		dialect = sqlite.Open(config.GetString("server.database.dsn"))
	} else {
		dialect = mysql.Open(config.GetString("server.database.dsn"))
	}

	db, err := gorm.Open(dialect, &gorm.Config{})

	if err != nil {
		log.Fatalf("%v\n", err)
	}

	db.AutoMigrate(&models.User{}, &models.Post{}, &models.MediaItem{})

	r := NewRouter(db)
	r.LoadHTMLGlob("templates/*.tmpl")
	r.Run(fmt.Sprintf("%v:%v", config.GetString("server.host"), config.GetString("server.port")))
}

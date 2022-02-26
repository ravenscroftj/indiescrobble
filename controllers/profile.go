package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ravenscroftj/indiescrobble/models"
	"gorm.io/gorm"
)

type UserProfileController struct {
	db *gorm.DB
}

func NewUserProfileController(db *gorm.DB) *UserProfileController{
	return &UserProfileController{
		db: db,
	}
}

func (u *UserProfileController) GetConfig(c *gin.Context){

	// this is an authed endpoint so 'user' must be set and if not panicking is fair
	currentUser := c.MustGet("user").(*models.BaseUser)

	// render user config
	c.HTML(http.StatusOK, "profile/config.tmpl", gin.H{
		"user":                currentUser,
	})
}

func (u *UserProfileController) SaveConfig(c *gin.Context){

	err := c.Request.ParseForm()

	if err != nil{
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"message": err,
		})
		return
	}
	
	// this is an authed endpoint so 'user' must be set and if not panicking is fair
	currentUser := c.MustGet("user").(*models.BaseUser)

	fmt.Printf("Form content: %v\n",c.Request.Form.Encode())

	currentUser.UserRecord.DefaultSharePost = c.Request.Form.Get("default_share_posts") == "1"
	currentUser.UserRecord.DefaultEnableWatchOf = c.Request.Form.Get("default_enable_watchof") == "1"

	tx := u.db.Save(currentUser.UserRecord)

	if tx.Error != nil{
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"message": tx.Error,
		})
		return
	}
	

	// render user config
	c.HTML(http.StatusOK, "profile/config.tmpl", gin.H{
		"user":                currentUser,
	})
}
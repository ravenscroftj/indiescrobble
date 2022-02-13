package controllers

import (
	"net/http"

	"git.jamesravey.me/ravenscroftj/indiescrobble/models"
	"git.jamesravey.me/ravenscroftj/indiescrobble/scrobble"
	"github.com/gin-gonic/gin"
)

func Index(c *gin.Context) {

	// this is an authed endpoint so 'user' must be set and if not panicking is fair
	currentUser, exists := c.Get("user")

	var user *models.BaseUser

	if exists {
		user = currentUser.(*models.BaseUser)
	}else{
		user = nil
	}

	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"title":         "test",
		"user":          user,
		"scrobbleTypes": scrobble.ScrobbleTypeNames,
	})
}

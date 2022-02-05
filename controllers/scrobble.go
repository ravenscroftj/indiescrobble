package controllers

import (
	"net/http"

	"git.jamesravey.me/ravenscroftj/indiescrobble/scrobble"
	"github.com/gin-gonic/gin"
)


func Scrobble(c *gin.Context){

	err := c.Request.ParseForm()

	if err != nil{
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"message": err,
		})
	}

	// TODO: add validation of type
	scrobbleType := c.Request.Form.Get("type")

	c.HTML(http.StatusOK, "scrobble.tmpl", gin.H{
		"user": c.GetString("user"),
		"scrobbleType": scrobble.ScrobbleTypeNames[scrobbleType],
	})

}
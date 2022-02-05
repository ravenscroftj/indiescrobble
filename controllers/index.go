package controllers

import (
	"net/http"

	"git.jamesravey.me/ravenscroftj/indiescrobble/scrobble"
	"github.com/gin-gonic/gin"
)

func Index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"title": "test",
		"user": c.GetString("user"),
		"scrobbleTypes": scrobble.ScrobbleTypeNames,
	})
}

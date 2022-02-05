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

	searchEngine := scrobble.NewSearchProvider(scrobbleType)

	var searchResults []scrobble.ScrobbleMetaRecord = nil
	var item scrobble.ScrobbleMetaRecord = nil

	query := c.Request.Form.Get("q")
	itemID := c.Request.Form.Get("item")

	if itemID != "" {
		
		item, err = searchEngine.SearchProvider.GetItem(itemID)

		if err != nil{
			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
				"message": err,
			})
			return
		}
	}else if query != "" {
		var err error = nil
		searchResults, err = searchEngine.SearchProvider.Search(query)

		if err != nil{
			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
				"message": err,
			})
			return
		}
	}

	c.HTML(http.StatusOK, "scrobble.tmpl", gin.H{
		"user": c.GetString("user"),
		"scrobbleType": scrobbleType,
		"scrobblePlaceholder":  scrobble.ScrobblePlaceholders[scrobbleType],
		"scrobbleTypeName": scrobble.ScrobbleTypeNames[scrobbleType],
		"searchEngine": searchEngine.SearchProvider.GetName(),
		"searchResults": searchResults,
		"item": item,
	})

}
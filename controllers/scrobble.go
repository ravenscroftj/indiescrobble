package controllers

import (
	"net/http"

	"git.jamesravey.me/ravenscroftj/indiescrobble/models"
	"git.jamesravey.me/ravenscroftj/indiescrobble/scrobble"
	"git.jamesravey.me/ravenscroftj/indiescrobble/services/micropub"
	"github.com/gin-gonic/gin"
)


func Scrobble(c *gin.Context){

	err := c.Request.ParseForm()

	// this is an authed endpoint so 'user' must be set and if not panicking is fair
	currentUser := c.MustGet("user").(*models.BaseUser)

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
		"user": currentUser,
		"scrobbleType": scrobbleType,
		"scrobblePlaceholder":  scrobble.ScrobblePlaceholders[scrobbleType],
		"scrobbleTypeName": scrobble.ScrobbleTypeNames[scrobbleType],
		"searchEngine": searchEngine.SearchProvider.GetName(),
		"searchResults": searchResults,
		"item": item,
	})

}

func PreviewScrobble(c *gin.Context){

	err := c.Request.ParseForm()

	// this is an authed endpoint so 'user' must be set and if not panicking is fair
	currentUser := c.MustGet("user").(*models.BaseUser)

	if err != nil{
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"message": err,
		})
	}

	scrobbleType := c.Request.Form.Get("type")

	searchEngine := scrobble.NewSearchProvider(scrobbleType)

	itemID := c.Request.Form.Get("item")

	item, err := searchEngine.SearchProvider.GetItem(itemID)

	if err != nil{
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"message": err,
		})
		return
	}

	discovery := micropub.MicropubDiscoveryService{}
	
	discovery.Discover(currentUser.Me, currentUser.Token )

	c.HTML(http.StatusOK, "preview.tmpl", gin.H{
		"user": currentUser,
		"scrobbleType": scrobbleType,
		"scrobbleTypeName": scrobble.ScrobbleTypeNames[scrobbleType],
		"item": item,
		"when": c.Request.Form.Get("when"),
		"rating": c.Request.Form.Get("rating"),
		"content": c.Request.Form.Get("content"),
	})

}
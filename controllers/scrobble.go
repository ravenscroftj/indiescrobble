package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ravenscroftj/indiescrobble/config"
	"github.com/ravenscroftj/indiescrobble/models"
	"github.com/ravenscroftj/indiescrobble/services/micropub"
	"github.com/ravenscroftj/indiescrobble/services/scrobble"
	"gorm.io/gorm"
)

type ScrobbleController struct {
	db        *gorm.DB
	scrobbler *scrobble.Scrobbler
}

func NewScrobbleController(db *gorm.DB) *ScrobbleController {
	return &ScrobbleController{db, scrobble.NewScrobbler(db)}
}

/*Do the actual post to the user's site*/
func (s *ScrobbleController) DoScrobble(c *gin.Context) {

	err := c.Request.ParseForm()

	// this is an authed endpoint so 'user' must be set and if not panicking is fair
	currentUser := c.MustGet("user").(*models.BaseUser)

	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"message": err,
		})
		return
	}

	post, err := s.scrobbler.Scrobble(&c.Request.Form, currentUser)

	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"message": err,
		})
		return
	}

	c.HTML(http.StatusOK, "scrobble/done.tmpl", gin.H{
		"user":             currentUser,
		"scrobbleTypeName": scrobble.ScrobbleTypeNames[post.PostType],
		"post":             post,
		"title": 			"Preview Post",
	})
}

/*Display the scrobble form and allow user to search for and add media*/
func (s *ScrobbleController) ScrobbleForm(c *gin.Context) {

	err := c.Request.ParseForm()

	// this is an authed endpoint so 'user' must be set and if not panicking is fair
	currentUser := c.MustGet("user").(*models.BaseUser)

	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"message": err,
		})
		return
	}

	scrobbleType := c.Request.Form.Get("type")

	if c.Request.Form.Get("item") != "" {

		item, err := s.scrobbler.GetItemByID(&c.Request.Form)

		if err != nil {
			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
				"message": err,
			})
			return
		} else {
			c.HTML(http.StatusOK, "scrobble/compose.tmpl", gin.H{
				"user":                currentUser,
				"scrobbleType":        scrobbleType,
				"scrobblePlaceholder": scrobble.ScrobblePlaceholders[scrobbleType],
				"scrobbleTypeName":    scrobble.ScrobbleTypeNames[scrobbleType],
				"item":                item,
				"now":                 time.Now().Format(config.BROWSER_TIME_FORMAT),
				"title": 			   "Compose a Post",
			})
			return
		}

	} else if query := c.Request.Form.Get("q"); query != "" {

		searchResults, err := s.scrobbler.Search(&c.Request.Form)

		if err != nil {
			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
				"message": err,
			})
			return
		}

		c.HTML(http.StatusOK, "scrobble/search.tmpl", gin.H{
			"user":                currentUser,
			"scrobbleType":        scrobbleType,
			"scrobblePlaceholder": scrobble.ScrobblePlaceholders[scrobbleType],
			"scrobbleTypeName":    scrobble.ScrobbleTypeNames[scrobbleType],
			"searchEngine":        s.scrobbler.GetSearchEngineNameForType(scrobbleType),
			"searchResults":       searchResults,
			"now":                 time.Now().Format("2006-01-02T15:04"),
			"title":               "Add A Post",
		})
	} else if scrobbleType := c.Request.Form.Get("type"); scrobbleType != "" {
		c.HTML(http.StatusOK, "scrobble/search.tmpl", gin.H{
			"user":                currentUser,
			"scrobbleType":        scrobbleType,
			"scrobblePlaceholder": scrobble.ScrobblePlaceholders[scrobbleType],
			"scrobbleTypeName":    scrobble.ScrobbleTypeNames[scrobbleType],
			"now":                 time.Now().Format("2006-01-02T15:04"),
			"title":               "Add A Post",
		})
	} else {
		c.HTML(http.StatusOK, "scrobble/begin.tmpl", gin.H{
			"user":          currentUser,
			"scrobbleTypes": scrobble.ScrobbleTypeNames,
			"now":           time.Now().Format("2006-01-02T15:04"),
			"title":         "Add A Post",
		})
	}

}

/*Preview the content of a scrobble to be submitted to \*/
func (s *ScrobbleController) PreviewScrobble(c *gin.Context) {

	err := c.Request.ParseForm()

	// this is an authed endpoint so 'user' must be set and if not panicking is fair
	currentUser := c.MustGet("user").(*models.BaseUser)

	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"message": err,
		})
	}

	post, err := s.scrobbler.Preview(&c.Request.Form)

	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"message": err,
		})
	}

	scrobbleType := c.Request.Form.Get("type")

	discovery := micropub.MicropubDiscoveryService{}

	config, err := discovery.Discover(currentUser.Me, currentUser.Token)

	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"message": err,
		})
		return
	}

	postBody, err := s.scrobbler.BuildMicroPubPayload(post)

	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"message": err,
		})
		return
	}

	c.HTML(http.StatusOK, "scrobble/preview.tmpl", gin.H{
		"user":             currentUser,
		"scrobbleType":     scrobbleType,
		"scrobbleTypeName": scrobble.ScrobbleTypeNames[scrobbleType],
		"post":             post,
		"config":           config,
		"summary":          s.scrobbler.GenerateSummary(post),
		"postBody":         string(postBody),
		"title": fmt.Sprintf("Post Preview: %v", post.MediaItem.DisplayName.String),
	})

}

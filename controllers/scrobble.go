package controllers

import (
	"net/http"
	"time"

	"git.jamesravey.me/ravenscroftj/indiescrobble/models"
	"git.jamesravey.me/ravenscroftj/indiescrobble/services/scrobble"
	"git.jamesravey.me/ravenscroftj/indiescrobble/services/micropub"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)


type ScrobbleController struct{
	db *gorm.DB
	scrobbler *scrobble.Scrobbler
}

func NewScrobbleController(db *gorm.DB) *ScrobbleController{
	return &ScrobbleController{db, scrobble.NewScrobbler(db)}
}

/*Do the actual post to the user's site*/
func (s *ScrobbleController) DoScrobble(c *gin.Context){

	err := c.Request.ParseForm()

	// this is an authed endpoint so 'user' must be set and if not panicking is fair
	currentUser := c.MustGet("user").(*models.BaseUser)

	if err != nil{
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"message": err,
		})
		return
	}

	s.scrobbler.Scrobble(&c.Request.Form, currentUser)
}



/*Display the scrobble form and allow user to search for and add media*/
func (s *ScrobbleController) ScrobbleForm(c *gin.Context){

	err := c.Request.ParseForm()

	// this is an authed endpoint so 'user' must be set and if not panicking is fair
	currentUser := c.MustGet("user").(*models.BaseUser)

	if err != nil{
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
		}else{
			c.HTML(http.StatusOK, "scrobble.tmpl", gin.H{
				"user": currentUser,
				"scrobbleType": scrobbleType,
				"scrobblePlaceholder":  scrobble.ScrobblePlaceholders[scrobbleType],
				"scrobbleTypeName": scrobble.ScrobbleTypeNames[scrobbleType],
				"item": item,
				"now": time.Now().Format("2006-01-02T15:04"),
			})
			return
		}

	}else if query := c.Request.Form.Get("q"); query != "" {

		searchResults, err := s.scrobbler.Search(&c.Request.Form)

		if err != nil{
			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
				"message": err,
			})
			return
		}

		c.HTML(http.StatusOK, "scrobble.tmpl", gin.H{
			"user": currentUser,
			"scrobbleType": scrobbleType,
			"scrobblePlaceholder":  scrobble.ScrobblePlaceholders[scrobbleType],
			"scrobbleTypeName": scrobble.ScrobbleTypeNames[scrobbleType],
			"searchEngine": s.scrobbler.GetSearchEngineNameForType(scrobbleType),
			"searchResults": searchResults,
			"now": time.Now().Format("2006-01-02T15:04"),
		})
	}else{
		c.HTML(http.StatusOK, "scrobble.tmpl", gin.H{
			"user": currentUser,
			"scrobbleType": scrobbleType,
			"scrobblePlaceholder":  scrobble.ScrobblePlaceholders[scrobbleType],
			"scrobbleTypeName": scrobble.ScrobbleTypeNames[scrobbleType],
			"now": time.Now().Format("2006-01-02T15:04"),
		})
	}

}

func (s *ScrobbleController) PreviewScrobble(c *gin.Context){

	err := c.Request.ParseForm()

	// this is an authed endpoint so 'user' must be set and if not panicking is fair
	currentUser := c.MustGet("user").(*models.BaseUser)

	if err != nil{
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"message": err,
		})
	}

	scrobbleType := c.Request.Form.Get("type")

	searchEngine := scrobble.NewSearchProvider(scrobbleType, s.db)

	itemID := c.Request.Form.Get("item")

	item, err := searchEngine.SearchProvider.GetItem(itemID)

	if err != nil{
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"message": err,
		})
		return
	}

	discovery := micropub.MicropubDiscoveryService{}
	
	config, err := discovery.Discover(currentUser.Me, currentUser.Token )

	if err != nil{
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"message": err,
		})
		return
	}

	c.HTML(http.StatusOK, "preview.tmpl", gin.H{
		"user": currentUser,
		"scrobbleType": scrobbleType,
		"scrobbleTypeName": scrobble.ScrobbleTypeNames[scrobbleType],
		"item": item,
		"when": c.Request.Form.Get("when"),
		"rating": c.Request.Form.Get("rating"),
		"content": c.Request.Form.Get("content"),
		"config": config,
	})

}
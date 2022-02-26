package controllers

import (
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ravenscroftj/indiescrobble/models"
	"gorm.io/gorm"
)

type UserProfileController struct {
	db *gorm.DB
}

func NewUserProfileController(db *gorm.DB) *UserProfileController {
	return &UserProfileController{
		db: db,
	}
}

func (u *UserProfileController) GetConfig(c *gin.Context) {

	// this is an authed endpoint so 'user' must be set and if not panicking is fair
	currentUser := c.MustGet("user").(*models.BaseUser)

	// render user config
	c.HTML(http.StatusOK, "profile/config.tmpl", gin.H{
		"user": currentUser,
	})
}

func (u *UserProfileController) SaveConfig(c *gin.Context) {

	err := c.Request.ParseForm()

	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"message": err,
		})
		return
	}

	// this is an authed endpoint so 'user' must be set and if not panicking is fair
	currentUser := c.MustGet("user").(*models.BaseUser)

	fmt.Printf("Form content: %v\n", c.Request.Form.Encode())

	currentUser.UserRecord.DefaultSharePost = c.Request.Form.Get("default_share_posts") == "1"
	currentUser.UserRecord.DefaultEnableWatchOf = c.Request.Form.Get("default_enable_watchof") == "1"

	tx := u.db.Save(currentUser.UserRecord)

	if tx.Error != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"message": tx.Error,
		})
		return
	}

	// render user config
	c.HTML(http.StatusOK, "profile/config.tmpl", gin.H{
		"user": currentUser,
	})
}

func (u *UserProfileController) ViewUserPosts(c *gin.Context) {

	var err error  = c.Request.ParseForm()
	pageLimit := 10
	page := 0
	offset := 0


	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"message": err,
		})
		return
	}

	// this is an authed endpoint so 'user' must be set and if not panicking is fair
	currentUser := c.MustGet("user").(*models.BaseUser)


	

	if c.Request.Form.Get("limit") != ""{
		pageLimit, err = strconv.Atoi(c.Request.Form.Get("limit"))

		if err != nil {
			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
				"message": err,
			})
			return
		}

		pageLimit = int(math.Min(float64(100), float64(pageLimit)))
	}

	if c.Request.Form.Get("page") != ""{
		page, err = strconv.Atoi(c.Request.Form.Get("page"))

		if err != nil {
			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
				"message": err,
			})
			return
		}
		
		offset = page * pageLimit
	}

	query := u.db.Model(&models.Post{}).Where(&models.Post{UserID: currentUser.UserRecord.ID}).Order("posts.created_at DESC")

	var postCount int64 = 0

	tx := query.Count(&postCount)

	if tx.Error != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"message": tx.Error,
		})
		return
	}

	posts := []models.Post{}

	tx = query.Limit(pageLimit).Offset(offset).Joins("MediaItem").Find(&posts)

	if tx.Error != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"message": tx.Error,
		})
		return
	}




	// render user config
	c.HTML(http.StatusOK, "profile/posts.tmpl", gin.H{
		"user":  currentUser,
		"count": postCount,
		"posts": posts,
		"pageLimit": pageLimit,
		"nextLink": (offset + pageLimit) < int(postCount),
		"prevLink": offset > 0,
		"nextPage": page+1,
		"prevPage": page-1,
	})
}

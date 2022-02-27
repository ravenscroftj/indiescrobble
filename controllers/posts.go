package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ravenscroftj/indiescrobble/models"
	"gorm.io/gorm"
)

/* Controllers concerning accessing user posts */

type PostsController struct{
	db *gorm.DB
}

func NewPostsController(db *gorm.DB) *PostsController {
	return &PostsController{db:db}
}

func (p *PostsController) ViewPost(c *gin.Context) {

	currentUserIface, exists := c.Get("user")
	var currentUser *models.BaseUser = nil

	if exists {
		currentUser = currentUserIface.(*models.BaseUser)
	}

	postID := c.Param("postID")

	post := models.Post{}

	tx := p.db.Model(&models.Post{}).Joins("MediaItem").First(&post, "posts.id = ?", postID)

	if tx.Error != nil{
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"message": tx.Error,
		})
		return
	}

	// make sure user has permission to see the post
	if !post.SharePost && (currentUser == nil || currentUser.UserRecord.ID != post.UserID) {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"message": "You are not permitted to view this post",
		})
		return
	}

	// render user config
	c.HTML(http.StatusOK, "posts/single.tmpl", gin.H{
		"post": post,
		"user": currentUser,
	})



}
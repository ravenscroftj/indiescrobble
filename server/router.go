package server

import (
	"git.jamesravey.me/ravenscroftj/indiescrobble/config"
	"git.jamesravey.me/ravenscroftj/indiescrobble/controllers"
	"git.jamesravey.me/ravenscroftj/indiescrobble/middlewares"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	config := config.GetConfig()

	health := new(controllers.HealthController)

	iam := controllers.NewIndieAuthManager(db)

	router.GET("/health", health.Status)

	router.Use(middlewares.AuthMiddleware(false, iam))

	router.GET("/", controllers.Index)

	router.Static("/static", config.GetString("server.static_path"))

	// add auth endpoints

	router.POST("/indieauth", iam.IndieAuthLoginPost)
	router.GET("/auth", iam.LoginCallbackGet)
	router.GET("/logout", iam.Logout)

	authed := router.Use(middlewares.AuthMiddleware(true, iam))

	// add scrobble endpoints
	scrobbleController := controllers.NewScrobbleController(db)

	authed.GET("/scrobble", scrobbleController.ScrobbleForm)
	
	authed.POST("/scrobble/preview", scrobbleController.PreviewScrobble)

	// v1 := router.Group("v1")
	// {
	// 	userGroup := v1.Group("user")
	// 	{
	// 		user := new(controllers.UserController)
	// 		userGroup.GET("/:id", user.Retrieve)
	// 	}
	// }
	return router

}

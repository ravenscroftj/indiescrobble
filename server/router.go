package server

import (
	"github.com/gin-gonic/gin"
	"github.com/ravenscroftj/indiescrobble/config"
	"github.com/ravenscroftj/indiescrobble/controllers"
	"github.com/ravenscroftj/indiescrobble/middlewares"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	config := config.GetConfig()

	health := new(controllers.HealthController)

	iam := controllers.NewIndieAuthManager(db)

	profile := controllers.NewUserProfileController(db)

	router.GET("/health", health.Status)

	router.Use(middlewares.AuthMiddleware(false, iam))

	router.GET("/", controllers.Index)

	router.Static("/static", config.GetString("server.static_path"))

	// add auth endpoints

	router.POST("/indieauth", iam.IndieAuthLoginPost)
	router.GET("/auth", iam.LoginCallbackGet)
	router.GET("/logout", iam.Logout)

	router.GET("/profile/config", profile.GetConfig)
	router.POST("/profile/config", profile.SaveConfig)
	router.GET("/profile", profile.ViewUserPosts)

	authed := router.Use(middlewares.AuthMiddleware(true, iam))

	// add scrobble endpoints
	scrobbleController := controllers.NewScrobbleController(db)

	authed.GET("/scrobble", scrobbleController.ScrobbleForm)

	authed.POST("/scrobble/preview", scrobbleController.PreviewScrobble)

	authed.POST("/scrobble/do", scrobbleController.DoScrobble)

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

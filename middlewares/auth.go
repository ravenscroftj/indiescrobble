package middlewares

import (
	"net/http"

	"git.jamesravey.me/ravenscroftj/indiescrobble/controllers"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(requireValidUser bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// config := config.GetConfig()

		iam := controllers.NewIndieAuthManager()
		
		currentUser := iam.GetCurrentUser(c)

		if requireValidUser && (currentUser == nil) {
			c.SetCookie("jwt", "", -1, "/", "", c.Request.URL.Scheme == "https", true)
			c.Redirect(http.StatusSeeOther, "/")
		}

		c.Set("user", currentUser)

		// reqKey := c.Request.Header.Get("X-Auth-Key")
		// reqSecret := c.Request.Header.Get("X-Auth-Secret")

		// var key string
		// var secret string
		// if key = config.GetString("http.auth.key"); len(strings.TrimSpace(key)) == 0 {
		// 	c.AbortWithStatus(500)
		// }
		// if secret = config.GetString("http.auth.secret"); len(strings.TrimSpace(secret)) == 0 {
		// 	c.AbortWithStatus(401)
		// }
		// if key != reqKey || secret != reqSecret {
		// 	c.AbortWithStatus(401)
		// 	return
		// }
		c.Next()
	}
}


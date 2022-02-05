package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)


func IndieAuthLoginPost(c *gin.Context) {

	err := c.Request.ParseForm()

	if err != nil{
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"message": err,
		})
	}

}

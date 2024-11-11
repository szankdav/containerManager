package main

import (
	handler "container_manager/handlers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	e := gin.Default()
	e.LoadHTMLGlob("templates/*")
	e.Static("/static", "./static")

	e.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	e.POST("/url", handler.StartContainer)

	e.Run(":8080")
}

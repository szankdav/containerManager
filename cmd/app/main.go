package main

import (
	"net/http"

	handler "example.com/container-manager/internal/api/http/handlers"

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
	e.POST("/test", handler.SpinUpTest)

	e.Run(":8080")
}

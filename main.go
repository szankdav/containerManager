package main

import (
	handler "container_manager/handlers"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	corsConfig := cors.DefaultConfig()

	corsConfig.AllowOrigins = []string{
		"http://localhost:8080",
	}

	e := gin.Default()
	e.LoadHTMLGlob("templates/*")
	e.Static("/static", "./static")
	e.Use(cors.New(corsConfig))

	e.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	e.POST("/url", handler.GetUrlFromHeader)

	e.Run(":8080")
}

package main

import (
	handler "container_manager/internal/api/http/handlers"
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
	e.GET("/increaseTest", handler.SpinUpIncreaseTest)
	e.GET("/decreaseTest", handler.SpinUpDecreaseTest)

	e.Run(":8080")
}

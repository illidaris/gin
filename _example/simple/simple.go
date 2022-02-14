package main

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	ginEx "github.com/illidaris/gin"
	"github.com/illidaris/logger"
)

func main() {
	// init log core
	logger.OnlyConsole()
	// init gin
	router := gin.Default()
	router.Use(ginEx.LoggerHandler())
	router.Use(ginEx.RecoverHandler())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})
	router.GET("/error", func(c *gin.Context) {
		panic(errors.New("this is an error"))
	})
	router.Run(":8080")
}

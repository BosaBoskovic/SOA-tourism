package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/encounters", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Encounters service radi",
		})
	})

	r.Run(":8083")
}
package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/stakeholders", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Stakeholders service radi",
		})
	})

	r.Run(":8081")
}
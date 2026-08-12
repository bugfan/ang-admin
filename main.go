package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Initialize SQLite database
	_, err := xorm.NewEngine("sqlite3", "./ang.db")
	if err != nil {
		log.Fatalf("Failed to create engine: %v", err)
	}

	// Initialize Gin router
	r := gin.Default()

	// Enable CORS for frontend development
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// Start the server
	log.Println("ang-admin server started on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

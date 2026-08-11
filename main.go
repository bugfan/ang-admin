package main

import (
	"log"
	"net/http"

	"github.com/bugfan/ang-admin/models"
	"github.com/bugfan/rest"
	"github.com/gin-gonic/gin"
	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Initialize SQLite database
	engine, err := xorm.NewEngine("sqlite3", "./ang.db")
	if err != nil {
		log.Fatalf("Failed to create engine: %v", err)
	}

	if err := engine.Sync2(new(models.AppProxy)); err != nil {
		log.Fatalf("Failed to sync database: %v", err)
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

	apiGroup := r.Group("/api")

	// Create REST API Backend
	rest.NewAPIBackend(apiGroup, engine, "")

	// Register models to auto-generate REST endpoints
	rest.Register(new(models.AppProxy), new(models.AppProxyContent), rest.RouteTypeALL, nil, "app_proxy")

	// Start the server
	log.Println("ang-admin server started on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

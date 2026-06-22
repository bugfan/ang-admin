package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/bugfan/empty015/internal/api"
	"github.com/bugfan/empty015/internal/store"
	"github.com/gin-gonic/gin"
)

//go:embed web/*
var webFS embed.FS

func main() {
	// Ensure data directory exists
	dataDir := "./data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("failed to create data directory: %v", err)
	}

	// Initialize database
	if err := store.Init(dataDir + "/knowledge.db"); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	log.Println("✅ Database initialized")

	// Setup Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Register API routes
	h := api.New()
	api.RegisterRoutes(r, h)

	// Serve embedded web files
	webFiles, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatalf("failed to load web files: %v", err)
	}
	r.NoRoute(func(c *gin.Context) {
		http.FileServer(http.FS(webFiles)).ServeHTTP(c.Writer, c.Request)
	})

	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	log.Printf("🚀 KnowledgeAI started at http://localhost%s", addr)
	log.Printf("📚 Open your browser and navigate to http://localhost%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

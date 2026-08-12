package api

import (
	"log"
	"net/http"

	"github.com/bugfan/ang-admin/models"
	"github.com/bugfan/rest"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
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

	api := r.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "pong"})
		})
		api.POST("/login", LoginHandler)
		api.POST("/register", RegisterHandler)
		api.GET("/captcha", CaptchaHandler)
		api.POST("/users", ListUsersHandler)
		api.GET("/mine", MineHandler)
		api.POST("/refresh-token", RefreshTokenHandler)
	}

	// Register RESTful APIs using bugfan/rest
	rest.NewAPIBackend(api, models.GetEngine(), "/s")

	return r
}

func StartServer(addr string) {
	r := SetupRouter()
	log.Printf("ang-admin server started on %s\n", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

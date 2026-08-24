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
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, X-Token")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// 1. 无需鉴权的通用/公共非 REST 接口 (Public / Custom APIs)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})
	r.POST("/login", LoginHandler)
	r.GET("/captcha", CaptchaHandler)
	r.GET("/avatar.png", AvatarHandler)
	r.GET("/get-async-routes", AsyncRoutesHandler)

	// 2. 需要鉴权的通用/业务非 REST 接口 (Protected Custom APIs)
	authorized := r.Group("/")
	authorized.Use(AuthMiddleware())
	{
		authorized.GET("/mine", MineHandler)
		authorized.POST("/refresh-token", RefreshTokenHandler)
		authorized.POST("/api/certificate/generate", GenerateCertHandler)
		authorized.POST("/api/certificate/acme-issue", IssueAcmeCertHandler)
		authorized.POST("/api/certificate/acme-issue-by-config/:id", IssueAcmeCertByConfigHandler)
		authorized.GET("/api/tunnel-client/active-connections", GetActiveTunnelConnectionsHandler)
		authorized.POST("/api/cluster-node/verify", VerifyClusterNodeHandler)
		authorized.POST("/api/cluster-node/:id/ping", PingClusterNodeHandler)
		authorized.POST("/api/cluster-node/:id/sync", SyncClusterNodeHandler)
		authorized.POST("/api/cluster-node/sync-all", SyncAllClusterNodesHandler)
		authorized.GET("/api/cluster-node/:id/tunnel", GetClusterNodeTunnelHandler)
	}


	// 3. 基于 bugfan/rest 自动生成的模型 RESTful 增删改查接口 (Auto-generated Model RESTful APIs)
	apiGroup := r.Group("/api")
	apiGroup.Use(AuthMiddleware())
	rest.NewAPIBackend(apiGroup, models.GetEngine(), "")

	return r
}

func StartServer(addr string) {
	r := SetupRouter()
	log.Printf("ang-admin server started on %s\n", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

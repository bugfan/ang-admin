package api

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes sets up all API routes
func RegisterRoutes(r *gin.Engine, h *Handler) {
	api := r.Group("/api")
	{
		// Books
		api.GET("/books", h.ListBooks)
		api.POST("/books/upload", h.UploadBook)
		api.DELETE("/books/:id", h.DeleteBook)
		api.GET("/books/:id/status", h.GetBookStatus)

		// Chat
		api.POST("/chat", h.Chat)
		api.GET("/chat/history", h.GetHistory)
		api.DELETE("/chat/history", h.ClearHistory)

		// Config
		api.GET("/config", h.GetConfig)
		api.POST("/config", h.SaveConfig)
	}
}

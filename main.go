package main

import (
	"github.com/bugfan/ang-admin/api"
	"github.com/bugfan/ang-admin/models"
)

func main() {
	// 1. Initialize Database
	models.InitDB("./ang.db")

	// 2. Start HTTP Server
	api.StartServer(":8080")
}

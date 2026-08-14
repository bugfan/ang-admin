package main

import (
	"github.com/bugfan/ang-admin/api"
	"github.com/bugfan/ang-admin/models"
	"github.com/bugfan/ang-admin/service"
)

func main() {
	// 1. Initialize Database
	models.InitDB("./ang.db")

	// 2. Sync configuration entities to cluster
	service.SyncAllToCluster()

	// 3. Start HTTP Server
	api.StartServer(":8080")
}


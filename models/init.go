package models

import (
	"log"

	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
)

var engine *xorm.Engine

func InitDB(dsn string) {
	var err error
	engine, err = xorm.NewEngine("sqlite3", dsn)
	if err != nil {
		log.Fatalf("Failed to create engine: %v", err)
	}

	// Automatically sync database schemas if necessary
	err = engine.Sync2(new(User))
	if err != nil {
		log.Fatalf("Failed to sync database: %v", err)
	}
}

func GetEngine() *xorm.Engine {
	return engine
}

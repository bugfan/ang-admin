package models

import (
	"log"

	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
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

	// Initialize default admin user
	admin := &User{Username: "admin"}
	has, err := engine.Get(admin)
	if err != nil {
		log.Fatalf("Failed to query admin user: %v", err)
	}
	if !has {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin@9527"), bcrypt.DefaultCost)
		admin.Password = string(hashedPassword)
		_, err = engine.Insert(admin)
		if err != nil {
			log.Fatalf("Failed to insert default admin: %v", err)
		}
	}
}

func GetEngine() *xorm.Engine {
	return engine
}

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

	// Drop old tunnel table if it still contains obsolete key_name column
	if isExist, _ := engine.IsTableExist("tunnel"); isExist {
		results, err := engine.QueryString("PRAGMA table_info(tunnel)")
		if err == nil {
			for _, row := range results {
				if row["name"] == "key_name" {
					_ = engine.DropTables("tunnel")
					break
				}
			}
		}
	}

	// Automatically sync database schemas if necessary
	err = engine.Sync2(new(AdminUser), new(Tunnel), new(Certificate), new(TunnelClient), new(DnsProxy), new(Rule))

	if err != nil {
		log.Fatalf("Failed to sync database: %v", err)
	}

	// Backfill certificate parsed metadata for all certificates
	var allCerts []Certificate
	if err := engine.Where("cert_content != ''").Find(&allCerts); err == nil {
		for _, cert := range allCerts {
			cert.ParseCertInfo()
			_, _ = engine.ID(cert.Id).Cols("subject_cn", "sans", "not_before", "not_after", "issuer", "serial_number").Update(&cert)
		}
	}

	// Initialize default admin user
	admin := &AdminUser{Username: "admin"}
	has, err := engine.Get(admin)
	if err != nil {
		log.Fatalf("Failed to query admin user: %v", err)
	}
	if !has {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin@9527"), bcrypt.DefaultCost)
		admin.Password = string(hashedPassword)
		admin.IsSuperAdmin = true
		_, err = engine.Insert(admin)
		if err != nil {
			log.Fatalf("Failed to insert default admin: %v", err)
		}
	} else if !admin.IsSuperAdmin {
		admin.IsSuperAdmin = true
		_, _ = engine.ID(admin.Id).Cols("is_super_admin").Update(admin)
	}
}

func GetEngine() *xorm.Engine {
	return engine
}

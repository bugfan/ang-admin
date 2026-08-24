package models

import (
	"log"
	"time"

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
	err = engine.Sync2(new(AdminUser), new(Tunnel), new(Certificate), new(TunnelClient), new(DnsProxy), new(Rule), new(HttpProxy), new(ClusterNode), new(AcmeConfig))

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

	// Backfill missing or zero created_at dates across tables
	nowStr := time.Now().Format("2006-01-02 15:04:05")
	_, _ = engine.Exec("UPDATE http_proxy SET created_at = ? WHERE created_at IS NULL OR created_at = '' OR created_at LIKE '0001-01-01%'", nowStr)
	_, _ = engine.Exec("UPDATE rule SET created_at = ? WHERE created_at IS NULL OR created_at = '' OR created_at LIKE '0001-01-01%'", nowStr)
	_, _ = engine.Exec("UPDATE dns_proxy SET created_at = ? WHERE created_at IS NULL OR created_at = '' OR created_at LIKE '0001-01-01%'", nowStr)
	_, _ = engine.Exec("UPDATE tunnel SET created_at = ? WHERE created_at IS NULL OR created_at = '' OR created_at LIKE '0001-01-01%'", nowStr)
	_, _ = engine.Exec("UPDATE certificate SET source = 'MANUAL' WHERE source IS NULL OR source = ''")
	_, _ = engine.Exec("UPDATE certificate SET source = 'SELF_SIGNED' WHERE type LIKE 'SELF%'")
	_, _ = engine.Exec("UPDATE certificate SET source = 'ACME' WHERE cert_id LIKE 'acme-%'")

	// Initialize default admin user
	admin := &AdminUser{Username: "admin"}
	has, err := engine.Get(admin)
	if err != nil {
		log.Fatalf("Failed to query admin user: %v", err)
	}
	if !has {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin@123"), bcrypt.DefaultCost)
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

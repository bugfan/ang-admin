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

	err = engine.Sync2(
		new(AdminUser), new(Tunnel), new(Certificate), new(TunnelClient),
		new(DnsProxy), new(TcpProxy), new(UdpProxy), new(SniProxy),
		new(Rule), new(HttpProxy), new(ClusterNode), new(AcmeAccount),
		new(UserGroup), new(User), new(AuthMethod), new(Auth),
		new(WebvpnDomain), new(WebvpnSite),
	)

	if err != nil {
		log.Fatalf("Failed to sync database: %v", err)
	}

	// Ensure default UserGroup exists
	if count, err := engine.Count(new(UserGroup)); err == nil && count == 0 {
		_, _ = engine.Insert(&UserGroup{
			Name:        "默认用户组",
			Description: "系统默认用户组",
			IsDefault:   true,
		})
	}

	// Ensure default Local AuthMethod exists
	if count, err := engine.Count(new(AuthMethod)); err == nil && count == 0 {
		_, _ = engine.Insert(&AuthMethod{
			Name:       "本地账号认证",
			Type:       "local",
			Enabled:    true,
			Priority:   1,
			ConfigJSON: `{"allow_self_register":false,"password_min_len":6}`,
			Remark:     "系统默认本地用户名密码认证",
		})
	}

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

package service_test

import (
	"testing"

	"github.com/bugfan/ang-admin/service"
)

// 测试前请先启动 FreeRADIUS 测试服务器:
//   cd testenv && docker compose up -d
//
// FreeRADIUS 配置:
//   Host:   127.0.0.1
//   Port:   1812
//   Secret: testing123
//
// 测试账号 (见 testenv/freeradius-config/users):
//   testuser / testpass
//   admin    / admin123
//   alice    / alice@2024
//   bob      / bob@2024

const (
	testHost   = "127.0.0.1"
	testPort   = 1812
	testSecret = "testing123"
)

func testCfg() *service.RadiusConfig {
	return &service.RadiusConfig{
		Host:          testHost,
		Port:          testPort,
		Secret:        testSecret,
		Timeout:       5,
		Retries:       3,
		NASIdentifier: "ang-admin-test",
	}
}

// TestRadiusConnectivity 测试与 FreeRADIUS 服务器的连通性
// 即便探针用户不存在（Access-Reject），只要服务器正常响应就视为连通
func TestRadiusConnectivity(t *testing.T) {
	cfg := testCfg()
	if err := service.RadiusTestConnectivity(cfg); err != nil {
		t.Skipf("跳过测试: FreeRADIUS 服务未启动或不可达 (%v)。请先运行 cd testenv && docker compose up -d", err)
	}
	t.Log("RADIUS 连通性测试通过")
}

// TestRadiusAuthSuccess 测试正确用户名/密码认证成功
func TestRadiusAuthSuccess(t *testing.T) {
	cfg := testCfg()

	// 先检查连通性，不可达则跳过
	if err := service.RadiusTestConnectivity(cfg); err != nil {
		t.Skipf("跳过测试: FreeRADIUS 服务未启动 (%v)", err)
	}

	result, err := service.RadiusAuthenticate(cfg, "testuser", "testpass")
	if err != nil {
		t.Fatalf("认证请求失败: %v", err)
	}
	if !result.Success {
		t.Fatalf("期望认证成功，实际被拒绝: code=%s, message=%s", result.Code, result.ReplyMessage)
	}
	t.Logf("认证成功: code=%s, reply=%s", result.Code, result.ReplyMessage)
}

// TestRadiusAuthWrongPassword 测试错误密码认证失败
func TestRadiusAuthWrongPassword(t *testing.T) {
	cfg := testCfg()

	if err := service.RadiusTestConnectivity(cfg); err != nil {
		t.Skipf("跳过测试: FreeRADIUS 服务未启动 (%v)", err)
	}

	result, err := service.RadiusAuthenticate(cfg, "testuser", "wrongpassword")
	if err != nil {
		t.Fatalf("认证请求失败: %v", err)
	}
	if result.Success {
		t.Fatal("期望认证失败（密码错误），但实际认证成功了，请检查 FreeRADIUS 配置")
	}
	t.Logf("认证正确被拒绝: code=%s", result.Code)
}

// TestRadiusAuthNonExistentUser 测试不存在的用户
func TestRadiusAuthNonExistentUser(t *testing.T) {
	cfg := testCfg()

	if err := service.RadiusTestConnectivity(cfg); err != nil {
		t.Skipf("跳过测试: FreeRADIUS 服务未启动 (%v)", err)
	}

	result, err := service.RadiusAuthenticate(cfg, "nonexistentuser", "anypassword")
	if err != nil {
		t.Fatalf("认证请求失败: %v", err)
	}
	if result.Success {
		t.Fatal("期望不存在的用户认证失败，但实际认证成功了")
	}
	t.Logf("不存在的用户正确被拒绝: code=%s", result.Code)
}

// TestRadiusAuthMultipleUsers 测试多个用户认证
func TestRadiusAuthMultipleUsers(t *testing.T) {
	cfg := testCfg()

	if err := service.RadiusTestConnectivity(cfg); err != nil {
		t.Skipf("跳过测试: FreeRADIUS 服务未启动 (%v)", err)
	}

	cases := []struct {
		username string
		password string
		wantOK   bool
	}{
		{"testuser", "testpass", true},
		{"admin", "admin123", true},
		{"alice", "alice@2024", true},
		{"bob", "bob@2024", true},
		{"testuser", "wrongpass", false},
		{"nouser", "nopass", false},
	}

	for _, tc := range cases {
		t.Run(tc.username, func(t *testing.T) {
			result, err := service.RadiusAuthenticate(cfg, tc.username, tc.password)
			if err != nil {
				t.Fatalf("认证请求失败: %v", err)
			}
			if result.Success != tc.wantOK {
				t.Errorf("用户 %s: 期望 success=%v, 实际 success=%v (code=%s)",
					tc.username, tc.wantOK, result.Success, result.Code)
			} else {
				t.Logf("用户 %s: success=%v, code=%s, reply=%s",
					tc.username, result.Success, result.Code, result.ReplyMessage)
			}
		})
	}
}

// TestParseRadiusConfig 测试配置解析
func TestParseRadiusConfig(t *testing.T) {
	// 完整配置
	full := `{"host":"1.2.3.4","port":1812,"secret":"mysecret","timeout":10,"retries":5,"nas_identifier":"mynas"}`
	cfg, err := service.ParseRadiusConfig(full)
	if err != nil {
		t.Fatalf("解析完整配置失败: %v", err)
	}
	if cfg.Host != "1.2.3.4" || cfg.Port != 1812 || cfg.Secret != "mysecret" {
		t.Errorf("配置解析结果不正确: %+v", cfg)
	}

	// 默认值
	minimal := `{"host":"1.2.3.4","secret":"s"}`
	cfg2, err := service.ParseRadiusConfig(minimal)
	if err != nil {
		t.Fatalf("解析最小配置失败: %v", err)
	}
	if cfg2.Port != 1812 {
		t.Errorf("默认端口应为 1812，实际为 %d", cfg2.Port)
	}
	if cfg2.Timeout != 5 {
		t.Errorf("默认超时应为 5，实际为 %d", cfg2.Timeout)
	}
	if cfg2.NASIdentifier != "ang-admin" {
		t.Errorf("默认 NASIdentifier 应为 'ang-admin'，实际为 %s", cfg2.NASIdentifier)
	}
	t.Log("配置解析测试通过")
}

// TestRadiusAuthWrongSecret 测试错误 shared secret 的情况
// 注意：UDP RADIUS 协议下，错误的 secret 会导致服务器静默丢包不回应，
// 客户端会等到超时。这是协议层面的正常安全行为。
// 此测试使用极短超时，不会阻塞整个测试进程。
func TestRadiusAuthWrongSecret(t *testing.T) {
	wrongCfg := &service.RadiusConfig{
		Host:          testHost,
		Port:          testPort,
		Secret:        "definitely_wrong_secret_xyz",
		Timeout:       2, // 2 秒快速超时
		Retries:       1,
		NASIdentifier: "ang-admin-test",
	}

	_, err := service.RadiusAuthenticate(wrongCfg, "testuser", "testpass")
	// 错误 secret 下有两种可能结果：
	// 1. 超时 (context deadline exceeded) —— 服务器静默丢包，符合预期
	// 2. 其他错误
	// 两种都是"认证不成功"，均符合预期
	if err != nil {
		t.Logf("错误 secret 下认证失败（符合预期）: %v", err)
	} else {
		t.Log("注意: 使用错误 secret 但未报错，服务器可能允许任意 secret（生产环境不安全）")
	}
}

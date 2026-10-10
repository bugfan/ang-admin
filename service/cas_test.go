package service_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bugfan/ang-admin/service"
)

// TestParseCasConfig 测试配置解析与默认值
func TestParseCasConfig(t *testing.T) {
	raw := `{"base_url":"http://cas.example.com:8080/cas/","version":2,"login_path":"/custom_login","skip_tls_verify":true}`
	cfg, err := service.ParseCasConfig(raw)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	if cfg.BaseURL != "http://cas.example.com:8080/cas" {
		t.Errorf("BaseURL 尾部斜杠应被移除，实际为: %s", cfg.BaseURL)
	}
	if cfg.LoginPath != "/custom_login" {
		t.Errorf("LoginPath 应为 /custom_login，实际为: %s", cfg.LoginPath)
	}
	if cfg.ValidatePath != "/serviceValidate" {
		t.Errorf("ValidatePath 默认值应为 /serviceValidate，实际为: %s", cfg.ValidatePath)
	}
	if !cfg.SkipTLSVerify {
		t.Errorf("SkipTLSVerify 应为 true")
	}
}

// TestBuildCasLoginURL 测试登录重定向 URL 构造
func TestBuildCasLoginURL(t *testing.T) {
	cfg := &service.CasConfig{
		BaseURL:   "http://127.0.0.1:8082/cas",
		LoginPath: "/login",
	}

	u := service.BuildCasLoginURL(cfg, "http://ang.local/callback?redirect=http%3A%2F%2Fapp.local")
	expectedPrefix := "http://127.0.0.1:8082/cas/login?service="
	if len(u) < len(expectedPrefix) || u[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("生成的登录 URL 格式错误: %s", u)
	}
}

// TestCasIntegrationWithMockServer 端到端模拟 CAS 流程测试
func TestCasIntegrationWithMockServer(t *testing.T) {
	validTicket := "ST-test-valid-123"
	invalidTicket := "ST-test-invalid-999"
	expectedUser := "test_student_01"

	// 启动一个内存 CAS Mock Server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch path {
		case "/cas/login":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("<html>CAS Login Page</html>"))
		case "/cas/serviceValidate":
			ticket := r.URL.Query().Get("ticket")
			w.Header().Set("Content-Type", "application/xml; charset=utf-8")
			if ticket == validTicket {
				xml := fmt.Sprintf(`<cas:serviceResponse xmlns:cas='http://www.yale.edu/tp/cas'>
    <cas:authenticationSuccess>
        <cas:user>%s</cas:user>
        <cas:attributes>
            <cas:email>student@school.edu</cas:email>
            <cas:department>Computer Science</cas:department>
        </cas:attributes>
    </cas:authenticationSuccess>
</cas:serviceResponse>`, expectedUser)
				_, _ = w.Write([]byte(xml))
			} else {
				xml := fmt.Sprintf(`<cas:serviceResponse xmlns:cas='http://www.yale.edu/tp/cas'>
    <cas:authenticationFailure code="INVALID_TICKET">
        Ticket '%s' not recognized
    </cas:authenticationFailure>
</cas:serviceResponse>`, ticket)
				_, _ = w.Write([]byte(xml))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	cfg := &service.CasConfig{
		BaseURL:      ts.URL + "/cas",
		Version:      2,
		LoginPath:    "/login",
		ValidatePath: "/serviceValidate",
		Timeout:      3,
	}

	// 1. 连通性测试
	if err := service.CasTestConnectivity(cfg); err != nil {
		t.Fatalf("连通性测试应成功，但返回错误: %v", err)
	}

	// 2. 有效 Ticket 验证
	resValid, err := service.ValidateCasTicket(cfg, validTicket, "http://ang.local/callback")
	if err != nil {
		t.Fatalf("有效 Ticket 校验请求失败: %v", err)
	}
	if !resValid.Success {
		t.Fatalf("期望验证成功，实际返回失败: %s", resValid.ErrorMsg)
	}
	if resValid.Username != expectedUser {
		t.Errorf("期望用户名 %s，实际为 %s", expectedUser, resValid.Username)
	}
	if resValid.Attributes["email"] != "student@school.edu" {
		t.Errorf("未正确解析属性 email: %v", resValid.Attributes)
	}

	// 3. 无效 Ticket 验证
	resInvalid, err := service.ValidateCasTicket(cfg, invalidTicket, "http://ang.local/callback")
	if err != nil {
		t.Fatalf("无效 Ticket 校验请求报错: %v", err)
	}
	if resInvalid.Success {
		t.Fatalf("无效 Ticket 期望返回失败，但返回了成功")
	}
	if resInvalid.ErrorCode != "INVALID_TICKET" {
		t.Errorf("期望错误码 INVALID_TICKET，实际为: %s", resInvalid.ErrorCode)
	}
}

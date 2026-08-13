package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bugfan/ang-admin/models"
	"github.com/bugfan/ang-admin/service"
)

func TestNonSuperAdminListOnlySelf(t *testing.T) {
	models.InitDB(":memory:")
	r := SetupRouter()

	// 1. 注册一个非超管管理员 testUser
	testUser := &models.AdminUser{
		Username:     "testUser",
		Password:     "password123",
		IsSuperAdmin: false,
	}
	err := service.RegisterAdminUser(testUser)
	if err != nil {
		t.Fatalf("failed to register testUser: %v", err)
	}

	testToken, _ := service.GenerateToken("testUser", 3600)

	// 2. testUser 请求 GET /api/admin
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/admin", nil)
	req.Header.Set("Authorization", "Bearer "+testToken)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	t.Logf("Response body: %s", w.Body.String())

	var res []map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &res)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// 检查返回的数据列表数量和内容
	t.Logf("Returned %d users", len(res))
	for _, u := range res {
		t.Logf("User in list: %v", u["Username"])
	}

	if len(res) != 1 {
		t.Errorf("EXPECTED 1 user (testUser only), BUT GOT %d users!", len(res))
	} else if res[0]["Username"] != "testUser" {
		t.Errorf("EXPECTED username 'testUser', BUT GOT '%v'", res[0]["Username"])
	}
}

func TestSuperAdminPermissions(t *testing.T) {
	models.InitDB(":memory:")
	r := SetupRouter()

	superToken, _ := service.GenerateToken("admin", 3600)

	commonUser := &models.AdminUser{
		Username:     "commonUser",
		Password:     "password123",
		IsSuperAdmin: false,
	}
	err := service.RegisterAdminUser(commonUser)
	if err != nil {
		t.Fatalf("failed to register common user: %v", err)
	}

	commonToken, _ := service.GenerateToken("commonUser", 3600)

	// 超级管理员 GET /api/admin
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/api/admin", nil)
	req1.Header.Set("Authorization", "Bearer "+superToken)
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("expected status 200 for super admin, got %d", w1.Code)
	}

	// 非超管 POST /api/admin -> 403
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/api/admin", strings.NewReader(`{"username":"test11","password":"123"}`))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+commonToken)
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusForbidden {
		t.Fatalf("expected status 403 for common user POST, got %d", w2.Code)
	}

	// 非超管 DELETE /api/admin/1 -> 403
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("DELETE", "/api/admin/1", nil)
	req3.Header.Set("Authorization", "Bearer "+commonToken)
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusForbidden {
		t.Fatalf("expected status 403 for common user DELETE, got %d", w3.Code)
	}
}

package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// TicketInfo 存储 Service Ticket 详细信息
type TicketInfo struct {
	Ticket    string
	Username  string
	Service   string
	CreatedAt time.Time
	Email     string
}

// 模拟账号密码数据库
var mockUsers = map[string]struct {
	Password string
	Email    string
}{
	"admin":    {"admin123", "admin@ang.local"},
	"testuser": {"testpass", "testuser@ang.local"},
	"alice":    {"alice@2024", "alice@ang.local"},
	"casuser":  {"cas123", "casuser@ang.local"},
}

var (
	ticketsLock sync.Mutex
	tickets     = make(map[string]*TicketInfo)
)

func generateTicket() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	return "ST-" + hex.EncodeToString(bytes)
}

func main() {
	mux := http.NewServeMux()

	// 1. CAS 登录接口 (GET/POST)
	mux.HandleFunc("/cas/login", handleLogin)

	// 2. CAS 2.0 / 3.0 Service Ticket 校验接口
	mux.HandleFunc("/cas/serviceValidate", handleServiceValidate)
	mux.HandleFunc("/cas/p3/serviceValidate", handleServiceValidate)

	// 3. CAS 1.0 票据校验接口
	mux.HandleFunc("/cas/validate", handleValidateV1)

	// 4. 注销接口
	mux.HandleFunc("/cas/logout", handleLogout)

	// 5. 健康检查
	mux.HandleFunc("/cas/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","server":"Mock CAS Server 2.0/3.0"}`))
	})

	addr := ":8082"
	log.Printf("==================================================")
	log.Printf("🚀 Mock CAS Server started on http://127.0.0.1%s/cas", addr)
	log.Printf("📋 Pre-configured Test Users:")
	for u, p := range mockUsers {
		log.Printf("   - %s / %s (%s)", u, p.Password, p.Email)
	}
	log.Printf("==================================================")

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Mock CAS Server error: %v", err)
	}
}

const loginHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>企业统一身份认证服务 (CAS SSO)</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; }
        body { min-height: 100vh; display: flex; align-items: center; justify-content: center; background: linear-gradient(135deg, #0f172a 0%, #1e1b4b 100%); color: #f8fafc; padding: 1.5rem; }
        .card { width: 100%; max-width: 440px; background: rgba(30, 41, 59, 0.9); backdrop-filter: blur(16px); border: 1px solid rgba(255, 255, 255, 0.12); border-radius: 16px; padding: 2.5rem 2rem; box-shadow: 0 25px 30px -5px rgba(0, 0, 0, 0.6); }
        .logo { width: 56px; height: 56px; margin: 0 auto 1.25rem; background: linear-gradient(135deg, #6366f1, #a855f7); border-radius: 14px; display: flex; align-items: center; justify-content: center; box-shadow: 0 10px 15px -3px rgba(168, 85, 247, 0.4); font-size: 26px; }
        h1 { text-align: center; font-size: 1.5rem; font-weight: 600; margin-bottom: 0.35rem; color: #fff; }
        p.sub { text-align: center; font-size: 0.875rem; color: #94a3b8; margin-bottom: 1.75rem; }
        .service-box { background: rgba(15, 23, 42, 0.6); border: 1px dashed rgba(255,255,255,0.15); border-radius: 8px; padding: 0.65rem 0.85rem; font-size: 0.8rem; color: #cbd5e1; margin-bottom: 1.25rem; word-break: break-all; }
        .service-box strong { color: #38bdf8; }
        .form-group { margin-bottom: 1.25rem; }
        label { display: block; font-size: 0.875rem; font-weight: 500; color: #cbd5e1; margin-bottom: 0.5rem; }
        input { width: 100%; padding: 0.75rem 1rem; border-radius: 8px; border: 1px solid #475569; background: #0f172a; color: #fff; font-size: 0.95rem; outline: none; }
        input:focus { border-color: #a855f7; box-shadow: 0 0 0 3px rgba(168, 85, 247, 0.25); }
        .err { padding: 0.75rem 1rem; border-radius: 8px; font-size: 0.875rem; margin-bottom: 1.25rem; background: rgba(239, 68, 68, 0.15); border: 1px solid rgba(239, 68, 68, 0.3); color: #fca5a5; }
        button { width: 100%; padding: 0.85rem; border: none; border-radius: 8px; background: linear-gradient(135deg, #6366f1, #a855f7); color: white; font-size: 1rem; font-weight: 600; cursor: pointer; margin-top: 0.5rem; }
        button:hover { opacity: 0.92; }
        .tips { margin-top: 1.5rem; background: rgba(15, 23, 42, 0.4); border-radius: 8px; padding: 0.75rem; font-size: 0.75rem; color: #94a3b8; }
        .tips table { width: 100%; border-collapse: collapse; margin-top: 0.25rem; }
        .tips td { padding: 0.2rem 0.4rem; }
    </style>
</head>
<body>
    <div class="card">
        <div class="logo">🔐</div>
        <h1>统一身份认证</h1>
        <p class="sub">Central Authentication Service (CAS 2.0/3.0)</p>
        
        {{if .Service}}
        <div class="service-box">
            回调业务地址: <strong>{{.Service}}</strong>
        </div>
        {{end}}

        {{if .Error}}
        <div class="err">{{.Error}}</div>
        {{end}}

        <form method="POST" action="/cas/login?service={{.Service}}">
            <div class="form-group">
                <label for="username">用户名 / 学号 / 工号</label>
                <input type="text" id="username" name="username" value="{{.Username}}" required autofocus placeholder="请输入您的账号">
            </div>
            <div class="form-group">
                <label for="password">密码</label>
                <input type="password" id="password" name="password" required placeholder="请输入您的密码">
            </div>
            <button type="submit">立即登录</button>
        </form>

        <div class="tips">
            <div>💡 <strong>测试账号列表：</strong></div>
            <table>
                <tr><td>testuser / testpass</td><td>alice / alice@2024</td></tr>
                <tr><td>admin / admin123</td><td>casuser / cas123</td></tr>
            </table>
        </div>
    </div>
</body>
</html>`

func handleLogin(w http.ResponseWriter, r *http.Request) {
	serviceURL := r.URL.Query().Get("service")
	tmpl, _ := template.New("login").Parse(loginHTML)

	if r.Method == http.MethodGet {
		_ = tmpl.Execute(w, map[string]interface{}{
			"Service":  serviceURL,
			"Username": "",
			"Error":    "",
		})
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		username := strings.TrimSpace(r.FormValue("username"))
		password := r.FormValue("password")

		user, exists := mockUsers[username]
		if !exists || user.Password != password {
			w.WriteHeader(http.StatusUnauthorized)
			_ = tmpl.Execute(w, map[string]interface{}{
				"Service":  serviceURL,
				"Username": username,
				"Error":    "用户名或密码错误，请检查后重试",
			})
			return
		}

		// 验证通过，生成 Service Ticket (ST)
		st := generateTicket()
		ticketsLock.Lock()
		tickets[st] = &TicketInfo{
			Ticket:    st,
			Username:  username,
			Service:   serviceURL,
			CreatedAt: time.Now(),
			Email:     user.Email,
		}
		ticketsLock.Unlock()

		log.Printf("[CAS Server] User '%s' logged in successfully. Issued ticket: %s for service: %s", username, st, serviceURL)

		// 如果没有 service，直接显示已登录
		if serviceURL == "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(fmt.Sprintf("<h1>登录成功</h1><p>用户: %s，生成的 Ticket: %s</p>", username, st)))
			return
		}

		// 302 重定向带上 ticket
		redirectTarget := serviceURL
		if strings.Contains(redirectTarget, "?") {
			redirectTarget += "&ticket=" + url.QueryEscape(st)
		} else {
			redirectTarget += "?ticket=" + url.QueryEscape(st)
		}

		http.Redirect(w, r, redirectTarget, http.StatusFound)
	}
}

// handleServiceValidate 处理 CAS 2.0 / 3.0 标准 serviceValidate
func handleServiceValidate(w http.ResponseWriter, r *http.Request) {
	ticket := r.URL.Query().Get("ticket")
	serviceURL := r.URL.Query().Get("service")

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")

	ticketsLock.Lock()
	info, exists := tickets[ticket]
	if exists {
		// CAS 标准：Ticket 必须一次性消费
		delete(tickets, ticket)
	}
	ticketsLock.Unlock()

	// 检查过期 (5分钟)
	if !exists || time.Since(info.CreatedAt) > 5*time.Minute {
		log.Printf("[CAS Server] Validation FAILED: ticket '%s' invalid or expired", ticket)
		xmlResp := fmt.Sprintf(`<cas:serviceResponse xmlns:cas='http://www.yale.edu/tp/cas'>
    <cas:authenticationFailure code="INVALID_TICKET">
        Ticket '%s' not recognized
    </cas:authenticationFailure>
</cas:serviceResponse>`, ticket)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(xmlResp))
		return
	}

	log.Printf("[CAS Server] Validation SUCCESS: ticket '%s' validated for user '%s', service: '%s'", ticket, info.Username, serviceURL)

	xmlSuccess := fmt.Sprintf(`<cas:serviceResponse xmlns:cas='http://www.yale.edu/tp/cas'>
    <cas:authenticationSuccess>
        <cas:user>%s</cas:user>
        <cas:attributes>
            <cas:email>%s</cas:email>
            <cas:authenticationDate>%s</cas:authenticationDate>
        </cas:attributes>
    </cas:authenticationSuccess>
</cas:serviceResponse>`, info.Username, info.Email, info.CreatedAt.Format(time.RFC3339))

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(xmlSuccess))
}

// handleValidateV1 处理 CAS 1.0 validate
func handleValidateV1(w http.ResponseWriter, r *http.Request) {
	ticket := r.URL.Query().Get("ticket")

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	ticketsLock.Lock()
	info, exists := tickets[ticket]
	if exists {
		delete(tickets, ticket)
	}
	ticketsLock.Unlock()

	if !exists || time.Since(info.CreatedAt) > 5*time.Minute {
		_, _ = w.Write([]byte("no\n\n"))
		return
	}

	_, _ = w.Write([]byte(fmt.Sprintf("yes\n%s\n", info.Username)))
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	serviceURL := r.URL.Query().Get("service")
	if serviceURL != "" {
		http.Redirect(w, r, serviceURL, http.StatusFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte("<h1>CAS 会话已注销</h1><p>您已成功退出统一身份认证系统。</p>"))
}

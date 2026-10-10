package service

import (
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// CasConfig CAS 认证方式的配置结构
// 对应 AuthMethod.ConfigJSON 字段内容
type CasConfig struct {
	BaseURL       string `json:"base_url"`        // CAS 服务器 Base URL，如 http://127.0.0.1:8082/cas
	Version       int    `json:"version"`         // CAS 协议版本: 1, 2, 3，默认 2
	LoginPath     string `json:"login_path"`      // 登录页路径，默认 /login
	ValidatePath  string `json:"validate_path"`   // 票据验证路径，默认 /serviceValidate (或 /p3/serviceValidate)
	SkipTLSVerify bool   `json:"skip_tls_verify"`  // 是否跳过 HTTPS 证书验证
	Timeout       int    `json:"timeout"`         // 请求超时秒数，默认 5
}

// ParseCasConfig 从 JSON 字符串解析 CAS 配置
func ParseCasConfig(configJSON string) (*CasConfig, error) {
	cfg := &CasConfig{
		Version:      2,
		LoginPath:    "/login",
		ValidatePath: "/serviceValidate",
		Timeout:      5,
	}
	if configJSON == "" || configJSON == "{}" {
		return cfg, nil
	}
	if err := json.Unmarshal([]byte(configJSON), cfg); err != nil {
		return nil, fmt.Errorf("解析 CAS 配置失败: %w", err)
	}

	cfg.BaseURL = strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if cfg.LoginPath == "" {
		cfg.LoginPath = "/login"
	}
	if !strings.HasPrefix(cfg.LoginPath, "/") {
		cfg.LoginPath = "/" + cfg.LoginPath
	}
	if cfg.ValidatePath == "" {
		cfg.ValidatePath = "/serviceValidate"
	}
	if !strings.HasPrefix(cfg.ValidatePath, "/") {
		cfg.ValidatePath = "/" + cfg.ValidatePath
	}
	if cfg.Version <= 0 {
		cfg.Version = 2
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 5
	}
	return cfg, nil
}

// CasAuthResult CAS 票据验证结果
type CasAuthResult struct {
	Success    bool              // 是否认证成功
	Username   string            // CAS 返回的用户名/账号
	Attributes map[string]string // CAS 返回的用户属性字典
	ErrorCode  string            // 失败时的错误代码，如 INVALID_TICKET
	ErrorMsg   string            // 失败时的错误信息
}

// BuildCasLoginURL 生成跳转到 CAS Server 登录页的重定向 URL
func BuildCasLoginURL(cfg *CasConfig, serviceURL string) string {
	loginURL := cfg.BaseURL + cfg.LoginPath
	if serviceURL == "" {
		return loginURL
	}
	if strings.Contains(loginURL, "?") {
		return loginURL + "&service=" + url.QueryEscape(serviceURL)
	}
	return loginURL + "?service=" + url.QueryEscape(serviceURL)
}

// ValidateCasTicket 向 CAS Server 验证 Service Ticket
func ValidateCasTicket(cfg *CasConfig, ticket, serviceURL string) (*CasAuthResult, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("CAS Base URL 不能为空")
	}
	if ticket == "" {
		return nil, fmt.Errorf("CAS Ticket 不能为空")
	}

	// 拼装验证接口 URL
	validateEndpoint := cfg.BaseURL + cfg.ValidatePath
	valURL, err := url.Parse(validateEndpoint)
	if err != nil {
		return nil, fmt.Errorf("解析 CAS 验证地址失败: %w", err)
	}

	q := valURL.Query()
	q.Set("ticket", ticket)
	if serviceURL != "" {
		q.Set("service", serviceURL)
	}
	valURL.RawQuery = q.Encode()

	// 构造 HTTP 客户端
	client := &http.Client{
		Timeout: time.Duration(cfg.Timeout) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.SkipTLSVerify},
		},
	}

	req, err := http.NewRequest(http.MethodGet, valURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("创建 CAS 验证请求失败: %w", err)
	}
	req.Header.Set("Accept", "application/xml, text/xml, */*")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("CAS 服务器请求失败 [%s]: %w", valURL.String(), err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取 CAS 响应数据失败: %w", err)
	}
	bodyStr := string(bodyBytes)

	// 如果配置为 CAS 1.0 (纯文本响应: yes\nusername\n)
	if cfg.Version == 1 || strings.HasSuffix(cfg.ValidatePath, "/validate") {
		lines := strings.Split(strings.TrimSpace(bodyStr), "\n")
		if len(lines) > 0 && strings.TrimSpace(lines[0]) == "yes" {
			username := ""
			if len(lines) > 1 {
				username = strings.TrimSpace(lines[1])
			}
			return &CasAuthResult{
				Success:  true,
				Username: username,
			}, nil
		}
		return &CasAuthResult{
			Success:  false,
			ErrorMsg: "CAS 1.0 验证失败",
		}, nil
	}

	// CAS 2.0 / 3.0 (XML 格式)
	return parseCasXMLResponse(bodyStr)
}

// parseCasXMLResponse 解析 CAS 2.0 / 3.0 的 XML 响应
func parseCasXMLResponse(xmlStr string) (*CasAuthResult, error) {
	result := &CasAuthResult{
		Attributes: make(map[string]string),
	}

	// 1. 判断是否包含成功标签
	if strings.Contains(xmlStr, "authenticationSuccess") {
		result.Success = true

		// 正则/字符串精确提取 <cas:user> 或 <user> 标签内容（去除命名空间干扰）
		userRe := regexp.MustCompile(`<(?:[a-zA-Z0-9_-]+:)?user>\s*([^<\s]+)\s*</(?:[a-zA-Z0-9_-]+:)?user>`)
		if matches := userRe.FindStringSubmatch(xmlStr); len(matches) > 1 {
			result.Username = strings.TrimSpace(matches[1])
		}

		// 提取 attributes 属性（如果有）
		attrRe := regexp.MustCompile(`<(?:[a-zA-Z0-9_-]+:)?([a-zA-Z0-9_-]+)>\s*([^<]+)\s*</(?:[a-zA-Z0-9_-]+)?:?([a-zA-Z0-9_-]+)>`)
		allMatches := attrRe.FindAllStringSubmatch(xmlStr, -1)
		for _, m := range allMatches {
			if len(m) >= 3 {
				tag := strings.ToLower(m[1])
				val := strings.TrimSpace(m[2])
				if tag != "user" && tag != "authenticationsuccess" && tag != "serviceresponse" && tag != "attributes" {
					result.Attributes[tag] = val
				}
			}
		}

		if result.Username == "" {
			return nil, fmt.Errorf("CAS 响应成功但未提取到有效用户名: %s", xmlStr)
		}
		return result, nil
	}

	// 2. 判断是否包含失败标签
	if strings.Contains(xmlStr, "authenticationFailure") {
		result.Success = false
		codeRe := regexp.MustCompile(`code=["']([^"']+)["']`)
		if codeMatches := codeRe.FindStringSubmatch(xmlStr); len(codeMatches) > 1 {
			result.ErrorCode = codeMatches[1]
		}
		msgRe := regexp.MustCompile(`<(?:[a-zA-Z0-9_-]+:)?authenticationFailure[^>]*>\s*([^<]+)\s*</`)
		if msgMatches := msgRe.FindStringSubmatch(xmlStr); len(msgMatches) > 1 {
			result.ErrorMsg = strings.TrimSpace(msgMatches[1])
		} else {
			result.ErrorMsg = "CAS 票据校验失败"
		}
		return result, nil
	}

	// 3. 既无 Success 也无 Failure，尝试 XML 反序列化
	var raw struct {
		XMLName xml.Name `xml:"serviceResponse"`
	}
	_ = xml.Unmarshal([]byte(xmlStr), &raw)

	return &CasAuthResult{
		Success:  false,
		ErrorMsg: fmt.Sprintf("未知 CAS 响应结构: %s", xmlStr),
	}, nil
}

// CasTestConnectivity 连通性测试：检查 CAS Server 服务是否可达
func CasTestConnectivity(cfg *CasConfig) error {
	if cfg.BaseURL == "" {
		return fmt.Errorf("CAS Base URL 不能为空")
	}

	targetURL := cfg.BaseURL + cfg.LoginPath
	client := &http.Client{
		Timeout: time.Duration(cfg.Timeout) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.SkipTLSVerify},
		},
		// 不自动跟随重定向，200 或 302 均说明服务正常
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(targetURL)
	if err != nil {
		return fmt.Errorf("连接 CAS 服务 [%s] 失败: %w", targetURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return nil
	}

	return fmt.Errorf("CAS 服务返回了异常状态码: %d", resp.StatusCode)
}

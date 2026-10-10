package service

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
)

// RadiusConfig RADIUS 认证方式的配置结构
// 对应 AuthMethod.ConfigJSON 字段内容
type RadiusConfig struct {
	Host          string `json:"host"`            // RADIUS 服务器地址
	Port          int    `json:"port"`            // RADIUS 服务器端口，默认 1812
	Secret        string `json:"secret"`          // 共享密钥 (Shared Secret)
	Timeout       int    `json:"timeout"`         // 超时秒数，默认 5
	Retries       int    `json:"retries"`         // 重试次数，默认 3
	NASIdentifier string `json:"nas_identifier"`  // NAS 标识符，默认 "ang-admin"
	NASIPAddress  string `json:"nas_ip_address"`  // NAS IP 地址（可选）
	NASIP         string `json:"nas_ip"`          // 兼容前端配置字段 nas_ip
	AuthProtocol  string `json:"auth_protocol"`   // PAP / CHAP 认证协议，默认 PAP
}

// ParseRadiusConfig 从 JSON 字符串解析 RADIUS 配置
func ParseRadiusConfig(configJSON string) (*RadiusConfig, error) {
	cfg := &RadiusConfig{
		Port:          1812,
		Timeout:       5,
		Retries:       3,
		NASIdentifier: "ang-admin",
		AuthProtocol:  "PAP",
	}
	if configJSON == "" || configJSON == "{}" {
		return cfg, nil
	}
	if err := json.Unmarshal([]byte(configJSON), cfg); err != nil {
		return nil, fmt.Errorf("解析 RADIUS 配置失败: %w", err)
	}
	if cfg.Port <= 0 || cfg.Port > 65535 {
		cfg.Port = 1812
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 5
	}
	if cfg.Retries <= 0 {
		cfg.Retries = 3
	}
	if cfg.NASIdentifier == "" {
		cfg.NASIdentifier = "ang-admin"
	}
	if cfg.NASIPAddress == "" && cfg.NASIP != "" {
		cfg.NASIPAddress = cfg.NASIP
	}
	if cfg.AuthProtocol == "" {
		cfg.AuthProtocol = "PAP"
	}
	return cfg, nil
}

// SetMessageAuthenticator 按照 RFC 2869 / RFC 3579 计算并填充 Message-Authenticator 属性
// 兼容 FreeRADIUS 3.2+、Cisco ISE、Windows NPS 等开启了强制校验的环境
func SetMessageAuthenticator(packet *radius.Packet, secret []byte) error {
	zeros := make([]byte, 16)
	if err := rfc2869.MessageAuthenticator_Set(packet, zeros); err != nil {
		return err
	}
	b, err := packet.Encode()
	if err != nil {
		return err
	}
	mac := hmac.New(md5.New, secret)
	mac.Write(b)
	hash := mac.Sum(nil)
	return rfc2869.MessageAuthenticator_Set(packet, hash)
}

// RadiusAuthResult RADIUS 认证结果
type RadiusAuthResult struct {
	Success bool
	// RADIUS 服务器返回的 Reply-Message（如果有）
	ReplyMessage string
	// 原始响应码描述
	Code string
}

// RadiusAuthenticate 使用 RADIUS 协议验证用户名/密码
// 默认采用 RFC 2865 PAP 认证，同时附加 RFC 2869 Message-Authenticator 确保兼容各类企业 RADIUS 服务
func RadiusAuthenticate(cfg *RadiusConfig, username, password string) (*RadiusAuthResult, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("RADIUS 服务器地址不能为空")
	}
	if cfg.Secret == "" {
		return nil, fmt.Errorf("RADIUS 共享密钥不能为空")
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	timeout := time.Duration(cfg.Timeout) * time.Second

	// 构造 RADIUS Access-Request 包
	packet := radius.New(radius.CodeAccessRequest, []byte(cfg.Secret))

	// 设置用户名属性 (User-Name, Attribute 1)
	if err := rfc2865.UserName_SetString(packet, username); err != nil {
		return nil, fmt.Errorf("设置 User-Name 失败: %w", err)
	}

	// 设置用户密码属性 (User-Password, Attribute 2)，PAP 方式
	if err := rfc2865.UserPassword_SetString(packet, password); err != nil {
		return nil, fmt.Errorf("设置 User-Password 失败: %w", err)
	}

	// 设置 NAS 标识符 (NAS-Identifier, Attribute 32)
	if cfg.NASIdentifier != "" {
		if err := rfc2865.NASIdentifier_SetString(packet, cfg.NASIdentifier); err != nil {
			return nil, fmt.Errorf("设置 NAS-Identifier 失败: %w", err)
		}
	}

	// 设置 NAS IP 地址 (NAS-IP-Address, Attribute 4)
	if cfg.NASIPAddress != "" {
		if ip := net.ParseIP(strings.TrimSpace(cfg.NASIPAddress)); ip != nil {
			if ip4 := ip.To4(); ip4 != nil {
				_ = rfc2865.NASIPAddress_Set(packet, ip4)
			}
		}
	}

	// 设置 NAS-Port-Type 为 Virtual (19)
	if err := rfc2865.NASPortType_Set(packet, rfc2865.NASPortType_Value_Virtual); err != nil {
		return nil, fmt.Errorf("设置 NAS-Port-Type 失败: %w", err)
	}

	// 计算并添加 Message-Authenticator 属性 (RFC 2869 / RFC 3579)
	_ = SetMessageAuthenticator(packet, []byte(cfg.Secret))

	// 发送请求并等待响应
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	response, err := radius.Exchange(ctx, packet, addr)
	if err != nil {
		return nil, fmt.Errorf("RADIUS 请求失败 [%s]: %w", addr, err)
	}

	result := &RadiusAuthResult{
		Code: response.Code.String(),
	}

	// 提取 Reply-Message（如果 RADIUS 服务器有返回）
	if msg := rfc2865.ReplyMessage_GetString(response); msg != "" {
		result.ReplyMessage = msg
	}

	switch response.Code {
	case radius.CodeAccessAccept:
		result.Success = true
	case radius.CodeAccessReject:
		result.Success = false
	case radius.CodeAccessChallenge:
		result.Success = false
		result.ReplyMessage = "RADIUS 服务器发起 Challenge，暂不支持多步认证"
	default:
		result.Success = false
		result.ReplyMessage = fmt.Sprintf("收到未知响应码: %s", response.Code.String())
	}

	return result, nil
}

// RadiusTestConnectivity 仅测试 RADIUS 服务器连通性
// 发送 Access-Request，只要服务器有合法回包（无论 Accept/Reject）即视为连通
func RadiusTestConnectivity(cfg *RadiusConfig) error {
	if cfg.Host == "" {
		return fmt.Errorf("RADIUS 服务器地址不能为空")
	}
	if cfg.Secret == "" {
		return fmt.Errorf("RADIUS 共享密钥不能为空")
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	timeout := time.Duration(cfg.Timeout) * time.Second

	packet := radius.New(radius.CodeAccessRequest, []byte(cfg.Secret))
	_ = rfc2865.UserName_SetString(packet, "__ang_connectivity_probe__")
	_ = rfc2865.UserPassword_SetString(packet, "__probe__")
	_ = rfc2865.NASIdentifier_SetString(packet, cfg.NASIdentifier)
	if cfg.NASIPAddress != "" {
		if ip := net.ParseIP(strings.TrimSpace(cfg.NASIPAddress)); ip != nil {
			if ip4 := ip.To4(); ip4 != nil {
				_ = rfc2865.NASIPAddress_Set(packet, ip4)
			}
		}
	}
	_ = rfc2865.NASPortType_Set(packet, rfc2865.NASPortType_Value_Virtual)
	_ = SetMessageAuthenticator(packet, []byte(cfg.Secret))

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := radius.Exchange(ctx, packet, addr)
	if err != nil {
		return fmt.Errorf("RADIUS 服务 [%s] 连通性测试失败: %w", addr, err)
	}

	switch resp.Code {
	case radius.CodeAccessAccept, radius.CodeAccessReject, radius.CodeAccessChallenge:
		return nil
	default:
		return fmt.Errorf("RADIUS 服务响应了未知的响应码: %s", resp.Code.String())
	}
}

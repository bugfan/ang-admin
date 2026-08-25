package service

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bugfan/ang-admin/models"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns"
	"github.com/go-acme/lego/v4/registration"
)

type AcmeUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *AcmeUser) GetEmail() string {
	return u.Email
}

func (u *AcmeUser) GetRegistration() *registration.Resource {
	return u.Registration
}

func (u *AcmeUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

type IssueCertificateRequest struct {
	CertId        string            `json:"cert_id"`
	Email         string            `json:"email"`
	DirectoryUrl  string            `json:"directory_url"`
	KeyType       string            `json:"key_type"`
	ChallengeType string            `json:"challenge_type"` // "DNS-01", "HTTP-01"
	DnsProvider   string            `json:"dns_provider"`   // e.g. "alidns", "dnspod", "cloudflare", "huaweicloud"
	DnsEnvMap     map[string]string `json:"dns_env_map"`
	Domains       []string          `json:"domains"`
	DisableCname  bool              `json:"disable_cname"` // 禁用 CNAME 别名跳转 (防止跨账号找不到 Zone)
	SaveCert      bool              `json:"save_cert"`
	AcmeConfigId  int64             `json:"acme_config_id"` // 关联的配置项 ID (选填，用于自动更新状态)
}

type IssueCertificateResponse struct {
	CertId      string    `json:"cert_id"`
	Domain      string    `json:"domain"`
	CertContent string    `json:"cert_content"`
	KeyContent  string    `json:"key_content"`
	Issuer      string    `json:"issuer"`
	NotBefore   time.Time `json:"not_before"`
	NotAfter    time.Time `json:"not_after"`
	SANs        string    `json:"sans"`
}

func normalizeDnsEnv(provider string, envMap map[string]string) (string, map[string]string) {
	normProvider := strings.ToLower(strings.TrimSpace(provider))
	resultEnv := make(map[string]string)
	for k, v := range envMap {
		kClean := strings.TrimSpace(k)
		vClean := strings.TrimSpace(v)
		if kClean != "" {
			resultEnv[kClean] = vClean
		}
	}

	// 智能适配腾讯云与 DNSPod 凭证
	if normProvider == "tencentcloud" || normProvider == "tencent" || normProvider == "qcloud" {
		normProvider = "tencentcloud"
		if id, ok := resultEnv["SECRET_ID"]; ok {
			resultEnv["TENCENTCLOUD_SECRET_ID"] = id
		}
		if key, ok := resultEnv["SECRET_KEY"]; ok {
			resultEnv["TENCENTCLOUD_SECRET_KEY"] = key
		}
		if id, ok := resultEnv["TENCENTCLOUD_SECRET_ID"]; ok {
			resultEnv["TENCENTCLOUD_SECRET_ID"] = id
		}
		if key, ok := resultEnv["TENCENTCLOUD_SECRET_KEY"]; ok {
			resultEnv["TENCENTCLOUD_SECRET_KEY"] = key
		}
	} else if normProvider == "dnspod" {
		// 若用户在 DNSPod 中填写了以 AKID 开头的腾讯云 SecretId，自动转换为 tencentcloud 驱动
		if token, ok := resultEnv["DNSPOD_API_KEY"]; ok && strings.HasPrefix(token, "AKID") {
			normProvider = "tencentcloud"
			resultEnv["TENCENTCLOUD_SECRET_ID"] = token
			if secretKey, has := resultEnv["DNSPOD_SECRET_KEY"]; has {
				resultEnv["TENCENTCLOUD_SECRET_KEY"] = secretKey
			}
		}
		if secId, ok := resultEnv["TENCENTCLOUD_SECRET_ID"]; ok && secId != "" {
			normProvider = "tencentcloud"
		}
	}

	// 智能适配阿里云 AccessKey 凭证
	if normProvider == "alidns" || normProvider == "aliyun" || normProvider == "alicloud" {
		normProvider = "alidns"
		if ak, ok := resultEnv["ALICLOUD_ACCESS_KEY_ID"]; ok {
			resultEnv["ALICLOUD_ACCESS_KEY"] = ak
		} else if ak, ok := resultEnv["ALICLOUD_ACCESS_KEY"]; ok {
			resultEnv["ALICLOUD_ACCESS_KEY_ID"] = ak
		}
		if sk, ok := resultEnv["ALICLOUD_ACCESS_KEY_SECRET"]; ok {
			resultEnv["ALICLOUD_SECRET_KEY"] = sk
		} else if sk, ok := resultEnv["ALICLOUD_SECRET_KEY"]; ok {
			resultEnv["ALICLOUD_ACCESS_KEY_SECRET"] = sk
		}
	}

	// 智能适配 Cloudflare Token 凭证
	if normProvider == "cloudflare" || normProvider == "cf" {
		normProvider = "cloudflare"
		if token, ok := resultEnv["CLOUDFLARE_API_TOKEN"]; ok {
			resultEnv["CLOUDFLARE_DNS_API_TOKEN"] = token
		} else if token, ok := resultEnv["CF_DNS_API_TOKEN"]; ok {
			resultEnv["CLOUDFLARE_DNS_API_TOKEN"] = token
		}
	}

	// 智能适配华为云凭证
	if normProvider == "huaweicloud" || normProvider == "huawei" {
		normProvider = "huaweicloud"
		if ak, ok := resultEnv["HUAWEICLOUD_ACCESS_KEY"]; ok {
			resultEnv["HUAWEICLOUD_ACCESS_KEY_ID"] = ak
		}
		if sk, ok := resultEnv["HUAWEICLOUD_SECRET_KEY"]; ok {
			resultEnv["HUAWEICLOUD_SECRET_ACCESS_KEY"] = sk
		}
	}

	return normProvider, resultEnv
}

func IssueAcmeCertificate(req *IssueCertificateRequest) (*IssueCertificateResponse, error) {
	if len(req.Domains) == 0 {
		return nil, fmt.Errorf("必须提供至少一个有效域名")
	}
	if req.Email == "" {
		return nil, fmt.Errorf("必须提供 ACME 注册邮箱")
	}

	// 1. 生成 ACME 账户私钥
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("生成 ACME 私钥失败: %w", err)
	}

	user := &AcmeUser{
		Email: req.Email,
		key:   privateKey,
	}

	// 2. 配置 lego Client
	config := lego.NewConfig(user)
	if req.DirectoryUrl != "" {
		config.CADirURL = req.DirectoryUrl
	} else {
		config.CADirURL = lego.LEDirectoryProduction
	}

	switch strings.ToUpper(req.KeyType) {
	case "RSA2048":
		config.Certificate.KeyType = certcrypto.RSA2048
	case "RSA4096":
		config.Certificate.KeyType = certcrypto.RSA4096
	case "EC384":
		config.Certificate.KeyType = certcrypto.EC384
	default:
		config.Certificate.KeyType = certcrypto.EC256
	}

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("创建 ACME 客户端失败: %w", err)
	}

	// 3. 注册 ACME 账户
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return nil, fmt.Errorf("注册 ACME 账户失败: %w", err)
	}
	user.Registration = reg

	// 4. 配置验证模式 (DNS-01 或 HTTP-01)
	challengeType := strings.ToUpper(strings.TrimSpace(req.ChallengeType))
	if challengeType == "" || challengeType == "DNS-01" {
		if req.DnsProvider == "" {
			return nil, fmt.Errorf("DNS-01 验证模式下必须指定 DNS 提供商")
		}

		providerName, normalizedEnv := normalizeDnsEnv(req.DnsProvider, req.DnsEnvMap)

		// 默认禁用 CNAME 别名跳转 (防止跨账号找不到 Zone 报错)，或根据请求配置
		if req.DisableCname || !strings.EqualFold(os.Getenv("LEGO_DISABLE_CNAME_SUPPORT"), "false") {
			os.Setenv("LEGO_DISABLE_CNAME_SUPPORT", "true")
		}

		// 注入 DNS 环境变量
		for k, v := range normalizedEnv {
			if strings.TrimSpace(k) != "" {
				os.Setenv(strings.TrimSpace(k), strings.TrimSpace(v))
			}
		}

		provider, err := dns.NewDNSChallengeProviderByName(providerName)
		if err != nil {
			return nil, fmt.Errorf("获取 DNS 提供商 '%s' 校验逻辑失败 (请检查 API 密钥配置是否完整): %w", providerName, err)
		}

		dnsOpts := []dns01.ChallengeOption{
			dns01.AddRecursiveNameservers([]string{"223.5.5.5:53", "119.29.29.29:53", "8.8.8.8:53", "1.1.1.1:53"}),
			dns01.DisableAuthoritativeNssPropagationRequirement(),
		}

		err = client.Challenge.SetDNS01Provider(provider, dnsOpts...)
		if err != nil {
			return nil, fmt.Errorf("设置 DNS-01 验证器失败: %w", err)
		}
	} else {
		return nil, fmt.Errorf("不支持的验证方式: %s, 目前仅支持 DNS-01", challengeType)
	}

	// 5. 申请签发证书
	obtainReq := certificate.ObtainRequest{
		Domains: req.Domains,
		Bundle:  true,
	}

	certificates, err := client.Certificate.Obtain(obtainReq)
	if err != nil {
		return nil, fmt.Errorf("ACME 验证/签发证书失败: %w", err)
	}

	certPEM := string(certificates.Certificate)
	keyPEM := string(certificates.PrivateKey)

	// 6. 解析证书元数据
	certModel := &models.Certificate{
		CertContent: certPEM,
		KeyContent:  keyPEM,
	}
	certModel.ParseCertInfo()

	certId := strings.TrimSpace(req.CertId)
	if certId == "" {
		certId = fmt.Sprintf("acme-%s", strings.ReplaceAll(req.Domains[0], "*.", "wildcard-"))
	}

	// 7. 保存至数据库并同步集群 (如勾选)
	if req.SaveCert {
		engine := models.GetEngine()
		var existing models.Certificate
		has, err := engine.Where("cert_id = ?", certId).Get(&existing)
		if err == nil && has {
			existing.Type = "STD"
			existing.KeyContent = keyPEM
			existing.CertContent = certPEM
			existing.Source = "ACME"
			existing.Remark = fmt.Sprintf("ACME Auto Issued via %s (%s)", req.DnsProvider, time.Now().Format("2006-01-02 15:04:05"))
			existing.ParseCertInfo()
			_, _ = engine.ID(existing.Id).AllCols().Update(&existing)
		} else {
			newCert := &models.Certificate{
				CertId:      certId,
				Type:        "STD",
				KeyContent:  keyPEM,
				CertContent: certPEM,
				Source:      "ACME",
				Remark:      fmt.Sprintf("ACME Auto Issued via %s (%s)", req.DnsProvider, time.Now().Format("2006-01-02 15:04:05")),
			}
			newCert.ParseCertInfo()
			_, _ = engine.Insert(newCert)
		}
		SyncCertificateToCluster()
	}


	return &IssueCertificateResponse{
		CertId:      certId,
		Domain:      certModel.SubjectCN,
		CertContent: certPEM,
		KeyContent:  keyPEM,
		Issuer:      certModel.Issuer,
		NotBefore:   certModel.NotBefore,
		NotAfter:    certModel.NotAfter,
		SANs:        certModel.SANs,
	}, nil
}

// 根据已保存的 Certificate ID 执行一键签发
func IssueAcmeCertificateForId(certId int64) (*IssueCertificateResponse, error) {
	var cert models.Certificate
	has, err := models.GetEngine().ID(certId).Get(&cert)
	if err != nil || !has {
		return nil, fmt.Errorf("未找到对应的证书 (ID: %d)", certId)
	}

	if cert.Source != "ACME" {
		return nil, fmt.Errorf("该证书不是 ACME 自动签发的证书")
	}

	var email, directoryUrl, keyType, challengeType, provider, dnsEnv string
	var disableCname bool

	// If an ACME Issuance Config (DnsProvider) is linked, load its settings
	if cert.AcmeAccountId > 0 {
		var dp models.AcmeAccount
		hasDP, errDP := models.GetEngine().ID(cert.AcmeAccountId).Get(&dp)
		if errDP == nil && hasDP {
			provider = dp.Provider
			dnsEnv = dp.DnsEnv
			email = dp.Email
			directoryUrl = dp.DirectoryUrl
			keyType = dp.KeyType
			challengeType = dp.ChallengeType
		} else {
			return nil, fmt.Errorf("关联的 DNS 凭证不存在")
		}
	} else {
		return nil, fmt.Errorf("未关联 DNS 凭证，无法签发 ACME 证书")
	}

	var envMap map[string]string
	if dnsEnv != "" {
		_ = json.Unmarshal([]byte(dnsEnv), &envMap)
	}
	if envMap == nil {
		envMap = make(map[string]string)
	}

	domains := strings.FieldsFunc(cert.Domains, func(r rune) bool {
		return r == ',' || r == '\n' || r == ';' || r == ' '
	})

	if len(domains) == 0 {
		return nil, fmt.Errorf("未指定签发域名")
	}

	req := &IssueCertificateRequest{
		CertId:        cert.CertId,
		Email:         email,
		DirectoryUrl:  directoryUrl,
		KeyType:       keyType,
		ChallengeType: challengeType,
		DnsProvider:   provider,
		DnsEnvMap:     envMap,
		Domains:       domains,
		DisableCname:  disableCname,
		SaveCert:      true,
		AcmeConfigId:  0, // We no longer have AcmeConfigId
	}

	resp, err := IssueAcmeCertificate(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// 定时巡检与自动续签逻辑
func CheckAndRenewAcmeCertificates() {
	var certs []models.Certificate
	err := models.GetEngine().Where("auto_renew = ? AND source = ?", true, "ACME").Find(&certs)
	if err != nil {
		log.Printf("[ACME Auto-Renew] 查询自动续签证书列表失败: %v\n", err)
		return
	}

	if len(certs) == 0 {
		return
	}

	log.Printf("[ACME Auto-Renew] 发现 %d 个启用了自动续签的 ACME 证书，正在检查到期状态...\n", len(certs))
	for _, cert := range certs {
		renewDays := cert.RenewDays
		if renewDays <= 0 {
			renewDays = 30
		}

		shouldRenew := false
		if cert.KeyContent == "" || cert.CertContent == "" {
			log.Printf("[ACME Auto-Renew] 证书【%s】尚未生成内容，触发首次自动签发\n", cert.CertId)
			shouldRenew = true
		} else {
			remainingTime := time.Until(cert.NotAfter)
			threshold := time.Duration(renewDays) * 24 * time.Hour
			if remainingTime <= threshold {
				log.Printf("[ACME Auto-Renew] 证书【%s】将于 %v 后到期 (设定阈值: %d 天)，触发自动续签覆盖更新\n", cert.CertId, remainingTime.Round(time.Hour), renewDays)
				shouldRenew = true
			}
		}

		if shouldRenew {
			go func(c models.Certificate) {
				log.Printf("[ACME Auto-Renew] 正在执行自动续签: %s (ID: %d)\n", c.CertId, c.Id)
				resp, err := IssueAcmeCertificateForId(c.Id)
				if err != nil {
					log.Printf("[ACME Auto-Renew] 自动续签失败: %s, 错误: %v\n", c.CertId, err)
				} else {
					log.Printf("[ACME Auto-Renew] 自动续签成功: %s, 证书: %s, 新有效期至: %s\n", c.CertId, resp.CertId, resp.NotAfter.Format("2006-01-02 15:04:05"))
				}
			}(cert)
		}
	}
}

// 启动自动续签后台 Cron 定时巡检
func StartAcmeAutoRenewCron() {
	// 启动后 1 分钟执行首次巡检
	time.AfterFunc(1*time.Minute, func() {
		CheckAndRenewAcmeCertificates()
	})

	// 每 12 小时自动巡检一次
	ticker := time.NewTicker(12 * time.Hour)
	go func() {
		for range ticker.C {
			CheckAndRenewAcmeCertificates()
		}
	}()
	log.Println("[ACME Auto-Renew] 证书自动续签定时巡检服务已启动 (每 12 小时检查一次到期状态)")
}

package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bugfan/ang-admin/cluster"
	"github.com/bugfan/ang-admin/entity"
	"github.com/bugfan/ang-admin/models"
)

func buildCertMap() map[string]entity.CertConfig {
	engine := models.GetEngine()
	if engine == nil {
		return nil
	}
	var certs []models.Certificate
	err := engine.Find(&certs)
	if err != nil {
		log.Printf("buildCertMap error: %v\n", err)
		return nil
	}

	certMap := make(map[string]entity.CertConfig)
	for _, item := range certs {
		certId := item.CertId
		if certId == "" {
			certId = fmt.Sprintf("id-%d", item.Id)
		}
		certMap[certId] = entity.CertConfig{
			Type: item.Type,
			Data: entity.CertData{
				Key:  item.KeyContent,
				Cert: item.CertContent,
			},
		}
	}
	return certMap
}

func buildTunnelMaps() (map[string]entity.TunnelConfig, map[string]entity.TunnelConfig) {
	engine := models.GetEngine()
	if engine == nil {
		return nil, nil
	}
	var tunnels []models.Tunnel
	err := engine.Find(&tunnels)
	if err != nil {
		log.Printf("buildTunnelMaps error: %v\n", err)
		return nil, nil
	}

	tlsMap := make(map[string]entity.TunnelConfig)
	quicMap := make(map[string]entity.TunnelConfig)

	for _, item := range tunnels {
		keyStr := strconv.FormatInt(item.Id, 10)
		cfg := entity.TunnelConfig{
			Port:        item.Port,
			Certificate: item.Certificate,
		}

		tType := strings.ToUpper(strings.TrimSpace(item.Type))
		if tType == "TLS-TUNNEL" || tType == "TLS" {
			tlsMap[keyStr] = cfg
		} else if tType == "QUIC-TUNNEL" || tType == "QUIC" {
			quicMap[keyStr] = cfg
		} else {
			tlsMap[keyStr] = cfg
		}
	}
	return tlsMap, quicMap
}

func buildRulesDBMap() map[string]models.Rule {
	engine := models.GetEngine()
	if engine == nil {
		return nil
	}
	var rules []models.Rule
	err := engine.Find(&rules)
	if err != nil {
		log.Printf("buildRulesDBMap error: %v\n", err)
		return nil
	}

	rulesMap := make(map[string]models.Rule)
	for _, r := range rules {
		if r.Name != "" {
			rulesMap[r.Name] = r
		}
		rulesMap[strconv.FormatInt(r.Id, 10)] = r
	}
	return rulesMap
}

func buildDNSMap(rulesMap map[string]models.Rule) map[string]entity.DNSConfig {
	engine := models.GetEngine()
	if engine == nil {
		return nil
	}
	var dnsList []models.DnsProxy
	err := engine.Find(&dnsList)
	if err != nil {
		log.Printf("buildDNSMap error: %v\n", err)
		return nil
	}

	dnsMap := make(map[string]entity.DNSConfig)
	for _, item := range dnsList {
		keyStr := strconv.FormatInt(item.Id, 10)

		// Parse Hosts
		var hosts entity.DNSHosts
		if item.HostsJSON != "" {
			_ = json.Unmarshal([]byte(item.HostsJSON), &hosts)
		}

		// Parse Rules (Rule Set expansion)
		var ruleConfigs []entity.RuleConfig
		if item.Rules != "" {
			var ruleNames []string
			_ = json.Unmarshal([]byte(item.Rules), &ruleNames)
			for _, rName := range ruleNames {
				rName = strings.TrimSpace(rName)
				if dbRule, exists := rulesMap[rName]; exists {
					if dbRule.Items != "" {
						var items []entity.RuleConfig
						if err := json.Unmarshal([]byte(dbRule.Items), &items); err == nil {
							ruleConfigs = append(ruleConfigs, items...)
						}
					}
				} else if rName == "ip_matcher" {
					ruleConfigs = append(ruleConfigs, entity.RuleConfig{
						Matcher: entity.MatcherConfig{
							Name: "ip_matcher",
							Config: map[string]interface{}{
								"Address": []string{"121.0.0.1"},
							},
						},
						Action: entity.ActionConfig{
							Name: "reset_conn_action",
							Config: map[string]interface{}{
								"Content": "reset you",
							},
						},
					})
				}
			}
		}

		// Parse Backend
		var backend entity.DNSBackend
		if item.TunnelId != "" {
			backend.Tunnel = &entity.BackendTunnel{
				Type:  item.TunnelType,
				ID:    item.TunnelId,
				Token: item.TunnelToken,
			}
		}

		if item.UpstreamServers != "" {
			var servers []entity.UpstreamServer
			_ = json.Unmarshal([]byte(item.UpstreamServers), &servers)
			if len(servers) > 0 {
				method := item.UpstreamMethod
				if method == "" {
					method = "round_robin"
				}
				backend.Upstream = &entity.UpstreamConfig{
					Method: method,
					Data: entity.UpstreamData{
						Servers: servers,
					},
				}
			}
		}

		dnsMap[keyStr] = entity.DNSConfig{
			Address: item.Address,
			Port:    item.Port,
			Rule:    ruleConfigs,
			Hosts:   &hosts,
			Backend: &backend,
		}
	}
	return dnsMap
}

func buildHTTPMap(rulesMap map[string]models.Rule) map[string]entity.HTTPConfig {
	engine := models.GetEngine()
	if engine == nil {
		return nil
	}
	var httpList []models.HttpProxy
	err := engine.Find(&httpList)
	if err != nil {
		log.Printf("buildHTTPMap error: %v\n", err)
		return nil
	}

	httpMap := make(map[string]entity.HTTPConfig)
	for _, item := range httpList {
		keyStr := strconv.FormatInt(item.Id, 10)

		// Parse ProxyHeaders
		var proxyHeaders []string
		if item.ProxyHeaders != "" {
			_ = json.Unmarshal([]byte(item.ProxyHeaders), &proxyHeaders)
		}

		// Parse Rules
		var ruleConfigs []entity.RuleConfig
		if item.Rules != "" {
			var ruleNames []string
			_ = json.Unmarshal([]byte(item.Rules), &ruleNames)
			for _, rName := range ruleNames {
				rName = strings.TrimSpace(rName)
				if dbRule, exists := rulesMap[rName]; exists {
					if dbRule.Items != "" {
						var items []entity.RuleConfig
						if err := json.Unmarshal([]byte(dbRule.Items), &items); err == nil {
							ruleConfigs = append(ruleConfigs, items...)
						}
					}
				}
			}
		}

		// Parse Backend Locations
		var rawLocations []entity.HTTPLocation
		var locations []entity.HTTPLocation
		if item.LocationJSON != "" {
			if err := json.Unmarshal([]byte(item.LocationJSON), &rawLocations); err == nil {
				for _, loc := range rawLocations {
					uType := loc.Upstream.Type
					if uType == "root" || uType == "alias" {
						// Clean legacy fields if any
						var dir string
						if mData, ok := loc.Upstream.Data.(map[string]interface{}); ok {
							if d, exists := mData["Dir"].(string); exists {
								dir = d
							}
						}
						if dir == "" {
							dir = "./static"
						}
						loc.Upstream.Data = map[string]interface{}{
							"Dir": dir,
						}
					} else {
						// proxy_pass
						var method string
						var servers interface{}
						if mData, ok := loc.Upstream.Data.(map[string]interface{}); ok {
							if m, exists := mData["Method"].(string); exists {
								method = m
							}
							if s, exists := mData["Servers"]; exists {
								servers = s
							}
						}
						if method == "" {
							method = "round_robin"
						}
						loc.Upstream.Type = "proxy_pass"
						loc.Upstream.Data = map[string]interface{}{
							"Method":  method,
							"Servers": servers,
						}
					}
					locations = append(locations, loc)
				}
			}
		}

		// Parse Backend Tunnel
		var backendTunnel *entity.BackendTunnel
		if item.TunnelId != "" {
			backendTunnel = &entity.BackendTunnel{
				Type:  item.TunnelType,
				ID:    item.TunnelId,
				Token: item.TunnelToken,
			}
		}

		var dnsResolver []string
		if item.DNSResolver != "" {
			if err := json.Unmarshal([]byte(item.DNSResolver), &dnsResolver); err != nil {
				// Fallback for old single string data
				dnsResolver = []string{item.DNSResolver}
			}
		}

		httpMap[keyStr] = entity.HTTPConfig{
			Front: entity.HTTPFront{
				Port:         item.Port,
				Hostname:     item.Hostname,
				HTTP:         item.HTTP,
				TLS:          item.TLS,
				H2:           item.H2,
				HSTS:         item.HSTS,
				Certificate:  item.Certificate,
				ProxyHeaders: proxyHeaders,
			},
			Feature: entity.HTTPFeature{
				Compress: item.Compress,
			},
			Rule: ruleConfigs,
			Backend: entity.HTTPBackend{
				RealIp:      item.RealIp,
				Tunnel:      backendTunnel,
				DNSResolver: dnsResolver,
				Location:    locations,
			},
		}
	}
	return httpMap
}

// BuildFullServerConfig builds the combined server.json data structure matching ang engine
func BuildFullServerConfig() entity.ServerConfig {
	certMap := buildCertMap()
	tlsMap, quicMap := buildTunnelMaps()
	rulesMap := buildRulesDBMap()
	dnsMap := buildDNSMap(rulesMap)
	httpMap := buildHTTPMap(rulesMap)

	return entity.ServerConfig{
		TLSTunnel:   tlsMap,
		QUICTunnel:  quicMap,
		DNS:         dnsMap,
		Certificate: certMap,
		HTTP:        httpMap,
	}
}

// PushServerConfigToNodes pushes the compiled server.json to all registered ang engine nodes
func PushServerConfigToNodes(cfg entity.ServerConfig) {
	engine := models.GetEngine()
	if engine == nil {
		return
	}
	var nodes []models.ClusterNode
	if err := engine.Find(&nodes); err != nil || len(nodes) == 0 {
		return
	}

	payload := map[string]interface{}{
		"server_config": cfg,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, node := range nodes {
		addr := strings.TrimRight(node.Addr, "/")
		if addr == "" {
			continue
		}
		syncURL := addr + "/api/config/sync"
		req, err := http.NewRequest("POST", syncURL, bytes.NewBuffer(data))
		if err != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		if node.Secret != "" {
			req.Header.Set("X-Ang-Secret", node.Secret)
		}

		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			node.Status = 1
			node.LastPing = time.Now()
			_, _ = engine.ID(node.Id).Cols("status", "last_ping").Update(&node)
			log.Printf("[cluster_sync] Successfully pushed config to node %s (%s)", node.Name, node.Addr)
		} else {
			if resp != nil {
				resp.Body.Close()
			}
			node.Status = 0
			_, _ = engine.ID(node.Id).Cols("status").Update(&node)
			log.Printf("[cluster_sync] Failed to push config to node %s (%s)", node.Name, node.Addr)
		}
	}
}

// PingNode sends a heartbeat check to an ang engine node
func PingNode(node *models.ClusterNode) (bool, string) {
	return VerifyNode(node.Addr, node.Secret)
}

// SyncCertificateToCluster queries all certificates, updates cluster and prints server.json
func SyncCertificateToCluster() {
	certMap := buildCertMap()
	cluster.Put("Certificate", certMap)
	cfg := BuildFullServerConfig()
	cluster.PrintFullServerConfig(cfg)
	go PushServerConfigToNodes(cfg)
}

// SyncTunnelToCluster queries all tunnels, updates cluster and prints server.json
func SyncTunnelToCluster() {
	tlsMap, quicMap := buildTunnelMaps()
	cluster.Put("TLS-TUNNEL", tlsMap)
	cluster.Put("QUIC-TUNNEL", quicMap)
	cfg := BuildFullServerConfig()
	cluster.PrintFullServerConfig(cfg)
	go PushServerConfigToNodes(cfg)
}

// SyncDNSToCluster queries all DNS proxies, updates cluster and prints server.json
func SyncDNSToCluster() {
	rulesMap := buildRulesDBMap()
	dnsMap := buildDNSMap(rulesMap)
	cluster.Put("DNS", dnsMap)
	cfg := BuildFullServerConfig()
	cluster.PrintFullServerConfig(cfg)
	go PushServerConfigToNodes(cfg)
}

// SyncRuleToCluster queries all rules, updates cluster and prints server.json
func SyncRuleToCluster() {
	rulesMap := buildRulesDBMap()
	cluster.Put("Rule", rulesMap)
	cfg := BuildFullServerConfig()
	cluster.PrintFullServerConfig(cfg)
	go PushServerConfigToNodes(cfg)
}

// SyncHTTPToCluster queries all HTTP proxies, updates cluster and prints server.json
func SyncHTTPToCluster() {
	rulesMap := buildRulesDBMap()
	httpMap := buildHTTPMap(rulesMap)
	cluster.Put("HTTP", httpMap)
	cfg := BuildFullServerConfig()
	cluster.PrintFullServerConfig(cfg)
	go PushServerConfigToNodes(cfg)
}

// SyncAllToCluster syncs all implemented entities to cluster and prints overall server.json
func SyncAllToCluster() {
	SyncCertificateToCluster()
	SyncTunnelToCluster()
	SyncDNSToCluster()
	SyncRuleToCluster()
	SyncHTTPToCluster()
}

func VerifyNode(addr, secret string) (bool, string) {
	addr = strings.TrimRight(addr, "/")
	if addr == "" {
		return false, "empty address"
	}
	verifyURL := addr + "/api/verify"
	client := &http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequest("GET", verifyURL, nil)
	if err != nil {
		return false, err.Error()
	}
	if secret != "" {
		req.Header.Set("X-Ang-Secret", secret)
	}
	resp, err := client.Do(req)
	if err != nil {
		return false, err.Error()
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return true, "success"
	}
	if resp.StatusCode == 401 {
		return false, "auth_failed"
	}
	return false, fmt.Sprintf("http_status_%d", resp.StatusCode)
}

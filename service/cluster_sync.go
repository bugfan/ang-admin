package service

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

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
			SNI:         item.SNI,
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

		// Parse Rules
		var ruleConfigs []entity.RuleConfig
		if item.Rules != "" {
			var ruleNames []string
			_ = json.Unmarshal([]byte(item.Rules), &ruleNames)
			for _, rName := range ruleNames {
				rName = strings.TrimSpace(rName)
				if dbRule, exists := rulesMap[rName]; exists {
					var matcherObj entity.MatcherConfig
					var actionObj entity.ActionConfig
					_ = json.Unmarshal([]byte(dbRule.Matcher), &matcherObj)
					_ = json.Unmarshal([]byte(dbRule.Action), &actionObj)
					ruleConfigs = append(ruleConfigs, entity.RuleConfig{
						Matcher: matcherObj,
						Action:  actionObj,
					})
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
				} else if rName == "reset_conn_action" {
					ruleConfigs = append(ruleConfigs, entity.RuleConfig{
						Matcher: entity.MatcherConfig{
							Name: "ip_matcher",
							Config: map[string]interface{}{
								"Address": []string{"127.0.0.1"},
							},
						},
						Action: entity.ActionConfig{
							Name: "reset_conn_action",
							Config: map[string]interface{}{
								"Content": "reset connection",
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

// BuildFullServerConfig builds the combined server.json data structure matching ang engine
func BuildFullServerConfig() entity.ServerConfig {
	certMap := buildCertMap()
	tlsMap, quicMap := buildTunnelMaps()
	rulesMap := buildRulesDBMap()
	dnsMap := buildDNSMap(rulesMap)

	return entity.ServerConfig{
		TLSTunnel:   tlsMap,
		QUICTunnel:  quicMap,
		DNS:         dnsMap,
		Certificate: certMap,
	}
}

// SyncCertificateToCluster queries all certificates, updates cluster and prints server.json
func SyncCertificateToCluster() {
	certMap := buildCertMap()
	cluster.Put("Certificate", certMap)
	cluster.PrintFullServerConfig(BuildFullServerConfig())
}

// SyncTunnelToCluster queries all tunnels, updates cluster and prints server.json
func SyncTunnelToCluster() {
	tlsMap, quicMap := buildTunnelMaps()
	cluster.Put("TLS-TUNNEL", tlsMap)
	cluster.Put("QUIC-TUNNEL", quicMap)
	cluster.PrintFullServerConfig(BuildFullServerConfig())
}

// SyncDNSToCluster queries all DNS proxies, updates cluster and prints server.json
func SyncDNSToCluster() {
	rulesMap := buildRulesDBMap()
	dnsMap := buildDNSMap(rulesMap)
	cluster.Put("DNS", dnsMap)
	cluster.PrintFullServerConfig(BuildFullServerConfig())
}

// SyncRuleToCluster queries all rules, updates cluster and prints server.json
func SyncRuleToCluster() {
	rulesMap := buildRulesDBMap()
	cluster.Put("Rule", rulesMap)
	cluster.PrintFullServerConfig(BuildFullServerConfig())
}

// SyncAllToCluster syncs all implemented entities to cluster and prints overall server.json
func SyncAllToCluster() {
	SyncCertificateToCluster()
	SyncTunnelToCluster()
	SyncDNSToCluster()
	SyncRuleToCluster()
}

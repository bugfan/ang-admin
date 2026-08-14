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

// SyncCertificateToCluster queries all certificates from database and calls cluster.Put("Certificate", certMap)
func SyncCertificateToCluster() {
	engine := models.GetEngine()
	if engine == nil {
		return
	}

	var certs []models.Certificate
	err := engine.Find(&certs)
	if err != nil {
		log.Printf("SyncCertificateToCluster error: %v\n", err)
		return
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

	cluster.Put("Certificate", certMap)
}

// SyncTunnelToCluster queries all tunnels from database and calls cluster.Put("TLS-TUNNEL", tlsMap) and cluster.Put("QUIC-TUNNEL", quicMap)
func SyncTunnelToCluster() {
	engine := models.GetEngine()
	if engine == nil {
		return
	}

	var tunnels []models.Tunnel
	err := engine.Find(&tunnels)
	if err != nil {
		log.Printf("SyncTunnelToCluster error: %v\n", err)
		return
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

	cluster.Put("TLS-TUNNEL", tlsMap)
	cluster.Put("QUIC-TUNNEL", quicMap)
}

// SyncDNSToCluster queries all DNS proxies from database and calls cluster.Put("DNS", dnsMap)
func SyncDNSToCluster() {
	engine := models.GetEngine()
	if engine == nil {
		return
	}

	var dnsList []models.DnsProxy
	err := engine.Find(&dnsList)
	if err != nil {
		log.Printf("SyncDNSToCluster error: %v\n", err)
		return
	}

	dnsMap := make(map[string]entity.DNSConfig)
	for _, item := range dnsList {
		keyStr := strconv.FormatInt(item.Id, 10)

		// Parse Hosts
		var hosts entity.DNSHosts
		if item.HostsJSON != "" {
			_ = json.Unmarshal([]byte(item.HostsJSON), &hosts)
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
			Hosts:   &hosts,
			Backend: &backend,
		}
	}

	cluster.Put("DNS", dnsMap)
}

// SyncAllToCluster syncs all implemented entities to cluster
func SyncAllToCluster() {
	SyncCertificateToCluster()
	SyncTunnelToCluster()
	SyncDNSToCluster()
}

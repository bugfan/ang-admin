package service

import (
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

// SyncAllToCluster syncs all implemented entities to cluster
func SyncAllToCluster() {
	SyncCertificateToCluster()
	SyncTunnelToCluster()
}

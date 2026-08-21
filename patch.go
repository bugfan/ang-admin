package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"github.com/bugfan/ang-admin/models"
)

func fetchActiveConnectionsFromAng() ([]ActiveConnItem, error) {
	var nodes []models.ClusterNode
	_ = models.GetEngine().Find(&nodes)

	if len(nodes) == 0 {
		return []ActiveConnItem{}, fmt.Errorf("no cluster nodes configured")
	}

	var allItems []ActiveConnItem
	var lastErr error
	client := http.Client{Timeout: 3 * time.Second}

	// Fetch DB Tunnel records to dynamically resolve real TunnelId
	var dbTunnels []models.Tunnel
	_ = models.GetEngine().Find(&dbTunnels)

	resolveTunnelID := func(clientType string, rawGroupID string, localAddr string, sni string) string {
		rawGroupIDStr := strings.TrimSpace(rawGroupID)
		normType := strings.ToLower(clientType) // "tls" or "quic"

		// 1. Direct ID match first
		for _, t := range dbTunnels {
			tIDStr := fmt.Sprintf("%d", t.Id)
			if tIDStr == rawGroupIDStr {
				return tIDStr
			}
		}

		// 2. Fallback match by Port or SNI and Type
		connPort := ""
		if idx := strings.LastIndex(localAddr, ":"); idx != -1 {
			connPort = localAddr[idx+1:]
		}

		for _, t := range dbTunnels {
			tTypeNorm := strings.ToLower(t.Type)
			matchType := (strings.Contains(tTypeNorm, "tls") && normType == "tls") ||
				(strings.Contains(tTypeNorm, "quic") && normType == "quic")
			if !matchType {
				continue
			}

			if connPort != "" && t.Port == connPort {
				return fmt.Sprintf("%d", t.Id)
			}
			if sni != "" && t.SNI == sni {
				return fmt.Sprintf("%d", t.Id)
			}
		}

		return rawGroupIDStr
	}

	// Fetch from all nodes and aggregate
	for _, node := range nodes {
		addr := strings.TrimRight(node.Addr, "/")
		if addr == "" {
			continue
		}

		req, err := http.NewRequest("GET", addr+"/tunnel", nil)
		if err != nil {
			lastErr = err
			continue
		}
		if node.Secret != "" {
			req.Header.Set("X-Ang-Secret", node.Secret)
		}
		
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("node %s returned status %d", node.Name, resp.StatusCode)
			continue
		}

		var angData AngTunnelResponse
		err = json.NewDecoder(resp.Body).Decode(&angData)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		// Process TLS
		for _, group := range angData.TLS {
			for _, conn := range group.Connections {
				realTunnelID := resolveTunnelID("tls", group.ID, conn.LocalAddr, conn.SNI)
				label := fmt.Sprintf("[%s] [TLS] Tunnel: %s | Token: %s | Remote: %s", node.Name, realTunnelID, conn.Token, conn.RemoteAddr)
				if conn.SNI != "" {
					label += fmt.Sprintf(" (%s)", conn.SNI)
				}
				allItems = append(allItems, ActiveConnItem{
					Type:       "tls",
					TunnelId:   realTunnelID,
					Token:      conn.Token,
					RemoteAddr: conn.RemoteAddr,
					LocalAddr:  conn.LocalAddr,
					SNI:        conn.SNI,
					Label:      label,
				})
			}
		}

		// Process QUIC
		for _, group := range angData.QUIC {
			for _, conn := range group.Connections {
				realTunnelID := resolveTunnelID("quic", group.ID, conn.LocalAddr, conn.SNI)
				label := fmt.Sprintf("[%s] [QUIC] Tunnel: %s | Token: %s | Remote: %s", node.Name, realTunnelID, conn.Token, conn.RemoteAddr)
				if conn.SNI != "" {
					label += fmt.Sprintf(" (%s)", conn.SNI)
				}
				allItems = append(allItems, ActiveConnItem{
					Type:       "quic",
					TunnelId:   realTunnelID,
					Token:      conn.Token,
					RemoteAddr: conn.RemoteAddr,
					LocalAddr:  conn.LocalAddr,
					SNI:        conn.SNI,
					Label:      label,
				})
			}
		}
	}

	if len(allItems) == 0 && lastErr != nil {
		return allItems, lastErr
	}
	return allItems, nil
}

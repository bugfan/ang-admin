package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bugfan/ang-admin/models"
	"github.com/bugfan/rest"
	"github.com/gin-gonic/gin"
	"github.com/go-xorm/xorm"
)

func init() {
	rest.Register(&models.TunnelClient{}, &tunnelClientHandler{}, rest.RouteTypeALL, nil, "tunnel-client")
}

type tunnelClientHandler struct {
	Id         int64     `json:"id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	TunnelId   string    `json:"tunnel_id"`
	Token      string    `json:"token"`
	Remark     string    `json:"remark"`
	IsOnline   bool      `json:"is_online"`
	IsSaved    bool      `json:"is_saved"`
	RemoteAddr string    `json:"remote_addr"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (t *tunnelClientHandler) Before(g *gin.Context, x *xorm.Engine) bool {
	method := g.Request.Method
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
		name := strings.TrimSpace(t.Name)
		if name == "" {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{"code": 1, "message": "节点名称不能为空"})
			return false
		}

		token := strings.TrimSpace(t.Token)
		if token == "" {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{"code": 1, "message": "请输入 Token"})
			return false
		}

		tunnelId := strings.TrimSpace(t.TunnelId)
		if tunnelId == "" {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{"code": 1, "message": "请选择所属服务器"})
			return false
		}

		// 1. 检查节点名称唯一性
		nameSess := x.Where("name = ?", name)
		if t.Id > 0 {
			nameSess.And("id != ?", t.Id)
		}
		var nameExist models.TunnelClient
		hasName, err := nameSess.Get(&nameExist)
		if err == nil && hasName {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{"code": 1, "message": fmt.Sprintf("节点名称 [%s] 已存在，请输入唯一的节点名称", name)})
			return false
		}

		// 2. 检查同一 Tunnel 下 Token 唯一性 (防止重复添加/关联)
		tokenSess := x.Where("tunnel_id = ? AND token = ?", tunnelId, token)
		if t.Id > 0 {
			tokenSess.And("id != ?", t.Id)
		}
		var tokenExist models.TunnelClient
		hasToken, err := tokenSess.Get(&tokenExist)
		if err == nil && hasToken {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{"code": 1, "message": fmt.Sprintf("该客户端节点 (Token: %s) 已经在当前隧道下绑定过，不能重复添加", token)})
			return false
		}
	}
	return true
}

func (t *tunnelClientHandler) List(c *gin.Context) {
	var clients []models.TunnelClient
	session := models.GetEngine().NewSession()
	defer session.Close()

	if name := c.Query("name"); name != "" {
		session.Where("name LIKE ?", "%"+name+"%")
	}
	if clientType := c.Query("type"); clientType != "" {
		session.Where("type = ?", clientType)
	}
	if tunnelId := c.Query("tunnel_id"); tunnelId != "" {
		session.Where("tunnel_id = ?", tunnelId)
	}
	if token := c.Query("token"); token != "" {
		session.Where("token LIKE ?", "%"+token+"%")
	}

	err := session.Desc("id").Find(&clients)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": "查询TunnelClient列表失败: " + err.Error()})
		return
	}

	// Fetch live active connections from ang engine
	activeList, _ := fetchActiveConnectionsFromAng()
	activeMap := make(map[string]string) // key: "type|tunnel_id|token" -> remote_addr
	for _, item := range activeList {
		clientTypeLower := strings.ToLower(item.Type)
		if clientTypeLower == "tls-tunnel" {
			clientTypeLower = "tls"
		} else if clientTypeLower == "quic-tunnel" {
			clientTypeLower = "quic"
		}
		key := fmt.Sprintf("%s|%s|%s", clientTypeLower, item.TunnelId, item.Token)
		activeMap[key] = item.RemoteAddr
	}

	resList := make([]tunnelClientHandler, 0, len(clients))
	for _, item := range clients {
		clientTypeLower := strings.ToLower(item.Type)
		if clientTypeLower == "tls-tunnel" {
			clientTypeLower = "tls"
		} else if clientTypeLower == "quic-tunnel" {
			clientTypeLower = "quic"
		}
		key := fmt.Sprintf("%s|%s|%s", clientTypeLower, item.TunnelId, item.Token)
		remoteAddr, isOnline := activeMap[key]

		resList = append(resList, tunnelClientHandler{
			Id:         item.Id,
			Name:       item.Name,
			Type:       item.Type,
			TunnelId:   item.TunnelId,
			Token:      item.Token,
			Remark:     item.Remark,
			IsOnline:   isOnline,
			IsSaved:    true,
			RemoteAddr: remoteAddr,
			CreatedAt:  item.CreatedAt,
			UpdatedAt:  item.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, resList)
}

type ActiveConnItem struct {
	Type       string `json:"type"`        // "tls" 或 "quic"
	TunnelId   string `json:"tunnel_id"`   // Tunnel ID (如 "1")
	Token      string `json:"token"`       // Token
	RemoteAddr string `json:"remote_addr"` // 客户端 remote address
	LocalAddr  string `json:"local_addr"`  // 本地监听 address
	SNI        string `json:"sni"`         // SNI
	Label      string `json:"label"`       // 界面下拉展示 label
}

type AngTunnelConn struct {
	Token      string `json:"token"`
	RemoteAddr string `json:"remote_addr"`
	LocalAddr  string `json:"local_addr"`
	SNI        string `json:"sni"`
}

type AngTunnelGroup struct {
	ID          string          `json:"id"`
	Connections []AngTunnelConn `json:"connections"`
}

type AngTunnelResponse struct {
	TLS  []AngTunnelGroup `json:"tls"`
	QUIC []AngTunnelGroup `json:"quic"`
}

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
func GetActiveTunnelConnectionsHandler(c *gin.Context) {
	list, err := fetchActiveConnectionsFromAng()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "success (ang engine offline or unready)",
			"data":    []ActiveConnItem{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    list,
	})
}

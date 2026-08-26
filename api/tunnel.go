package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bugfan/ang-admin/models"
	"github.com/bugfan/ang-admin/service"
	"github.com/bugfan/rest"
	"github.com/gin-gonic/gin"
	"github.com/go-xorm/xorm"
)

func init() {
	rest.Register(&models.Tunnel{}, &tunnelHandler{}, rest.RouteTypeALL, nil, "tunnel")
}

type tunnelHandler struct {
	Id           int64                 `json:"id"`
	Name         string                `json:"name"`
	Type         string                `json:"type"`
	Port         string                `json:"port"`
	Certificate  string                `json:"certificate,omitempty"`
	Auth         bool                  `json:"auth"`
	Remark       string                `json:"remark"`
	ClientNodes  []tunnelClientHandler `json:"client_nodes"`
	OnlineCount  int                   `json:"online_count"`
	UnsavedCount int                   `json:"unsaved_count"`
	TotalCount   int                   `json:"total_count"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
}

func (t *tunnelHandler) Before(g *gin.Context, x *xorm.Engine) bool {
	return true
}

func (t *tunnelHandler) After(g *gin.Context, x *xorm.Engine, args ...interface{}) {
	if g.Request.Method == http.MethodPost || g.Request.Method == http.MethodPut || g.Request.Method == http.MethodPatch || g.Request.Method == http.MethodDelete {
		service.SyncTunnelToCluster()
	}
}

func (t *tunnelHandler) List(c *gin.Context) {
	var tunnels []models.Tunnel
	session := models.GetEngine().NewSession()
	defer session.Close()

	if name := c.Query("name"); name != "" {
		session.Where("name LIKE ?", "%"+name+"%")
	}

		if tunnelType := strings.TrimSpace(c.Query("type")); tunnelType != "" {
			upperType := strings.ToUpper(tunnelType)
			if upperType == "TLS" || upperType == "TLS-TUNNEL" {
				session.Where("type LIKE ? OR type LIKE ?", "%TLS%", "%tls%")
			} else if upperType == "QUIC" || upperType == "QUIC-TUNNEL" {
				session.Where("type LIKE ? OR type LIKE ?", "%QUIC%", "%quic%")
			} else {
				session.Where("type = ?", tunnelType)
			}
		}
	if sni := c.Query("sni"); sni != "" {
		session.Where("sni LIKE ?", "%"+sni+"%")
	}
	if port := c.Query("port"); port != "" {
		session.Where("port LIKE ?", "%"+port+"%")
	}

	err := session.Desc("id").Find(&tunnels)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": "查询Tunnel列表失败: " + err.Error()})
		return
	}

	// Fetch all tunnel client records from DB and index by Token
	var allClients []models.TunnelClient
	_ = models.GetEngine().Find(&allClients)
	clientByToken := make(map[string]models.TunnelClient)
	for _, c := range allClients {
		if c.Token != "" {
			clientByToken[c.Token] = c
		}
	}

	// Fetch active connections from ang engine
	activeList, _ := fetchActiveConnectionsFromAng()

	resList := make([]tunnelHandler, 0, len(tunnels))
	for _, item := range tunnels {
		tidStr := fmt.Sprintf("%d", item.Id)
		srvTypeNorm := strings.ToLower(item.Type)
		if strings.Contains(srvTypeNorm, "tls") {
			srvTypeNorm = "tls"
		} else if strings.Contains(srvTypeNorm, "quic") {
			srvTypeNorm = "quic"
		}

		var matchedNodes []tunnelClientHandler
		onlineCount := 0
		unsavedCount := 0

		for _, activeItem := range activeList {
			actTypeNorm := strings.ToLower(activeItem.Type)
			if strings.Contains(actTypeNorm, "tls") {
				actTypeNorm = "tls"
			} else if strings.Contains(actTypeNorm, "quic") {
				actTypeNorm = "quic"
			}

			if activeItem.TunnelId == tidStr && actTypeNorm == srvTypeNorm {
				onlineCount++
				if savedClient, isSaved := clientByToken[activeItem.Token]; isSaved {
					matchedNodes = append(matchedNodes, tunnelClientHandler{
						Id:         savedClient.Id,
						Name:       savedClient.Name,
						Token:      savedClient.Token,
						Remark:     savedClient.Remark,
						IsOnline:   true,
						IsSaved:    true,
						RemoteAddr: activeItem.RemoteAddr,
						CreatedAt:  savedClient.CreatedAt,
						UpdatedAt:  savedClient.UpdatedAt,
					})
				} else {
					unsavedCount++
					matchedNodes = append(matchedNodes, tunnelClientHandler{
						Id:         0,
						Name:       "",
						Token:      activeItem.Token,
						Remark:     "",
						IsOnline:   true,
						IsSaved:    false,
						RemoteAddr: activeItem.RemoteAddr,
					})
				}
			}
		}

		if matchedNodes == nil {
			matchedNodes = make([]tunnelClientHandler, 0)
		}

		resList = append(resList, tunnelHandler{
			Id:           item.Id,
			Name:         item.Name,
			Type:         item.Type,
			Port:         item.Port,
			Certificate:  item.Certificate,
			Auth:         item.Auth,
			Remark:       item.Remark,
			ClientNodes:  matchedNodes,
			OnlineCount:  onlineCount,
			UnsavedCount: unsavedCount,
			TotalCount:   len(matchedNodes),
			CreatedAt:    item.CreatedAt,
			UpdatedAt:    item.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, resList)
}

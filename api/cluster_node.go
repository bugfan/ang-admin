package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bugfan/ang-admin/models"
	"github.com/bugfan/ang-admin/service"
	"github.com/bugfan/rest"
	"github.com/gin-gonic/gin"
	"github.com/go-xorm/xorm"
)

func init() {
	rest.Register(&models.ClusterNode{}, &clusterNodeHandler{}, rest.RouteTypeALL, nil, "cluster-node")
}

type clusterNodeHandler struct {
	Id        int64     `json:"id"`
	Name      string    `json:"name"`
	Addr      string    `json:"addr"`
	Secret    string    `json:"secret"`
	Status    int       `json:"status"`
	LastPing  time.Time `json:"last_ping"`
	Remark    string    `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (h *clusterNodeHandler) Before(g *gin.Context, x *xorm.Engine) bool {
	method := g.Request.Method
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
		if h.Name == "" {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":      1,
				"error_key": "nameRequired",
				"message":   "节点名称不能为空",
			})
			return false
		}
		if h.Addr == "" {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":      1,
				"error_key": "apiAddrRequired",
				"message":   "节点 API 地址不能为空",
			})
			return false
		}
		// Standardize address format
		if !strings.HasPrefix(h.Addr, "http://") && !strings.HasPrefix(h.Addr, "https://") {
			h.Addr = "http://" + h.Addr
		}
	}
	return true
}

func (h *clusterNodeHandler) List(c *gin.Context) {
	var list []models.ClusterNode
	session := models.GetEngine().NewSession()
	defer session.Close()

	if name := c.Query("name"); name != "" {
		session.Where("name LIKE ?", "%"+name+"%")
	}

	err := session.Desc("id").Find(&list)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": "查询集群节点列表失败: " + err.Error()})
		return
	}

	// Live ping check for each node
	resList := make([]clusterNodeHandler, 0, len(list))
	for _, item := range list {
		ok, msg := service.PingNode(&item)
		status := 0
		if ok {
			status = 1
			item.LastPing = time.Now()
			_, _ = models.GetEngine().ID(item.Id).Cols("status", "last_ping").Update(&item)
		} else {
			if msg == "auth_failed" {
				status = 2
			}
			item.Status = status
			_, _ = models.GetEngine().ID(item.Id).Cols("status").Update(&item)
		}

		resList = append(resList, clusterNodeHandler{
			Id:        item.Id,
			Name:      item.Name,
			Addr:      item.Addr,
			Secret:    item.Secret,
			Status:    status,
			LastPing:  item.LastPing,
			Remark:    item.Remark,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, resList)
}

// Handler: Manual Ping Cluster Node
func PingClusterNodeHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	var node models.ClusterNode
	has, err := models.GetEngine().ID(id).Get(&node)
	if err != nil || !has {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "message": "节点不存在"})
		return
	}

	ok, msg := service.PingNode(&node)
	if ok {
		node.Status = 1
		node.LastPing = time.Now()
		_, _ = models.GetEngine().ID(node.Id).Cols("status", "last_ping").Update(&node)
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": msg, "data": gin.H{"status": 1, "last_ping": node.LastPing}})
	} else {
		status := 0
		if msg == "auth_failed" {
			status = 2
		}
		node.Status = status
		_, _ = models.GetEngine().ID(node.Id).Cols("status").Update(&node)
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": msg, "data": gin.H{"status": status}})
	}
}

// Handler: Sync Config to a specific node
func SyncClusterNodeHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	var node models.ClusterNode
	has, err := models.GetEngine().ID(id).Get(&node)
	if err != nil || !has {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "message": "节点不存在"})
		return
	}

	serverCfg := service.BuildFullServerConfig()
	tunnelCfg := service.BuildFullTunnelConfig()
	certCfg := service.BuildFullCertificateConfig()
	payload := map[string]interface{}{
		"server_config":      serverCfg,
		"tunnel_config":      tunnelCfg,
		"certificate_config": certCfg,
	}
	data, _ := json.Marshal(payload)

	addr := strings.TrimRight(node.Addr, "/")
	syncURL := addr + "/api/config/sync"
	client := &http.Client{Timeout: 5 * time.Second}

	req, err := http.NewRequest("POST", syncURL, bytes.NewBuffer(data))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": "创建请求失败: " + err.Error()})
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if node.Secret != "" {
		req.Header.Set("X-Ang-Secret", node.Secret)
	}

	resp, err := client.Do(req)
	if err != nil {
		node.Status = 0
		_, _ = models.GetEngine().ID(node.Id).Cols("status").Update(&node)
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": "下发配置失败: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == 200 {
		node.Status = 1
		node.LastPing = time.Now()
		_, _ = models.GetEngine().ID(node.Id).Cols("status", "last_ping").Update(&node)
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": fmt.Sprintf("成功下发配置至节点 %s 兼已触发热重载", node.Name), "data": string(body)})
	} else {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": fmt.Sprintf("节点返回错误 (HTTP %d): %s", resp.StatusCode, string(body))})
	}
}

// Handler: Sync Config to All Online Nodes
func SyncAllClusterNodesHandler(c *gin.Context) {
	service.SyncAllToCluster()
	log.Printf("[cluster_sync] Manual trigger sync config to all nodes executed")
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "全集群配置下发任务已触发"})
}

// Handler: Query Active Tunnel Connections from node
func GetClusterNodeTunnelHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	var node models.ClusterNode
	has, err := models.GetEngine().ID(id).Get(&node)
	if err != nil || !has {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "message": "节点不存在"})
		return
	}

	addr := strings.TrimRight(node.Addr, "/")
	tunnelURL := addr + "/tunnel"
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(tunnelURL)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": "查询节点隧道失败: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": "解析节点隧道响应失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": result})
}

// Handler: Verify Node connection (Address + Secret)
func VerifyClusterNodeHandler(c *gin.Context) {
	var req struct {
		Addr   string `json:"addr"`
		Secret string `json:"secret"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": "Invalid parameters"})
		return
	}

	ok, msg := service.VerifyNode(req.Addr, req.Secret)
	if ok {
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": msg})
	} else {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": msg})
	}
}

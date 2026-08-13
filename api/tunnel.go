package api

import (
	"net/http"
	"time"

	"github.com/bugfan/ang-admin/models"
	"github.com/bugfan/rest"
	"github.com/gin-gonic/gin"
	"github.com/go-xorm/xorm"
)

func init() {
	rest.Register(&models.Tunnel{}, &tunnelHandler{}, rest.RouteTypeALL, nil, "tunnel")
}

type tunnelHandler struct {
	Id          int64     `json:"id"`
	Type        string    `json:"type"`
	Port        string    `json:"port"`
	SNI         string    `json:"sni"`
	Certificate string    `json:"certificate"`
	Remark      string    `json:"remark"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (t *tunnelHandler) Before(g *gin.Context, x *xorm.Engine) bool {
	return true
}

func (t *tunnelHandler) List(c *gin.Context) {
	var tunnels []models.Tunnel
	session := models.GetEngine().NewSession()
	defer session.Close()

	if tunnelType := c.Query("type"); tunnelType != "" {
		session.Where("type = ?", tunnelType)
	}
	if sni := c.Query("sni"); sni != "" {
		session.Where("sni LIKE ?", "%"+sni+"%")
	}
	if port := c.Query("port"); port != "" {
		session.Where("port LIKE ?", "%"+port+"%")
	}

	err := session.Find(&tunnels)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": "查询Tunnel列表失败: " + err.Error()})
		return
	}

	resList := make([]tunnelHandler, 0, len(tunnels))
	for _, item := range tunnels {
		resList = append(resList, tunnelHandler{
			Id:          item.Id,
			Type:        item.Type,
			Port:        item.Port,
			SNI:         item.SNI,
			Certificate: item.Certificate,
			Remark:      item.Remark,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, resList)
}

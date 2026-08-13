package api

import (
	"net/http"
	"time"

	"github.com/bugfan/ang-admin/models"
	"github.com/bugfan/ang-admin/service"
	"github.com/bugfan/rest"
	"github.com/gin-gonic/gin"
	"github.com/go-xorm/xorm"
)

func init() {
	rest.Register(&models.Certificate{}, &certHandler{}, rest.RouteTypeALL, nil, "certificate")
}

type certHandler struct {
	Id           int64     `json:"id"`
	CertId       string    `json:"cert_id"`
	Type         string    `json:"type"`
	KeyContent   string    `json:"key_content"`
	CertContent  string    `json:"cert_content"`
	Remark       string    `json:"remark"`
	SubjectCN    string    `json:"subject_cn"`
	SANs         string    `json:"sans"`
	NotBefore    time.Time `json:"not_before"`
	NotAfter     time.Time `json:"not_after"`
	Issuer       string    `json:"issuer"`
	SerialNumber string    `json:"serial_number"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (c *certHandler) Before(g *gin.Context, x *xorm.Engine) bool {
	// Automatically parse certificate metadata before saving
	if g.Request.Method == http.MethodPost || g.Request.Method == http.MethodPut || g.Request.Method == http.MethodPatch {
		mCert := &models.Certificate{
			CertContent: c.CertContent,
		}
		mCert.ParseCertInfo()
		c.SubjectCN = mCert.SubjectCN
		c.SANs = mCert.SANs
		c.NotBefore = mCert.NotBefore
		c.NotAfter = mCert.NotAfter
		c.Issuer = mCert.Issuer
		c.SerialNumber = mCert.SerialNumber
	}
	return true
}

func (c *certHandler) List(ctx *gin.Context) {
	var certs []models.Certificate
	session := models.GetEngine().NewSession()
	defer session.Close()

	if certId := ctx.Query("cert_id"); certId != "" {
		session.Where("cert_id LIKE ?", "%"+certId+"%")
	}
	if certType := ctx.Query("type"); certType != "" {
		session.Where("type = ?", certType)
	}
	if cn := ctx.Query("subject_cn"); cn != "" {
		session.Where("subject_cn LIKE ?", "%"+cn+"%")
	}
	if sans := ctx.Query("sans"); sans != "" {
		session.Where("sans LIKE ?", "%"+sans+"%")
	}

	err := session.Find(&certs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": "查询证书列表失败: " + err.Error()})
		return
	}

	resList := make([]certHandler, 0, len(certs))
	for _, item := range certs {
		resList = append(resList, certHandler{
			Id:           item.Id,
			CertId:       item.CertId,
			Type:         item.Type,
			KeyContent:   item.KeyContent,
			CertContent:  item.CertContent,
			Remark:       item.Remark,
			SubjectCN:    item.SubjectCN,
			SANs:         item.SANs,
			NotBefore:    item.NotBefore,
			NotAfter:     item.NotAfter,
			Issuer:       item.Issuer,
			SerialNumber: item.SerialNumber,
			CreatedAt:    item.CreatedAt,
			UpdatedAt:    item.UpdatedAt,
		})
	}

	ctx.JSON(http.StatusOK, resList)
}

type GenerateCertRequest struct {
	CommonName string   `json:"common_name"`
	DNSNames   []string `json:"dns_names"`
	ValidDays  int      `json:"valid_days"`
}

func GenerateCertHandler(c *gin.Context) {
	var req GenerateCertRequest
	_ = c.ShouldBindJSON(&req)

	if req.CommonName == "" {
		req.CommonName = "local.i443.cn"
	}
	if req.ValidDays <= 0 {
		req.ValidDays = 365
	}

	keyPEM, certPEM, err := service.GenerateSelfSignedCert(req.CommonName, req.DNSNames, req.ValidDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": "生成证书失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"key_content":  keyPEM,
			"cert_content": certPEM,
		},
	})
}

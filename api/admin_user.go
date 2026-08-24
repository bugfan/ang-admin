package api

import (
	"errors"
	"image/png"
	"net/http"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/bugfan/ang-admin/models"
	"github.com/bugfan/ang-admin/service"
	"github.com/bugfan/rest"
	"github.com/disintegration/letteravatar"
	"github.com/gin-gonic/gin"
	"github.com/go-xorm/xorm"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	rest.Register(&models.AdminUser{}, &adminUser{}, rest.RouteTypeALL, []string{"Password"}, "admin")
}

type adminUser struct {
	Id           int64
	Username     string
	Password     string
	Description  string
	IsSuperAdmin bool `json:"is_super_admin"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (u *adminUser) Before(g *gin.Context, x *xorm.Engine) bool {
	// 1. 获取当前登录的管理员用户
	currentUsername := ""
	if uname, ok := g.Get("username"); ok {
		if s, ok := uname.(string); ok {
			currentUsername = s
		}
	}

	var currAdminUser models.AdminUser
	if currentUsername != "" {
		_, _ = x.Where("username = ?", currentUsername).Get(&currAdminUser)
	}

	// 2. 如果登录的用户不是超管，进行权限拦截
	if !currAdminUser.IsSuperAdmin {
		switch g.Request.Method {
		case http.MethodGet:
			// 如果是单条获取，如果不是自己的账号，拒绝访问
			if g.Param("id") != "" {
				paramID := g.Param("id")
				var targetAdminUser models.AdminUser
				has, _ := x.Where("id = ?", paramID).Get(&targetAdminUser)
				if has && targetAdminUser.Username != currAdminUser.Username {
					g.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 403, "message": "非超级管理员只能查看自己的账号"})
					return false
				}
			}
		case http.MethodPost:
			// 非超管无法新增管理员
			g.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 403, "message": "非超级管理员无法新增管理员"})
			return false
		case http.MethodDelete:
			// 非超管无法删除管理员
			g.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 403, "message": "非超级管理员无法删除管理员"})
			return false
		case http.MethodPut, http.MethodPatch:
			// 非超管只能修改自己，且无法修改超管属性
			paramID := g.Param("id")
			var targetAdminUser models.AdminUser
			has, _ := x.Where("id = ?", paramID).Get(&targetAdminUser)
			if has && targetAdminUser.Username != currAdminUser.Username {
				g.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 403, "message": "非超级管理员无法修改其他管理员账号"})
				return false
			}
			// 保持非超管用户的超管属性不变 (防止提权)
			u.IsSuperAdmin = currAdminUser.IsSuperAdmin
		}
	}

	// 3. 处理密码 Hash
	if g.Request.Method == http.MethodPost || g.Request.Method == http.MethodPut || g.Request.Method == http.MethodPatch {
		if len(u.Password) > 0 {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
			if err != nil {
				g.AbortWithError(http.StatusInternalServerError, errors.New("failed to hash password"))
				return false
			}
			u.Password = string(hashedPassword)
		} else {
			// Require password for new and edit
			g.AbortWithError(http.StatusBadRequest, errors.New("password is required"))
			return false
		}
	}
	return true
}

func (u *adminUser) List(c *gin.Context) {
	currentUsername := ""
	if uname, ok := c.Get("username"); ok {
		if s, ok := uname.(string); ok {
			currentUsername = s
		}
	}

	var currAdminUser models.AdminUser
	has, err := models.GetEngine().Where("username = ?", currentUsername).Get(&currAdminUser)
	if err != nil || !has {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "管理员不存在"})
		return
	}

	var adminUsers []models.AdminUser
	session := models.GetEngine().NewSession()
	defer session.Close()

	if !currAdminUser.IsSuperAdmin {
		// 非超级管理员，绝对只能查到自己的账号
		session.Where("username = ?", currAdminUser.Username)
	} else {
		// 超级管理员，如果有搜索条件，按搜索条件模糊匹配
		searchUsername := c.Query("username")
		if searchUsername != "" {
			session.Where("username LIKE ?", "%"+searchUsername+"%")
		}
	}

	err = session.Find(&adminUsers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": "查询列表失败: " + err.Error()})
		return
	}

	resList := make([]adminUser, 0, len(adminUsers))
	for _, item := range adminUsers {
		resList = append(resList, adminUser{
			Id:           item.Id,
			Username:     item.Username,
			Description:  item.Description,
			IsSuperAdmin: item.IsSuperAdmin,
			CreatedAt:    item.CreatedAt,
			UpdatedAt:    item.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, resList)
}

type LoginRequest struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	VerifyCode string `json:"verifyCode"`
	CaptchaId  string `json:"captchaId"`
}

func LoginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": "invalid request format"})
		return
	}

	if !service.VerifyCaptcha(req.CaptchaId, req.VerifyCode) {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": "用户名、密码或验证码错误"})
		return
	}

	data, err := service.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    data,
	})
}

func MineHandler(c *gin.Context) {
	usernameStr := ""
	if uname, ok := c.Get("username"); ok {
		if s, ok := uname.(string); ok {
			usernameStr = s
		}
	}

	data, err := service.GetAdminUserInfo(usernameStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    data,
	})
}

func RefreshTokenHandler(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}
	_ = c.ShouldBindJSON(&req)
	data, err := service.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    data,
	})
}

func CaptchaHandler(c *gin.Context) {
	id, b64s, err := service.GenerateCaptcha()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": "failed to generate captcha"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": map[string]string{
			"captchaId": id,
			"b64s":      b64s,
		},
	})
}

func AvatarHandler(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		username = "Admin"
	}

	firstLetter, _ := utf8.DecodeRuneInString(username)
	firstLetter = unicode.ToUpper(firstLetter)

	img, err := letteravatar.Draw(75, firstLetter, &letteravatar.Options{
		PaletteKey: username,
	})
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.Header("Content-Type", "image/png")
	err = png.Encode(c.Writer, img)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
	}
}

func AsyncRoutesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": []gin.H{
			{
				"path": "/app",
				"meta": gin.H{
					"icon":  "ri:apps-line",
					"title": "menus.pureApp",
					"rank":  5,
				},
				"children": []gin.H{
					{
						"path":      "/app/cert",
						"name":      "AppCert",
						"component": "app/cert/index",
						"meta": gin.H{
							"icon":  "ri:lock-line",
							"title": "menus.pureCert",
							"roles": []string{"admin", "common"},
						},
					},
						{
							"path": "/app/tunnel-parent",
							"name": "AppTunnelParent",
							"meta": gin.H{
								"icon":  "ri:route-line",
								"title": "menus.pureTunnel",
								"roles": []string{"admin", "common"},
							},
							"children": []gin.H{
								{
									"path":      "/app/tunnel",
									"name":      "AppTunnel",
									"component": "app/tunnel/index",
									"meta": gin.H{
										"title": "menus.pureTunnelServer",
										"roles": []string{"admin", "common"},
									},
								},
								{
									"path":      "/app/tunnel-client",
									"name":      "AppTunnelClient",
									"component": "app/tunnel-client/index",
									"meta": gin.H{
										"title": "menus.pureTunnelClient",
										"roles": []string{"admin", "common"},
									},
								},
							},
						},
					{
						"path":      "/app/rule",
						"name":      "AppRule",
						"component": "app/rule/index",
						"meta": gin.H{
							"icon":  "ri:filter-3-line",
							"title": "menus.pureRule",
							"roles": []string{"admin", "common"},
						},
					},
					{
						"path":      "/app/dns",
						"name":      "AppDns",
						"component": "app/dns/index",
						"meta": gin.H{
							"icon":  "ri:global-line",
							"title": "menus.pureDns",
							"roles": []string{"admin", "common"},
						},
					},
					{
						"path":      "/app/http",
						"name":      "AppHttpProxy",
						"component": "app/http/index",
						"meta": gin.H{
							"icon":  "ri:links-line",
							"title": "menus.pureHttpProxy",
							"roles": []string{"admin", "common"},
						},
					},
					{
						"path":      "/app/cluster",
						"name":      "AppCluster",
						"component": "app/cluster/index",
						"meta": gin.H{
							"icon":  "ri:server-line",
							"title": "menus.pureCluster",
							"roles": []string{"admin", "common"},
						},
					},
				},
			},


			{
				"path": "/admin",
				"meta": gin.H{
					"icon":  "ri:admin-line",
					"title": "menus.pureAdminManagement",
					"rank":  10,
				},
				"children": []gin.H{
					{
						"path":      "/admin/index",
						"name":      "SystemAdmin",
						"component": "system/admin/index",
						"meta": gin.H{
							"icon":  "ri:admin-line",
							"title": "menus.pureAdminManagement",
							"roles": []string{"admin"},
						},
					},
				},
			},
		},
	})
}

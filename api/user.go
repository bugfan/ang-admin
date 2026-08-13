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
	rest.Register(&models.User{}, &adminUser{}, rest.RouteTypeALL, []string{"Password"}, "admin")
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
	// 1. 获取当前登录的用户
	currentUsername := ""
	if uname, ok := g.Get("username"); ok {
		if s, ok := uname.(string); ok {
			currentUsername = s
		}
	}

	var currentUser models.User
	if currentUsername != "" {
		_, _ = x.Where("username = ?", currentUsername).Get(&currentUser)
	}

	// 2. 如果登录的用户不是超管，进行权限判断
	if !currentUser.IsSuperAdmin {
		switch g.Request.Method {
		case http.MethodGet:
			// 如果是列表查询 (如 /api/admin)，强制约束只查询自己的账号
			if g.Param("id") == "" {
				q := g.Request.URL.Query()
				q.Set("username", currentUser.Username)
				g.Request.URL.RawQuery = q.Encode()
			} else {
				// 单条获取，如果不是自己的账号，拒绝访问
				paramID := g.Param("id")
				var targetUser models.User
				has, _ := x.Where("id = ?", paramID).Get(&targetUser)
				if has && targetUser.Username != currentUser.Username {
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
			var targetUser models.User
			has, _ := x.Where("id = ?", paramID).Get(&targetUser)
			if has && targetUser.Username != currentUser.Username {
				g.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 403, "message": "非超级管理员无法修改其他管理员账号"})
				return false
			}
			// 保持非超管用户的超管属性不变 (防止提权)
			u.IsSuperAdmin = currentUser.IsSuperAdmin
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



func (u *adminUser) List(c *gin.Context) {
	currentUsername := ""
	if uname, ok := c.Get("username"); ok {
		if s, ok := uname.(string); ok {
			currentUsername = s
		}
	}

	var currentUser models.User
	has, err := models.GetEngine().Where("username = ?", currentUsername).Get(&currentUser)
	if err != nil || !has {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "用户不存在"})
		return
	}

	var users []models.User
	session := models.GetEngine().NewSession()
	defer session.Close()

	if !currentUser.IsSuperAdmin {
		// 非超级管理员，绝对只能查到自己的账号
		session.Where("username = ?", currentUser.Username)
	} else {
		// 超级管理员，如果有搜索条件，按搜索条件模糊匹配
		searchUsername := c.Query("username")
		if searchUsername != "" {
			session.Where("username LIKE ?", "%"+searchUsername+"%")
		}
	}

	err = session.Find(&users)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": "查询列表失败: " + err.Error()})
		return
	}

	resList := make([]adminUser, 0, len(users))
	for _, user := range users {
		resList = append(resList, adminUser{
			Id:           user.Id,
			Username:     user.Username,
			Description:  user.Description,
			IsSuperAdmin: user.IsSuperAdmin,
			CreatedAt:    user.CreatedAt,
			UpdatedAt:    user.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, resList)
}

func MineHandler(c *gin.Context) {
	usernameStr := ""
	if uname, ok := c.Get("username"); ok {
		if s, ok := uname.(string); ok {
			usernameStr = s
		}
	}

	data, err := service.GetUserInfo(usernameStr)
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

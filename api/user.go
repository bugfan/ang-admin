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
)

func init() {
	rest.Register(&models.User{}, &adminUser{}, rest.RouteTypeALL, []string{"Password"}, "admin")
}

type adminUser struct {
	Id          int64
	Username    string
	Password    string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (u *adminUser) Before(g *gin.Context, x *xorm.Engine) bool {
	if g.Request.Method == http.MethodPost {
		if len(u.Password) > 0 && len(u.Password) < 3 {
			g.AbortWithError(http.StatusBadRequest, errors.New("密码不少于3位"))
			return false
		}
	}
	return true
}

type LoginRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	VerifyCode string `json:"verifyCode"`
	CaptchaId string `json:"captchaId"`
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

type ListUsersRequest struct {
	Username    string `json:"username"`
	CurrentPage int    `json:"currentPage"`
	PageSize    int    `json:"pageSize"`
}

func ListUsersHandler(c *gin.Context) {
	var req ListUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": "invalid request format"})
		return
	}
	if req.CurrentPage <= 0 {
		req.CurrentPage = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	users, total, err := service.ListUsers(req.Username, req.CurrentPage, req.PageSize)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": map[string]interface{}{
			"list":        users,
			"total":       total,
			"pageSize":    req.PageSize,
			"currentPage": req.CurrentPage,
		},
	})
}

type RegisterRequest struct {
	Username       string `json:"username"`
	Password       string `json:"password"`
	RepeatPassword string `json:"repeatPassword"`
	Description    string `json:"description"`
}

func RegisterHandler(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": "invalid request format"})
		return
	}

	if req.Password != req.RepeatPassword {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": "passwords do not match"})
		return
	}

	user := &models.User{
		Username:    req.Username,
		Password:    req.Password,
		Description: req.Description,
	}

	err := service.Register(user)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "register success",
	})
}

func UpdateUserHandler(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": "invalid request format"})
		return
	}

	if req.Password != "" && req.Password != req.RepeatPassword {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": "passwords do not match"})
		return
	}

	user := &models.User{
		Username:    req.Username,
		Password:    req.Password,
		Description: req.Description,
	}

	err := service.UpdateUser(user)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "update success",
	})
}

type DeleteUserRequest struct {
	Id  int64   `json:"id"`
	Ids []int64 `json:"ids"`
}

func DeleteUserHandler(c *gin.Context) {
	var req DeleteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": "invalid request format"})
		return
	}

	engine := models.GetEngine()
	if len(req.Ids) > 0 {
		_, err := engine.In("id", req.Ids).Delete(&models.User{})
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 1, "message": err.Error()})
			return
		}
	} else if req.Id > 0 {
		_, err := engine.ID(req.Id).Delete(&models.User{})
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 1, "message": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "delete success",
	})
}

func MineHandler(c *gin.Context) {
	data, err := service.GetUserInfo()
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
	data, err := service.RefreshToken("")
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

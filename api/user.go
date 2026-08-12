package api

import (
	"net/http"

	"github.com/bugfan/ang-admin/service"
	"github.com/gin-gonic/gin"
)

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

	err := service.Register(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "register success",
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

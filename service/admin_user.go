package service

import (
	"errors"
	"time"

	"github.com/bugfan/ang-admin/models"
	"golang.org/x/crypto/bcrypt"
)

func ListAdminUsers(username string, page, pageSize int) ([]models.AdminUser, int64, error) {
	var adminUsers []models.AdminUser
	session := models.GetEngine().NewSession()
	defer session.Close()

	if username != "" {
		session.Where("username LIKE ?", "%"+username+"%")
	}
	total, err := session.Clone().Count(new(models.AdminUser))
	if err != nil {
		return nil, 0, err
	}

	err = session.Limit(pageSize, (page-1)*pageSize).Find(&adminUsers)
	return adminUsers, total, err
}

func RegisterAdminUser(adminUser *models.AdminUser) error {
	engine := models.GetEngine()

	has, err := engine.Where("username = ?", adminUser.Username).Exist(&models.AdminUser{})
	if err != nil {
		return err
	}
	if has {
		return errors.New("admin user already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	adminUser.Password = string(hashedPassword)
	adminUser.Roles = []string{"common"}

	_, err = engine.Insert(adminUser)
	return err
}

func Register(adminUser *models.AdminUser) error {
	return RegisterAdminUser(adminUser)
}

func UpdateAdminUser(adminUser *models.AdminUser) error {
	engine := models.GetEngine()
	existing := new(models.AdminUser)
	has, err := engine.Where("username = ?", adminUser.Username).Get(existing)
	if err != nil {
		return err
	}
	if !has {
		return errors.New("admin user not found")
	}

	existing.Description = adminUser.Description

	if adminUser.Password != "" {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(adminUser.Password), bcrypt.DefaultCost)
		existing.Password = string(hashedPassword)
	}

	_, err = engine.ID(existing.Id).Cols("description", "password").Update(existing)
	return err
}

func Login(username, password string) (map[string]interface{}, error) {
	adminUser := new(models.AdminUser)
	has, err := models.GetEngine().Where("username = ?", username).Get(adminUser)
	if err != nil {
		return nil, errors.New("用户名、密码或验证码错误")
	}
	if !has {
		return nil, errors.New("用户名、密码或验证码错误")
	}

	err = bcrypt.CompareHashAndPassword([]byte(adminUser.Password), []byte(password))
	if err != nil {
		return nil, errors.New("用户名、密码或验证码错误")
	}

	accessToken, err := GenerateToken(adminUser.Username, 7*24*time.Hour)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	refreshToken, err := GenerateToken(adminUser.Username, 30*24*time.Hour)
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	return map[string]interface{}{
		"id":             adminUser.Id,
		"avatar":         "/avatar.png?username=" + adminUser.Username,
		"username":       adminUser.Username,
		"is_super_admin": adminUser.IsSuperAdmin,
		"roles":          []string{"admin"}, // For now, hardcode admin role
		"permissions":    []string{"*:*:*"},
		"accessToken":    accessToken,
		"refreshToken":   refreshToken,
		"expires":        time.Now().Add(7 * 24 * time.Hour).Format("2006/01/02 15:04:05"),
	}, nil
}

func GetAdminUserInfo(username string) (map[string]interface{}, error) {
	if username == "" {
		username = "admin"
	}
	adminUserObj := new(models.AdminUser)
	has, err := models.GetEngine().Where("username = ?", username).Get(adminUserObj)
	if err != nil || !has {
		return map[string]interface{}{
			"avatar":         "/avatar.png?username=" + username,
			"username":       username,
			"is_super_admin": false,
			"email":          "",
			"phone":          "",
			"description":    "",
		}, nil
	}
	return map[string]interface{}{
		"id":             adminUserObj.Id,
		"avatar":         "/avatar.png?username=" + adminUserObj.Username,
		"username":       adminUserObj.Username,
		"is_super_admin": adminUserObj.IsSuperAdmin,
		"email":          "",
		"phone":          "",
		"description":    adminUserObj.Description,
	}, nil
}

func GetUserInfo(username string) (map[string]interface{}, error) {
	return GetAdminUserInfo(username)
}

func RefreshToken(oldToken string) (map[string]interface{}, error) {
	username := "admin"
	if oldToken != "" {
		claims, err := ParseToken(oldToken)
		if err == nil && claims.Username != "" {
			username = claims.Username
		}
	}

	accessToken, err := GenerateToken(username, 7*24*time.Hour)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}
	refreshToken, err := GenerateToken(username, 30*24*time.Hour)
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	return map[string]interface{}{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
		"expires":      time.Now().Add(7 * 24 * time.Hour).Format("2006/01/02 15:04:05"),
	}, nil
}

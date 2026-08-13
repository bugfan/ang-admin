package service

import (
	"errors"
	"time"

	"github.com/bugfan/ang-admin/models"
	"golang.org/x/crypto/bcrypt"
)

func ListUsers(username string, page, pageSize int) ([]models.User, int64, error) {
	var users []models.User
	session := models.GetEngine().NewSession()
	defer session.Close()

	if username != "" {
		session.Where("username LIKE ?", "%"+username+"%")
	}
	total, err := session.Clone().Count(new(models.User))
	if err != nil {
		return nil, 0, err
	}

	err = session.Limit(pageSize, (page-1)*pageSize).Find(&users)
	return users, total, err
}

func Register(user *models.User) error {
	engine := models.GetEngine()

	has, err := engine.Where("username = ?", user.Username).Exist(&models.User{})
	if err != nil {
		return err
	}
	if has {
		return errors.New("user already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)
	user.Roles = []string{"common"}

	_, err = engine.Insert(user)
	return err
}

func UpdateUser(user *models.User) error {
	engine := models.GetEngine()
	existing := new(models.User)
	has, err := engine.Where("username = ?", user.Username).Get(existing)
	if err != nil {
		return err
	}
	if !has {
		return errors.New("user not found")
	}

	existing.Description = user.Description

	if user.Password != "" {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		existing.Password = string(hashedPassword)
	}

	_, err = engine.ID(existing.Id).Cols("description", "password").Update(existing)
	return err
}

func Login(username, password string) (map[string]interface{}, error) {
	user := new(models.User)
	has, err := models.GetEngine().Where("username = ?", username).Get(user)
	if err != nil {
		return nil, errors.New("用户名、密码或验证码错误")
	}
	if !has {
		return nil, errors.New("用户名、密码或验证码错误")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("用户名、密码或验证码错误")
	}

	accessToken, err := GenerateToken(user.Username, 7*24*time.Hour)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	refreshToken, err := GenerateToken(user.Username, 30*24*time.Hour)
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	return map[string]interface{}{
		"id":             user.Id,
		"avatar":         "/avatar.png?username=" + user.Username,
		"username":       user.Username,
		"is_super_admin": user.IsSuperAdmin,
		"roles":          []string{"admin"}, // For now, hardcode admin role
		"permissions":    []string{"*:*:*"},
		"accessToken":    accessToken,
		"refreshToken":   refreshToken,
		"expires":        time.Now().Add(7 * 24 * time.Hour).Format("2006/01/02 15:04:05"),
	}, nil
}

func GetUserInfo(username string) (map[string]interface{}, error) {
	if username == "" {
		username = "admin"
	}
	user := new(models.User)
	has, err := models.GetEngine().Where("username = ?", username).Get(user)
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
		"id":             user.Id,
		"avatar":         "/avatar.png?username=" + user.Username,
		"username":       user.Username,
		"is_super_admin": user.IsSuperAdmin,
		"email":          "",
		"phone":          "",
		"description":    user.Description,
	}, nil
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

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

func Register(username, password string) error {
	engine := models.GetEngine()

	has, err := engine.Where("username = ?", username).Exist(&models.User{})
	if err != nil {
		return err
	}
	if has {
		return errors.New("user already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &models.User{
		Username: username,
		Password: string(hashedPassword),
		Nickname: username,
		Roles:    []string{"common"},
	}

	_, err = engine.Insert(user)
	return err
}

func Login(username, password string) (map[string]interface{}, error) {
	user := new(models.User)
	has, err := models.GetEngine().Where("username = ?", username).Get(user)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("user not found")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("invalid password")
	}

	return map[string]interface{}{
		"id":           user.Id,
		"avatar":       user.Avatar,
		"username":     user.Username,
		"nickname":     user.Nickname,
		"roles":        []string{"admin"}, // For now, hardcode admin role
		"permissions":  []string{"*:*:*"},
		"accessToken":  "eyJhbGciOiJIUzUxMiJ9." + user.Username,
		"refreshToken": "eyJhbGciOiJIUzUxMiJ9." + user.Username + "Refresh",
		"expires":      time.Now().Add(24 * 7 * time.Hour).Format("2006/01/02 15:04:05"),
	}, nil
}

func GetUserInfo() (map[string]interface{}, error) {
	return map[string]interface{}{
		"avatar":      "",
		"username":    "admin",
		"nickname":    "admin",
		"email":       "",
		"phone":       "",
		"description": "",
	}, nil
}

func RefreshToken(oldToken string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"accessToken":  "eyJhbGciOiJIUzUxMiJ9.admin.new",
		"refreshToken": "eyJhbGciOiJIUzUxMiJ9.adminRefresh.new",
		"expires":      time.Now().Add(24 * 7 * time.Hour).Format("2006/01/02 15:04:05"),
	}, nil
}

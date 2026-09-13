package services

import (
	"cold-chain-trace/backend/database"
	"cold-chain-trace/backend/models"
	"errors"
	"regexp"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
)

// 登录与用户查询相关错误（供 controllers 用 errors.Is 区分提示）
var (
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidPassword = errors.New("invalid password")
	ErrUserInactive    = errors.New("user is inactive")
	// ErrUsernameExists 用户名唯一约束冲突（注册等场景）
	ErrUsernameExists = errors.New("username already exists")
	// ErrEmailExists 邮箱唯一约束冲突（注册等场景）
	ErrEmailExists = errors.New("email already exists")
)

type UserService struct{}

var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// CreateUser 创建用户
func (s *UserService) CreateUser(user *models.User) error {
	// When binding JSON, omitted fields keep zero values. New accounts should be active by default.
	if !user.IsActive {
		user.IsActive = true
	}
	if err := user.HashPassword(user.Password); err != nil {
		return err
	}
	if err := database.DB.Create(user).Error; err != nil {
		return translateUserCreateError(err)
	}
	return nil
}

// translateUserCreateError 将数据库唯一约束等错误转为业务错误，避免向前端暴露 pq 原始信息。
func translateUserCreateError(err error) error {
	if err == nil {
		return nil
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		c := strings.ToLower(pqErr.Constraint)
		switch {
		case strings.Contains(c, "username"):
			return ErrUsernameExists
		case strings.Contains(c, "email"):
			return ErrEmailExists
		}
	}
	msg := err.Error()
	if strings.Contains(msg, "uix_users_username") {
		return ErrUsernameExists
	}
	if strings.Contains(msg, "uix_users_email") {
		return ErrEmailExists
	}
	return err
}

// GetUserByUsername 根据用户名获取用户
func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail 根据邮箱获取用户
func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// ValidateUser 验证用户登录
func (s *UserService) ValidateUser(username, password string) (*models.User, error) {
	user, err := s.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	if !user.CheckPassword(password) {
		return nil, ErrInvalidPassword
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	return user, nil
}

// GetUserByID 根据 ID 获取用户
func (s *UserService) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUsername 修改当前用户用户名（需保证新用户名未被占用）
func (s *UserService) UpdateUsername(userID uint, newUsername string) error {
	newUsername = strings.TrimSpace(newUsername)
	if newUsername == "" {
		return errors.New("用户名不能为空")
	}
	var count int64
	database.DB.Model(&models.User{}).Where("username = ? AND id != ?", newUsername, userID).Count(&count)
	if count > 0 {
		return errors.New("该用户名已被使用")
	}
	tx := database.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := tx.Model(&models.User{}).Where("id = ?", userID).Update("username", newUsername).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Model(&models.Product{}).
		Where("consumer_id = ? AND is_deleted = ?", userID, false).
		Update("consignee", newUsername).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// UpdateEmail 修改当前用户邮箱（需保证唯一）
func (s *UserService) UpdateEmail(userID uint, newEmail string) error {
	normalized, err := normalizeAndValidateEmail(newEmail)
	if err != nil {
		return err
	}
	var count int64
	database.DB.Model(&models.User{}).Where("email = ? AND id != ?", normalized, userID).Count(&count)
	if count > 0 {
		return errors.New("该邮箱已被使用")
	}
	return database.DB.Model(&models.User{}).Where("id = ?", userID).Update("email", normalized).Error
}

func normalizeAndValidateEmail(email string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" {
		return "", errors.New("邮箱不能为空")
	}
	if !emailPattern.MatchString(normalized) {
		return "", errors.New("邮箱格式不正确")
	}
	return normalized, nil
}

// UpdatePassword 修改当前用户密码（需验证旧密码）
func (s *UserService) UpdatePassword(userID uint, oldPassword, newPassword string) error {
	if newPassword == "" || len(newPassword) < 6 {
		return errors.New("新密码长度至少6位")
	}
	user, err := s.GetUserByID(userID)
	if err != nil {
		return err
	}
	if !user.CheckPassword(oldPassword) {
		return errors.New("原密码错误")
	}
	user.Password = ""
	if err := user.HashPassword(newPassword); err != nil {
		return err
	}
	return database.DB.Model(&models.User{}).Where("id = ?", userID).Update("password", user.Password).Error
}

// ResetPassword 直接重置用户密码（用于忘记密码流程）
func (s *UserService) ResetPassword(userID uint, newPassword string) error {
	if newPassword == "" || len(newPassword) < 6 {
		return errors.New("新密码长度至少6位")
	}
	user, err := s.GetUserByID(userID)
	if err != nil {
		return err
	}
	user.Password = ""
	if err := user.HashPassword(newPassword); err != nil {
		return err
	}
	return database.DB.Model(&models.User{}).Where("id = ?", userID).Update("password", user.Password).Error
}

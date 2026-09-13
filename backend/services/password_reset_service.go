package services

import (
	"cold-chain-trace/backend/database"
	"cold-chain-trace/backend/models"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

const passwordResetCodeTTL = 10 * time.Minute

var (
	ErrResetEmailRequired    = errors.New("请输入邮箱")
	ErrResetCodeRequired     = errors.New("请输入验证码")
	ErrResetCodeInvalid      = errors.New("验证码无效或已过期")
	ErrResetPasswordTooShort = errors.New("新密码长度至少6位")
	// ErrForgotUsernameNotFound 找回密码时登录用户名不存在
	ErrForgotUsernameNotFound = errors.New("forgot: username not found")
	// ErrForgotEmailMismatch 输入邮箱与账号注册邮箱不一致
	ErrForgotEmailMismatch = errors.New("forgot: email mismatch")
	// ErrForgotAccountNoEmail 账号未绑定邮箱，无法邮件找回
	ErrForgotAccountNoEmail = errors.New("forgot: account has no email")
	// ErrForgotUsernameRequired 未提供登录用户名
	ErrForgotUsernameRequired = errors.New("请输入登录用户名")
)

type PasswordResetService struct {
	userService *UserService
	mailer      PasswordResetMailer
}

func NewPasswordResetService(mailer PasswordResetMailer) *PasswordResetService {
	if mailer == nil {
		mailer = NewEmailService()
	}
	return &PasswordResetService{
		userService: &UserService{},
		mailer:      mailer,
	}
}

// SendResetCodeWithUsername 根据「待登录用户名 + 邮箱」发送验证码：邮箱须与该用户注册邮箱一致。
func (s *PasswordResetService) SendResetCodeWithUsername(username, email string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return ErrForgotUsernameRequired
	}
	normalizedEmail := normalizeEmail(email)
	if normalizedEmail == "" {
		return ErrResetEmailRequired
	}

	user, err := s.userService.GetUserByUsername(username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return ErrForgotUsernameNotFound
		}
		return err
	}

	reg := normalizeEmail(user.Email)
	if strings.TrimSpace(user.Email) == "" || reg == "" {
		return ErrForgotAccountNoEmail
	}
	if reg != normalizedEmail {
		return ErrForgotEmailMismatch
	}

	return s.deliverPasswordResetCode(user, normalizedEmail)
}

// deliverPasswordResetCode 写入验证码记录并发送邮件（调用方已校验用户与邮箱一致）。
func (s *PasswordResetService) deliverPasswordResetCode(user *models.User, normalizedEmail string) error {
	code, err := generateNumericCode(6)
	if err != nil {
		return err
	}
	codeHash, err := hashResetCode(code)
	if err != nil {
		return err
	}

	record := models.PasswordResetCode{
		UserID:    user.ID,
		Email:     normalizedEmail,
		CodeHash:  codeHash,
		ExpiresAt: time.Now().Add(passwordResetCodeTTL),
	}
	if err := database.DB.Create(&record).Error; err != nil {
		return err
	}

	// 若发送邮件失败，则删除刚写入的记录
	if err := s.mailer.SendPasswordResetCode(normalizedEmail, code); err != nil {
		_ = database.DB.Delete(&record).Error
		return err
	}

	now := time.Now()
	return database.DB.Model(&models.PasswordResetCode{}).
		Where("email = ? AND id <> ? AND used_at IS NULL", normalizedEmail, record.ID).
		Update("used_at", &now).Error
}

// ResetPasswordWithCode 使用邮箱验证码重置密码。
func (s *PasswordResetService) ResetPasswordWithCode(email, code, newPassword string) error {
	normalizedEmail := normalizeEmail(email)
	trimmedCode := strings.TrimSpace(code)
	if normalizedEmail == "" {
		return ErrResetEmailRequired
	}
	if trimmedCode == "" {
		return ErrResetCodeRequired
	}
	user, err := s.userService.GetUserByEmail(normalizedEmail)
	if err != nil {
		return ErrResetCodeInvalid
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	var rollbackErr error
	defer func() {
		if rollbackErr != nil {
			tx.Rollback()
		}
	}()

	var record models.PasswordResetCode
	now := time.Now()
	err = tx.Set("gorm:query_option", "FOR UPDATE").
		Where("email = ? AND used_at IS NULL AND expires_at > ?", normalizedEmail, now).
		Order("created_at DESC").
		First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			rollbackErr = ErrResetCodeInvalid
			return rollbackErr
		}
		rollbackErr = err
		return rollbackErr
	}

	if err := isResetCodeUsable(&record, trimmedCode, now); err != nil {
		rollbackErr = err
		return rollbackErr
	}

	if err := applyResetPassword(user, newPassword); err != nil {
		rollbackErr = err
		return rollbackErr
	}
	if err := tx.Model(&models.User{}).Where("id = ?", user.ID).Update("password", user.Password).Error; err != nil {
		rollbackErr = err
		return rollbackErr
	}
	if err := tx.Model(&models.PasswordResetCode{}).
		Where("email = ? AND used_at IS NULL", normalizedEmail).
		Update("used_at", &now).Error; err != nil {
		rollbackErr = err
		return rollbackErr
	}
	if err := tx.Commit().Error; err != nil {
		rollbackErr = err
		return rollbackErr
	}

	rollbackErr = nil
	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func generateNumericCode(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("invalid code length")
	}

	var builder strings.Builder
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		builder.WriteByte(byte('0' + n.Int64()))
	}
	return builder.String(), nil
}

func hashResetCode(code string) (string, error) {
	codeHash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(codeHash), nil
}

func isResetCodeUsable(record *models.PasswordResetCode, code string, now time.Time) error {
	if record == nil || record.UsedAt != nil || !record.ExpiresAt.After(now) {
		return ErrResetCodeInvalid
	}
	if bcrypt.CompareHashAndPassword([]byte(record.CodeHash), []byte(strings.TrimSpace(code))) != nil {
		return ErrResetCodeInvalid
	}
	return nil
}

func applyResetPassword(user *models.User, newPassword string) error {
	if len(newPassword) < 6 {
		return ErrResetPasswordTooShort
	}
	user.Password = ""
	return user.HashPassword(newPassword)
}

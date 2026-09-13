package models

import "time"

// PasswordResetCode 邮箱找回密码验证码记录
type PasswordResetCode struct {
	ID        uint       `gorm:"primary_key" json:"id"`
	UserID    uint       `gorm:"index;not null" json:"user_id"`
	Email     string     `gorm:"index;not null" json:"email"`
	CodeHash  string     `gorm:"not null" json:"-"`
	ExpiresAt time.Time  `gorm:"index;not null" json:"expires_at"`
	UsedAt    *time.Time `gorm:"index" json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// TableName 指定表名
func (PasswordResetCode) TableName() string {
	return "password_reset_codes"
}

package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// UserRole 用户角色枚举
type UserRole string

const (
	RoleProducer  UserRole = "producer"  // 生产商
	RoleWarehouse UserRole = "warehouse" // 仓储商
	RoleLogistics UserRole = "logistics" // 物流商
	RoleRegulator UserRole = "regulator" // 监管者
	RoleConsumer  UserRole = "consumer"  // 消费者
)

// User 用户模型
type User struct {
	ID          uint      `gorm:"primary_key" json:"id"`
	Username    string    `gorm:"unique_index;not null" json:"username"`
	Password    string    `gorm:"not null" json:"password,omitempty"`
	Email       string    `gorm:"unique_index" json:"email"`
	Role        UserRole  `gorm:"type:varchar(20);not null;index" json:"role"`
	CompanyName string    `json:"company_name"`
	CompanyID   string    `json:"company_id"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// HashPassword 加密密码
func (u *User) HashPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword 验证密码
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

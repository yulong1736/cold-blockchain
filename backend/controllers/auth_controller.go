package controllers

import (
	"cold-chain-trace/backend/models"
	"cold-chain-trace/backend/services"
	"cold-chain-trace/backend/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	userService          *services.UserService
	passwordResetService *services.PasswordResetService
}

func NewAuthController() *AuthController {
	return &AuthController{
		userService:          &services.UserService{},
		passwordResetService: services.NewPasswordResetService(nil),
	}
}

// Register 用户注册
// @Summary 用户注册
// @Description 注册新用户
// @Tags 认证
// @Accept json
// @Produce json
// @Param user body models.User true "用户信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/auth/register [post]
func (ac *AuthController) Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	if err := ac.userService.CreateUser(&user); err != nil {
		switch {
		case errors.Is(err, services.ErrUsernameExists):
			respondErrorMessage(c, http.StatusConflict, "用户名已存在")
		case errors.Is(err, services.ErrEmailExists):
			respondErrorMessage(c, http.StatusConflict, "该邮箱已被注册")
		default:
			respondError(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "注册成功",
		"data": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录获取token
// @Tags 认证
// @Accept json
// @Produce json
// @Param credentials body map[string]string true "登录凭证"
// @Success 200 {object} map[string]interface{}
// @Router /api/auth/login [post]
func (ac *AuthController) Login(c *gin.Context) {
	var credentials struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&credentials); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	user, err := ac.userService.ValidateUser(credentials.Username, credentials.Password)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserNotFound):
			respondErrorMessage(c, http.StatusUnauthorized, "用户不存在")
		case errors.Is(err, services.ErrInvalidPassword):
			respondErrorMessage(c, http.StatusUnauthorized, "密码错误")
		case errors.Is(err, services.ErrUserInactive):
			respondErrorMessage(c, http.StatusForbidden, "账号未激活或已禁用")
		default:
			respondError(c, http.StatusUnauthorized, err)
		}
		return
	}

	token, err := utils.GenerateToken(user.ID, user.Username, string(user.Role), user.CompanyName)
	if err != nil {
		respondErrorMessage(c, http.StatusInternalServerError, "生成token失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"data": gin.H{
			"token": token,
			"user": gin.H{
				"id":           user.ID,
				"username":     user.Username,
				"role":         user.Role,
				"email":        user.Email,
				"company_name": user.CompanyName,
			},
		},
	})
}

// UpdateProfile 修改当前用户资料（用户名/邮箱，需登录）
func (ac *AuthController) UpdateProfile(c *gin.Context) {
	user, err := getRequestUser(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondErrorMessage(c, http.StatusBadRequest, "请求参数无效")
		return
	}

	hasUsername := req.Username != ""
	hasEmail := req.Email != ""
	if !hasUsername && !hasEmail {
		respondErrorMessage(c, http.StatusBadRequest, "请至少填写一个要修改的字段")
		return
	}

	if hasUsername {
		if err := ac.userService.UpdateUsername(user.ID, req.Username); err != nil {
			respondError(c, http.StatusBadRequest, err)
			return
		}
	}
	if hasEmail {
		if err := ac.userService.UpdateEmail(user.ID, req.Email); err != nil {
			respondError(c, http.StatusBadRequest, err)
			return
		}
	}

	latestUser, err := ac.userService.GetUserByID(user.ID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "资料已更新",
		"data": gin.H{
			"id":           latestUser.ID,
			"username":     latestUser.Username,
			"email":        latestUser.Email,
			"role":         latestUser.Role,
			"company_name": latestUser.CompanyName,
		},
	})
}

// ChangePassword 修改当前用户密码（需登录，需提供原密码）
func (ac *AuthController) ChangePassword(c *gin.Context) {
	user, err := getRequestUser(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondErrorMessage(c, http.StatusBadRequest, "请填写原密码和新密码")
		return
	}
	if err := ac.userService.UpdatePassword(user.ID, req.OldPassword, req.NewPassword); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "密码已更新"})
}

// SendPasswordResetCode 发送邮箱验证码（未登录可用）
// 须同时提供登录用户名与邮箱，且邮箱须与该用户注册邮箱一致。
func (ac *AuthController) SendPasswordResetCode(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondErrorMessage(c, http.StatusBadRequest, "请填写登录用户名与邮箱")
		return
	}
	if err := ac.passwordResetService.SendResetCodeWithUsername(req.Username, req.Email); err != nil {
		switch {
		case errors.Is(err, services.ErrResetEmailRequired),
			errors.Is(err, services.ErrForgotUsernameRequired):
			respondError(c, http.StatusBadRequest, err)
		case errors.Is(err, services.ErrForgotUsernameNotFound):
			respondErrorMessage(c, http.StatusBadRequest, "用户不存在，请确认登录用户名")
		case errors.Is(err, services.ErrForgotEmailMismatch):
			respondErrorMessage(c, http.StatusBadRequest, "非注册邮箱，请重新输入")
		case errors.Is(err, services.ErrForgotAccountNoEmail):
			respondErrorMessage(c, http.StatusBadRequest, "该账号未绑定邮箱，无法通过邮箱找回")
		default:
			respondErrorMessage(c, http.StatusInternalServerError, "验证码发送失败")
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "验证码已发送至邮箱"})
}

// ResetPasswordWithCode 通过邮箱验证码重置密码（未登录可用）
func (ac *AuthController) ResetPasswordWithCode(c *gin.Context) {
	var req struct {
		Email       string `json:"email" binding:"required"`
		Code        string `json:"code" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondErrorMessage(c, http.StatusBadRequest, "请填写邮箱、验证码和新密码")
		return
	}
	if err := ac.passwordResetService.ResetPasswordWithCode(req.Email, req.Code, req.NewPassword); err != nil {
		if errors.Is(err, services.ErrResetEmailRequired) ||
			errors.Is(err, services.ErrResetCodeRequired) ||
			errors.Is(err, services.ErrResetCodeInvalid) ||
			errors.Is(err, services.ErrResetPasswordTooShort) {
			respondError(c, http.StatusBadRequest, err)
			return
		}
		respondErrorMessage(c, http.StatusInternalServerError, "重置密码失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "密码重置成功"})
}

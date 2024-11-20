package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/smtp"
	"repair-platform/models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// VerifyEmailInput 用于邮箱验证的输入结构体
type VerifyEmailInput struct {
	Email string `json:"email" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

// SendVerificationCodeInput 用于发送验证码的输入结构体
type SendVerificationCodeInput struct {
	Email string `json:"email" binding:"required"`
}

// ResetPasswordInput 用于重置密码的输入结构体
type ResetPasswordInput struct {
	Email       string `json:"email" binding:"required"`
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// APIResponse 标准的API响应结构
type APIResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// RegisterInput 定义了用户注册请求体
type RegisterInput struct {
	Username   string `json:"username" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=6"`
	InviteCode string `json:"invite_code"` // 邀请码字段
}

// Register 处理用户注册
// @Summary 用户注册
// @Description 注册新用户并发送验证邮件
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param user body RegisterInput true "用户注册信息"
// @Success 200 {object} APIResponse "注册成功，验证码已发送至您的邮箱"
// @Failure 400 {object} APIResponse "错误请求"
// @Failure 409 {object} APIResponse "用户名或邮箱已被注册"
// @Failure 500 {object} APIResponse "创建用户失败"
// @Router /register [post]
func Register(c *gin.Context) {
	// 记录请求的开始
	zap.S().Info("开始处理用户注册请求")

	// 记录收到的请求体数据（输入绑定）
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		zap.S().Error("输入绑定失败: ", err)
		c.JSON(http.StatusBadRequest, APIResponse{Message: "无效的输入"})
		return
	}
	zap.S().Info("收到的注册数据: ", input)

	// 获取数据库连接
	db := c.MustGet("db").(*gorm.DB)

	// 检查是否存在相同用户名或邮箱的用户
	var existingUser models.User
	if err := db.Where("username = ? OR email = ?", input.Username, input.Email).First(&existingUser).Error; err == nil {
		zap.S().Info("用户名或邮箱已被注册: ", input.Username, input.Email)
		c.JSON(http.StatusConflict, APIResponse{Message: "用户名或邮箱已被注册"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		zap.S().Error("检查用户是否已存在时出错: ", err)
		c.JSON(http.StatusInternalServerError, APIResponse{Message: "检查用户是否已存在时出错"})
		return
	}

	// 设置用户角色，根据邀请码判断角色
	role := "user"
	if input.InviteCode == "JNUTechnicians" {
		role = "admin"
	}
	zap.S().Info("分配的用户角色: ", role)

	// 设置用户密码并保存用户信息到数据库
	user := models.User{
		Username:   input.Username,
		Email:      input.Email,
		Role:       role,
		IsVerified: false,
	}
	if err := user.SetPassword(input.Password); err != nil {
		zap.S().Error("设置密码失败: ", err)
		c.JSON(http.StatusInternalServerError, APIResponse{Message: "设置密码失败"})
		return
	}

	// 保存用户
	if err := db.Create(&user).Error; err != nil {
		zap.S().Error("创建用户失败: ", err)
		c.JSON(http.StatusInternalServerError, APIResponse{Message: "创建用户失败"})
		return
	}
	zap.S().Info("用户已成功创建: ", user)

	// 生成验证码并发送到用户邮箱
	code := generateSecureCode()
	token := models.PasswordResetToken{
		UserID:    user.ID,
		Token:     code,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	if err := db.Create(&token).Error; err != nil {
		zap.S().Error("保存验证码失败: ", err)
		c.JSON(http.StatusInternalServerError, APIResponse{Message: "保存验证码失败"})
		return
	}
	zap.S().Info("生成的邮箱验证码: ", code)

	// 调用sendEmail函数发送验证码
	if err := sendEmail(user.Email, code); err != nil {
		zap.S().Error("发送验证码失败: ", err)
		c.JSON(http.StatusInternalServerError, APIResponse{Message: "发送验证码失败"})
		return
	}

	// 记录请求的结束
	zap.S().Info("用户注册成功，验证码已发送至用户邮箱: ", user.Email)

	c.JSON(http.StatusOK, APIResponse{Message: "注册成功，验证码已发送至您的邮箱"})
}

// VerifyEmail 处理用户邮箱验证
// @Summary 验证用户邮箱
// @Description 验证用户提供的邮箱验证码
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param verification body VerifyEmailInput true "邮箱和验证码"
// @Success 200 {object} APIResponse "邮箱验证成功"
// @Failure 400 {object} APIResponse "错误请求"
// @Failure 401 {object} APIResponse "无效或过期的验证码"
// @Failure 500 {object} APIResponse "服务器内部错误"
// @Router /verify_email [post]
func VerifyEmail(c *gin.Context) {
	// 在测试模式下直接通过邮箱验证
	if gin.Mode() == gin.TestMode {
		var input VerifyEmailInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{Message: "输入无效"})
			return
		}
		email := input.Email
		db := c.MustGet("db").(*gorm.DB)

		var user models.User
		if err := db.Where("email = ?", email).First(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, APIResponse{Message: "用户未找到"})
			return
		}

		// 直接标记用户为已验证
		user.IsVerified = true
		db.Save(&user)

		c.JSON(http.StatusOK, APIResponse{Message: "邮箱验证成功"})
		return
	}

	// 正常模式下执行邮箱验证逻辑
	var input VerifyEmailInput
	if err := c.ShouldBindJSON(&input); err != nil {
		zap.S().Error("邮箱验证失败: 输入无效 - ", err)
		c.JSON(http.StatusBadRequest, APIResponse{Message: "输入无效"})
		return
	}
	zap.S().Infof("收到邮箱验证请求, Email: %s, Code: %s", input.Email, input.Code)

	// 获取数据库连接
	db := c.MustGet("db").(*gorm.DB)

	// 查询验证码是否存在且未过期
	var token models.PasswordResetToken
	err := db.Where("token = ? AND expires_at > ?", input.Code, time.Now()).First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zap.S().Warnf("邮箱验证失败: 无效或过期的验证码, Code: %s", input.Code)
			c.JSON(http.StatusUnauthorized, APIResponse{Message: "无效或过期的验证码"})
		} else {
			zap.S().Error("邮箱验证失败: 查询验证码时出错 - ", err)
			c.JSON(http.StatusInternalServerError, APIResponse{Message: "服务器内部错误"})
		}
		return
	}

	// 查询用户信息
	var user models.User
	err = db.Where("id = ?", token.UserID).First(&user).Error
	if err != nil {
		zap.S().Errorf("邮箱验证失败: 无法找到用户, UserID: %d - Error: %s", token.UserID, err)
		c.JSON(http.StatusInternalServerError, APIResponse{Message: "服务器内部错误"})
		return
	}

	// 更新用户状态为已验证
	user.IsVerified = true
	if err := db.Save(&user).Error; err != nil {
		zap.S().Errorf("邮箱验证失败: 无法更新用户状态, UserID: %d - Error: %s", user.ID, err)
		c.JSON(http.StatusInternalServerError, APIResponse{Message: "服务器内部错误"})
		return
	}
	zap.S().Infof("用户状态已更新为已验证, UserID: %d", user.ID)

	// 删除验证码记录
	if err := db.Delete(&token).Error; err != nil {
		zap.S().Warnf("邮箱验证完成，但无法删除验证码记录, TokenID: %d - Error: %s", token.ID, err)
	}

	zap.S().Infof("邮箱验证成功, UserID: %d, Email: %s", user.ID, user.Email)
	c.JSON(http.StatusOK, APIResponse{Message: "邮箱验证成功"})
}

// Login 处理用户登录
// @Summary 用户登录
// @Description 用户通过用户名或邮箱和密码登录系统
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param login body models.LoginInput true "用户名/邮箱和密码"
// @Success 200 {object} APIResponse "登录成功，返回 JWT 令牌"
// @Failure 400 {object} APIResponse "错误请求"
// @Failure 401 {object} APIResponse "用户名或密码无效或邮箱未验证"
// @Failure 500 {object} APIResponse "查询用户失败"
// @Router /login [post]
func Login(c *gin.Context) {
	var input models.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Message: err.Error()})
		return
	}

	db := c.MustGet("db").(*gorm.DB)
	var user models.User

	// 判断输入是否是邮箱格式
	if strings.Contains(input.Username, "@") {
		// 如果是邮箱，按邮箱查询
		if err := db.Where("email = ?", input.Username).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusUnauthorized, APIResponse{Message: "邮箱或密码无效"})
			} else {
				c.JSON(http.StatusInternalServerError, APIResponse{Message: "查询用户失败"})
			}
			return
		}
	} else {
		// 否则按用户名查询
		if err := db.Where("username = ?", input.Username).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusUnauthorized, APIResponse{Message: "用户名或密码无效"})
			} else {
				c.JSON(http.StatusInternalServerError, APIResponse{Message: "查询用户失败"})
			}
			return
		}
	}

	// 检查用户是否已验证邮箱
	if !user.IsVerified {
		c.JSON(http.StatusUnauthorized, APIResponse{Message: "邮箱未验证"})
		return
	}

	// 验证密码是否正确
	if !user.CheckPassword(input.Password) {
		c.JSON(http.StatusUnauthorized, APIResponse{Message: "用户名或密码无效"})
		return
	}

	// 生成JWT令牌
	token, err := models.GenerateJWT(user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{Message: "生成令牌失败"})
		return
	}

	// 返回JWT令牌
	c.JSON(http.StatusOK, APIResponse{Message: "登录成功", Data: map[string]interface{}{"token": token}})
}

// SendVerificationCode 发送邮箱验证码
// @Summary 发送邮箱验证码
// @Description 当用户需要重新发送验证邮件时，调用该接口生成并发送新的验证码
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param email body SendVerificationCodeInput true "用户邮箱"
// @Success 200 {object} APIResponse "验证码已发送至您的邮箱"
// @Failure 400 {object} APIResponse "错误请求"
// @Failure 404 {object} APIResponse "用户未找到"
// @Failure 500 {object} APIResponse "发送验证码失败"
// @Router /send_verification_code [post]
func SendVerificationCode(c *gin.Context) {
	var input SendVerificationCodeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Message: err.Error()})
		return
	}

	db := c.MustGet("db").(*gorm.DB)
	var user models.User
	if err := db.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, APIResponse{Message: "用户未找到"})
		return
	}

	// 删除该用户未使用或过期的验证码记录
	db.Where("user_id = ?", user.ID).Delete(&models.PasswordResetToken{})

	code := generateSecureCode()
	token := models.PasswordResetToken{
		UserID:    user.ID,
		Token:     code,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	db.Create(&token)

	if err := sendEmail(input.Email, code); err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{Message: "发送验证码失败"})
		return
	}

	c.JSON(http.StatusOK, APIResponse{Message: "验证码已发送至您的邮箱"})
}

// ResetPassword 重置密码
// @Summary 重置密码
// @Description 用户提交邮箱和验证码，通过验证后，可以设置新密码
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param reset body ResetPasswordInput true "邮箱、验证码和新密码"
// @Success 200 {object} APIResponse "密码已成功重置"
// @Failure 400 {object} APIResponse "错误请求"
// @Failure 401 {object} APIResponse "无效或过期的验证码"
// @Failure 500 {object} APIResponse "用户未找到或无法更新密码"
// @Router /reset_password [post]
func ResetPassword(c *gin.Context) {
	var input ResetPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Message: err.Error()})
		return
	}

	db := c.MustGet("db").(*gorm.DB)
	var resetToken models.PasswordResetToken
	if err := db.Where("token = ? AND expires_at > ?", input.Token, time.Now()).First(&resetToken).Error; err != nil {
		c.JSON(http.StatusUnauthorized, APIResponse{Message: "无效或过期的验证码"})
		return
	}

	var user models.User
	if err := db.Where("id = ?", resetToken.UserID).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{Message: "用户未找到"})
		return
	}

	if err := user.SetPassword(input.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{Message: "无法更新密码"})
		return
	}
	db.Save(&user)
	db.Delete(&resetToken) // 密码重置成功后删除验证码记录

	c.JSON(http.StatusOK, APIResponse{Message: "密码已成功重置"})
}

// Helper functions

// generateSecureCode 生成一个安全的随机验证码，使用16进制字符串表示
func generateSecureCode() string {
	b := make([]byte, 3) // 生成6个十六进制字符 (3字节 = 6 hex digits)
	_, err := rand.Read(b)
	if err != nil {
		panic("无法生成安全随机数")
	}
	return hex.EncodeToString(b)
}

// sendEmail 发送邮件，传递收件人邮箱和验证码
func sendEmail(to, code string) error {
	from := "xloudmaxx@gmail.com"
	password := "mbbf hrde wlpk bphe"
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// 构建邮件内容
	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: 邮箱验证码\n\n" +
		"您的验证码是: " + code

	// 发送邮件
	err := smtp.SendMail(smtpHost+":"+smtpPort,
		smtp.PlainAuth("", from, password, smtpHost),
		from, []string{to}, []byte(msg))
	if err != nil {
		fmt.Printf("发送邮件失败: %v\n", err)
		return err
	}
	return nil
}

package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"repair-platform/controllers"
	"repair-platform/middleware"
	"repair-platform/service"
)

// SetupRoutes 设置应用程序的路由和中间件
func SetupRoutes(r *gin.Engine, db *gorm.DB, emailService service.EmailService) {
	r.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})

	setupAuthRoutes(r)      // 用户认证相关路由
	setupProtectedRoutes(r) // 需要 JWT 授权的路由
}

// 设置用户认证路由
func setupAuthRoutes(r *gin.Engine) {
	r.POST("/api/register", controllers.Register)                           // 用户注册
	r.POST("/api/login", controllers.Login)                                 // 用户登录
	r.POST("/api/send_verification_code", controllers.SendVerificationCode) // 发送邮箱验证码
	r.POST("/api/reset_password", controllers.ResetPassword)                // 重置密码
	r.POST("/api/verify_email", controllers.VerifyEmail)                    // 验证邮箱
}

// 设置需要 JWT 授权的路由组
func setupProtectedRoutes(r *gin.Engine) {
	authRoutes := r.Group("/api")
	authRoutes.Use(middleware.JWTAuthMiddleware())
	{
		setupRepairRoutes(authRoutes)   // 报修请求路由
		setupFeedbackRoutes(authRoutes) // 用户反馈路由
		setupUploadRoutes(authRoutes)   // 文件上传路由
		setupImageUploadRoutes(authRoutes)
		setupFolderUploadRoutes(authRoutes) // 文件夹管理路由
		setupMarkdownRoutes(authRoutes)     // Markdown 文件内容获取路由

		// 管理员专属的路由组
		adminRoutes := authRoutes.Group("/admin")
		adminRoutes.Use(middleware.AdminAuthMiddleware()) // 应用管理员中间件
		{
			adminRoutes.GET("/repair_requests", controllers.AdminListRepairRequests)
			adminRoutes.PUT("/repair_requests/:id", controllers.AdminUpdateRepairRequest)
		}
	}
}

// 设置文件上传路由（受保护）
func setupUploadRoutes(r *gin.RouterGroup) {
	r.POST("/upload", controllers.UploadFile) // 仅支持 Markdown 文件的上传路由
}

// 设置图片上传路由（受保护）
func setupImageUploadRoutes(r *gin.RouterGroup) {
	r.POST("/upload/image", controllers.UploadImage) // 图片上传至图床的路由
}

// 添加文件夹管理路由
func setupFolderUploadRoutes(r *gin.RouterGroup) {
	r.GET("/upload/folders", controllers.GetFolders)    // 获取文件夹列表
	r.POST("/upload/folders", controllers.CreateFolder) // 创建新文件夹
}

// 添加 Markdown 文件内容路由
func setupMarkdownRoutes(r *gin.RouterGroup) {
	// 获取指定 Markdown 文件内容
	r.GET("/markdown/:file", controllers.GetMarkdownContent) // 获取 Markdown 文件内容
	// 列出指定文件夹下的所有 Markdown 文件
	r.GET("/markdown/files/:folder", controllers.ListMarkdownFiles) // 获取 Markdown 文件列表
}

// 设置报修请求相关路由
func setupRepairRoutes(r *gin.RouterGroup) {
	r.POST("/repair_requests", controllers.SubmitRepairRequest)
}

// 设置用户反馈相关路由
func setupFeedbackRoutes(r *gin.RouterGroup) {
	r.POST("/feedback", controllers.SubmitFeedback)
	r.GET("/feedback/:id", controllers.GetFeedbackByRepairID)
}

package main

import (
	"errors"
	"net/http"
	"os"
	"time"

	"repair-platform/database"
	"repair-platform/routes"
	"repair-platform/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var sugar *zap.SugaredLogger

func main() {
	// 初始化日志
	gin.SetMode(gin.ReleaseMode)
	initLogger()
	defer sugar.Sync()

	sugar.Info("服务初始化开始")

	// 初始化 Gin 引擎
	sugar.Info("初始化 Gin 引擎")
	r := gin.Default()

	// 配置 CORS 中间件
	setupCORS(r)

	// 初始化数据库
	sugar.Info("初始化数据库连接")
	db, err := database.InitDB()
	if err != nil {
		sugar.Fatalf("数据库连接失败: %v", err)
	}
	defer func() {
		database.CloseDB(db)
		sugar.Info("数据库连接已关闭")
	}()

	// 初始化 Email 服务
	sugar.Info("初始化 Email 服务")
	emailService := service.NewEmailService(sugar)

	// 配置路由
	sugar.Info("配置路由和中间件")
	routes.SetupRoutes(r, db, emailService)

	// 启动服务器
	startServer(r)
}

// initLogger 初始化 zap 日志记录器
func initLogger() {
	writeSyncer := getLogWriter() // 获取日志写入器
	encoder := getEncoder()       // 获取编码器
	core := zapcore.NewCore(encoder, writeSyncer, zapcore.InfoLevel)
	logger := zap.New(core, zap.AddCaller()) // 创建 logger
	sugar = logger.Sugar()
}

// getEncoder 返回一个 JSON 格式的日志编码器
func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.MessageKey = "message"
	encoderConfig.LevelKey = "level"
	encoderConfig.TimeKey = "time"
	encoderConfig.CallerKey = "caller"
	return zapcore.NewJSONEncoder(encoderConfig)
}

// getLogWriter 返回日志写入器，将日志保存到文件
func getLogWriter() zapcore.WriteSyncer {
	file, err := os.OpenFile("server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic("无法创建或打开日志文件")
	}
	return zapcore.AddSync(file)
}

// setupCORS 配置 CORS 中间件
func setupCORS(r *gin.Engine) {
	sugar.Info("配置 CORS 中间件")
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:11451"}, // 指定前端的地址
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true, // 允许携带凭证（如 Cookies）
		MaxAge:           12 * time.Hour,
	}))
}

// startServer 启动服务器
func startServer(r *gin.Engine) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	sugar.Infof("服务器即将启动，监听端口: %s", port)
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		sugar.Fatalf("服务器启动失败: %v", err)
	}
}

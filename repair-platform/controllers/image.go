package controllers

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// SMMSResponse sm.ms API 响应结构
type SMMSResponse struct {
	Success bool     `json:"success"`
	Code    string   `json:"code"`
	Message string   `json:"message"`
	Data    SMMSData `json:"data"`
}

// SMMSData 响应数据
type SMMSData struct {
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Filename  string `json:"filename"`
	Storename string `json:"storename"`
	Size      int    `json:"size"`
	Path      string `json:"path"`
	Hash      string `json:"hash"`
	URL       string `json:"url"`
	Delete    string `json:"delete"`
	Page      string `json:"page"`
}

// SmmsApiUrl sm.ms API 配置
const SmmsApiUrl = "https://sm.ms/api/v2/upload"

// SmmsApiToken 硬编码的 API Token
const SmmsApiToken = "fnRtUbOCD3cXnO0GMbcp54EAsNEL5wXW"

// MaxFileSize 文件大小限制（10MB）
const MaxFileSize = 10 * 1024 * 1024

// UploadImage 处理图片上传到 sm.ms 的请求
func UploadImage(c *gin.Context) {
	logger := c.MustGet("logger").(*zap.SugaredLogger)
	logger.Info("开始处理图片上传请求")

	// 获取上传的文件
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		logger.Warn("图片上传失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "图片上传失败", "details": err.Error()})
		return
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			logger.Warn("关闭文件失败", zap.Error(cerr))
		}
	}()
	logger.Infof("接收到文件: %s, 大小: %d bytes", header.Filename, header.Size)

	// 验证文件大小
	if header.Size > MaxFileSize {
		logger.Warn("文件大小超过限制", zap.Int64("大小", header.Size))
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "文件大小超过限制（最大 10MB）"})
		return
	}

	// 验证文件类型
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		logger.Warn("文件类型不支持", zap.String("文件类型", ext))
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持 JPG 和 PNG 格式的图片"})
		return
	}

	// 创建请求体
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	fw, err := writer.CreateFormFile("smfile", header.Filename)
	if err != nil {
		logger.Error("创建表单文件失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建表单文件失败", "details": err.Error()})
		return
	}
	if _, err = io.Copy(fw, file); err != nil {
		logger.Error("复制图片失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "复制图片失败", "details": err.Error()})
		return
	}
	if err = writer.Close(); err != nil {
		logger.Error("关闭表单写入器失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "关闭表单写入器失败", "details": err.Error()})
		return
	}
	logger.Info("表单数据准备完成")

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", SmmsApiUrl, &b)
	if err != nil {
		logger.Error("创建 HTTP 请求失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建请求失败", "details": err.Error()})
		return
	}
	req.Header.Set("Authorization", SmmsApiToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 设置 HTTP 客户端超时
	client := &http.Client{Timeout: 10 * time.Second}

	// 发送请求到 sm.ms API
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("上传到 sm.ms 失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "上传到 sm.ms 失败", "details": err.Error()})
		return
	}
	defer resp.Body.Close()

	// 解析响应数据
	var result SMMSResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logger.Error("解析响应失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解析响应失败", "details": err.Error()})
		return
	}
	if !result.Success {
		logger.Warn("sm.ms API 响应错误", zap.String("错误信息", result.Message))
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Message})
		return
	}

	logger.Infof("图片上传成功, URL: %s, 删除链接: %s", result.Data.URL, result.Data.Delete)
	c.JSON(http.StatusOK, gin.H{
		"image_url":  result.Data.URL,
		"delete_url": result.Data.Delete,
	})
}

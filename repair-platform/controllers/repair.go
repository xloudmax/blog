package controllers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"repair-platform/models"
	"strings"
	"time"
)

// 文件上传配置
const (
	MaxFileSize2   = 5 << 20 // 5MB
	UploadDir      = "./uploads/"
	AllowedFormats = "jpg,jpeg,png,pdf"
)

// RepairRequestForm 用于绑定维修请求表单数据
type RepairRequestForm struct {
	Description string                `form:"description" binding:"required"`
	File        *multipart.FileHeader `form:"file"`
}

// SubmitRepairRequest 提交维修请求
// @Summary 提交维修请求
// @Description 用户提交新的维修请求，并上传相关文件
// @Tags 维修请求
// @Accept multipart/form-data
// @Produce json
// @Param description formData string true "维修请求描述"
// @Param file formData file false "上传的文件（图片或PDF）"
// @Success 200 {object} map[string]string "维修请求提交成功"
// @Failure 400 {object} map[string]string "输入数据无效或文件上传失败"
// @Failure 500 {object} map[string]string "提交维修请求失败"
// @Router /repair_requests [post]
func SubmitRepairRequest(c *gin.Context) {
	var form RepairRequestForm

	// 绑定请求数据
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "输入数据无效"})
		return
	}

	var request models.RepairRequest
	request.Description = form.Description

	// 处理文件上传
	if form.File != nil {
		if err := handleFileUpload(c, form.File, &request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("文件上传失败: %v", err)})
			return
		}
	}

	// 获取数据库连接
	db := c.MustGet("db").(*gorm.DB)

	// 创建新的维修请求
	if err := db.Create(&request).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交维修请求失败"})
		return
	}

	// 返回提交成功消息
	c.JSON(http.StatusOK, gin.H{"message": "维修请求提交成功"})
}

// handleFileUpload 处理文件上传，包含类型检查、大小限制和路径安全性
func handleFileUpload(c *gin.Context, file *multipart.FileHeader, request *models.RepairRequest) error {
	// 检查文件大小
	if file.Size > MaxFileSize2 {
		return errors.New("文件大小超过限制")
	}

	// 检查文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !strings.Contains(AllowedFormats, ext[1:]) {
		return errors.New("文件格式不支持，仅允许上传 " + AllowedFormats)
	}

	// 确保上传目录存在
	if err := os.MkdirAll(UploadDir, os.ModePerm); err != nil {
		return errors.New("无法创建上传目录")
	}

	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// 保存文件
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	filePath := filepath.Join(UploadDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		return errors.New("文件保存失败")
	}
	defer dst.Close()

	// 将文件内容复制到目标文件
	if _, err := io.Copy(dst, src); err != nil {
		return errors.New("文件保存失败")
	}

	// 将文件路径保存到维修请求中
	request.ImageURL = filePath
	return nil
}

// AdminListRepairRequests 管理员查看维修请求列表
// @Summary 查看维修请求列表
// @Description 管理员查看所有维修请求的列表
// @Tags 维修请求
// @Produce json
// @Success 200 {array} models.RepairRequest "维修请求列表"
// @Failure 500 {object} map[string]string "检索维修请求失败"
// @Router /repair_requests [get]
func AdminListRepairRequests(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	var requests []models.RepairRequest

	// 查找所有维修请求
	if err := db.Find(&requests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检索维修请求失败"})
		return
	}

	// 返回维修请求列表
	c.JSON(http.StatusOK, requests)
}

// AdminUpdateRepairRequest 管理员更新维修请求
// @Summary 更新维修请求
// @Description 管理员根据请求ID更新维修请求信息
// @Tags 维修请求
// @Accept json
// @Produce json
// @Param id path string true "维修请求ID"
// @Param request body models.RepairRequest true "更新的维修请求内容"
// @Success 200 {object} map[string]string "维修请求更新成功"
// @Failure 400 {object} map[string]string "输入数据无效"
// @Failure 404 {object} map[string]string "未找到维修请求"
// @Failure 500 {object} map[string]string "更新维修请求失败"
// @Router /repair_requests/{id} [put]
func AdminUpdateRepairRequest(c *gin.Context) {
	id := c.Param("id")
	var request models.RepairRequest

	// 获取数据库连接
	db := c.MustGet("db").(*gorm.DB)

	// 根据ID查找维修请求
	if err := db.Where("id = ?", id).First(&request).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到维修请求"})
		return
	}

	// 绑定JSON数据到维修请求模型
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "输入数据无效"})
		return
	}

	// 更新维修请求
	if err := db.Save(&request).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新维修请求失败"})
		return
	}

	// 返回更新成功消息
	c.JSON(http.StatusOK, gin.H{"message": "维修请求更新成功"})
}

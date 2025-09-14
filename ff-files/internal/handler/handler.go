package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-files/internal/service"
)

type Handler struct {
	svc *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

type UploadResponse struct {
	FileID string `json:"file_id"`
}

type TemporaryURLResponse struct {
	URL string `json:"url"`
}

func (h *Handler) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	// Получаем метаданные из тела запроса
	var metadata map[string]interface{}
	if err := c.ShouldBindJSON(&metadata); err != nil && err != http.ErrNotSupported {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid metadata format"})
		return
	}

	// Открываем файл
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer src.Close()

	// Получаем ID сервиса из контекста
	serviceID, exists := c.Get("service_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "service ID not found"})
		return
	}

	// Загружаем файл
	result, err := h.svc.UploadFile(c.Request.Context(), src, file.Size, file.Header.Get("Content-Type"), strconv.Itoa(serviceID.(int)), metadata)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, UploadResponse{FileID: result.ID.String()})
}

func (h *Handler) GetFile(c *gin.Context) {
	fileID, err := uuid.Parse(c.Param("file_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file ID"})
		return
	}

	file, err := h.svc.GetFile(c.Request.Context(), fileID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	// Получаем файл из MinIO
	object, err := h.svc.GetObject(c.Request.Context(), file.StoragePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get file"})
		return
	}

	c.Header("Content-Type", file.MimeType)
	c.Header("Content-Length", strconv.FormatInt(file.Size, 10))
	c.DataFromReader(http.StatusOK, file.Size, file.MimeType, object, nil)
}

func (h *Handler) DeleteFile(c *gin.Context) {
	fileID, err := uuid.Parse(c.Param("file_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file ID"})
		return
	}

	if err := h.svc.DeleteFile(c.Request.Context(), fileID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) GetFileMetadata(c *gin.Context) {
	fileID, err := uuid.Parse(c.Param("file_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file ID"})
		return
	}

	file, err := h.svc.GetFileMetadata(c.Request.Context(), fileID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.JSON(http.StatusOK, file)
}

func (h *Handler) GenerateTemporaryURL(c *gin.Context) {
	fileID, err := uuid.Parse(c.Param("file_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file ID"})
		return
	}

	expiresIn, err := strconv.Atoi(c.Query("expires_in"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expires_in parameter"})
		return
	}

	url, err := h.svc.GenerateTemporaryURL(c.Request.Context(), fileID, time.Duration(expiresIn)*time.Second)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, TemporaryURLResponse{URL: url})
}

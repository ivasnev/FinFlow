package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-files/internal/service"
	"net/http"
	"time"
)

type FileHandler struct {
	fileService service.FileService
}

func NewFileHandler(fileService service.FileService) *FileHandler {
	return &FileHandler{
		fileService: fileService,
	}
}

type UploadURLRequest struct {
	URL string `json:"url" binding:"required,url"`
}

type GenerateTemporaryURLRequest struct {
	Duration string `json:"duration" binding:"required"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *FileHandler) RegisterRoutes(router *gin.Engine) {
	router.POST("/upload", h.UploadFile)
	router.POST("/upload/url", h.UploadFileFromURL)
	router.GET("/file/:id", h.GetFile)
	router.GET("/file/:id/meta", h.GetFileMetadata)
	router.DELETE("/file/:id", h.DeleteFile)
	router.POST("/file/:id/url", h.GenerateTemporaryURL)
	router.GET("/temp/:id", h.GetFileByTemporaryURL)
}

func (h *FileHandler) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "failed to get file from form"})
		return
	}

	uploader := c.GetString("user_id") // Предполагается, что middleware устанавливает user_id
	if uploader == "" {
		uploader = "anonymous"
	}

	fileModel, err := h.fileService.UploadFile(c.Request.Context(), file, uploader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, fileModel)
}

func (h *FileHandler) UploadFileFromURL(c *gin.Context) {
	var req UploadURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	uploader := c.GetString("user_id")
	if uploader == "" {
		uploader = "anonymous"
	}

	fileModel, err := h.fileService.UploadFileFromURL(c.Request.Context(), req.URL, uploader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, fileModel)
}

func (h *FileHandler) GetFile(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "file id is required"})
		return
	}

	file, reader, err := h.fileService.GetFile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	defer reader.Close()

	c.Header("Content-Disposition", "attachment; filename="+file.Name)
	c.Header("Content-Type", file.MimeType)
	c.Header("Content-Length", string(file.Size))

	c.DataFromReader(http.StatusOK, file.Size, file.MimeType, reader, nil)
}

func (h *FileHandler) GetFileMetadata(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "file id is required"})
		return
	}

	file, metadata, err := h.fileService.GetFileMetadata(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"file":     file,
		"metadata": metadata,
	})
}

func (h *FileHandler) DeleteFile(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "file id is required"})
		return
	}

	if err := h.fileService.DeleteFile(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *FileHandler) GenerateTemporaryURL(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "file id is required"})
		return
	}

	var req GenerateTemporaryURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid duration format"})
		return
	}

	tempURL, err := h.fileService.GenerateTemporaryURL(c.Request.Context(), id, duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, tempURL)
}

func (h *FileHandler) GetFileByTemporaryURL(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "temporary url id is required"})
		return
	}

	file, reader, err := h.fileService.GetFileByTemporaryURL(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	defer reader.Close()

	c.Header("Content-Disposition", "attachment; filename="+file.Name)
	c.Header("Content-Type", file.MimeType)
	c.Header("Content-Length", string(file.Size))

	c.DataFromReader(http.StatusOK, file.Size, file.MimeType, reader, nil)
} 
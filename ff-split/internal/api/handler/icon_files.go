package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-auth/pkg/auth"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

const maxIconSize = 5 << 20

// IconFileGateway хранит серверный TVM-клиент файлового сервиса.
type IconFileGateway struct {
	Client  *http.Client
	BaseURL string
}

// UploadIconFile принимает изображение администратора и передаёт его через TVM.
func (s *ServerHandler) UploadIconFile(c *gin.Context) {
	if !requireIconAdmin(c) {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxIconSize+(1<<20))
	if err := c.Request.ParseMultipartForm(maxIconSize); err != nil {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "invalid or oversized multipart body"})
		return
	}
	defer c.Request.MultipartForm.RemoveAll()
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxIconSize+1))
	if err != nil || len(data) > maxIconSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "icon exceeds 5 MiB"})
		return
	}
	switch http.DetectContentType(data) {
	case "image/png", "image/jpeg", "image/gif", "image/webp":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "PNG, JPEG, GIF or WebP image required"})
		return
	}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", header.Filename)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
		return
	}
	if _, err = part.Write(data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encode upload"})
		return
	}
	if err = writer.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encode upload"})
		return
	}
	s.forwardIconFile(c, http.MethodPost, "/api/v1/files", &body, writer.FormDataContentType())
}

// GetIconFileURL выдаёт ссылку только на файл зарегистрированной общей иконки.
func (s *ServerHandler) GetIconFileURL(c *gin.Context, fileId openapi_types.UUID) {
	icons, err := s.iconService.GetIcons(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read icons"})
		return
	}
	for _, icon := range icons {
		if icon.FileUUID == fileId.String() {
			s.forwardIconFile(c, http.MethodGet, "/api/v1/files/"+fileId.String(), nil, "")
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "icon file not found"})
}

// forwardIconFile ограничивает ответ и не пересылает пользовательские заголовки сервису.
func (s *ServerHandler) forwardIconFile(c *gin.Context, method, path string, body io.Reader, contentType string) {
	if s.iconFiles == nil || s.iconFiles.Client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "file gateway unavailable"})
		return
	}
	req, err := http.NewRequestWithContext(c.Request.Context(), method, strings.TrimRight(s.iconFiles.BaseURL, "/")+path, body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "invalid file service configuration"})
		return
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	response, err := s.iconFiles.Client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "file service unavailable"})
		return
	}
	defer response.Body.Close()
	data, err := io.ReadAll(io.LimitReader(response.Body, (1<<20)+1))
	if err != nil || len(data) > 1<<20 || !json.Valid(data) {
		c.JSON(http.StatusBadGateway, gin.H{"error": "invalid file service response"})
		return
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		c.JSON(http.StatusBadGateway, gin.H{"error": "file service rejected request"})
		return
	}
	c.Data(response.StatusCode, "application/json", data)
}

// requireIconAdmin защищает публикацию файлов в общем списке иконок.
func requireIconAdmin(c *gin.Context) bool {
	actor, ok := auth.GetUserData(c)
	if !ok || !actor.IsAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin role required"})
		return false
	}
	return true
}

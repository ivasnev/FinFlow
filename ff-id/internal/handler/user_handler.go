package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-id/internal/models"
	"github.com/ivasnev/FinFlow/ff-id/internal/service"
	"net/http"
	"io"
	"fmt"
	"strings"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type UpdateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
}

// UploadAvatarResponse представляет ответ на загрузку аватара
type UploadAvatarResponse struct {
	AvatarURL string `json:"avatar_url"`
}

func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.userService.Register(c.Request.Context(), req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		if err == service.ErrEmailTaken {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "user registered successfully"})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, refreshToken, err := h.userService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, refreshToken, err := h.userService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetUint("user_id")
	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"email":      user.Email,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"phone":      user.Phone,
		"role":       user.Role,
	})
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")
	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user profile"})
		return
	}

	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}

	if err := h.userService.UpdateUser(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "profile updated successfully"})
}

func (h *UserHandler) DeleteProfile(c *gin.Context) {
	userID := c.GetUint("user_id")
	if err := h.userService.DeleteUser(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "profile deleted successfully"})
}

// UploadAvatar загружает аватар пользователя
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "failed to get avatar file"})
		return
	}

	// Проверяем размер и тип файла
	if file.Size > 5*1024*1024 { // 5MB
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "file size exceeds 5MB limit"})
		return
	}

	contentType := file.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "file must be an image"})
		return
	}

	// Открываем файл
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to open file"})
		return
	}
	defer src.Close()

	// Читаем содержимое файла
	fileData, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read file"})
		return
	}

	// Загружаем файл в сервис файлов
	avatarID, err := h.filesClient.UploadFile(c.Request.Context(), fileData, file.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to upload file"})
		return
	}

	// Обновляем ID аватара в профиле пользователя
	if err := h.userService.UpdateAvatar(c.Request.Context(), userID, avatarID); err != nil {
		// Если не удалось обновить профиль, удаляем загруженный файл
		_ = h.filesClient.DeleteFile(c.Request.Context(), avatarID)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update avatar"})
		return
	}

	c.JSON(http.StatusOK, UploadAvatarResponse{
		AvatarURL: fmt.Sprintf("/file/%s", avatarID),
	})
}

// DeleteAvatar удаляет аватар пользователя
func (h *UserHandler) DeleteAvatar(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	// Получаем текущий ID аватара
	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get user"})
		return
	}

	if user.AvatarID != "" {
		// Удаляем файл из сервиса файлов
		if err := h.filesClient.DeleteFile(c.Request.Context(), user.AvatarID); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to delete avatar file"})
			return
		}

		// Очищаем ID аватара в профиле
		if err := h.userService.UpdateAvatar(c.Request.Context(), userID, ""); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update avatar"})
			return
		}
	}

	c.Status(http.StatusNoContent)
}

func (h *UserHandler) RegisterRoutes(router *gin.Engine) {
	// ... existing routes ...
	
	// Маршруты для работы с аватарами
	router.POST("/users/avatar", h.UploadAvatar)
	router.DELETE("/users/avatar", h.DeleteAvatar)
} 
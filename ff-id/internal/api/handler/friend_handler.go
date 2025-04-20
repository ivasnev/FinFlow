package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-id/internal/api/dto"
	"github.com/ivasnev/FinFlow/ff-id/internal/service"
)

// FriendHandler обрабатывает запросы, связанные с друзьями пользователей
type FriendHandler struct {
	friendService service.FriendServiceInterface
}

// NewFriendHandler создает новый FriendHandler
func NewFriendHandler(friendService service.FriendServiceInterface) *FriendHandler {
	return &FriendHandler{
		friendService: friendService,
	}
}

// AddFriend обрабатывает запрос на добавление друга
// @Summary Добавление друга
// @Description Добавляет пользователя в список друзей
// @Tags friends
// @Accept json
// @Produce json
// @Param request body dto.AddFriendRequest true "Данные для добавления друга"
// @Success 201 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /users/me/friends [post]
func (h *FriendHandler) AddFriend(c *gin.Context) {
	// Получаем ID пользователя из контекста
	userIDStr, exist := c.Get("user_id")
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "отсутствует ID пользователя в контексте"})
		return
	}
	userID, ok := userIDStr.(int64)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный формат ID пользователя"})
		return
	}

	// Парсим данные из запроса
	var req dto.AddFriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Добавляем друга
	err := h.friendService.AddFriend(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем успешный ответ
	c.JSON(http.StatusCreated, gin.H{"message": "заявка в друзья успешно отправлена"})
}

// FriendAction обрабатывает запрос на действие с заявкой в друзья
// @Summary Действие с заявкой в друзья
// @Description Принимает, отклоняет или блокирует пользователя
// @Tags friends
// @Accept json
// @Produce json
// @Param request body dto.FriendActionRequest true "Данные для действия с заявкой"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /users/me/friends/action [post]
func (h *FriendHandler) FriendAction(c *gin.Context) {
	// Получаем ID пользователя из контекста
	userIDStr, exist := c.Get("user_id")
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "отсутствует ID пользователя в контексте"})
		return
	}
	userID, ok := userIDStr.(int64)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный формат ID пользователя"})
		return
	}

	// Парсим данные из запроса
	var req dto.FriendActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var err error
	var message string

	// Выполняем нужное действие в зависимости от параметра action
	switch req.Action {
	case "accept":
		err = h.friendService.AcceptFriendRequest(c.Request.Context(), userID, req)
		message = "заявка в друзья принята"
	case "reject":
		err = h.friendService.RejectFriendRequest(c.Request.Context(), userID, req)
		message = "заявка в друзья отклонена"
	case "block":
		err = h.friendService.BlockUser(c.Request.Context(), userID, req)
		message = "пользователь заблокирован"
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректное действие"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем успешный ответ
	c.JSON(http.StatusOK, gin.H{"message": message})
}

// RemoveFriend обрабатывает запрос на удаление друга
// @Summary Удаление друга
// @Description Удаляет пользователя из списка друзей
// @Tags friends
// @Produce json
// @Param friend_id path int true "ID друга для удаления"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /users/me/friends/{friend_id} [delete]
func (h *FriendHandler) RemoveFriend(c *gin.Context) {
	// Получаем ID пользователя из контекста
	userIDStr, exist := c.Get("user_id")
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "отсутствует ID пользователя в контексте"})
		return
	}
	userID, ok := userIDStr.(int64)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный формат ID пользователя"})
		return
	}

	// Получаем ID друга из URL
	var friendID int64
	if _, err := fmt.Sscanf(c.Param("friend_id"), "%d", &friendID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный ID друга"})
		return
	}

	// Удаляем друга
	err := h.friendService.RemoveFriend(c.Request.Context(), userID, friendID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем успешный ответ
	c.JSON(http.StatusOK, gin.H{"message": "друг успешно удален"})
}

// GetFriends обрабатывает запрос на получение списка друзей
// @Summary Получение списка друзей
// @Description Возвращает список друзей пользователя с пагинацией и опциональной фильтрацией по имени
// @Tags friends
// @Produce json
// @Param nickname path string true "Никнейм пользователя"
// @Param page query int false "Номер страницы (по умолчанию 1)"
// @Param page_size query int false "Размер страницы (по умолчанию 20, максимум 100)"
// @Param friend_name query string false "Фильтр по имени друга (ILIKE)"
// @Param status query string false "Фильтр по статусу друга (по умолчанию 'accepted')"
// @Success 200 {object} dto.FriendsListResponse
// @Failure 400 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /users/{nickname}/friends [get]
func (h *FriendHandler) GetFriends(c *gin.Context) {
	// Получаем никнейм пользователя из URL
	nickname := c.Param("nickname")

	// Получаем параметры запроса для пагинации и фильтрации
	var params dto.FriendsQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем список друзей
	response, err := h.friendService.GetFriends(c.Request.Context(), nickname, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем успешный ответ
	c.JSON(http.StatusOK, response)
}

// GetFriendRequests обрабатывает запрос на получение списка заявок в друзья
// @Summary Получение списка заявок в друзья
// @Description Возвращает список входящих или исходящих заявок в друзья
// @Tags friends
// @Produce json
// @Param incoming query bool false "Тип заявок (true - входящие, false - исходящие)"
// @Param page query int false "Номер страницы (по умолчанию 1)"
// @Param page_size query int false "Размер страницы (по умолчанию 20, максимум 100)"
// @Success 200 {object} dto.FriendsListResponse
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /users/me/friend-requests [get]
func (h *FriendHandler) GetFriendRequests(c *gin.Context) {
	// Получаем ID пользователя из контекста
	userIDStr, exist := c.Get("user_id")
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "отсутствует ID пользователя в контексте"})
		return
	}
	userID, ok := userIDStr.(int64)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный формат ID пользователя"})
		return
	}

	// Получаем параметры запроса
	incoming := c.DefaultQuery("incoming", "true") == "true"

	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		fmt.Sscanf(pageStr, "%d", &page)
		if page < 1 {
			page = 1
		}
	}

	pageSize := 20
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		fmt.Sscanf(pageSizeStr, "%d", &pageSize)
		if pageSize < 1 {
			pageSize = 20
		}
	}

	// Получаем список заявок
	response, err := h.friendService.GetFriendRequests(c.Request.Context(), userID, page, pageSize, incoming)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем успешный ответ
	c.JSON(http.StatusOK, response)
}

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-auth/pkg/auth"
	"github.com/ivasnev/FinFlow/ff-id/internal/service"
	"github.com/ivasnev/FinFlow/ff-id/pkg/api"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// ServerHandler реализует интерфейс api.ServerInterface
type ServerHandler struct {
	friendService service.FriendServiceInterface
	userService   service.UserServiceInterface
}

// NewServerHandler создает новый ServerHandler
func NewServerHandler(
	friendService service.FriendServiceInterface,
	userService service.UserServiceInterface,
) *ServerHandler {
	return &ServerHandler{
		friendService: friendService,
		userService:   userService,
	}
}

// AddFriend обрабатывает запрос на отправку заявки в друзья
func (h *ServerHandler) AddFriend(c *gin.Context) {
	// Получаем данные пользователя из контекста
	userData, exists := auth.GetUserData(c)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "отсутствует ID пользователя в контексте"})
		return
	}

	// Парсим данные из запроса
	var req api.AddFriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	serviceReq := service.AddFriendRequest{
		FriendNickname: req.FriendNickname,
	}

	// Добавляем друга
	err := h.friendService.AddFriend(c.Request.Context(), userData.UserID, serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем успешный ответ
	c.JSON(http.StatusCreated, gin.H{"message": "заявка в друзья успешно отправлена"})
}

// FriendAction обрабатывает запрос на действие с заявкой в друзья
func (h *ServerHandler) FriendAction(c *gin.Context) {
	// Получаем данные пользователя из контекста
	userData, exists := auth.GetUserData(c)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "отсутствует ID пользователя в контексте"})
		return
	}

	// Парсим данные из запроса
	var req api.FriendActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	serviceReq := service.FriendActionRequest{
		UserID: userData.UserID,
		Action: string(req.Action),
	}

	var err error
	var message string

	// Выполняем нужное действие в зависимости от параметра action
	switch req.Action {
	case "accept":
		err = h.friendService.AcceptFriendRequest(c.Request.Context(), userData.UserID, serviceReq)
		message = "заявка в друзья принята"
	case "reject":
		err = h.friendService.RejectFriendRequest(c.Request.Context(), userData.UserID, serviceReq)
		message = "заявка в друзья отклонена"
	case "block":
		err = h.friendService.BlockUser(c.Request.Context(), userData.UserID, serviceReq)
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
func (h *ServerHandler) RemoveFriend(c *gin.Context, friendId int64) {
	// Получаем данные пользователя из контекста
	userData, exists := auth.GetUserData(c)
	if !exists {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "отсутствует ID пользователя в контексте"})
		return
	}

	// Удаляем друга
	err := h.friendService.RemoveFriend(c.Request.Context(), userData.UserID, friendId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{Error: err.Error()})
		return
	}

	// Возвращаем успешный ответ
	c.JSON(http.StatusOK, gin.H{"message": "друг успешно удален"})
}

// GetFriends обрабатывает запрос на получение списка друзей
func (h *ServerHandler) GetFriends(c *gin.Context, nickname string, params api.GetFriendsParams) {
	// Формируем параметры для сервиса
	page := 1
	if params.Page != nil {
		page = *params.Page
	}

	pageSize := 20
	if params.PageSize != nil {
		pageSize = *params.PageSize
	}

	friendName := ""
	if params.FriendName != nil {
		friendName = *params.FriendName
	}

	status := ""
	if params.Status != nil {
		status = string(*params.Status)
	}

	serviceParams := service.FriendsQueryParams{
		Page:       page,
		PageSize:   pageSize,
		FriendName: friendName,
		Status:     status,
	}

	// Получаем список друзей
	serviceResponse, err := h.friendService.GetFriends(c.Request.Context(), nickname, serviceParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{Error: err.Error()})
		return
	}

	// Конвертируем в API типы
	apiResponse := convertToAPIFriendsListResponse(serviceResponse)

	// Возвращаем успешный ответ
	c.JSON(http.StatusOK, apiResponse)
}

// GetFriendRequests обрабатывает запрос на получение списка заявок в друзья
func (h *ServerHandler) GetFriendRequests(c *gin.Context, params api.GetFriendRequestsParams) {
	// Получаем данные пользователя из контекста
	userData, exists := auth.GetUserData(c)
	if !exists {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "отсутствует ID пользователя в контексте"})
		return
	}

	// Получаем параметры запроса
	incoming := true
	if params.Incoming != nil {
		incoming = *params.Incoming
	}

	page := 1
	if params.Page != nil {
		page = *params.Page
	}

	pageSize := 20
	if params.PageSize != nil {
		pageSize = *params.PageSize
	}

	// Получаем список заявок
	serviceResponse, err := h.friendService.GetFriendRequests(c.Request.Context(), userData.UserID, page, pageSize, incoming)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{Error: err.Error()})
		return
	}

	// Конвертируем в API типы
	apiResponse := convertToAPIFriendsListResponse(serviceResponse)

	// Возвращаем успешный ответ
	c.JSON(http.StatusOK, apiResponse)
}

// GetUsersByIds обрабатывает запрос на получение информации о пользователях по их ID
func (h *ServerHandler) GetUsersByIds(c *gin.Context, params api.GetUsersByIdsParams) {
	// Получаем пользователей
	serviceUsers, err := h.userService.GetUsersByIds(c.Request.Context(), params.UserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{Error: err.Error()})
		return
	}

	// Конвертируем в API типы
	apiUsers := make([]api.UserDTO, len(serviceUsers))
	for i, user := range serviceUsers {
		apiUsers[i] = convertToAPIUserDTO(user)
	}

	c.JSON(http.StatusOK, apiUsers)
}

// GetUserByNickname обрабатывает запрос на получение информации о пользователе по никнейму
func (h *ServerHandler) GetUserByNickname(c *gin.Context, nickname string) {
	serviceUser, err := h.userService.GetUserByNickname(c.Request.Context(), nickname)
	if err != nil {
		c.JSON(http.StatusNotFound, api.ErrorResponse{Error: "user not found"})
		return
	}

	// Конвертируем в API типы
	apiUser := convertToAPIUserDTO(serviceUser)

	c.JSON(http.StatusOK, apiUser)
}

// UpdateUser обрабатывает запрос на обновление профиля пользователя
func (h *ServerHandler) UpdateUser(c *gin.Context) {
	// Получаем данные пользователя из контекста
	userData, exists := auth.GetUserData(c)
	if !exists {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "user not found in context"})
		return
	}

	var req api.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: err.Error()})
		return
	}

	// Конвертируем в сервисный тип
	var email *string
	if req.Email != nil {
		emailStr := string(*req.Email)
		email = &emailStr
	}

	serviceReq := service.UpdateUserRequest{
		Email:     email,
		Phone:     req.Phone,
		Name:      req.Name,
		Birthdate: req.Birthdate,
		Nickname:  req.Nickname,
	}

	serviceUser, err := h.userService.UpdateUser(c.Request.Context(), userData.UserID, serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{Error: err.Error()})
		return
	}

	// Конвертируем в API типы
	apiUser := convertToAPIUserDTO(serviceUser)

	c.JSON(http.StatusOK, apiUser)
}

// RegisterUser обрабатывает запрос на регистрацию пользователя от клиента с токеном авторизации
func (h *ServerHandler) RegisterUser(c *gin.Context) {
	// Получаем данные пользователя из контекста (установлен middleware)
	userData, exists := auth.GetUserData(c)
	if !exists {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "отсутствует ID пользователя в контексте"})
		return
	}

	// Парсим данные из запроса
	var req api.RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: err.Error()})
		return
	}

	// Конвертируем в сервисный тип
	serviceReq := &service.RegisterUserRequest{
		Email:     string(req.Email),
		Nickname:  req.Nickname,
		Phone:     req.Phone,
		Birthdate: req.Birthdate,
		AvatarID:  req.AvatarId,
		Name:      req.Name,
	}

	// Регистрируем пользователя
	serviceUser, err := h.userService.RegisterUser(c.Request.Context(), userData.UserID, serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{Error: err.Error()})
		return
	}

	// Конвертируем в API типы
	apiUser := convertToAPIUserDTO(serviceUser)

	// Возвращаем успешный ответ
	c.JSON(http.StatusCreated, apiUser)
}

// RegisterUserFromService обрабатывает запрос на регистрацию пользователя от другого сервиса
func (h *ServerHandler) RegisterUserFromService(c *gin.Context) {
	// Парсим данные из запроса
	var req api.ServiceRegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: err.Error()})
		return
	}

	// Конвертируем в сервисный тип
	serviceReq := &service.RegisterUserRequest{
		Email:    string(req.Email),
		Nickname: req.Nickname,
		Name:     req.Name,
	}

	// Регистрируем пользователя
	serviceUser, err := h.userService.RegisterUser(c.Request.Context(), req.UserId, serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{Error: err.Error()})
		return
	}

	// Конвертируем в API типы
	apiUser := convertToAPIUserDTO(serviceUser)

	// Возвращаем успешный ответ
	c.JSON(http.StatusCreated, apiUser)
}

// convertToAPIUserDTO конвертирует сервисный UserDTO в API UserDTO
func convertToAPIUserDTO(serviceUser *service.UserDTO) api.UserDTO {
	apiUser := api.UserDTO{
		Id:        serviceUser.ID,
		Email:     openapi_types.Email(serviceUser.Email),
		Nickname:  serviceUser.Nickname,
		CreatedAt: serviceUser.CreatedAt,
		UpdatedAt: serviceUser.UpdatedAt,
	}

	if serviceUser.Phone != nil {
		apiUser.Phone = serviceUser.Phone
	}

	if serviceUser.Name != nil {
		apiUser.Name = serviceUser.Name
	}

	if serviceUser.Birthdate != nil {
		apiUser.Birthdate = serviceUser.Birthdate
	}

	if serviceUser.AvatarID != nil {
		avatarID := openapi_types.UUID(*serviceUser.AvatarID)
		apiUser.AvatarId = &avatarID
	}

	return apiUser
}

// convertToAPIFriendsListResponse конвертирует сервисный FriendsListResponse в API FriendsListResponse
func convertToAPIFriendsListResponse(serviceResponse *service.FriendsListResponse) api.FriendsListResponse {
	apiFriends := make([]api.FriendDTO, len(serviceResponse.Objects))
	for i, serviceFriend := range serviceResponse.Objects {
		apiFriend := api.FriendDTO{
			UserId: serviceFriend.UserID,
			Name:   serviceFriend.Name,
		}

		if serviceFriend.PhotoID != (uuid.UUID{}) {
			photoID := openapi_types.UUID(serviceFriend.PhotoID)
			apiFriend.PhotoId = &photoID
		}

		if serviceFriend.Status != "" {
			status := api.FriendDTOStatus(serviceFriend.Status)
			apiFriend.Status = &status
		}

		apiFriends[i] = apiFriend
	}

	return api.FriendsListResponse{
		Page:     serviceResponse.Page,
		PageSize: serviceResponse.PageSize,
		Total:    serviceResponse.Total,
		Objects:  apiFriends,
	}
}

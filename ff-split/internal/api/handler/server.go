package handler

import (
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
)

// ServerHandler реализует сгенерированный интерфейс api.ServerInterface
type ServerHandler struct {
	eventService       service.Event
	userService        service.User
	transactionService service.Transaction
	activityService    service.Activity
	taskService        service.Task
	categoryService    service.Category
	iconService        service.Icon
}

// NewServerHandler создает новый экземпляр ServerHandler
func NewServerHandler(
	eventService service.Event,
	userService service.User,
	transactionService service.Transaction,
	activityService service.Activity,
	taskService service.Task,
	categoryService service.Category,
	iconService service.Icon,
) *ServerHandler {
	return &ServerHandler{
		eventService:       eventService,
		userService:        userService,
		transactionService: transactionService,
		activityService:    activityService,
		taskService:        taskService,
		categoryService:    categoryService,
		iconService:        iconService,
	}
}

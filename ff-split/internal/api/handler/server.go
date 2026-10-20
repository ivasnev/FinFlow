package handler

import (
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
)

// ServerHandler реализует сгенерированный интерфейс api.ServerInterface
type ServerHandler struct {
	historyService     service.History
	settlementService  service.Settlement
	eventService       service.Event
	userService        service.User
	transactionService service.Transaction
	activityService    service.Activity
	taskService        service.Task
	categoryService    service.Category
	iconService        service.Icon
	iconFiles          *IconFileGateway
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
	iconFiles *IconFileGateway,
) *ServerHandler {
	return &ServerHandler{
		eventService:       eventService,
		userService:        userService,
		transactionService: transactionService,
		activityService:    activityService,
		taskService:        taskService,
		categoryService:    categoryService,
		iconService:        iconService,
		iconFiles:          iconFiles,
	}
}

// SetSettlementService подключает журнал возвратов к обработчику.
func (s *ServerHandler) SetSettlementService(ledger service.Settlement) { s.settlementService = ledger }

// SetHistoryService подключает доступ только для чтения журнала.
func (s *ServerHandler) SetHistoryService(history service.History) { s.historyService = history }

package handler

import "github.com/gin-gonic/gin"

// TaskHandlerInterface интерфейс для обработки запросов по задачам
type TaskHandlerInterface interface {
	GetTasksByEventID(c *gin.Context)
	GetTaskByID(c *gin.Context)
	CreateTask(c *gin.Context)
	UpdateTask(c *gin.Context)
	DeleteTask(c *gin.Context)
}

// IconHandlerInterface интерфейс для обработки запросов по иконкам
type IconHandlerInterface interface {
	GetIcons(c *gin.Context)
	GetIconByID(c *gin.Context)
	CreateIcon(c *gin.Context)
	UpdateIcon(c *gin.Context)
	DeleteIcon(c *gin.Context)
}

// CategoryHandlerInterface интерфейс для обработки запросов по категориям
type CategoryHandlerInterface interface {
	Options(c *gin.Context)
	GetCategories(c *gin.Context)
	GetCategoryByID(c *gin.Context)
	CreateCategory(c *gin.Context)
	UpdateCategory(c *gin.Context)
	DeleteCategory(c *gin.Context)
}

// EventHandlerInterface интерфейс для обработки запросов по мероприятиям
type EventHandlerInterface interface {
	GetEvents(c *gin.Context)
	GetEventByID(c *gin.Context)
	CreateEvent(c *gin.Context)
	UpdateEvent(c *gin.Context)
	DeleteEvent(c *gin.Context)
}

// ActivityHandlerInterface интерфейс для обработки запросов по активностям
type ActivityHandlerInterface interface {
	GetActivitiesByEventID(c *gin.Context)
	GetActivityByID(c *gin.Context)
	CreateActivity(c *gin.Context)
	UpdateActivity(c *gin.Context)
	DeleteActivity(c *gin.Context)
}

// TransactionHandlerInterface интерфейс для обработки запросов по транзакциям
type TransactionHandlerInterface interface {
	GetTransactionsByEventID(c *gin.Context)
	GetTransactionByID(c *gin.Context)
	CreateTransaction(c *gin.Context)
	UpdateTransaction(c *gin.Context)
	DeleteTransaction(c *gin.Context)
	GetDebtsByEventID(c *gin.Context)
}

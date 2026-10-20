package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

// ArchiveAccess отключает изменения в архиве; проверки под SQL-блокировкой закрывают гонку с архивированием.
func ArchiveAccess(database *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.Param("id_event")
		if raw == "" || c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead {
			return
		}
		if c.FullPath() == "/api/v1/event/:id_event/restore" || (c.Request.Method == http.MethodDelete && c.FullPath() == "/api/v1/event/:id_event") {
			return
		}
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return
		}
		var archived bool
		if err := database.WithContext(c.Request.Context()).Raw("SELECT deleted_at IS NOT NULL FROM events WHERE id=?", id).Scan(&archived).Error; err != nil {
			errors.HTTPErrorHandler(c, err)
			c.Abort()
			return
		}
		if archived {
			errors.HTTPErrorHandler(c, errors.ErrEventArchived)
			c.Abort()
		}
	}
}

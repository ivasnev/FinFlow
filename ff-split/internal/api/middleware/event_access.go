package middleware

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-auth/pkg/auth"
	"github.com/ivasnev/FinFlow/ff-split/internal/service/access"
)

// EventAccess проверяет все маршруты с id_event после аутентификации.
func EventAccess(checker *access.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawEvent := c.Param("id_event")
		if rawEvent == "" {
			return
		}
		actor, ok := auth.GetUserData(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
		eventID, err := strconv.ParseInt(rawEvent, 10, 64)
		if err != nil || eventID <= 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid event ID"})
			return
		}
		var kind string
		var resourceID int64
		for _, candidate := range []string{"transaction", "task", "activity", "user", "settlement", "invitation"} {
			if raw := c.Param("id_" + candidate); raw != "" {
				kind = candidate
				resourceID, err = strconv.ParseInt(raw, 10, 64)
				if err != nil || resourceID <= 0 {
					c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid resource ID"})
					return
				}
				break
			}
		}
		if err := checker.Authorize(c.Request.Context(), actor.UserID, eventID, kind, resourceID); err != nil {
			status := http.StatusInternalServerError
			message := "failed to check event access"
			if errors.Is(err, access.ErrDenied) {
				status = http.StatusForbidden
				message = "event access denied"
			}
			c.AbortWithStatusJSON(status, gin.H{"error": message})
		}
	}
}

package errors

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestOptimizationHTTPError проверяет явный отказ при перегрузке и изменившемся снимке.
func TestOptimizationHTTPError(t *testing.T) {
	for _, tc := range []struct {
		name   string
		err    error
		status int
		retry  string
	}{
		{"busy", ErrOptimizationBusy, 429, "1"}, {"changed", ErrDebtRevisionChanged, 409, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			res := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(res)
			ctx.Request = httptest.NewRequest("GET", "/optimized-debts", nil)
			HTTPErrorHandler(ctx, fmt.Errorf("optimization: %w", tc.err))
			if res.Code != tc.status || res.Header().Get("Retry-After") != tc.retry {
				t.Fatalf("status=%d headers=%v", res.Code, res.Header())
			}
		})
	}
}

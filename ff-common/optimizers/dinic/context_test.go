package dinic

import (
	"context"
	"errors"
	"testing"

	"github.com/ivasnev/FinFlow/ff-common/optimizers"
)

// TestOptimizeContextCancellation проверяет, что отмена до расчёта не выдаёт результат.
func TestOptimizeContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got, err := New().OptimizeContext(ctx, []optimizers.Transfer{{From: "a", To: "b", Amount: 50}})
	if !errors.Is(err, context.Canceled) || got != nil {
		t.Fatalf("cancelled optimization: %+v %v", got, err)
	}
}

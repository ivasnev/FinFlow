package tests

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/config"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
)

// TestOptimizationPreservesCents проверяет дробные суммы от расхода до сохранённого результата оптимизации.
func TestOptimizationPreservesCents(t *testing.T) {
	database := setupTestDB(t)
	t.Cleanup(func() { teardownTestDB(t, database) })
	if err := database.DB.Exec(`INSERT INTO users(id,is_dummy) VALUES(1,true),(2,true);
		INSERT INTO events(id,name) VALUES(1,'trip');
		INSERT INTO user_event(user_id,event_id) VALUES(1,1),(2,1);`).Error; err != nil {
		t.Fatal(err)
	}
	c, err := createTestContainer(t, &config.Config{}, gin.New(), database.DB, http.DefaultClient)
	if err != nil {
		t.Fatal(err)
	}
	for _, amount := range []float64{0.01, 0.4, 0.6, 1.23} {
		t.Run(fmt.Sprintf("amount=%.2f", amount), func(t *testing.T) {
			if err := database.DB.Exec("TRUNCATE optimized_debts, transactions CASCADE").Error; err != nil {
				t.Fatal(err)
			}
			request := &service.TransactionRequest{Name: "expense", Type: "amount", FromUser: 1, Users: []int64{1, 2}, Amount: amount * 2, Portion: map[string]float64{"1": amount, "2": amount}}
			if _, err := c.TransactionService.CreateTransaction(context.Background(), 1, request); err != nil {
				t.Fatal(err)
			}
			optimized, err := c.TransactionService.OptimizeDebts(context.Background(), 1)
			if err != nil {
				t.Fatal(err)
			}
			if len(optimized) != 1 || optimized[0].Amount != amount {
				t.Fatalf("amount=%.2f result=%+v", amount, optimized)
			}
			var stored float64
			if err := database.DB.Table("optimized_debts").Select("amount").Where("event_id = ?", 1).Scan(&stored).Error; err != nil {
				t.Fatal(err)
			}
			if stored != amount {
				t.Fatalf("stored=%.2f want=%.2f", stored, amount)
			}
		})
	}
}

package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/config"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
)

// TestTransactionAtomicity проверяет откат расходов, долей и долгов при поздней ошибке.
func TestTransactionAtomicity(t *testing.T) {
	database := setupTestDB(t)
	t.Cleanup(func() { teardownTestDB(t, database) })
	db := database.DB
	if err := db.Exec(`INSERT INTO users(id,is_dummy) VALUES(1,true),(2,true);
		INSERT INTO events(id,name) VALUES(1,'trip');
		INSERT INTO user_event(user_id,event_id) VALUES(1,1),(2,1);`).Error; err != nil {
		t.Fatal(err)
	}
	c, err := createTestContainer(t, &config.Config{}, gin.New(), db, http.DefaultClient)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	request := func(amount float64, valid bool) *service.TransactionRequest {
		portion := map[string]float64{"1": amount / 2, "2": amount / 2}
		if !valid {
			portion = map[string]float64{"1": 1, "2": 1}
		}
		return &service.TransactionRequest{Name: "expense", Type: "amount", FromUser: 1, Users: []int64{1, 2}, Amount: amount, Portion: portion}
	}
	assertRows := func(table string, want int64) {
		t.Helper()
		var count int64
		if err := db.Table(table).Count(&count).Error; err != nil {
			t.Fatal(err)
		}
		if count != want {
			t.Fatalf("%s: rows=%d want=%d", table, count, want)
		}
	}
	// Сценарий: ошибка калькулятора после INSERT не оставляет расхода.
	if _, err := c.TransactionService.CreateTransaction(ctx, 1, request(100, false)); err == nil {
		t.Fatal("invalid create accepted")
	}
	for _, table := range []string{"transactions", "transaction_shares", "debts"} {
		assertRows(table, 0)
	}
	created, err := c.TransactionService.CreateTransaction(ctx, 1, request(100, true))
	if err != nil {
		t.Fatal(err)
	}
	// Сценарий: ошибка после UPDATE и удаления прежних долей/долгов сохраняет прежнее состояние.
	if _, err := c.TransactionService.UpdateTransaction(ctx, created.ID, request(200, false)); err == nil {
		t.Fatal("invalid update accepted")
	}
	assertOriginal := func() {
		t.Helper()
		got, err := c.TransactionService.GetTransactionByID(ctx, created.ID)
		if err != nil {
			t.Fatal(err)
		}
		if got.Amount != 100 || len(got.Shares) != 2 || len(got.Debts) != 1 || got.Debts[0].Amount != 50 {
			t.Fatalf("original state lost: %+v", got)
		}
		for _, share := range got.Shares {
			if share.Value != 50 {
				t.Fatalf("share changed: %+v", share)
			}
		}
	}
	assertOriginal()
	// Ошибка INSERT долгов возникает уже после записи корректных долей.
	if err := db.Exec(`CREATE FUNCTION reject_debt() RETURNS trigger LANGUAGE plpgsql AS $$
		BEGIN RAISE EXCEPTION 'injected debt write failure'; END; $$;
		CREATE TRIGGER reject_debt BEFORE INSERT ON debts FOR EACH ROW EXECUTE FUNCTION reject_debt();`).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := c.TransactionService.CreateTransaction(ctx, 1, request(300, true)); err == nil {
		t.Fatal("database failure ignored on create")
	}
	assertRows("transactions", 1)
	assertRows("transaction_shares", 2)
	assertRows("debts", 1)
	if _, err := c.TransactionService.UpdateTransaction(ctx, created.ID, request(300, true)); err == nil {
		t.Fatal("database failure ignored on update")
	}
	assertOriginal()
}

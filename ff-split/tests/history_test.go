package tests

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/actor"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/history"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
	"gorm.io/gorm"
)

// TestHistoryExpenseLifecycle проверяет автора, снимки долей, удаления и атомарный отказ журнала.
func TestHistoryExpenseLifecycle(t *testing.T) {
	database, c := settlementFixture(t)
	ctx := actor.WithID(context.Background(), 101)
	request := &service.TransactionRequest{Name: "Дом", Type: "amount", FromUser: 1, Amount: 100.03, Users: []int64{2}, Portion: map[string]float64{"2": 100.03}}
	expense, err := c.TransactionService.CreateTransaction(ctx, 1, request)
	if err != nil {
		t.Fatal(err)
	}
	request.Name = "Дом и уборка"
	request.Amount = 120.05
	request.Portion["2"] = 120.05
	if _, err = c.TransactionService.UpdateTransaction(ctx, expense.ID, request); err != nil {
		t.Fatal(err)
	}
	rows, err := history.NewRepository(database.DB).List(ctx, 1, 0)
	if err != nil || len(rows) != 2 {
		t.Fatalf("rows=%+v err=%v", rows, err)
	}
	if rows[0].ActorID != 101 || rows[0].ActorName == "" || rows[0].BeforeValues["Сумма"] != "100.03 ₽" || rows[0].AfterValues["Сумма"] != "120.05 ₽" || rows[0].BeforeValues["Доли"] == rows[0].AfterValues["Доли"] {
		t.Fatalf("incorrect snapshot: %+v", rows[0])
	}
	if err = database.DB.Exec(`CREATE FUNCTION fail_history() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'history unavailable'; END $$; CREATE TRIGGER fail_history BEFORE INSERT ON event_history FOR EACH ROW EXECUTE FUNCTION fail_history();`).Error; err != nil {
		t.Fatal(err)
	}
	request.Amount = 150
	request.Portion["2"] = 150
	if _, err = c.TransactionService.UpdateTransaction(ctx, expense.ID, request); err == nil {
		t.Fatal("mutation succeeded without history")
	}
	actual, err := c.TransactionService.GetTransactionByID(ctx, expense.ID)
	if err != nil || actual.Amount != 120.05 {
		t.Fatalf("rollback=%+v %v", actual, err)
	}
	if err = database.DB.Exec("DROP TRIGGER fail_history ON event_history; DROP FUNCTION fail_history()").Error; err != nil {
		t.Fatal(err)
	}
	if err = c.TransactionService.DeleteTransaction(ctx, expense.ID); err != nil {
		t.Fatal(err)
	}
	rows, err = history.NewRepository(database.DB).List(ctx, 1, 0)
	if err != nil || len(rows) != 3 || rows[0].Kind != "expense.deleted" || rows[0].BeforeValues["Название"] != "Дом и уборка" || len(rows[0].AfterValues) != 0 {
		t.Fatalf("deleted snapshot: %+v %v", rows, err)
	}
}

// TestHistoryConcurrentEdits проверяет непрерывность before/after при двух одновременных редакторах.
func TestHistoryConcurrentEdits(t *testing.T) {
	database, c := settlementFixture(t)
	ctx := actor.WithID(context.Background(), 101)
	expense := settlementExpense(t, c, 1, 2, 100)
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for _, amount := range []float64{150, 200} {
		wg.Add(1)
		go func(amount float64) {
			defer wg.Done()
			_, err := c.TransactionService.UpdateTransaction(ctx, expense.ID, &service.TransactionRequest{Name: "Дом", Type: "amount", FromUser: 1, Amount: amount, Users: []int64{2}, Portion: map[string]float64{"2": amount}})
			errs <- err
		}(amount)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	rows, err := history.NewRepository(database.DB).List(ctx, 1, 0)
	if err != nil || len(rows) != 2 {
		t.Fatalf("history=%+v %v", rows, err)
	}
	if rows[0].BeforeValues["Сумма"] != rows[1].AfterValues["Сумма"] || rows[1].BeforeValues["Сумма"] != "100.00 ₽" {
		t.Fatalf("broken chain=%+v", rows)
	}
}

// TestHistoryMembershipAndPayments проверяет отсутствие вымышленных повторов и курсор страницы.
func TestHistoryMembershipAndPayments(t *testing.T) {
	database, c := settlementFixture(t)
	ctx := actor.WithID(context.Background(), 101)
	if err := c.UserService.AddUsersToEvent(ctx, []int64{1, 2}, 1); err != nil {
		t.Fatal(err)
	}
	settlementExpense(t, c, 1, 2, 100)
	req := service.SettlementRequest{FromUserID: 2, AmountCents: 1000, RequestID: uuid.NewString()}
	row, err := c.SettlementService.Receive(ctx, 1, 101, req)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = c.SettlementService.Receive(ctx, 1, 101, req); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err = c.SettlementService.Cancel(ctx, 1, 101, row.ID); err != nil {
			t.Fatal(err)
		}
	}
	rows, err := history.NewRepository(database.DB).List(ctx, 1, 0)
	if err != nil || len(rows) != 2 {
		t.Fatalf("duplicate history=%+v %v", rows, err)
	}
	for i := 0; i < 55; i++ {
		if err = database.DB.Transaction(func(tx *gorm.DB) error {
			return history.Record(ctx, tx, 1, "event.updated", 1, nil, map[string]string{"Название": fmt.Sprint(i)})
		}); err != nil {
			t.Fatal(err)
		}
	}
	rows, err = history.NewRepository(database.DB).List(ctx, 1, 0)
	if err != nil || len(rows) != 50 {
		t.Fatalf("page=%d err=%v", len(rows), err)
	}
	next, err := history.NewRepository(database.DB).List(ctx, 1, rows[49].ID)
	if err != nil || len(next) != 7 || next[0].ID >= rows[49].ID {
		t.Fatalf("next=%+v err=%v", next, err)
	}
}

// TestHistoryNamesakes фиксирует смену реальных ID даже при одинаковых отображаемых именах.
func TestHistoryNamesakes(t *testing.T) {
	database, c := settlementFixture(t)
	ctx := actor.WithID(context.Background(), 101)
	if err := database.DB.Exec("UPDATE users SET name_cashed='Саша'").Error; err != nil {
		t.Fatal(err)
	}
	request := &service.TransactionRequest{Name: "Ужин", Type: "amount", FromUser: 1, Amount: 10, Users: []int64{2}, Portion: map[string]float64{"2": 10}}
	expense, err := c.TransactionService.CreateTransaction(ctx, 1, request)
	if err != nil {
		t.Fatal(err)
	}
	request.FromUser = 3
	if _, err = c.TransactionService.UpdateTransaction(ctx, expense.ID, request); err != nil {
		t.Fatal(err)
	}
	request.Users = []int64{1}
	request.Portion = map[string]float64{"1": 10}
	if _, err = c.TransactionService.UpdateTransaction(ctx, expense.ID, request); err != nil {
		t.Fatal(err)
	}
	rows, err := history.NewRepository(database.DB).List(ctx, 1, 0)
	if err != nil || len(rows) != 3 {
		t.Fatalf("namesakes lost in history: %+v %v", rows, err)
	}
	if rows[0].BeforeValues["_shares"] == rows[0].AfterValues["_shares"] || rows[1].BeforeValues["_payer_id"] == rows[1].AfterValues["_payer_id"] {
		t.Fatalf("canonical identities missing: %+v", rows)
	}
}

// TestHistoryGuestAtomicity проверяет, что отказ журнала не оставляет созданного гостя вне события.
func TestHistoryGuestAtomicity(t *testing.T) {
	database, c := settlementFixture(t)
	ctx := actor.WithID(context.Background(), 101)
	if err := database.DB.Exec("SELECT setval(pg_get_serial_sequence('users','id'),(SELECT MAX(id) FROM users))").Error; err != nil {
		t.Fatal(err)
	}
	var before int64
	database.DB.Table("users").Count(&before)
	if err := database.DB.Exec(`CREATE FUNCTION fail_history() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'history unavailable'; END $$; CREATE TRIGGER fail_history BEFORE INSERT ON event_history FOR EACH ROW EXECUTE FUNCTION fail_history();`).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := c.UserService.CreateDummyUser(ctx, "Не должен сохраниться", 1); err == nil {
		t.Fatal("guest succeeded without history")
	}
	var after int64
	database.DB.Table("users").Count(&after)
	if before != after {
		t.Fatalf("orphan guest persisted: before=%d after=%d", before, after)
	}
}

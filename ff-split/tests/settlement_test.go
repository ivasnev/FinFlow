package tests

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"

	"github.com/google/uuid"
	domainerrors "github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/container"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
)

// settlementFixture создаёт настоящие сервисы с различающимися внешними и внутренними ID.
func settlementFixture(t *testing.T) (*TestDBContainer, *container.Container) {
	t.Helper()
	database, c := setupOptimizationTest(t)
	if err := database.DB.Exec("UPDATE users SET user_id=id*101,is_dummy=false").Error; err != nil {
		t.Fatal(err)
	}
	return database, c
}

// settlementExpense создаёт расход без участия плательщика в его потреблении.
func settlementExpense(t *testing.T, c *container.Container, payer, consumer int64, amount float64) *service.TransactionResponse {
	t.Helper()
	row, err := c.TransactionService.CreateTransaction(context.Background(), 1, &service.TransactionRequest{Name: "expense", Type: "amount", Amount: amount, FromUser: payer, Users: []int64{consumer}, Portion: map[string]float64{strconv.FormatInt(consumer, 10): amount}})
	if err != nil {
		t.Fatal(err)
	}
	return row
}

// assertSettlementBalance проверяет единый баланс из расходов и независимых возвратов.
func assertSettlementBalance(t *testing.T, c *container.Container, id int64, want float64) {
	t.Helper()
	got, err := c.EventService.GetBalanceByEventID(context.Background(), id, 1)
	if err != nil || got != want {
		t.Fatalf("balance user=%d got=%v want=%v error=%v", id, got, want, err)
	}
}

// TestSettlementIdempotencyAndCancellation проверяет денежный остаток, безопасный повтор и сохранение отменённой записи.
func TestSettlementIdempotencyAndCancellation(t *testing.T) {
	database, c := settlementFixture(t)
	ctx := context.Background()
	expense := settlementExpense(t, c, 1, 2, 100.03)
	req := service.SettlementRequest{FromUserID: 2, AmountCents: 10003, RequestID: uuid.NewString()}
	row, err := c.SettlementService.Receive(ctx, 1, 101, req)
	if err != nil {
		t.Fatal(err)
	}
	if row.CreatedBy != 1 || row.ToUserID != 1 || row.AmountCents != 10003 || row.CreatedAt.IsZero() {
		t.Fatalf("record=%+v", row)
	}
	assertSettlementBalance(t, c, 1, 0)
	assertSettlementBalance(t, c, 2, 0)
	var revision int64
	database.DB.Table("events").Select("debt_revision").Where("id=1").Scan(&revision)
	again, err := c.SettlementService.Receive(ctx, 1, 101, req)
	if err != nil || again.ID != row.ID {
		t.Fatalf("retry=%+v error=%v", again, err)
	}
	var after int64
	database.DB.Table("events").Select("debt_revision").Where("id=1").Scan(&after)
	if after != revision {
		t.Fatal("retry changed revision")
	}
	changed := req
	changed.AmountCents--
	var conflict *domainerrors.AlreadyExistsError
	if _, err := c.SettlementService.Receive(ctx, 1, 101, changed); !errors.As(err, &conflict) {
		t.Fatalf("key reuse: %v", err)
	}
	debts, err := c.TransactionService.GetDebtsByEventID(ctx, 1, nil)
	if err != nil || len(debts) != 0 {
		t.Fatalf("residual=%+v %v", debts, err)
	}
	optimized, err := c.TransactionService.OptimizeDebts(ctx, 1)
	if err != nil || len(optimized) != 0 {
		t.Fatalf("optimized=%+v %v", optimized, err)
	}
	original, err := c.TransactionService.GetTransactionByID(ctx, expense.ID)
	if err != nil || original.Amount != 100.03 {
		t.Fatalf("expense was changed: %+v %v", original, err)
	}
	if err := c.SettlementService.Cancel(ctx, 1, 202, row.ID); !errors.Is(err, domainerrors.ErrForbidden) {
		t.Fatalf("other actor: %v", err)
	}
	if err := c.SettlementService.Cancel(ctx, 2, 101, row.ID); err == nil {
		t.Fatal("cross-event cancellation succeeded")
	}
	if err := c.SettlementService.Cancel(ctx, 1, 101, row.ID); err != nil {
		t.Fatal(err)
	}
	if err := c.SettlementService.Cancel(ctx, 1, 101, row.ID); err != nil {
		t.Fatal(err)
	}
	database.DB.Table("events").Select("debt_revision").Where("id=1").Scan(&after)
	if after != revision+1 {
		t.Fatalf("cancel revision=%d want=%d", after, revision+1)
	}
	canceled, err := c.SettlementService.Receive(ctx, 1, 101, req)
	if err != nil || canceled.CanceledAt == nil {
		t.Fatalf("canceled retry=%+v %v", canceled, err)
	}
	assertSettlementBalance(t, c, 1, 100.03)
	rows, err := c.SettlementService.List(ctx, 1, 0)
	if err != nil || len(rows) != 1 || rows[0].CanceledAt == nil {
		t.Fatalf("history=%+v %v", rows, err)
	}
}

// TestConcurrentSettlementsCannotOverpay сериализует конкурирующие подтверждения без превышения остатка.
func TestConcurrentSettlementsCannotOverpay(t *testing.T) {
	database, c := settlementFixture(t)
	settlementExpense(t, c, 1, 2, 100.03)
	var group sync.WaitGroup
	success := make(chan bool, 12)
	for i := 0; i < 12; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			_, err := c.SettlementService.Receive(context.Background(), 1, 101, service.SettlementRequest{FromUserID: 2, AmountCents: 2000, RequestID: uuid.NewString()})
			success <- err == nil
		}()
	}
	group.Wait()
	close(success)
	count := 0
	for ok := range success {
		if ok {
			count++
		}
	}
	if count != 5 {
		t.Fatalf("successful payments=%d want5", count)
	}
	assertSettlementBalance(t, c, 1, 0.03)
	var rows int64
	database.DB.Table("settlements").Count(&rows)
	if rows != 5 {
		t.Fatalf("ledger rows=%d", rows)
	}
}

// TestConcurrentSettlementRetry создаёт одну запись при одновременной доставке одинакового ключа.
func TestConcurrentSettlementRetry(t *testing.T) {
	database, c := settlementFixture(t)
	settlementExpense(t, c, 1, 2, 100)
	req := service.SettlementRequest{FromUserID: 2, AmountCents: 10000, RequestID: uuid.NewString()}
	var group sync.WaitGroup
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			_, err := c.SettlementService.Receive(context.Background(), 1, 101, req)
			errs <- err
		}()
	}
	group.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	var rows int64
	database.DB.Table("settlements").Count(&rows)
	if rows != 1 {
		t.Fatalf("ledger rows=%d", rows)
	}
}

// TestSettlementSurvivesExpenseChanges сохраняет факт возврата после уменьшения и удаления расхода.
func TestSettlementSurvivesExpenseChanges(t *testing.T) {
	_, c := settlementFixture(t)
	ctx := context.Background()
	expense := settlementExpense(t, c, 1, 2, 100)
	row, err := c.SettlementService.Receive(ctx, 1, 101, service.SettlementRequest{FromUserID: 2, AmountCents: 10000, RequestID: uuid.NewString()})
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.TransactionService.UpdateTransaction(ctx, expense.ID, &service.TransactionRequest{Name: "corrected", Type: "amount", Amount: 60, FromUser: 1, Users: []int64{2}, Portion: map[string]float64{"2": 60}})
	if err != nil {
		t.Fatal(err)
	}
	assertSettlementBalance(t, c, 1, -40)
	assertSettlementBalance(t, c, 2, 40)
	if err := c.TransactionService.DeleteTransaction(ctx, expense.ID); err != nil {
		t.Fatal(err)
	}
	assertSettlementBalance(t, c, 1, -100)
	assertSettlementBalance(t, c, 2, 100)
	debts, err := c.TransactionService.GetDebtsByEventID(ctx, 1, nil)
	if err != nil || len(debts) != 1 || debts[0].FromUserID != 1 || debts[0].Amount != 100 {
		t.Fatalf("reverse debt=%+v %v", debts, err)
	}
	if err := c.SettlementService.Cancel(ctx, 1, 101, row.ID); err != nil {
		t.Fatal(err)
	}
	assertSettlementBalance(t, c, 1, 0)
}

// TestSettlementThroughIntermediateUser допускает показанные переводы через участника с чистым балансом ноль.
func TestSettlementThroughIntermediateUser(t *testing.T) {
	_, c := settlementFixture(t)
	ctx := context.Background()
	settlementExpense(t, c, 3, 2, 100)
	settlementExpense(t, c, 1, 3, 100)
	direct, err := c.SettlementService.Receive(ctx, 1, 101, service.SettlementRequest{FromUserID: 2, AmountCents: 10000, RequestID: uuid.NewString()})
	if err != nil {
		t.Fatal(err)
	}
	debts, err := c.TransactionService.GetDebtsByEventID(ctx, 1, nil)
	if err != nil || len(debts) != 0 {
		t.Fatalf("fictional cycle=%+v %v", debts, err)
	}
	if err := c.SettlementService.Cancel(ctx, 1, 101, direct.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := c.SettlementService.Receive(ctx, 1, 101, service.SettlementRequest{FromUserID: 3, AmountCents: 10000, RequestID: uuid.NewString()}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.SettlementService.Receive(ctx, 1, 303, service.SettlementRequest{FromUserID: 2, AmountCents: 10000, RequestID: uuid.NewString()}); err != nil {
		t.Fatal(err)
	}
	for _, id := range []int64{1, 2, 3} {
		assertSettlementBalance(t, c, id, 0)
	}
}

// TestSettlementRollbackKeepsLedgerAndRevisionAtomic откатывает вставку при ошибке смены версии.
func TestSettlementRollbackKeepsLedgerAndRevisionAtomic(t *testing.T) {
	database, c := settlementFixture(t)
	settlementExpense(t, c, 1, 2, 100)
	if err := database.DB.Exec(`CREATE FUNCTION fail_revision() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'revision failure'; END $$;
 CREATE TRIGGER reject_revision BEFORE UPDATE ON events FOR EACH ROW EXECUTE FUNCTION fail_revision();`).Error; err != nil {
		t.Fatal(err)
	}
	_, err := c.SettlementService.Receive(context.Background(), 1, 101, service.SettlementRequest{FromUserID: 2, AmountCents: 1000, RequestID: uuid.NewString()})
	if err == nil {
		t.Fatal("expected rollback")
	}
	var count int64
	database.DB.Table("settlements").Count(&count)
	if count != 0 {
		t.Fatalf("partial ledger insert: %d", count)
	}
	assertSettlementBalance(t, c, 1, 100)
}

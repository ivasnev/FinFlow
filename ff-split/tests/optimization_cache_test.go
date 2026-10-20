package tests

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/config"
	domainerrors "github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/container"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
	"gorm.io/gorm"
)

// setupOptimizationTest создаёт изолированные мероприятия и настоящие сервисы PostgreSQL.
func setupOptimizationTest(t *testing.T) (*TestDBContainer, *container.Container) {
	t.Helper()
	database := setupTestDB(t)
	t.Cleanup(func() { teardownTestDB(t, database) })
	if err := database.DB.Exec(`INSERT INTO users(id,is_dummy) VALUES(1,true),(2,true),(3,true);
		INSERT INTO events(id,name) VALUES(1,'trip'),(2,'empty');
		INSERT INTO user_event(user_id,event_id) VALUES(1,1),(2,1),(3,1);`).Error; err != nil {
		t.Fatal(err)
	}
	c, err := createTestContainer(t, &config.Config{}, gin.New(), database.DB, http.DefaultClient)
	if err != nil {
		t.Fatal(err)
	}
	return database, c
}

// optimizationRequest создаёт расход с одинаковыми долями двух участников.
func optimizationRequest(amount float64) *service.TransactionRequest {
	return &service.TransactionRequest{Name: "expense", Type: "amount", FromUser: 1, Users: []int64{1, 2}, Amount: amount, Portion: map[string]float64{"1": amount / 2, "2": amount / 2}}
}

// TestOptimizationCacheRevision проверяет отсутствие расчёта при записи и повторное использование пустого кэша.
func TestOptimizationCacheRevision(t *testing.T) {
	database, c := setupOptimizationTest(t)
	var calculations atomic.Int64
	if err := database.DB.Callback().Query().Before("gorm:query").Register("count_debt_reads", func(tx *gorm.DB) {
		if tx.Statement.Table == "debts" {
			calculations.Add(1)
		}
	}); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	expense, err := c.TransactionService.CreateTransaction(ctx, 1, optimizationRequest(100))
	if err != nil {
		t.Fatal(err)
	}
	if calculations.Load() != 0 {
		t.Fatal("create triggered optimization")
	}
	assertDebt := func(want float64) {
		t.Helper()
		got, err := c.TransactionService.GetOptimizedDebtsByEventID(ctx, 1, nil)
		if err != nil {
			t.Fatal(err)
		}
		if want == 0 {
			if len(got) != 0 {
				t.Fatalf("expected empty result: %+v", got)
			}
		} else if len(got) != 1 || got[0].Amount != want {
			t.Fatalf("want=%.2f result=%+v", want, got)
		}
	}
	assertDebt(50)
	for i := 0; i < 10; i++ {
		assertDebt(50)
	}
	if calculations.Load() != 1 {
		t.Fatalf("cached reads recomputed: %d", calculations.Load())
	}
	for i := 0; i < 5; i++ {
		got, err := c.TransactionService.GetOptimizedDebtsByUserID(ctx, 1, 3)
		if err != nil || len(got) != 0 {
			t.Fatalf("user without debts: %+v %v", got, err)
		}
	}
	if calculations.Load() != 1 {
		t.Fatal("user without debts triggered recomputation")
	}
	if _, err := c.TransactionService.UpdateTransaction(ctx, expense.ID, optimizationRequest(200)); err != nil {
		t.Fatal(err)
	}
	if calculations.Load() != 1 {
		t.Fatal("update triggered optimization")
	}
	var oldAmount float64
	if err := database.DB.Table("optimized_debts").Select("amount").Where("event_id = ?", 1).Scan(&oldAmount).Error; err != nil {
		t.Fatal(err)
	}
	if oldAmount != 50 {
		t.Fatal("write recalculated or replaced cached rows")
	}
	assertDebt(100)
	if calculations.Load() != 2 {
		t.Fatal("new revision was not computed once")
	}
	if err := c.TransactionService.DeleteTransaction(ctx, expense.ID); err != nil {
		t.Fatal(err)
	}
	if calculations.Load() != 2 {
		t.Fatal("delete triggered optimization")
	}
	for i := 0; i < 10; i++ {
		assertDebt(0)
	}
	if calculations.Load() != 3 {
		t.Fatalf("empty result was recomputed: %d", calculations.Load())
	}
}

// TestConcurrentOptimization проверяет объединение запросов, межпроцессный замок и отказ устаревшей публикации.
func TestConcurrentOptimization(t *testing.T) {
	database, c := setupOptimizationTest(t)
	ctx := context.Background()
	expense, err := c.TransactionService.CreateTransaction(ctx, 1, optimizationRequest(100))
	if err != nil {
		t.Fatal(err)
	}
	other, err := createTestContainer(t, &config.Config{}, gin.New(), database.DB, http.DefaultClient)
	if err != nil {
		t.Fatal(err)
	}
	var calculations atomic.Int64
	var workers sync.WaitGroup
	type answer struct {
		debts []service.OptimizedDebtDTO
		err   error
	}
	launch := func(callCtx context.Context) <-chan answer {
		ch := make(chan answer, 1)
		workers.Add(1)
		go func() {
			defer workers.Done()
			debts, err := c.TransactionService.GetOptimizedDebtsByEventID(callCtx, 1, nil)
			ch <- answer{debts, err}
		}()
		return ch
	}
	wait := func(ch <-chan answer) answer {
		t.Helper()
		select {
		case result := <-ch:
			return result
		case <-time.After(10 * time.Second):
			t.Fatal("optimization did not finish")
			return answer{}
		}
	}
	gate := func(name string) (<-chan struct{}, func()) {
		started := make(chan struct{})
		release := make(chan struct{})
		var once, releaseOnce sync.Once
		unblock := func() { releaseOnce.Do(func() { close(release) }) }
		t.Cleanup(func() { unblock(); workers.Wait() })
		if err := database.DB.Callback().Query().After("gorm:query").Register(name, func(tx *gorm.DB) {
			if tx.Statement.Table == "debts" {
				calculations.Add(1)
				once.Do(func() {
					close(started)
					select {
					case <-release:
					case <-tx.Statement.Context.Done():
					}
				})
			}
		}); err != nil {
			t.Fatal(err)
		}
		return started, unblock
	}
	started, release := gate("first_optimization")
	leader := launch(ctx)
	select {
	case <-started:
	case <-time.After(10 * time.Second):
		t.Fatal("debt snapshot was not read")
	}
	followers := make([]<-chan answer, 20)
	for i := range followers {
		followers[i] = launch(ctx)
	}
	if _, err := other.TransactionService.GetOptimizedDebtsByEventID(ctx, 1, nil); !errors.Is(err, domainerrors.ErrOptimizationBusy) {
		t.Fatalf("parallel instance was not rejected: %v", err)
	}
	cancelCtx, cancel := context.WithCancel(ctx)
	cancelled := launch(cancelCtx)
	cancel()
	if result := wait(cancelled); !errors.Is(result.err, context.Canceled) {
		t.Fatalf("cancelled waiter: %v", result.err)
	}
	release()
	for _, ch := range append(followers, leader) {
		result := wait(ch)
		if result.err != nil || len(result.debts) != 1 || result.debts[0].Amount != 50 {
			t.Fatalf("shared result: %+v %v", result.debts, result.err)
		}
	}
	if calculations.Load() != 1 {
		t.Fatalf("parallel requests computed %d times", calculations.Load())
	}
	if err := database.DB.Callback().Query().Remove("first_optimization"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.TransactionService.UpdateTransaction(ctx, expense.ID, optimizationRequest(200)); err != nil {
		t.Fatal(err)
	}
	started, release = gate("stale_optimization")
	stale := launch(ctx)
	select {
	case <-started:
	case <-time.After(10 * time.Second):
		t.Fatal("stale snapshot was not read")
	}
	// Запись расхода проходит, пока расчёт удерживает отдельный advisory lock.
	if _, err := c.TransactionService.UpdateTransaction(ctx, expense.ID, optimizationRequest(300)); err != nil {
		t.Fatal(err)
	}
	release()
	if result := wait(stale); !errors.Is(result.err, domainerrors.ErrDebtRevisionChanged) {
		t.Fatalf("stale result was published: %+v %v", result.debts, result.err)
	}
	fresh := wait(launch(ctx))
	if fresh.err != nil || len(fresh.debts) != 1 || fresh.debts[0].Amount != 150 {
		t.Fatalf("fresh result: %+v %v", fresh.debts, fresh.err)
	}
}

// TestPostCommitOptimizationRead не позволяет новому запросу присоединиться к старому чтению готового кэша.
func TestPostCommitOptimizationRead(t *testing.T) {
	database, c := setupOptimizationTest(t)
	ctx := context.Background()
	expense, err := c.TransactionService.CreateTransaction(ctx, 1, optimizationRequest(100))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.TransactionService.OptimizeDebts(ctx, 1); err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{})
	release := make(chan struct{})
	var once, releaseOnce sync.Once
	var workers sync.WaitGroup
	unblock := func() { releaseOnce.Do(func() { close(release) }) }
	t.Cleanup(func() { unblock(); workers.Wait() })
	if err := database.DB.Callback().Query().After("gorm:query").Register("hold_old_cache", func(tx *gorm.DB) {
		if tx.Statement.Table == "optimized_debts" {
			once.Do(func() { close(started); <-release })
		}
	}); err != nil {
		t.Fatal(err)
	}
	type answer struct {
		debts []service.OptimizedDebtDTO
		err   error
	}
	launch := func() <-chan answer {
		ch := make(chan answer, 1)
		workers.Add(1)
		go func() {
			defer workers.Done()
			debts, err := c.TransactionService.GetOptimizedDebtsByEventID(ctx, 1, nil)
			ch <- answer{debts, err}
		}()
		return ch
	}
	old := launch()
	select {
	case <-started:
	case <-time.After(10 * time.Second):
		t.Fatal("old cache read did not start")
	}
	if _, err := c.TransactionService.UpdateTransaction(ctx, expense.ID, optimizationRequest(200)); err != nil {
		t.Fatal(err)
	}
	// Этот запрос начался после commit и должен завершиться с новой суммой, пока прежний читатель остановлен.
	select {
	case fresh := <-launch():
		if fresh.err != nil || len(fresh.debts) != 1 || fresh.debts[0].Amount != 100 {
			t.Fatalf("post-commit read: %+v %v", fresh.debts, fresh.err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("post-commit read joined the stale flight")
	}
	unblock()
	select {
	case result := <-old:
		if result.err != nil {
			t.Fatal(result.err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("old read did not finish")
	}
}

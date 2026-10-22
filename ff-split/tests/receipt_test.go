package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/actor"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	archiverepo "github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/archive"
	receiptrepo "github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/receipt"
	archiveservice "github.com/ivasnev/FinFlow/ff-split/internal/service/archive"
	receiptservice "github.com/ivasnev/FinFlow/ff-split/internal/service/receipt"
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
	"time"
)

// receiptMigration проверяет откат и повторное применение миграции чеков.
func receiptMigration(t *testing.T, database *TestDBContainer) {
	t.Helper()
	if database.DB.Migrator().HasTable("receipt_image_reads") {
		if e := database.DB.Exec("DROP TABLE receipt_image_reads").Error; e != nil {
			t.Fatal(e)
		}
	}
	if database.DB.Migrator().HasTable("receipts") {
		if database.DB.Migrator().HasTable("receipt_admissions") {
			down8, e := os.ReadFile("../internal/repository/migrations/000008_receipt_actions.down.sql")
			if e != nil {
				t.Fatal(e)
			}
			if e = database.DB.Exec(string(down8)).Error; e != nil {
				t.Fatal(e)
			}
		}
		down, err := os.ReadFile("../internal/repository/migrations/000007_receipts.down.sql")
		if err != nil {
			t.Fatal(err)
		}
		if err = database.DB.Exec(string(down)).Error; err != nil {
			t.Fatal(err)
		}
	}
	raw, err := os.ReadFile("../internal/repository/migrations/000007_receipts.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	if err = database.DB.Exec(string(raw)).Error; err != nil {
		t.Fatal(err)
	}
	raw8, e := os.ReadFile("../internal/repository/migrations/000008_receipt_actions.up.sql")
	if e != nil {
		t.Fatal(e)
	}
	if e = database.DB.Exec(string(raw8)).Error; e != nil {
		t.Fatal(e)
	}
	raw9, e := os.ReadFile("../internal/repository/migrations/000009_receipt_read_budget.up.sql")
	if e != nil {
		t.Fatal(e)
	}
	if e = database.DB.Exec(string(raw9)).Error; e != nil {
		t.Fatal(e)
	}
}

// seedReceipt создаёт синтетическую задачу заданного события.
func seedReceipt(t *testing.T, database *TestDBContainer, event int64, status string) string {
	t.Helper()
	id := uuid.NewString()
	if err := database.DB.Exec("INSERT INTO receipts(id,event_id,created_by,image_key,image_sha256,content_type,status,result) VALUES(?,?,101,?,?,?,?,'{\"items\":[],\"total_cents\":null,\"warnings\":[],\"raw_text\":\"\",\"engine\":\"test\"}'::jsonb)", id, event, "receipts/"+id, fmt.Sprintf("%064s", id), "image/png", status).Error; err != nil {
		t.Fatal(err)
	}
	return id
}

// TestReceiptCanonicalRetry проверяет повтор подтверждения после канонизации черновика.
func TestReceiptCanonicalRetry(t *testing.T) {
	database, c := settlementFixture(t)
	// receiptMigration проверяет откат и повторное применение миграции чеков.
	receiptMigration(t, database)
	id := seedReceipt(t, database, 1, "ready")
	ctx := actor.WithID(context.Background(), 101)
	s := receiptservice.NewService(receiptrepo.NewRepository(database.DB), nil, nil, c.TransactionService)
	input := func() models.ReceiptDraft {
		var d models.ReceiptDraft
		if err := json.Unmarshal([]byte(`{"revision":1,"payer_id":1,"items":[{"name":"  Товар  ","amount_cents":101,"users":[3,2],"selected":true}]}`), &d); err != nil {
			t.Fatal(err)
		}
		return d
	}
	first, err := s.Edit(ctx, 1, id, input(), true)
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.Edit(ctx, 1, id, input(), true)
	if err != nil {
		t.Fatalf("same original confirmation rejected: %v", err)
	}
	if len(first.TransactionIDs) != 1 || fmt.Sprint(first.TransactionIDs) != fmt.Sprint(second.TransactionIDs) {
		t.Fatal("retry returned different transactions")
	}
	var count int64
	database.DB.Table("transactions").Count(&count)
	if count != 1 {
		t.Fatal("duplicate transactions")
	}
}

// TestReceiptBatchRollback проверяет откат всего пакета при ошибке одной позиции.
func TestReceiptBatchRollback(t *testing.T) {
	database, c := settlementFixture(t)
	// receiptMigration проверяет откат и повторное применение миграции чеков.
	receiptMigration(t, database)
	id := seedReceipt(t, database, 1, "ready")
	ctx := actor.WithID(context.Background(), 101)
	s := receiptservice.NewService(receiptrepo.NewRepository(database.DB), nil, nil, c.TransactionService)
	draft := models.ReceiptDraft{Revision: 1, PayerID: 1, Items: []models.ReceiptDraftItem{{Name: "First", AmountCents: 100, Users: []int64{2}, Selected: true}, {Name: "Bad", AmountCents: 100, Users: []int64{999}, Selected: true}}}
	if _, err := s.Edit(ctx, 1, id, draft, true); err == nil {
		t.Fatal("bad participant accepted")
	}
	for _, table := range []string{"transactions", "transaction_shares", "debts", "event_history"} {
		var count int64
		database.DB.Table(table).Count(&count)
		if count != 0 {
			t.Fatalf("partial write in %s", table)
		}
	}
	row, err := s.Get(ctx, 1, id)
	if err != nil || row.Status != "ready" || row.Revision != 1 || len(row.TransactionIDs) != 0 {
		t.Fatalf("partial confirmation %+v %v", row, err)
	}
}

// TestReceiptParkedQuota проверяет исключение архивных задач из активной квоты.
func TestReceiptParkedQuota(t *testing.T) {
	database, _ := settlementFixture(t)
	// receiptMigration проверяет откат и повторное применение миграции чеков.
	receiptMigration(t, database)
	for i := 0; i < 3; i++ {
		seedReceipt(t, database, 1, "queued")
	}
	ctx := actor.WithID(context.Background(), 101)
	archive := archiveservice.NewService(archiverepo.NewRepository(database.DB))
	if err := archive.Archive(ctx, 1); err != nil {
		t.Fatal(err)
	}
	repo := receiptrepo.NewRepository(database.DB)
	job, err := repo.Claim(ctx)
	if err != nil || job != nil {
		t.Fatalf("archived task claimed %+v %v", job, err)
	}
	if err := database.DB.Exec("INSERT INTO user_event(user_id,event_id) VALUES(1,2)").Error; err != nil {
		t.Fatal(err)
	}
	id := uuid.NewString()
	err = repo.InEvent(ctx, 2, func(ctx context.Context) error {
		_, err := repo.Create(ctx, &models.Receipt{ID: id, EventID: 2, CreatedBy: 101, ImageKey: "receipts/" + id, ImageSHA256: fmt.Sprintf("%064s", id), ContentType: "image/png"})
		return err
	})
	if err != nil {
		t.Fatalf("parked receipts consumed active quota: %v", err)
	}
	if err = archive.Restore(ctx, 1); err != nil {
		t.Fatal(err)
	}
	job, err = repo.Claim(ctx)
	if err != nil || job == nil || job.EventID != 1 {
		t.Fatalf("restored queue did not resume %+v %v", job, err)
	}
}

// receiptImageStore считает записи в хранилище для проверки предварительного допуска.
type receiptImageStore struct {
	puts     int
	deadline bool
}

// Put фиксирует обращение к хранилищу.
func (s *receiptImageStore) Put(ctx context.Context, _ string, _ []byte, _ string) error {
	s.puts++
	if d, ok := ctx.Deadline(); ok {
		s.deadline = time.Until(d) > 0 && time.Until(d) <= 30*time.Second
	}
	return nil
}

// Read реализует неиспользуемое чтение тестового хранилища.
func (s *receiptImageStore) Read(context.Context, string) ([]byte, error) { return nil, nil }

// Delete реализует неиспользуемое удаление тестового хранилища.
func (s *receiptImageStore) Delete(context.Context, string) error { return nil }

// TestReceiptAdmissionBeforeStorage проверяет отказ до загрузки в S3 и дедупликацию.
func TestReceiptAdmissionBeforeStorage(t *testing.T) {
	database, c := settlementFixture(t)
	// receiptMigration проверяет откат и повторное применение миграции чеков.
	receiptMigration(t, database)
	ctx := actor.WithID(context.Background(), 101)
	store := &receiptImageStore{}
	s := receiptservice.NewService(receiptrepo.NewRepository(database.DB), store, nil, c.TransactionService)
	pngData := func(n uint8) []byte {
		img := image.NewRGBA(image.Rect(0, 0, 1, 1))
		img.Set(0, 0, color.RGBA{R: n, A: 255})
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			t.Fatal(err)
		}
		return buf.Bytes()
	}
	first, err := s.Upload(ctx, 1, pngData(1))
	if err != nil {
		t.Fatal(err)
	}
	again, err := s.Upload(ctx, 1, pngData(1))
	if err != nil || again.ID != first.ID || store.puts != 1 {
		t.Fatalf("duplicate wrote an object puts=%d err=%v", store.puts, err)
	}
	for _, n := range []uint8{2, 3} {
		if _, err = s.Upload(ctx, 1, pngData(n)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = s.Upload(ctx, 1, pngData(4)); err == nil {
		t.Fatal("fourth pending accepted")
	}
	if store.puts != 3 {
		t.Fatalf("over-quota request wrote object puts=%d", store.puts)
	}
	if !store.deadline {
		t.Fatal("Put lacks 30s deadline")
	}
}

// TestReceiptLeaseFence проверяет отбрасывание результата потерявшего аренду обработчика.
func TestReceiptLeaseFence(t *testing.T) {
	database, _ := settlementFixture(t)
	// receiptMigration проверяет откат и повторное применение миграции чеков.
	receiptMigration(t, database)
	id := seedReceipt(t, database, 1, "queued")
	repo := receiptrepo.NewRepository(database.DB)
	ctx := context.Background()
	old, err := repo.Claim(ctx)
	if err != nil || old == nil {
		t.Fatalf("claim %+v %v", old, err)
	}
	if err = database.DB.Exec("UPDATE receipts SET lease_until=clock_timestamp()-interval '1 second' WHERE id=?", id).Error; err != nil {
		t.Fatal(err)
	}
	current, err := repo.Claim(ctx)
	if err != nil || current == nil || current.LeaseToken == old.LeaseToken {
		t.Fatalf("reclaim %+v %v", current, err)
	}
	result := &models.ReceiptResult{Items: []models.ReceiptItem{{Name: "Old", AmountCents: 1, Confidence: 1}}}
	if err = repo.Complete(ctx, old, result, nil); err != nil {
		t.Fatal(err)
	}
	row, err := repo.Get(ctx, 1, id, false)
	if err != nil || row.Status != "processing" || row.LeaseToken != current.LeaseToken {
		t.Fatalf("stale result published %+v %v", row, err)
	}
	result.Items[0].Name = "Current"
	if err = repo.Complete(ctx, current, result, nil); err != nil {
		t.Fatal(err)
	}
	row, err = repo.Get(ctx, 1, id, false)
	if err != nil || row.Status != "ready" || row.Result.Items[0].Name != "Current" {
		t.Fatalf("current result lost %+v %v", row, err)
	}
}

package tests

import (
	"context"
	"errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/actor"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	receiptrepo "github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/receipt"
	receiptservice "github.com/ivasnev/FinFlow/ff-split/internal/service/receipt"
	"gorm.io/gorm"
	"sync"
	"testing"
)

// TestReceiptRetryActions проверяет идемпотентность повтора, сохранение правок и защиту аренды.
func TestReceiptRetryActions(t *testing.T) {
	database, c := settlementFixture(t)
	receiptMigration(t, database)
	id := seedReceipt(t, database, 1, "processing")
	repo := receiptrepo.NewRepository(database.DB)
	ctx := actor.WithID(context.Background(), 101)
	// Завершаем предыдущую попытку и сохраняем исправленное человеком название.
	if err := database.DB.Exec("UPDATE receipts SET lease_token='aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa' WHERE id=?", id).Error; err != nil {
		t.Fatal(err)
	}
	old, _ := repo.Get(ctx, 1, id, false)
	if err := repo.Complete(ctx, old, &models.ReceiptResult{Items: []models.ReceiptItem{{Name: "OCR", AmountCents: 100}}, Warnings: []string{}}, nil); err != nil {
		t.Fatal(err)
	}
	store := &receiptHTTPStore{objects: map[string][]byte{"receipts/" + id: []byte("synthetic")}}
	service := receiptservice.NewService(repo, store, nil, c.TransactionService)
	draft := models.ReceiptDraft{Revision: 2, PayerID: 1, Items: []models.ReceiptDraftItem{{Name: "Ручная правка", AmountCents: 150, Users: []int64{1}, Selected: true}}}
	saved, err := service.Edit(ctx, 1, id, draft, false)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, e := service.Retry(ctx, 1, id, saved.Revision); errs <- e }()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		if e != nil {
			t.Fatal(e)
		}
	}
	var count int64
	database.DB.Table("receipt_admissions").Where("receipt_id=?", id).Count(&count)
	if count != 2 {
		t.Fatalf("retry consumed quota twice: %d", count)
	}
	job, err := repo.Claim(ctx)
	if err != nil || job == nil {
		t.Fatalf("claim: %v", err)
	}
	if err = repo.Complete(ctx, old, &models.ReceiptResult{Engine: "stale"}, nil); err != nil {
		t.Fatal(err)
	}
	if err = repo.Complete(ctx, job, &models.ReceiptResult{Items: []models.ReceiptItem{{Name: "New OCR", AmountCents: 200}}, Warnings: []string{}}, nil); err != nil {
		t.Fatal(err)
	}
	result, err := service.Retry(ctx, 1, id, saved.Revision)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "ready" || result.Draft.Items[0].Name != "Ручная правка" || result.Result.Items[0].Name != "New OCR" {
		t.Fatalf("retry lost draft or replayed: %+v", result)
	}
	if _, err = service.Retry(ctx, 1, id, 1); err == nil {
		t.Fatal("stale revision accepted")
	}
	database.DB.Table("receipt_admissions").Where("receipt_id=?", id).Count(&count)
	if count != 2 {
		t.Fatal("replay consumed quota")
	}
}

// TestReceiptRemoveActions проверяет отмену обработки без воскрешения и обхода суточной квоты.
func TestReceiptRemoveActions(t *testing.T) {
	database, c := settlementFixture(t)
	receiptMigration(t, database)
	ctx := actor.WithID(context.Background(), 101)
	id := seedReceipt(t, database, 1, "queued")
	repo := receiptrepo.NewRepository(database.DB)
	service := receiptservice.NewService(repo, nil, nil, c.TransactionService)
	job, err := repo.Claim(ctx)
	if err != nil || job == nil {
		t.Fatal(err)
	}
	if err = service.Remove(ctx, 1, id, job.Revision+1); err == nil {
		t.Fatal("stale delete accepted")
	}
	if err = service.Remove(ctx, 1, id, job.Revision); err != nil {
		t.Fatal(err)
	}
	if err = repo.Complete(ctx, job, &models.ReceiptResult{Engine: "stale"}, nil); err != nil {
		t.Fatal(err)
	}
	if _, err = service.Get(ctx, 1, id); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("deleted receipt visible: %v", err)
	}
	list, err := service.List(ctx, 1, 0)
	if err != nil || len(list.Receipts) != 0 {
		t.Fatal("deleted receipt listed")
	}
	if job, err = repo.Claim(ctx); err != nil || job != nil {
		t.Fatal("deleted receipt claimed")
	}
	var count int64
	database.DB.Table("receipt_admissions").Where("receipt_id=?", id).Count(&count)
	if count != 1 {
		t.Fatal("delete freed daily quota")
	}
	database.DB.Table("receipt_image_gc").Where("image_key=?", "receipts/"+id).Count(&count)
	if count != 1 {
		t.Fatal("image deletion missing from outbox")
	}
	// Та же фотография может быть загружена снова, при этом допуск списывается заново.
	database.DB.Exec("INSERT INTO receipt_admissions(receipt_id,actor_id) SELECT gen_random_uuid(),101 FROM generate_series(1,49)")
	next := &models.Receipt{ID: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", EventID: 1, CreatedBy: 101, ImageKey: "new", ImageSHA256: "new", ContentType: "image/png"}
	err = repo.InEvent(ctx, 1, func(ctx context.Context) error { _, e := repo.Create(ctx, next); return e })
	if err == nil {
		t.Fatal("deleted receipt freed admission budget")
	}
}

// TestReceiptActionBoundaries проверяет архив, чужое событие, отсутствующее фото и подтверждённые расходы.
func TestReceiptActionBoundaries(t *testing.T) {
	database, c := settlementFixture(t)
	receiptMigration(t, database)
	ctx := actor.WithID(context.Background(), 101)
	id := seedReceipt(t, database, 1, "failed")
	repo := receiptrepo.NewRepository(database.DB)
	store := &receiptHTTPStore{objects: map[string][]byte{}}
	service := receiptservice.NewService(repo, store, nil, c.TransactionService)
	if _, err := service.Retry(ctx, 1, id, 1); err == nil {
		t.Fatal("missing image retried")
	}
	store.objects["receipts/"+id] = []byte("synthetic")
	if _, err := service.Retry(actor.WithID(ctx, 404), 1, id, 1); err == nil {
		t.Fatal("outsider retried")
	}
	if err := service.Remove(actor.WithID(ctx, 404), 1, id, 1); err == nil {
		t.Fatal("outsider deleted")
	}
	database.DB.Exec("INSERT INTO receipt_image_gc(image_key) VALUES(?)", "receipts/"+id)
	if _, err := service.Retry(ctx, 1, id, 1); err == nil {
		t.Fatal("orphaned image retried")
	}
	database.DB.Exec("DELETE FROM receipt_image_gc")
	database.DB.Exec("UPDATE receipts SET status='confirmed' WHERE id=?", id)
	if _, err := service.Retry(ctx, 1, id, 1); err == nil {
		t.Fatal("confirmed receipt retried")
	}
	if err := service.Remove(ctx, 1, id, 1); err == nil {
		t.Fatal("confirmed receipt deleted")
	}
	database.DB.Exec("UPDATE receipts SET status='failed' WHERE id=?", id)
	database.DB.Exec("UPDATE events SET deleted_at=clock_timestamp() WHERE id=1")
	if _, err := service.Retry(ctx, 1, id, 1); err == nil {
		t.Fatal("archived receipt retried")
	}
	if err := service.Remove(ctx, 1, id, 1); err == nil {
		t.Fatal("archived receipt deleted")
	}
}

// disappearingReceiptImage воспроизводит GC после завершения чтения фото, до блокировки retry.
type disappearingReceiptImage struct {
	repo interface {
		DeleteGarbage(context.Context, string) error
	}
}

// Put не участвует в проверке завершённой очистки.
func (s *disappearingReceiptImage) Put(context.Context, string, []byte, string) error { return nil }

// Delete не участвует в проверке завершённой очистки.
func (s *disappearingReceiptImage) Delete(context.Context, string) error { return nil }

// Read фиксирует удаление объекта в окне между чтением и блокировкой.
func (s *disappearingReceiptImage) Read(ctx context.Context, key string) ([]byte, error) {
	// S3 уже вернул прочитанные байты; в этом окне другой воркер удалил объект и подтвердил outbox.
	if e := s.repo.DeleteGarbage(ctx, key); e != nil {
		return nil, e
	}
	return []byte("already read synthetic image"), nil
}

// TestReceiptRetryVsFinishedImageGC не ставит задание на уже удалённое GC изображение.
func TestReceiptRetryVsFinishedImageGC(t *testing.T) {
	d, c := settlementFixture(t)
	receiptMigration(t, d)
	id := seedReceipt(t, d, 1, "failed")
	repo := receiptrepo.NewRepository(d.DB)
	if e := d.DB.Exec("INSERT INTO receipt_image_gc(image_key,created_at) VALUES(?,clock_timestamp()-interval '2 hours')", "receipts/"+id).Error; e != nil {
		t.Fatal(e)
	}
	s := receiptservice.NewService(repo, &disappearingReceiptImage{repo: repo}, nil, c.TransactionService)
	if v, e := s.Retry(actor.WithID(context.Background(), 101), 1, id, 1); e == nil {
		t.Fatalf("retry queued a photo after acknowledged GC deletion: status=%s rev=%d", v.Status, v.Revision)
	}
}

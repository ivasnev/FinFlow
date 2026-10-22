package tests

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/png"
	"sync"
	"testing"

	"github.com/ivasnev/FinFlow/ff-split/internal/common/actor"
	commonerrors "github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	receiptrepo "github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/receipt"
	receiptservice "github.com/ivasnev/FinFlow/ff-split/internal/service/receipt"
)

// flakyReceiptStore воспроизводит прерванную первую запись без сохранения реальных фотографий.
type flakyReceiptStore struct {
	receiptHTTPStore
	failMu   sync.Mutex
	attempts int
	cancel   context.CancelFunc
}

// Put отклоняет первую запись и при необходимости отменяет исходный HTTP-контекст.
func (s *flakyReceiptStore) Put(ctx context.Context, key string, data []byte, kind string) error {
	s.failMu.Lock()
	s.attempts++
	attempt := s.attempts
	s.failMu.Unlock()
	if attempt == 1 {
		if s.cancel != nil {
			s.cancel()
		}
		return errors.New("synthetic image storage outage")
	}
	return s.receiptHTTPStore.Put(ctx, key, data, kind)
}

// TestReceiptFailedUploadRecovery восстанавливает тот же чек, не теряя правок и не переиспользуя старый ключ фото.
func TestReceiptFailedUploadRecovery(t *testing.T) {
	database, c := settlementFixture(t)
	receiptMigration(t, database)
	ctx := actor.WithID(context.Background(), 101)
	request, cancel := context.WithCancel(ctx)
	defer cancel()
	repo := receiptrepo.NewRepository(database.DB)
	store := &flakyReceiptStore{receiptHTTPStore: receiptHTTPStore{objects: map[string][]byte{}}, cancel: cancel}
	service := receiptservice.NewService(repo, store, nil, c.TransactionService)
	var photo bytes.Buffer
	if err := png.Encode(&photo, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Upload(request, 1, photo.Bytes()); !errors.Is(err, commonerrors.ErrReceiptStorageUnavailable) {
		t.Fatalf("unexpected outage error: %v", err)
	}
	list, err := service.List(ctx, 1, 0)
	if err != nil || len(list.Receipts) != 1 {
		t.Fatalf("failed reservation missing: %v", err)
	}
	failed, err := repo.Get(ctx, 1, list.Receipts[0].ID, false)
	if err != nil {
		t.Fatal(err)
	}
	if failed.Status != "failed" || failed.ImageAvailable {
		t.Fatal("request cancellation prevented failure finalization")
	}
	oldKey := failed.ImageKey
	draft := models.ReceiptDraft{Revision: failed.Revision, Items: []models.ReceiptDraftItem{{Name: "Saved manual draft"}}}
	if _, err = service.Edit(ctx, 1, failed.ID, draft, false); err != nil {
		t.Fatal(err)
	}
	resumed, err := service.Upload(ctx, 1, photo.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if resumed.ID != failed.ID || resumed.ImageKey == oldKey || resumed.Status != "queued" || !resumed.ImageAvailable || resumed.Draft.Items[0].Name != "Saved manual draft" {
		t.Fatalf("failed upload did not recover safely: %+v", resumed)
	}
	if err = repo.DeleteGarbage(ctx, oldKey); err != nil {
		t.Fatal(err)
	}
	current, err := repo.Get(ctx, 1, resumed.ID, false)
	if err != nil || !current.ImageAvailable {
		t.Fatal("late GC invalidated the replacement photo")
	}
	duplicate, err := service.Upload(ctx, 1, photo.Bytes())
	if err != nil || duplicate.ID != resumed.ID {
		t.Fatal("successful image was not deduplicated")
	}
	var count int64
	database.DB.Table("receipt_admissions").Where("receipt_id=?", resumed.ID).Count(&count)
	if count != 2 || store.attempts != 2 {
		t.Fatalf("wrong quota/storage attempts: %d/%d", count, store.attempts)
	}
}

// TestReceiptLegacyFailedUploadRecovery восстанавливает загрузку без достоверного старого флага доступности.
func TestReceiptLegacyFailedUploadRecovery(t *testing.T) {
	database, c := settlementFixture(t)
	receiptMigration(t, database)
	ctx := actor.WithID(context.Background(), 101)
	repo := receiptrepo.NewRepository(database.DB)
	store := &flakyReceiptStore{receiptHTTPStore: receiptHTTPStore{objects: map[string][]byte{}}}
	service := receiptservice.NewService(repo, store, nil, c.TransactionService)
	var photo bytes.Buffer
	if err := png.Encode(&photo, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatal(err)
	}
	_, _ = service.Upload(ctx, 1, photo.Bytes())
	list, err := service.List(ctx, 1, 0)
	if err != nil || len(list.Receipts) != 1 {
		t.Fatalf("failed reservation missing: %v", err)
	}
	failed, err := repo.Get(ctx, 1, list.Receipts[0].ID, false)
	if err != nil {
		t.Fatal(err)
	}
	// Состояние старого 000007 после 000008: failed, доступность по умолчанию true,
	// но PUT не завершился и записи outbox ещё нет.
	if err := database.DB.Exec("UPDATE receipts SET image_available=true WHERE id=?", failed.ID).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.DB.Exec("DELETE FROM receipt_image_gc WHERE image_key=?", failed.ImageKey).Error; err != nil {
		t.Fatal(err)
	}
	resumed, err := service.Upload(ctx, 1, photo.Bytes())
	if err != nil || resumed.ID != failed.ID || resumed.ImageKey == failed.ImageKey || resumed.Status != "queued" || !resumed.ImageAvailable {
		t.Fatalf("legacy missing photo did not recover: %+v, %v", resumed, err)
	}
}

// TestReceiptFailedUploadQuota не позволяет обходить допуск повторной записью того же фото.
func TestReceiptFailedUploadQuota(t *testing.T) {
	database, c := settlementFixture(t)
	receiptMigration(t, database)
	ctx := actor.WithID(context.Background(), 101)
	repo := receiptrepo.NewRepository(database.DB)
	store := &flakyReceiptStore{receiptHTTPStore: receiptHTTPStore{objects: map[string][]byte{}}}
	service := receiptservice.NewService(repo, store, nil, c.TransactionService)
	var photo bytes.Buffer
	if err := png.Encode(&photo, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatal(err)
	}
	_, _ = service.Upload(ctx, 1, photo.Bytes())
	if err := database.DB.Exec("INSERT INTO receipt_admissions(receipt_id,actor_id) SELECT gen_random_uuid(),101 FROM generate_series(1,49)").Error; err != nil {
		t.Fatal(err)
	}
	if _, err := service.Upload(ctx, 1, photo.Bytes()); err == nil {
		t.Fatal("recovery bypassed daily admission budget")
	}
	if store.attempts != 1 {
		t.Fatal("quota rejected only after S3 call")
	}
}

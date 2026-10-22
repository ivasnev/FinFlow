package tests

import (
	"context"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/actor"
	dbctx "github.com/ivasnev/FinFlow/ff-split/internal/common/db"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	receiptrepo "github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/receipt"
	receiptservice "github.com/ivasnev/FinFlow/ff-split/internal/service/receipt"
	"testing"
	"time"
)

// TestReceiptGCConcurrentConfirmation проверяет пропуск подтверждаемого чека при очистке.
func TestReceiptGCConcurrentConfirmation(t *testing.T) {
	database, c := settlementFixture(t)
	receiptMigration(t, database)
	id := seedReceipt(t, database, 1, "ready")
	if err := database.DB.Exec("UPDATE receipts SET updated_at=clock_timestamp()-interval '31 days',created_at=clock_timestamp()-interval '31 days' WHERE id=?", id).Error; err != nil {
		t.Fatal(err)
	}
	ctx := actor.WithID(context.Background(), 101)
	held := database.DB.Begin()
	defer held.Rollback()
	bound := context.WithValue(ctx, dbctx.TxContextKey{}, held)
	repo := receiptrepo.NewRepository(database.DB)
	s := receiptservice.NewService(repo, nil, nil, c.TransactionService)
	_, err := s.Edit(bound, 1, id, models.ReceiptDraft{Revision: 1, PayerID: 1, Items: []models.ReceiptDraftItem{{Name: "Confirm while GC starts", AmountCents: 100, Users: []int64{2}, Selected: true}}}, true)
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { _, err := repo.Garbage(context.Background()); done <- err }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("GC blocked on in-flight confirmation instead of skipping it")
	}
	if err = held.Commit().Error; err != nil {
		t.Fatal(err)
	}
	row, err := repo.Get(ctx, 1, id, false)
	if err != nil || row.Status != "confirmed" {
		t.Fatalf("GC deleted a freshly committed confirmation: receipt=%+v err=%v", row, err)
	}
}

// TestReceiptGCConcurrentClaim проверяет пропуск арендованной задачи при очистке.
func TestReceiptGCConcurrentClaim(t *testing.T) {
	database, _ := settlementFixture(t)
	receiptMigration(t, database)
	id := seedReceipt(t, database, 1, "queued")
	if err := database.DB.Exec("UPDATE receipts SET updated_at=clock_timestamp()-interval '31 days',created_at=clock_timestamp()-interval '31 days' WHERE id=?", id).Error; err != nil {
		t.Fatal(err)
	}
	held := database.DB.Begin()
	defer held.Rollback()
	job, err := receiptrepo.NewRepository(held).Claim(context.Background())
	if err != nil || job == nil {
		t.Fatalf("claim %+v %v", job, err)
	}
	repo := receiptrepo.NewRepository(database.DB)
	done := make(chan error, 1)
	go func() { _, err := repo.Garbage(context.Background()); done <- err }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("GC blocked on claim instead of skipping")
	}
	if err = held.Commit().Error; err != nil {
		t.Fatal(err)
	}
	row, err := repo.Get(context.Background(), 1, id, false)
	if err != nil || row.Status != "processing" || row.LeaseToken != job.LeaseToken {
		t.Fatalf("claimed receipt lost %+v %v", row, err)
	}
}

// TestReceiptGCRetention проверяет срок хранения черновика и надёжную очередь удаления фото.
func TestReceiptGCRetention(t *testing.T) {
	database, _ := settlementFixture(t)
	receiptMigration(t, database)
	old := seedReceipt(t, database, 1, "ready")
	confirmed := seedReceipt(t, database, 1, "confirmed")
	fresh := seedReceipt(t, database, 1, "ready")
	if err := database.DB.Exec("UPDATE receipts SET updated_at=clock_timestamp()-interval '31 days',created_at=clock_timestamp()-interval '31 days' WHERE id IN (?,?)", old, confirmed).Error; err != nil {
		t.Fatal(err)
	}
	repo := receiptrepo.NewRepository(database.DB)
	keys, err := repo.Garbage(context.Background())
	if err != nil || len(keys) != 0 {
		t.Fatalf("GC exposed fresh outbox key too early %v %v", keys, err)
	}
	for _, id := range []string{confirmed, fresh} {
		if _, err = repo.Get(context.Background(), 1, id, false); err != nil {
			t.Fatal("retained receipt missing")
		}
	}
	if _, err = repo.Get(context.Background(), 1, old, false); err == nil {
		t.Fatal("old draft retained")
	}
	var count int64
	database.DB.Table("receipt_image_gc").Where("image_key=?", "receipts/"+old).Count(&count)
	if count != 1 {
		t.Fatal("image not durably queued for deletion")
	}
	database.DB.Exec("UPDATE receipt_image_gc SET created_at=clock_timestamp()-interval '2 hours'")
	keys, err = repo.Garbage(context.Background())
	if err != nil || len(keys) != 1 || keys[0] != "receipts/"+old {
		t.Fatal("mature outbox mismatch")
	}
	if err = repo.DeleteGarbage(context.Background(), keys[0]); err != nil {
		t.Fatal(err)
	}
	if err = repo.DeleteGarbage(context.Background(), keys[0]); err != nil {
		t.Fatal(err)
	}
}

package tests

import (
	"context"
	"errors"
	commonerrors "github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	receiptrepo "github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/receipt"
	receiptservice "github.com/ivasnev/FinFlow/ff-split/internal/service/receipt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestReceiptBusyBudget проверяет временную перегрузку без расходования попыток и конечный срок ожидания.
func TestReceiptBusyBudget(t *testing.T) {
	database, _ := settlementFixture(t)
	receiptMigration(t, database)
	id := seedReceipt(t, database, 1, "queued")
	repo := receiptrepo.NewRepository(database.DB)
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		job, err := repo.Claim(ctx)
		if err != nil || job == nil {
			t.Fatalf("claim %d %+v %v", i, job, err)
		}
		if err = repo.Complete(ctx, job, nil, commonerrors.ErrOCRUnavailable); err != nil {
			t.Fatal(err)
		}
		var state struct {
			Status   string
			Attempts int
		}
		if err = database.DB.Table("receipts").Select("status,attempts").Where("id=?", id).Scan(&state).Error; err != nil {
			t.Fatal(err)
		}
		if state.Status != "queued" || state.Attempts != 0 {
			t.Fatalf("busy consumed budget %+v", state)
		}
		database.DB.Exec("UPDATE receipts SET available_at=clock_timestamp()-interval '1 second' WHERE id=?", id)
	}
	database.DB.Exec("UPDATE receipts SET queued_at=clock_timestamp()-interval '16 minutes',created_at=clock_timestamp()-interval '16 minutes' WHERE id=?", id)
	for i := 0; i < 3; i++ {
		job, err := repo.Claim(ctx)
		if err != nil || job == nil {
			t.Fatalf("expired claim%d %+v %v", i, job, err)
		}
		if err = repo.Complete(ctx, job, nil, commonerrors.ErrOCRUnavailable); err != nil {
			t.Fatal(err)
		}
		database.DB.Exec("UPDATE receipts SET available_at=clock_timestamp()-interval '1 second' WHERE id=?", id)
	}
	row, err := repo.Get(ctx, 1, id, false)
	if err != nil || row.Status != "failed" {
		t.Fatalf("busy lifetime not bounded %+v %v", row, err)
	}
}

// TestReceiptUnavailableHTTP проверяет различие временной недоступности и некорректного изображения.
func TestReceiptUnavailableHTTP(t *testing.T) {
	for _, status := range []int{429, 503, 422} {
		// Сценарий: проверка отдельного варианта входных данных.
		t.Run(http.StatusText(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(status) }))
			defer server.Close()
			r := &receiptservice.HTTPRecognizer{URL: server.URL, Token: "local-synthetic-token", Client: &http.Client{Timeout: time.Second}}
			_, err := r.Recognize(context.Background(), []byte("test"), "image/png")
			if (status == 429 || status == 503) != errors.Is(err, commonerrors.ErrOCRUnavailable) {
				t.Fatalf("HTTP%d classification %v", status, err)
			}
		})
	}
}

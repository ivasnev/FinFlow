package tests

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/png"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/actor"
	commonerrors "github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	receiptrepo "github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/receipt"
	receiptservice "github.com/ivasnev/FinFlow/ff-split/internal/service/receipt"
	"github.com/stretchr/testify/require"
)

func TestReceiptGlobalBudgets(t *testing.T) {
	database, c := settlementFixture(t)
	receiptMigration(t, database)
	ctx := actor.WithID(context.Background(), 101)
	repo := receiptrepo.NewRepository(database.DB)
	store := &receiptHTTPStore{objects: map[string][]byte{}}
	pipeline := receiptservice.NewService(repo, store, nil, c.TransactionService)
	// Other users consume the global allowance; the current user has no uploads.
	require.NoError(t, database.DB.Exec(`INSERT INTO receipt_admissions(receipt_id,actor_id) SELECT ?,10000+i FROM generate_series(1,500) i`, uuid.NewString()).Error)
	var photo bytes.Buffer
	require.NoError(t, png.Encode(&photo, image.NewRGBA(image.Rect(0, 0, 1, 1))))
	_, err := pipeline.Upload(ctx, 1, photo.Bytes())
	require.Error(t, err)
	require.Zero(t, store.puts)
	require.NoError(t, database.DB.Exec(`TRUNCATE receipt_admissions`).Error)
	id := seedReceipt(t, database, 1, "ready")
	store.objects["receipts/"+id] = photo.Bytes()
	// Simulate 999 reads by other accounts and race for the last global slot.
	require.NoError(t, database.DB.Exec(`INSERT INTO receipt_image_reads(actor_id) SELECT i+1000 FROM generate_series(1,999) i`).Error)
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for range 2 {
		wg.Add(1)
		go func() { defer wg.Done(); _, _, err := pipeline.Image(ctx, 1, id); results <- err }()
	}
	wg.Wait()
	close(results)
	successes, rejections := 0, 0
	for err := range results {
		if err == nil {
			successes++
		} else {
			require.ErrorIs(t, err, commonerrors.ErrReceiptReadBudget)
			rejections++
		}
	}
	require.Equal(t, 1, successes)
	require.Equal(t, 1, rejections)
	require.Equal(t, 1, store.reads)
	// Restarting the service and using a different actor cannot reset the global budget.
	pipeline = receiptservice.NewService(repo, store, nil, c.TransactionService)
	_, _, err = pipeline.Image(ctx, 1, id)
	require.ErrorIs(t, err, commonerrors.ErrReceiptReadBudget)
	require.Equal(t, 1, store.reads)
	require.ErrorIs(t, repo.AdmitImageRead(context.Background(), 99999), commonerrors.ErrReceiptReadBudget)
	// Denial occurs before the S3 preflight read of a retry too.
	current, err := repo.Get(ctx, 1, id, false)
	require.NoError(t, err)
	_, err = pipeline.Retry(ctx, 1, id, current.Revision)
	require.ErrorIs(t, err, commonerrors.ErrReceiptReadBudget)
	require.Equal(t, 1, store.reads)
	require.NoError(t, database.DB.Exec(`UPDATE receipt_image_reads SET created_at=now()-interval '3 days'`).Error)
	require.NoError(t, database.DB.Exec(`UPDATE receipt_admissions SET created_at=now()-interval '8 days'`).Error)
	_, err = repo.Garbage(ctx)
	require.NoError(t, err)
	require.NoError(t, repo.AdmitImageRead(ctx, 101))
}

// Keep repository validation at the same admission boundary for direct service callers.
func TestReceiptGlobalUploadBudgetAcrossActors(t *testing.T) {
	database, _ := settlementFixture(t)
	receiptMigration(t, database)
	repo := receiptrepo.NewRepository(database.DB)
	require.NoError(t, database.DB.Exec(`INSERT INTO receipt_admissions(receipt_id,actor_id) SELECT ?,10000+i FROM generate_series(1,500) i`, uuid.NewString()).Error)
	err := repo.InEvent(context.Background(), 1, func(ctx context.Context) error {
		_, err := repo.Create(ctx, &models.Receipt{ID: uuid.NewString(), EventID: 1, CreatedBy: 101, ImageKey: "budget", ImageSHA256: "budget", ContentType: "image/png"})
		return err
	})
	require.Error(t, err)
	require.False(t, errors.Is(err, commonerrors.ErrForbidden))
	var count int64
	require.NoError(t, database.DB.Table("receipts").Count(&count).Error)
	require.Zero(t, count)
}

func TestReceiptWorkerReadBudget(t *testing.T) {
	database, c := settlementFixture(t)
	receiptMigration(t, database)
	repo := receiptrepo.NewRepository(database.DB)
	id := seedReceipt(t, database, 1, "queued")
	require.NoError(t, database.DB.Exec(`INSERT INTO receipt_image_reads(actor_id) SELECT i+1000 FROM generate_series(1,1000) i`).Error)
	store := &receiptHTTPStore{objects: map[string][]byte{}}
	pipeline := receiptservice.NewService(repo, store, nil, c.TransactionService)
	stop := pipeline.Start(context.Background(), 1)
	defer stop()
	require.Eventually(t, func() bool {
		var status string
		require.NoError(t, database.DB.Table("receipts").Select("status").Where("id=?", id).Scan(&status).Error)
		return status == "failed"
	}, 5*time.Second, 20*time.Millisecond)
	stop() // Wait for the worker before inspecting the fake's counters.
	require.Zero(t, store.reads)
	var attempts int
	require.NoError(t, database.DB.Table("receipts").Select("attempts").Where("id=?", id).Scan(&attempts).Error)
	require.Equal(t, 1, attempts)
}

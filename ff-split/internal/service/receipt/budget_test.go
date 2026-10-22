package receipt

import (
	"context"
	"errors"
	commonerrors "github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository"
	"testing"
	"time"
)

type budgetRepo struct {
	repository.Receipt
	reads, completed int
	actor            int64
	processingErr    error
}

func (r *budgetRepo) Claim(context.Context) (*models.Receipt, error) {
	return &models.Receipt{ID: "independent-job", CreatedBy: 42, QueuedBy: 99, ImageKey: "synthetic/image", ContentType: "image/png", Attempts: 1, QueuedAt: time.Now()}, nil
}
func (r *budgetRepo) AdmitImageRead(_ context.Context, actor int64) error {
	r.reads++
	r.actor = actor
	return commonerrors.ErrReceiptReadBudget
}
func (r *budgetRepo) Complete(_ context.Context, _ *models.Receipt, _ *models.ReceiptResult, err error) error {
	r.completed++
	r.processingErr = err
	return nil
}

type budgetStore struct {
	ImageStore
	reads int
}

func (s *budgetStore) Read(context.Context, string) ([]byte, error) {
	s.reads++
	return []byte("synthetic image"), nil
}

type budgetOCR struct{ calls int }

func (o *budgetOCR) Recognize(context.Context, []byte, string) (*models.ReceiptResult, error) {
	o.calls++
	return nil, commonerrors.ErrOCRUnavailable
}

func TestWorkerReadBudget(t *testing.T) {
	repo := &budgetRepo{}
	store := &budgetStore{}
	ocr := &budgetOCR{}
	service := NewService(repo, store, ocr, nil)
	service.processOne(context.Background())
	if repo.actor != 99 {
		t.Fatalf("charged wrong actor: %d", repo.actor)
	}
	if repo.reads != 1 || store.reads != 0 || ocr.calls != 0 || repo.completed != 1 || !errors.Is(repo.processingErr, commonerrors.ErrReceiptReadBudget) {
		t.Fatalf("budgetCalls=%d S3Reads=%d OCRCalls=%d completed=%d error=%v", repo.reads, store.reads, ocr.calls, repo.completed, repo.processingErr)
	}
}

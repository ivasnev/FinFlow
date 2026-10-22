package receipt

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"math"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/actor"
	commonerrors "github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
	_ "golang.org/x/image/webp"
)

const MaxImageBytes = 10 << 20

// ImageStore хранит приватные объекты; URL от клиента никогда не загружается.
type ImageStore interface {
	// Put записывает фотографию по серверному приватному ключу.
	Put(context.Context, string, []byte, string) error
	// Read читает ограниченный размер фотографии.
	Read(context.Context, string) ([]byte, error)
	// Delete идемпотентно удаляет фотографию.
	Delete(context.Context, string) error
}

// Recognizer возвращает только структурированный ограниченный результат.
type Recognizer interface {
	// Recognize распознаёт изображение в пределах контекста запроса.
	Recognize(context.Context, []byte, string) (*models.ReceiptResult, error)
}

// Service связывает устойчивую очередь с обычным атомарным созданием расходов.
type Service struct {
	repo         repository.Receipt
	store        ImageStore
	ocr          Recognizer
	transactions service.Transaction
	slots        chan struct{}
}

// NewService ограничивает число одновременно загружаемых фотографий на экземпляр API.
func NewService(repo repository.Receipt, store ImageStore, ocr Recognizer, transactions service.Transaction) *Service {
	return &Service{repo: repo, store: store, ocr: ocr, transactions: transactions, slots: make(chan struct{}, 4)}
}

// ImageSlot ограничивает память при загрузке и выдаче фотографий.
func (s *Service) ImageSlot() (func(), bool) {
	select {
	case s.slots <- struct{}{}:
		return func() { <-s.slots }, true
	default:
		return nil, false
	}
}

// member проверяет участие повторно в сервисной границе, в том числе после блокировки.
func (s *Service) member(ctx context.Context, eventID int64) (int64, error) {
	id, ok := actor.ID(ctx)
	if !ok {
		return 0, commonerrors.ErrForbidden
	}
	member, err := s.repo.IsMember(ctx, eventID, id)
	if err != nil {
		return 0, err
	}
	if !member {
		return 0, commonerrors.ErrForbidden
	}
	return id, nil
}

// Upload сначала проверяет безопасные размеры, затем сохраняет объект и атомарно ставит задачу.
func (s *Service) Upload(ctx context.Context, eventID int64, data []byte) (*models.Receipt, error) {
	actorID, err := s.member(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 || len(data) > MaxImageBytes {
		return nil, commonerrors.NewValidationError("image", "фото должно быть не больше 10 МиБ")
	}
	cfg, kind, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || cfg.Width < 1 || cfg.Height < 1 || int64(cfg.Width)*int64(cfg.Height) > 16_000_000 || cfg.Width > 12000 || cfg.Height > 12000 {
		return nil, commonerrors.NewValidationError("image", "нужна корректная фотография JPEG, PNG или WebP до 16 мегапикселей")
	}
	types := map[string]string{"jpeg": "image/jpeg", "png": "image/png", "webp": "image/webp"}
	contentType := types[kind]
	if contentType == "" {
		return nil, commonerrors.NewValidationError("image", "поддерживаются JPEG, PNG и WebP")
	}
	id := uuid.NewString()
	key := "receipts/" + id
	hash := sha256.Sum256(data)
	var result *models.Receipt
	err = s.repo.InEvent(ctx, eventID, func(ctx context.Context) error {
		if _, err := s.member(ctx, eventID); err != nil {
			return err
		}
		var createErr error
		result, createErr = s.repo.Create(ctx, &models.Receipt{ID: id, EventID: eventID, CreatedBy: actorID, ImageKey: key, ImageSHA256: hex.EncodeToString(hash[:]), ContentType: contentType})
		return createErr
	})
	if err != nil {
		return nil, err
	}
	if result.ImageKey != key {
		return result, nil
	}
	upload, done := context.WithTimeout(ctx, 30*time.Second)
	defer done()
	if err := s.store.Put(upload, key, data, contentType); err != nil {
		finish, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = s.repo.FinishUpload(finish, result, false)
		return nil, commonerrors.ErrReceiptStorageUnavailable
	}
	if err := s.repo.FinishUpload(upload, result, true); err != nil {
		return nil, err
	}
	return result, nil
}

// List возвращает сохранённые задачи, включая завершённые и архивные.
func (s *Service) List(ctx context.Context, eventID, before int64) (*models.ReceiptList, error) {
	if _, err := s.member(ctx, eventID); err != nil {
		return nil, err
	}
	return s.repo.List(ctx, eventID, before)
}

// Get не принимает ключи S3 от клиента.
func (s *Service) Get(ctx context.Context, eventID int64, id string) (*models.Receipt, error) {
	if _, err := s.member(ctx, eventID); err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, eventID, id, false)
}

// Image отдаёт фотографию только после проверки события и текущего участия.
func (s *Service) Image(ctx context.Context, eventID int64, id string) ([]byte, string, error) {
	v, err := s.Get(ctx, eventID, id)
	if err != nil {
		return nil, "", err
	}
	read, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	data, err := s.store.Read(read, v.ImageKey)
	return data, v.ContentType, err
}

// validateDraft не заменяет деньги распознанными догадками и ограничивает пакет до 100 покупок.
func validateDraft(d *models.ReceiptDraft, confirm bool, total *int64) error {
	if len(d.Items) > 100 {
		return commonerrors.NewValidationError("receipt", "не более 100 позиций")
	}
	selected := 0
	sum := int64(0)
	var mergedUsers []int64
	for i := range d.Items {
		item := &d.Items[i]
		item.Name = strings.TrimSpace(item.Name)
		if utf8.RuneCountInString(item.Name) > 255 || item.AmountCents < 0 || item.AmountCents > 9_999_999_999 || len(item.Users) > 200 {
			return commonerrors.NewValidationError("receipt", "проверьте название, сумму и участников позиций")
		}
		sort.Slice(item.Users, func(i, j int) bool { return item.Users[i] < item.Users[j] })
		for j, id := range item.Users {
			if id <= 0 || (j > 0 && item.Users[j-1] == id) {
				return commonerrors.NewValidationError("receipt", "участники позиции не должны повторяться")
			}
		}
		if !item.Selected {
			continue
		}
		selected++
		sum += item.AmountCents
		if confirm && (item.Name == "" || item.AmountCents <= 0 || len(item.Users) == 0) {
			return commonerrors.NewValidationError("receipt", "каждой выбранной позиции нужны название, сумма и участники")
		}
		if mergedUsers == nil {
			mergedUsers = item.Users
		} else if confirm && d.Merge && !reflect.DeepEqual(mergedUsers, item.Users) {
			return commonerrors.NewValidationError("receipt", "объединять можно позиции с одинаковыми участниками")
		}
	}
	if !confirm {
		return nil
	}
	if selected == 0 || d.PayerID <= 0 {
		return commonerrors.NewValidationError("receipt", "выберите покупки и плательщика")
	}
	if d.Merge && sum > 9_999_999_999 {
		return commonerrors.NewValidationError("receipt", "сумма объединённого расхода слишком велика")
	}
	if total != nil && sum != *total && !d.AcknowledgeDifference {
		return commonerrors.NewValidationError("receipt", "сумма выбранного отличается от итога чека; подтвердите это явно")
	}
	return nil
}

// Edit сохраняет правки с контролем ревизии, Confirm создаёт весь пакет ровно одной транзакцией.
func (s *Service) Edit(ctx context.Context, eventID int64, id string, draft models.ReceiptDraft, confirm bool) (*models.Receipt, error) {
	var result *models.Receipt
	err := s.repo.InEvent(ctx, eventID, func(ctx context.Context) error {
		if _, err := s.member(ctx, eventID); err != nil {
			return err
		}
		v, err := s.repo.Get(ctx, eventID, id, true)
		if err != nil {
			return err
		}
		if err := validateDraft(&draft, false, v.Result.TotalCents); err != nil {
			return err
		}
		if v.Status == "confirmed" {
			before, after := v.Draft, draft
			before.Revision = 0
			after.Revision = 0
			if confirm && reflect.DeepEqual(before, after) {
				result = v
				return nil
			}
			return commonerrors.NewValidationError("receipt", "расходы по этому чеку уже созданы")
		}
		if v.Status != "ready" && v.Status != "failed" {
			return commonerrors.NewValidationError("receipt", "дождитесь завершения распознавания")
		}
		if draft.Revision != v.Revision {
			return commonerrors.NewValidationError("receipt", "чек уже изменён; обновите страницу перед сохранением")
		}
		if err := validateDraft(&draft, confirm, v.Result.TotalCents); err != nil {
			return err
		}
		v.Draft = draft
		if confirm {
			items := []models.ReceiptDraftItem{}
			for _, item := range draft.Items {
				if item.Selected {
					items = append(items, item)
				}
			}
			if draft.Merge {
				var sum int64
				for _, item := range items {
					sum += item.AmountCents
				}
				items = []models.ReceiptDraftItem{{Name: "Покупки по чеку", AmountCents: sum, Users: items[0].Users, Selected: true}}
			}
			for _, item := range items {
				transaction, err := s.transactions.CreateTransaction(ctx, eventID, &service.TransactionRequest{Type: "units", FromUser: draft.PayerID, Amount: float64(item.AmountCents) / 100, Name: item.Name, Users: item.Users})
				if err != nil {
					return err
				}
				v.TransactionIDs = append(v.TransactionIDs, transaction.ID)
			}
			v.Status = "confirmed"
		}
		if err := s.repo.Save(ctx, v); err != nil {
			return err
		}
		result = v
		return nil
	})
	return result, err
}

// validResult ограничивает недоверенный ответ OCR до сохранения в базе.
func validResult(result *models.ReceiptResult) bool {
	if result == nil || len(result.Items) > 100 || len(result.RawText) > 100000 || len(result.Warnings) > 20 || len(result.Engine) > 200 {
		return false
	}
	if result.TotalCents != nil && (*result.TotalCents < 0 || *result.TotalCents > 9_999_999_999) {
		return false
	}
	for _, item := range result.Items {
		if item.AmountCents <= 0 || item.AmountCents > 9_999_999_999 || len(item.Name) > 1020 || len(item.SourceText) > 4000 || math.IsNaN(item.Confidence) || item.Confidence < 0 || item.Confidence > 1 {
			return false
		}
	}
	return true
}

// Start запускает ограниченное число обработчиков с отменой и ожиданием остановки.
func (s *Service) Start(parent context.Context, workers int) func() {
	if workers < 1 {
		workers = 1
	}
	if workers > 8 {
		workers = 8
	}
	ctx, cancel := context.WithCancel(parent)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			timer := time.NewTimer(0)
			defer timer.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-timer.C:
					s.processOne(ctx)
					timer.Reset(time.Second)
				}
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		timer := time.NewTicker(time.Minute)
		defer timer.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				keys, err := s.repo.Garbage(ctx)
				if err != nil {
					log.Printf("receipt_gc outcome=failed")
					continue
				}
				for _, key := range keys {
					if s.store.Delete(ctx, key) == nil {
						_ = s.repo.DeleteGarbage(ctx, key)
					}
				}
			}
		}
	}()
	return func() { cancel(); wg.Wait() }
}

// processOne не держит SQL-транзакцию во время сетевой или вычислительной работы.
func (s *Service) processOne(parent context.Context) {
	ctx, cancel := context.WithTimeout(parent, 95*time.Second)
	defer cancel()
	job, err := s.repo.Claim(ctx)
	if err != nil {
		log.Printf("receipt_claim outcome=failed")
		return
	}
	if job == nil {
		return
	}
	started := time.Now()
	data, err := s.store.Read(ctx, job.ImageKey)
	var result *models.ReceiptResult
	if err == nil {
		result, err = s.ocr.Recognize(ctx, data, job.ContentType)
	}
	if err == nil && !validResult(result) {
		err = fmt.Errorf("invalid OCR result")
	}
	if parent.Err() != nil {
		return
	}
	// Отдельный короткий deadline позволяет вернуть задачу в очередь после timeout OCR.
	finish, done := context.WithTimeout(parent, 5*time.Second)
	defer done()
	if completeErr := s.repo.Complete(finish, job, result, err); completeErr != nil {
		log.Printf("receipt_complete outcome=failed")
	}
	log.Printf("receipt_processed id=%s success=%t elapsed_ms=%d", job.ID, err == nil, time.Since(started).Milliseconds())
}

// HTTPRecognizer обращается только к адресу сервиса из серверной конфигурации.
type HTTPRecognizer struct {
	URL    string
	Token  string
	Client *http.Client
}

// Recognize ограничивает размер и схему ответа, не следует перенаправлениям.
func (c *HTTPRecognizer) Recognize(ctx context.Context, data []byte, contentType string) (*models.ReceiptResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.URL, "/")+"/recognize", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", contentType)
	response, err := c.Client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, err
		}
		return nil, commonerrors.ErrOCRUnavailable
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusTooManyRequests || response.StatusCode == http.StatusServiceUnavailable {
		return nil, commonerrors.ErrOCRUnavailable
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OCR status %d", response.StatusCode)
	}
	limited := io.LimitReader(response.Body, (2<<20)+1)
	body, err := io.ReadAll(limited)
	if err != nil || len(body) > 2<<20 {
		return nil, fmt.Errorf("invalid OCR response size")
	}
	var result models.ReceiptResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

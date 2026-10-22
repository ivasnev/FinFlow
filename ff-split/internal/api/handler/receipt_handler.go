package handler

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	receipts "github.com/ivasnev/FinFlow/ff-split/internal/service/receipt"
	"github.com/ivasnev/FinFlow/ff-split/pkg/api"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// SetReceiptService подключает опциональный приватный OCR-конвейер.
func (s *ServerHandler) SetReceiptService(receipts *receipts.Service) { s.receipts = receipts }

// receiptReady возвращает явную ошибку отключённого OCR.
func (s *ServerHandler) receiptReady(c *gin.Context) bool {
	c.Header("Cache-Control", "no-store")
	if s.receipts == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "распознавание чеков пока не настроено"})
		return false
	}
	return true
}

// ListReceipts читает ограниченную страницу задач своего события.
func (s *ServerHandler) ListReceipts(c *gin.Context, eventID int64, params api.ListReceiptsParams) {
	if !s.receiptReady(c) {
		return
	}
	before := int64(0)
	if params.Before != nil {
		before = *params.Before
	}
	v, err := s.receipts.List(c.Request.Context(), eventID, before)
	if err != nil {
		errors.HTTPErrorHandler(c, err)
		return
	}
	c.JSON(http.StatusOK, v)
}

// UploadReceipt ограничивает тело и число одновременно принимаемых изображений до чтения файла.
func (s *ServerHandler) UploadReceipt(c *gin.Context, eventID int64) {
	if !s.receiptReady(c) {
		return
	}
	release, ok := s.receipts.ImageSlot()
	if !ok {
		c.Header("Retry-After", "2")
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "загрузка занята, попробуйте чуть позже"})
		return
	}
	defer release()
	// Deadline чтения защищает слоты от медленных незавершённых загрузок.
	controller := http.NewResponseController(c.Writer)
	_ = controller.SetReadDeadline(time.Now().Add(30 * time.Second))
	defer controller.SetReadDeadline(time.Time{})
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, receipts.MaxImageBytes)
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "фото должно быть не больше 10 МиБ"})
		return
	}
	v, err := s.receipts.Upload(c.Request.Context(), eventID, data)
	if err != nil {
		if stderrors.Is(err, errors.ErrReceiptStorageUnavailable) {
			c.Header("Retry-After", "5")
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": gin.H{"code": "receipt_storage_unavailable", "message": "Хранилище фотографий временно недоступно. Повторите отправку позже."}})
			return
		}
		errors.HTTPErrorHandler(c, err)
		return
	}
	c.JSON(http.StatusAccepted, v)
}

// GetReceipt связывает UUID чека с текущим событием.
func (s *ServerHandler) GetReceipt(c *gin.Context, eventID int64, id openapi_types.UUID) {
	if !s.receiptReady(c) {
		return
	}
	v, err := s.receipts.Get(c.Request.Context(), eventID, id.String())
	if err != nil {
		errors.HTTPErrorHandler(c, err)
		return
	}
	c.JSON(http.StatusOK, v)
}

// GetReceiptImage отдаёт байты через проверку участия без утечки доступа к S3.
func (s *ServerHandler) GetReceiptImage(c *gin.Context, eventID int64, id openapi_types.UUID) {
	if !s.receiptReady(c) {
		return
	}
	release, ok := s.receipts.ImageSlot()
	if !ok {
		c.Header("Retry-After", "2")
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "выдача фотографий занята, попробуйте чуть позже"})
		return
	}
	defer release()
	data, contentType, err := s.receipts.Image(c.Request.Context(), eventID, id.String())
	if err != nil {
		errors.HTTPErrorHandler(c, err)
		return
	}
	c.Header("X-Content-Type-Options", "nosniff")
	c.Data(http.StatusOK, contentType, data)
}

// editReceipt принимает один ограниченный JSON и не допускает подмену авторства.
func (s *ServerHandler) editReceipt(c *gin.Context, eventID int64, id openapi_types.UUID, confirm bool) {
	if !s.receiptReady(c) {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 256<<10)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	var draft models.ReceiptDraft
	if err := decoder.Decode(&draft); err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("receipt", "некорректный черновик"))
		return
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		errors.HTTPErrorHandler(c, errors.NewValidationError("receipt", "ожидается один черновик"))
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	v, err := s.receipts.Edit(ctx, eventID, id.String(), draft, confirm)
	if err != nil {
		errors.HTTPErrorHandler(c, err)
		return
	}
	c.JSON(http.StatusOK, v)
}

// EditReceipt сохраняет черновик без создания расходов.
func (s *ServerHandler) EditReceipt(c *gin.Context, eventID int64, id openapi_types.UUID) {
	s.editReceipt(c, eventID, id, false)
}

// ConfirmReceipt атомарно создаёт выбранные расходы и запоминает подтверждение.
func (s *ServerHandler) ConfirmReceipt(c *gin.Context, eventID int64, id openapi_types.UUID) {
	s.editReceipt(c, eventID, id, true)
}

// receiptRevision принимает ровно одну положительную ревизию без посторонних полей.
func receiptRevision(c *gin.Context) (int, bool) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1024)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	var body struct {
		Revision int `json:"revision"`
	}
	if err := decoder.Decode(&body); err != nil || body.Revision < 1 {
		errors.HTTPErrorHandler(c, errors.NewValidationError("receipt", "укажите ревизию чека"))
		return 0, false
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		errors.HTTPErrorHandler(c, errors.NewValidationError("receipt", "ожидается один JSON-объект"))
		return 0, false
	}
	return body.Revision, true
}

// RetryReceipt ставит завершённый черновик в очередь повторного распознавания.
func (s *ServerHandler) RetryReceipt(c *gin.Context, eventID int64, id openapi_types.UUID) {
	if !s.receiptReady(c) {
		return
	}
	revision, ok := receiptRevision(c)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()
	value, err := s.receipts.Retry(ctx, eventID, id.String(), revision)
	if err != nil {
		errors.HTTPErrorHandler(c, err)
		return
	}
	c.JSON(http.StatusAccepted, value)
}

// DeleteReceipt скрывает ненужный черновик, не затрагивая расходы события.
func (s *ServerHandler) DeleteReceipt(c *gin.Context, eventID int64, id openapi_types.UUID) {
	if !s.receiptReady(c) {
		return
	}
	revision, ok := receiptRevision(c)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	if err := s.receipts.Remove(ctx, eventID, id.String(), revision); err != nil {
		errors.HTTPErrorHandler(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

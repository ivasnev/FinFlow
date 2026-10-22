package models

import "time"

// ReceiptItem хранит денежную сумму строки в копейках независимо от OCR-формата.
type ReceiptItem struct {
	Name        string  `json:"name"`
	AmountCents int64   `json:"amount_cents"`
	Confidence  float64 `json:"confidence"`
	Quantity    *string `json:"quantity,omitempty"`
	SourceText  string  `json:"source_text,omitempty"`
}

// ReceiptResult сохраняет исходный результат отдельно от правок человека.
type ReceiptResult struct {
	Items      []ReceiptItem `json:"items"`
	TotalCents *int64        `json:"total_cents"`
	Warnings   []string      `json:"warnings"`
	RawText    string        `json:"raw_text"`
	Engine     string        `json:"engine"`
}

// ReceiptDraftItem задаёт выбранную покупку и внутренние ID участников.
type ReceiptDraftItem struct {
	Name        string  `json:"name"`
	AmountCents int64   `json:"amount_cents"`
	Users       []int64 `json:"users"`
	Selected    bool    `json:"selected"`
}

// ReceiptDraft задаёт проверенные человеком покупки и режим объединения.
type ReceiptDraft struct {
	Revision              int                `json:"revision"`
	PayerID               int64              `json:"payer_id"`
	Items                 []ReceiptDraftItem `json:"items"`
	Merge                 bool               `json:"merge"`
	AcknowledgeDifference bool               `json:"acknowledge_difference"`
}

// Receipt доступен только участникам события; ключ объекта не выходит наружу.
type Receipt struct {
	ID             string        `json:"id"`
	EventID        int64         `json:"event_id"`
	CreatedBy      int64         `json:"created_by"`
	Sequence       int64         `json:"sequence"`
	Status         string        `json:"status"`
	Revision       int           `json:"revision"`
	Result         ReceiptResult `json:"result"`
	Draft          ReceiptDraft  `json:"draft"`
	TransactionIDs []int         `json:"transaction_ids"`
	Error          string        `json:"error"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
	ImageKey       string        `json:"-"`
	ImageSHA256    string        `json:"-"`
	ContentType    string        `json:"-"`
	LeaseToken     string        `json:"-"`
	Attempts       int           `json:"-"`
	ImageAvailable bool          `json:"-"`
	RetryRevision  int           `json:"-"`
	QueuedBy       int64         `json:"-"`
	QueuedAt       time.Time     `json:"-"`
}

// ReceiptList ограничивает страницу и продолжает историю по монотонному номеру.
type ReceiptList struct {
	Receipts   []Receipt `json:"receipts"`
	NextBefore *int64    `json:"next_before,omitempty"`
}

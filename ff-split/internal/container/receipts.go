package container

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ivasnev/FinFlow/ff-split/internal/adapters/receiptimages"
	receipt_repository "github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/receipt"
	receipt_service "github.com/ivasnev/FinFlow/ff-split/internal/service/receipt"
)

// StartReceipts подключает опциональный OCR, сохраняя работоспособность установки без моделей.
func (c *Container) StartReceipts(ctx context.Context) (func(), error) {
	cfg := c.Config.Receipt
	if cfg.OCRURL == "" {
		return func() {}, nil
	}
	endpoint, err := url.Parse(cfg.OCRURL)
	if err != nil || endpoint.Host == "" || (endpoint.Scheme != "http" && endpoint.Scheme != "https") || endpoint.User != nil || endpoint.RawQuery != "" || endpoint.Fragment != "" || len(cfg.Token) < 32 || cfg.Workers < 1 || cfg.Workers > 8 {
		return nil, fmt.Errorf("invalid receipt OCR configuration")
	}
	startup, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	store, err := receiptimages.NewStore(startup, cfg.Endpoint, cfg.AccessKey, cfg.SecretKey, cfg.Bucket, cfg.Secure)
	if err != nil {
		return nil, fmt.Errorf("initialize receipt storage: %w", err)
	}
	client := &http.Client{Timeout: 90 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	pipeline := receipt_service.NewService(receipt_repository.NewRepository(c.DB), store, &receipt_service.HTTPRecognizer{URL: cfg.OCRURL, Token: cfg.Token, Client: client}, c.TransactionService)
	c.ServerHandler.SetReceiptService(pipeline)
	return pipeline.Start(ctx, cfg.Workers), nil
}

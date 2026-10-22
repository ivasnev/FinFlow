package receiptimages

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Store использует отдельный приватный бакет без публичных или подписанных ссылок.
type Store struct {
	client *minio.Client
	bucket string
}

// NewStore не меняет политику существующего бакета и требует явные учётные данные.
func NewStore(ctx context.Context, endpoint, key, secret, bucket string, secure bool) (*Store, error) {
	if endpoint == "" || key == "" || secret == "" || bucket == "" {
		return nil, fmt.Errorf("receipt storage configuration required")
	}
	client, err := minio.New(endpoint, &minio.Options{Creds: credentials.NewStaticV4(key, secret, ""), Secure: secure})
	if err != nil {
		return nil, err
	}
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			response := minio.ToErrorResponse(err)
			if response.Code != "BucketAlreadyOwnedByYou" {
				return nil, err
			}
		}
	}
	policy, err := client.GetBucketPolicy(ctx, bucket)
	if err != nil && minio.ToErrorResponse(err).Code != "NoSuchBucketPolicy" {
		return nil, fmt.Errorf("check receipt bucket policy: %w", err)
	}
	if strings.TrimSpace(policy) != "" {
		return nil, fmt.Errorf("receipt bucket must have no anonymous access policy")
	}
	return &Store{client: client, bucket: bucket}, nil
}

// Put сохраняет объект с непредсказуемым серверным ключом.
func (s *Store) Put(ctx context.Context, key string, data []byte, contentType string) error {
	if !strings.HasPrefix(key, "receipts/") {
		return fmt.Errorf("invalid object key")
	}
	_, err := s.client.PutObject(ctx, s.bucket, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{ContentType: contentType})
	return err
}

// Read ограничивает размер даже при повреждении метаданных в объектном хранилище.
func (s *Store) Read(ctx context.Context, key string) ([]byte, error) {
	object, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer object.Close()
	data, err := io.ReadAll(io.LimitReader(object, (10<<20)+1))
	if err != nil {
		return nil, err
	}
	if len(data) > 10<<20 {
		return nil, fmt.Errorf("receipt image exceeds limit")
	}
	return data, nil
}

// Delete безопасен при повторной обработке outbox.
func (s *Store) Delete(ctx context.Context, key string) error {
	return s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
}

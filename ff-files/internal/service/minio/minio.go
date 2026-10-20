package minio

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-files/internal/common/config"
	"github.com/ivasnev/FinFlow/ff-files/internal/service"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"golang.org/x/sync/errgroup"
)

// minioServiceImpl реализация интерфейса MinIO
type minioServiceImpl struct {
	mc  *minio.Client // Клиент Minio
	cfg *config.MinIO // Config Minio
}

// NewMinioService создает новый экземпляр MinIO сервиса
func NewMinioService() service.MinIO {
	return &minioServiceImpl{}
}

// InitMinio подключается к Minio и создает бакет, если не существует
func (m *minioServiceImpl) InitMinio(cfg *config.MinIO) error {
	ctx := context.Background()
	m.cfg = cfg

	// Указываем прокси-сервер
	proxyURL, err := url.Parse(cfg.InternalEndpoint)
	if err != nil {
		log.Fatalf("Invalid proxy URL: %v", err)
	}

	fmt.Println(proxyURL)

	// Создаём http.Transport с прокси
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	// Подключение к Minio с использованием имени пользователя и пароля
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(cfg.RootUser, cfg.RootPassword, ""),
		Secure:    cfg.UseSSL,
		Transport: transport,
	})
	if err != nil {
		return err
	}

	// Установка подключения Minio
	m.mc = client

	// Проверка наличия бакета и его создание, если не существует
	exists, err := m.mc.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return err
	}
	if !exists {
		err := m.mc.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateOne сохраняет один файл, учитывая отмену запроса и ограничение размера.
func (m *minioServiceImpl) CreateOne(ctx context.Context, file service.FileData) (*service.FileUploadResult, error) {
	files, err := prepareUploads([]service.FileData{file})
	if err != nil {
		return nil, err
	}
	return m.upload(ctx, files[0])
}

// upload передаёт ограниченный поток в SDK, сохраняя один ID объекта и ссылки.
func (m *minioServiceImpl) upload(ctx context.Context, file service.FileData) (*service.FileUploadResult, error) {
	objectID := uuid.NewString()
	_, err := m.mc.PutObject(ctx, m.cfg.BucketName, objectID, io.LimitReader(file.Reader, file.Size), file.Size, minio.PutObjectOptions{ContentType: file.ContentType})
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании объекта %s: %w", file.FileName, err)
	}
	link, err := m.mc.PresignedGetObject(ctx, m.cfg.BucketName, objectID, time.Second*time.Duration(m.cfg.FileTimeExpiration), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании URL для объекта %s: %w", file.FileName, err)
	}
	return &service.FileUploadResult{ObjectID: objectID, Link: link.String()}, nil
}

// prepareUploads проверяет весь пакет до первого обращения к хранилищу.
func prepareUploads(data []service.FileData) ([]service.FileData, error) {
	if len(data) == 0 || len(data) > service.MaxUploadFiles {
		return nil, fmt.Errorf("число файлов должно быть от 1 до %d", service.MaxUploadFiles)
	}
	files := append([]service.FileData(nil), data...)
	var total int64
	for i := range files {
		file := &files[i]
		if file.Reader == nil {
			file.Reader = bytes.NewReader(file.Data)
			file.Size = int64(len(file.Data))
			if file.ContentType == "" {
				file.ContentType = http.DetectContentType(file.Data)
			}
		}
		if file.Size < 0 || file.Size > service.MaxUploadBytes-total {
			return nil, fmt.Errorf("общий размер файлов превышает %d байт", service.MaxUploadBytes)
		}
		total += file.Size
	}
	return files, nil
}

// CreateMany загружает пакет максимум четырьмя параллельными потоками.
func (m *minioServiceImpl) CreateMany(ctx context.Context, data []service.FileData) ([]*service.FileUploadResult, error) {
	files, err := prepareUploads(data)
	if err != nil {
		return nil, err
	}
	objects := make([]*service.FileUploadResult, len(files))
	group, uploadCtx := errgroup.WithContext(ctx)
	group.SetLimit(4)
	for index, file := range files {
		group.Go(func() error {
			if err := uploadCtx.Err(); err != nil {
				return err
			}
			result, err := m.upload(uploadCtx, file)
			if err != nil {
				return err
			}
			objects[index] = result
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return nil, fmt.Errorf("ошибка при создании объектов: %w", err)
	}
	return objects, nil
}

// GetOne получает один объект из бакета Minio по его идентификатору
func (m *minioServiceImpl) GetOne(objectID string) (string, error) {
	// Получение предварительно подписанного URL для доступа к объекту Minio
	url, err := m.mc.PresignedGetObject(context.Background(), m.cfg.BucketName, objectID, time.Second*24*60*60, nil)
	if err != nil {
		return "", fmt.Errorf("ошибка при получении URL для объекта %s: %v", objectID, err)
	}

	return url.String(), nil
}

// GetMany получает несколько объектов из бакета Minio по их идентификаторам
func (m *minioServiceImpl) GetMany(objectIDs []string) ([]string, error) {
	// Создание каналов для передачи URL-адресов объектов и ошибок
	urlCh := make(chan string, len(objectIDs))
	errCh := make(chan error, len(objectIDs))

	var wg sync.WaitGroup
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запуск горутин для получения URL-адресов каждого объекта
	for _, objectID := range objectIDs {
		wg.Add(1)
		go func(objectID string) {
			defer wg.Done()
			url, err := m.GetOne(objectID)
			if err != nil {
				errCh <- fmt.Errorf("ошибка при получении объекта %s: %v", objectID, err)
				cancel()
				return
			}
			urlCh <- url
		}(objectID)
	}

	// Закрытие каналов после завершения всех горутин
	go func() {
		wg.Wait()
		close(urlCh)
		close(errCh)
	}()

	// Сбор URL-адресов объектов и ошибок из каналов
	var urls []string
	var errs []error
	for url := range urlCh {
		urls = append(urls, url)
	}
	for err := range errCh {
		errs = append(errs, err)
	}

	// Проверка наличия ошибок
	if len(errs) > 0 {
		return nil, fmt.Errorf("ошибки при получении объектов: %v", errs)
	}

	return urls, nil
}

// DeleteOne удаляет один объект из бакета Minio по его идентификатору
func (m *minioServiceImpl) DeleteOne(objectID string) error {
	// Удаление объекта из бакета Minio
	err := m.mc.RemoveObject(context.Background(), m.cfg.BucketName, objectID, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}

// DeleteMany удаляет несколько объектов из бакета Minio по их идентификаторам
func (m *minioServiceImpl) DeleteMany(objectIDs []string) error {
	// Создание канала для передачи ошибок
	errCh := make(chan error, len(objectIDs))
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запуск горутин для удаления каждого объекта
	for _, objectID := range objectIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			err := m.mc.RemoveObject(ctx, m.cfg.BucketName, id, minio.RemoveObjectOptions{})
			if err != nil {
				errCh <- fmt.Errorf("ошибка при удалении объекта %s: %v", id, err)
				cancel()
			}
		}(objectID)
	}

	// Ожидание завершения всех горутин и закрытие канала с ошибками
	go func() {
		wg.Wait()
		close(errCh)
	}()

	// Сбор ошибок из канала
	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}

	// Проверка наличия ошибок
	if len(errs) > 0 {
		return fmt.Errorf("ошибки при удалении объектов: %v", errs)
	}

	return nil
}

// GetMetadata получает метаданные файла из MinIO
func (m *minioServiceImpl) GetMetadata(objectID string) (*service.FileMetadata, error) {
	// Получаем информацию об объекте
	objInfo, err := m.mc.StatObject(context.Background(), m.cfg.BucketName, objectID, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении информации об объекте %s: %v", objectID, err)
	}

	// Создаем структуру метаданных
	metadata := &service.FileMetadata{
		FileID:      objectID,
		Filename:    objInfo.Key,
		Size:        objInfo.Size,
		ContentType: objInfo.ContentType,
		UploadDate:  objInfo.LastModified,
		OwnerID:     objInfo.Owner.DisplayName,
		Metadata:    make(map[string]interface{}),
	}

	// Добавляем дополнительные метаданные из пользовательских тегов
	if objInfo.UserMetadata != nil {
		for key, value := range objInfo.UserMetadata {
			metadata.Metadata[key] = value
		}
	}

	return metadata, nil
}

// GenerateTemporaryUrl генерирует временную ссылку для доступа к файлу
func (m *minioServiceImpl) GenerateTemporaryUrl(objectID string, expiresInSeconds int) (*service.TemporaryURLResult, error) {
	// Проверяем, что файл существует
	_, err := m.mc.StatObject(context.Background(), m.cfg.BucketName, objectID, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("файл с ID %s не найден: %v", objectID, err)
	}

	// Генерируем временную ссылку
	url, err := m.mc.PresignedGetObject(
		context.Background(),
		m.cfg.BucketName,
		objectID,
		time.Duration(expiresInSeconds)*time.Second,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка при генерации временной ссылки для объекта %s: %v", objectID, err)
	}

	// Вычисляем время истечения ссылки
	expiresAt := time.Now().Add(time.Duration(expiresInSeconds) * time.Second)

	return &service.TemporaryURLResult{
		URL:       url.String(),
		ExpiresAt: expiresAt,
	}, nil
}

package minio

import (
	"bytes"
	"context"
	"fmt"
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

// CreateOne создает один объект в бакете Minio
func (m *minioServiceImpl) CreateOne(file service.FileData) (*service.FileUploadResult, error) {
	// Генерация уникального идентификатора для нового объекта
	objectID := uuid.New().String()

	// Создание потока данных для загрузки в бакет Minio
	reader := bytes.NewReader(file.Data)

	// Определение ContentType на основе расширения файла
	contentType := http.DetectContentType(file.Data)

	// Параметры для загрузки объекта
	opts := minio.PutObjectOptions{
		ContentType: contentType,
	}

	// Загрузка данных в бакет Minio
	_, err := m.mc.PutObject(context.Background(), m.cfg.BucketName, objectID, reader, int64(len(file.Data)), opts)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании объекта %s: %v", file.FileName, err)
	}

	// Генерация подписанного URL для получения объекта
	url, err := m.mc.PresignedGetObject(
		context.Background(), m.cfg.BucketName, objectID,
		time.Second*time.Duration(m.cfg.FileTimeExpiration),
		nil)

	if err != nil {
		return nil, fmt.Errorf("ошибка при создании URL для объекта %s: %v", file.FileName, err)
	}

	return &service.FileUploadResult{
		ObjectID: objectID,
		Link:     url.String(),
	}, nil
}

// CreateMany создает несколько объектов в хранилище MinIO
func (m *minioServiceImpl) CreateMany(data map[string]service.FileData) ([]*service.FileUploadResult, error) {
	objects := make([]*service.FileUploadResult, 0, len(data))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	urlCh := make(chan string, len(data))
	errCh := make(chan error, 1)

	var wg sync.WaitGroup

	// Запуск горутин для создания каждого объекта
	for objectID, file := range data {
		wg.Add(1)
		go func(objectID string, file service.FileData) {
			defer wg.Done()

			// Определение ContentType на основе содержимого файла
			contentType := http.DetectContentType(file.Data)

			_, err := m.mc.PutObject(ctx, m.cfg.BucketName, objectID, bytes.NewReader(file.Data), int64(len(file.Data)), minio.PutObjectOptions{
				ContentType: contentType,
			})
			if err != nil {
				select {
				case errCh <- err:
					cancel()
				default:
				}
				return
			}

			// Получение URL для загруженного объекта
			url, err := m.mc.PresignedGetObject(ctx, m.cfg.BucketName, objectID, time.Hour*time.Duration(m.cfg.FileTimeExpiration), nil)
			if err != nil {
				select {
				case errCh <- err:
					cancel()
				default:
				}
				return
			}

			select {
			case urlCh <- url.String():
			case <-ctx.Done():
				return
			}
		}(objectID, file)
	}

	// Ожидание завершения всех горутин и закрытие каналов
	go func() {
		wg.Wait()
		close(urlCh)
		close(errCh)
	}()

	// Сбор URL-адресов из канала
	for url := range urlCh {
		object := &service.FileUploadResult{
			ObjectID: uuid.New().String(),
			Link:     url,
		}
		objects = append(objects, object)
	}

	// Проверка наличия ошибок
	if err := <-errCh; err != nil {
		return nil, fmt.Errorf("ошибка при создании объектов: %v", err)
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

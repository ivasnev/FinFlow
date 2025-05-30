package minio

import (
	"bytes"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-files/internal/common/config"
	helpers2 "github.com/ivasnev/FinFlow/ff-files/internal/service/minio/helpers"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// minioService реализация интерфейса MinioClient
type minioService struct {
	mc  *minio.Client // Клиент Minio
	cfg *config.MinIO // Config Minio
}

// NewMinioService создает новый экземпляр Minio Client
func NewMinioService() MinioServiceInterface {
	return &minioService{} // Возвращает новый экземпляр minioService с указанным именем бакета
}

// InitMinio подключается к Minio и создает бакет, если не существует
// Бакет - это контейнер для хранения объектов в Minio. Он представляет собой пространство имен, в котором можно хранить и организовывать файлы и папки.
func (m *minioService) InitMinio(cfg *config.MinIO) error {
	// Создание контекста с возможностью отмены операции
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

// CreateOne создает один объект в бакете Minio.
// Метод принимает структуру fileData, которая содержит имя файла и его данные.
// В случае успешной загрузки данных в бакет, метод возвращает nil, иначе возвращает ошибку.
// Все операции выполняются в контексте задачи.
func (m *minioService) CreateOne(file helpers2.FileDataType) (*helpers2.CreatedObject, error) {
	// Генерация уникального идентификатора для нового объекта.
	objectID := uuid.New().String()

	// Создание потока данных для загрузки в бакет Minio.
	reader := bytes.NewReader(file.Data)

	// Определение ContentType на основе расширения файла (например, jpg/jpeg, png, etc.)
	contentType := http.DetectContentType(file.Data)

	// Параметры для загрузки объекта
	opts := minio.PutObjectOptions{
		ContentType: contentType,
	}

	// Загрузка данных в бакет Minio с использованием контекста для возможности отмены операции.
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

	return &helpers2.CreatedObject{
		ObjectID: objectID,
		Link:     url.String(),
	}, nil
}

// CreateMany создает несколько объектов в хранилище MinIO из переданных данных.
// Если происходит ошибка при создании объекта, метод возвращает ошибку,
// указывающую на неудачные объекты.
func (m *minioService) CreateMany(data map[string]helpers2.FileDataType) ([]*helpers2.CreatedObject, error) {
	objects := make([]*helpers2.CreatedObject, 0, len(data)) // Массив для хранения созданных объектов

	ctx, cancel := context.WithCancel(context.Background()) // Создание контекста с возможностью отмены операции.
	defer cancel()                                          // Отложенный вызов функции отмены контекста при завершении функции CreateMany.

	urlCh := make(chan string, len(data)) // Канал для передачи URL-адресов с размером, равным количеству переданных данных.
	errCh := make(chan error, 1)          // Канал для передачи ошибок (буферизированный, чтобы избежать блокировок).

	var wg sync.WaitGroup // WaitGroup для ожидания завершения всех горутин.

	// Запуск горутин для создания каждого объекта.
	for objectID, file := range data {
		wg.Add(1) // Увеличение счетчика WaitGroup перед запуском каждой горутины.
		go func(objectID string, file helpers2.FileDataType) {
			defer wg.Done() // Уменьшение счетчика WaitGroup после завершения горутины.

			// Определение ContentType на основе содержимого файла
			contentType := http.DetectContentType(file.Data)

			_, err := m.mc.PutObject(ctx, m.cfg.BucketName, objectID, bytes.NewReader(file.Data), int64(len(file.Data)), minio.PutObjectOptions{
				ContentType: contentType,
			}) // Создание объекта в бакете MinIO.
			if err != nil {
				select {
				case errCh <- err: // Отправка ошибки в канал ошибок.
					cancel() // Отмена операции при возникновении ошибки.
				default:
				}
				return
			}

			// Получение URL для загруженного объекта
			url, err := m.mc.PresignedGetObject(ctx, m.cfg.BucketName, objectID, time.Hour*time.Duration(m.cfg.FileTimeExpiration), nil)
			if err != nil {
				select {
				case errCh <- err: // Отправка ошибки в канал ошибок.
					cancel() // Отмена операции при возникновении ошибки.
				default:
				}
				return
			}

			select {
			case urlCh <- url.String(): // Отправка URL-адреса в канал с URL-адресами.
			case <-ctx.Done(): // Выход, если контекст завершен.
				return
			}
		}(objectID, file) // Передача данных объекта в анонимную горутину.
	}

	// Ожидание завершения всех горутин и закрытие канала с URL-адресами.
	go func() {
		wg.Wait()    // Блокировка до тех пор, пока счетчик WaitGroup не станет равным 0.
		close(urlCh) // Закрытие канала с URL-адресами после завершения всех горутин.
		close(errCh) // Закрытие канала ошибок после завершения всех горутин.
	}()

	// Сбор URL-адресов из канала.
	for url := range urlCh {
		object := &helpers2.CreatedObject{
			ObjectID: uuid.New().String(), // Генерация уникального идентификатора для объекта
			Link:     url,
		}
		objects = append(objects, object) // Добавление объекта в массив объектов.
	}

	// Проверка наличия ошибок.
	if err := <-errCh; err != nil {
		return nil, fmt.Errorf("ошибка при создании объектов: %v", err)
	}

	return objects, nil
}

// GetOne получает один объект из бакета Minio по его идентификатору.
// Он принимает строку `objectID` в качестве параметра и возвращает срез байт данных объекта и ошибку, если такая возникает.
func (m *minioService) GetOne(objectID string) (string, error) {
	// Получение предварительно подписанного URL для доступа к объекту Minio.
	url, err := m.mc.PresignedGetObject(context.Background(), m.cfg.BucketName, objectID, time.Second*24*60*60, nil)
	if err != nil {
		return "", fmt.Errorf("ошибка при получении URL для объекта %s: %v", objectID, err)
	}

	return url.String(), nil
}

// GetMany получает несколько объектов из бакета Minio по их идентификаторам.
func (m *minioService) GetMany(objectIDs []string) ([]string, error) {
	// Создание каналов для передачи URL-адресов объектов и ошибок
	urlCh := make(chan string, len(objectIDs))                  // Канал для URL-адресов объектов
	errCh := make(chan helpers2.OperationError, len(objectIDs)) // Канал для ошибок

	var wg sync.WaitGroup                                 // WaitGroup для ожидания завершения всех горутин
	_, cancel := context.WithCancel(context.Background()) // Создание контекста с возможностью отмены операции
	defer cancel()                                        // Отложенный вызов функции отмены контекста при завершении функции GetMany

	// Запуск горутин для получения URL-адресов каждого объекта.
	for _, objectID := range objectIDs {
		wg.Add(1) // Увеличение счетчика WaitGroup перед запуском каждой горутины
		go func(objectID string) {
			defer wg.Done()                // Уменьшение счетчика WaitGroup после завершения горутины
			url, err := m.GetOne(objectID) // Получение URL-адреса объекта по его идентификатору с помощью метода GetOne
			if err != nil {
				errCh <- helpers2.OperationError{ObjectID: objectID, Error: fmt.Errorf("ошибка при получении объекта %s: %v", objectID, err)} // Отправка ошибки в канал с ошибками
				cancel()                                                                                                                      // Отмена операции при возникновении ошибки
				return
			}
			urlCh <- url // Отправка URL-адреса объекта в канал с URL-адресами
		}(objectID) // Передача идентификатора объекта в анонимную горутину
	}

	// Закрытие каналов после завершения всех горутин.
	go func() {
		wg.Wait()    // Блокировка до тех пор, пока счетчик WaitGroup не станет равным 0
		close(urlCh) // Закрытие канала с URL-адресами после завершения всех горутин
		close(errCh) // Закрытие канала с ошибками после завершения всех горутин
	}()

	// Сбор URL-адресов объектов и ошибок из каналов.
	var urls []string // Массив для хранения URL-адресов
	var errs []error  // Массив для хранения ошибок
	for url := range urlCh {
		urls = append(urls, url) // Добавление URL-адреса в массив URL-адресов
	}
	for opErr := range errCh {
		errs = append(errs, opErr.Error) // Добавление ошибки в массив ошибок
	}

	// Проверка наличия ошибок.
	if len(errs) > 0 {
		return nil, fmt.Errorf("ошибки при получении объектов: %v", errs) // Возврат ошибки, если возникли ошибки при получении объектов
	}

	return urls, nil // Возврат массива URL-адресов, если ошибок не возникло
}

// DeleteOne удаляет один объект из бакета Minio по его идентификатору.
func (m *minioService) DeleteOne(objectID string) error {
	// Удаление объекта из бакета Minio.
	err := m.mc.RemoveObject(context.Background(), m.cfg.BucketName, objectID, minio.RemoveObjectOptions{})
	if err != nil {
		return err // Возвращаем ошибку, если не удалось удалить объект.
	}
	return nil // Возвращаем nil, если объект успешно удалён.
}

// DeleteMany удаляет несколько объектов из бакета Minio по их идентификаторам с использованием горутин.
func (m *minioService) DeleteMany(objectIDs []string) error {
	// Создание канала для передачи ошибок с размером, равным количеству объектов для удаления
	errCh := make(chan helpers2.OperationError, len(objectIDs)) // Канал для ошибок
	var wg sync.WaitGroup                                       // WaitGroup для ожидания завершения всех горутин

	ctx, cancel := context.WithCancel(context.Background()) // Создание контекста с возможностью отмены операции
	defer cancel()                                          // Отложенный вызов функции отмены контекста при завершении функции DeleteMany

	// Запуск горутин для удаления каждого объекта.
	for _, objectID := range objectIDs {
		wg.Add(1) // Увеличение счетчика WaitGroup перед запуском каждой горутины
		go func(id string) {
			defer wg.Done()                                                                  // Уменьшение счетчика WaitGroup после завершения горутины
			err := m.mc.RemoveObject(ctx, m.cfg.BucketName, id, minio.RemoveObjectOptions{}) // Удаление объекта с использованием Minio клиента
			if err != nil {
				errCh <- helpers2.OperationError{ObjectID: id, Error: fmt.Errorf("ошибка при удалении объекта %s: %v", id, err)} // Отправка ошибки в канал с ошибками
				cancel()                                                                                                         // Отмена операции при возникновении ошибки
			}
		}(objectID) // Передача идентификатора объекта в анонимную горутину
	}

	// Ожидание завершения всех горутин и закрытие канала с ошибками.
	go func() {
		wg.Wait()    // Блокировка до тех пор, пока счетчик WaitGroup не станет равным 0
		close(errCh) // Закрытие канала с ошибками после завершения всех горутин
	}()

	// Сбор ошибок из канала.
	var errs []error // Массив для хранения ошибок
	for opErr := range errCh {
		errs = append(errs, opErr.Error) // Добавление ошибки в массив ошибок
	}

	// Проверка наличия ошибок.
	if len(errs) > 0 {
		return fmt.Errorf("ошибки при удалении объектов: %v", errs) // Возврат ошибки, если возникли ошибки при удалении объектов
	}

	return nil // Возврат nil, если ошибок не возникло
}

package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

// Storage определяет интерфейс для работы с файловым хранилищем
type Storage interface {
	// Save сохраняет файл в хранилище и возвращает путь к нему
	Save(ctx context.Context, file io.Reader, filename string) (string, error)
	
	// Get возвращает файл из хранилища
	Get(ctx context.Context, path string) (io.ReadCloser, error)
	
	// Delete удаляет файл из хранилища
	Delete(ctx context.Context, path string) error
	
	// GetURL возвращает URL для доступа к файлу
	GetURL(ctx context.Context, path string) (string, error)
}

// LocalStorage реализует Storage для локального хранения файлов
type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) (*LocalStorage, error) {
	// Создаем директорию, если она не существует
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, err
	}

	return &LocalStorage{
		basePath: basePath,
	}, nil
}

func (s *LocalStorage) Save(ctx context.Context, file io.Reader, filename string) (string, error) {
	// Генерируем путь для сохранения файла
	// Используем структуру директорий вида: basePath/ab/cd/abcdef...
	// где ab и cd - первые 2 и следующие 2 символа имени файла
	if len(filename) < 4 {
		filename = filename + "0000"
	}
	
	dir := filepath.Join(s.basePath, filename[:2], filename[2:4])
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	path := filepath.Join(dir, filename)
	
	// Создаем файл
	dst, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Копируем содержимое
	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	// Возвращаем относительный путь от basePath
	relPath, err := filepath.Rel(s.basePath, path)
	if err != nil {
		return "", err
	}

	return relPath, nil
}

func (s *LocalStorage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.basePath, path)
	return os.Open(fullPath)
}

func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	fullPath := filepath.Join(s.basePath, path)
	return os.Remove(fullPath)
}

func (s *LocalStorage) GetURL(ctx context.Context, path string) (string, error) {
	// Для локального хранилища возвращаем относительный путь
	return path, nil
} 
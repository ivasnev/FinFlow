package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-files/internal/service"
	"github.com/ivasnev/FinFlow/ff-files/pkg/api"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// ServerHandler реализует интерфейс api.ServerInterface
type ServerHandler struct {
	minioService service.MinIO
}

// NewServerHandler создает новый ServerHandler
func NewServerHandler(minioService service.MinIO) *ServerHandler {
	return &ServerHandler{
		minioService: minioService,
	}
}

// UploadFile обрабатывает запрос на загрузку одного файла
func (h *ServerHandler) UploadFile(c *gin.Context) {
	// Получаем файл из запроса
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: "No file is received",
			Code:  &[]int{http.StatusBadRequest}[0],
		})
		return
	}

	// Открываем файл для чтения
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Error: "Unable to open the file",
			Code:  &[]int{http.StatusInternalServerError}[0],
		})
		return
	}
	defer f.Close()

	// Читаем содержимое файла в байтовый срез
	fileBytes, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Error: "Unable to read the file",
			Code:  &[]int{http.StatusInternalServerError}[0],
		})
		return
	}

	// Создаем структуру FileData для хранения данных файла
	fileData := service.FileData{
		FileName: file.Filename,
		Data:     fileBytes,
	}

	// Сохраняем файл в MinIO
	result, err := h.minioService.CreateOne(fileData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Error: "Unable to save the file",
			Code:  &[]int{http.StatusInternalServerError}[0],
		})
		return
	}

	// Конвертируем в API тип
	objectID, err := uuid.Parse(result.ObjectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Error: "Unable to parse object ID",
			Code:  &[]int{http.StatusInternalServerError}[0],
		})
		return
	}

	// Возвращаем успешный ответ
	c.JSON(http.StatusOK, api.FileUploadResponse{
		Status:  http.StatusOK,
		Message: "File uploaded successfully",
		Data: api.CreatedObject{
			ObjectId: objectID,
			Link:     result.Link,
		},
	})
}

// UploadFiles обрабатывает запрос на загрузку нескольких файлов
func (h *ServerHandler) UploadFiles(c *gin.Context) {
	// Получаем multipart форму из запроса
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: "Invalid form",
			Code:  &[]int{http.StatusBadRequest}[0],
		})
		return
	}

	// Получаем файлы из формы
	files := form.File["files"]
	if files == nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: "No files received",
			Code:  &[]int{http.StatusBadRequest}[0],
		})
		return
	}

	// Подготавливаем данные для загрузки
	fileDataMap := make(map[string]service.FileData)
	for _, file := range files {
		// Открываем файл
		f, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, api.ErrorResponse{
				Error: "Unable to open file: " + file.Filename,
				Code:  &[]int{http.StatusInternalServerError}[0],
			})
			return
		}

		// Читаем содержимое файла
		fileBytes, err := io.ReadAll(f)
		f.Close()
		if err != nil {
			c.JSON(http.StatusInternalServerError, api.ErrorResponse{
				Error: "Unable to read file: " + file.Filename,
				Code:  &[]int{http.StatusInternalServerError}[0],
			})
			return
		}

		// Добавляем в карту
		fileDataMap[file.Filename] = service.FileData{
			FileName: file.Filename,
			Data:     fileBytes,
		}
	}

	// Загружаем файлы
	results, err := h.minioService.CreateMany(fileDataMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Error: "Unable to save files",
			Code:  &[]int{http.StatusInternalServerError}[0],
		})
		return
	}

	// Конвертируем результаты в API типы
	createdObjects := make([]api.CreatedObject, len(results))
	for i, result := range results {
		objectID, err := uuid.Parse(result.ObjectID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, api.ErrorResponse{
				Error: "Unable to parse object ID",
				Code:  &[]int{http.StatusInternalServerError}[0],
			})
			return
		}

		createdObjects[i] = api.CreatedObject{
			ObjectId: objectID,
			Link:     result.Link,
		}
	}

	// Возвращаем успешный ответ
	c.JSON(http.StatusOK, api.FilesUploadResponse{
		Status:  http.StatusOK,
		Message: "Files uploaded successfully",
		Data:    createdObjects,
	})
}

// GetFile обрабатывает запрос на получение одного файла
func (h *ServerHandler) GetFile(c *gin.Context, objectID openapi_types.UUID) {
	// Получаем URL файла
	url, err := h.minioService.GetOne(objectID.String())
	if err != nil {
		c.JSON(http.StatusNotFound, api.ErrorResponse{
			Error: "File not found",
			Code:  &[]int{http.StatusNotFound}[0],
		})
		return
	}

	// Возвращаем URL файла
	c.JSON(http.StatusOK, api.FileGetResponse{
		Status:  http.StatusOK,
		Message: "File received successfully",
		Data:    url,
	})
}

// GetFiles обрабатывает запрос на получение нескольких файлов
func (h *ServerHandler) GetFiles(c *gin.Context) {
	var req api.FileIDsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: "Invalid request body",
			Code:  &[]int{http.StatusBadRequest}[0],
		})
		return
	}

	// Конвертируем UUID в строки
	objectIDs := make([]string, len(req.FileIDs))
	for i, id := range req.FileIDs {
		objectIDs[i] = id.String()
	}

	// Получаем URL файлов
	urls, err := h.minioService.GetMany(objectIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Error: "Unable to retrieve files",
			Code:  &[]int{http.StatusInternalServerError}[0],
		})
		return
	}

	// Возвращаем URL файлов
	c.JSON(http.StatusOK, api.FilesGetResponse{
		Status:  http.StatusOK,
		Message: "Files received successfully",
		Data:    urls,
	})
}

// DeleteFile обрабатывает запрос на удаление одного файла
func (h *ServerHandler) DeleteFile(c *gin.Context, objectID openapi_types.UUID) {
	// Удаляем файл
	err := h.minioService.DeleteOne(objectID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Error: "Unable to delete file",
			Code:  &[]int{http.StatusInternalServerError}[0],
		})
		return
	}

	// Возвращаем успешный ответ
	c.JSON(http.StatusOK, api.FileDeleteResponse{
		Status:  http.StatusOK,
		Message: "File deleted successfully",
	})
}

// DeleteFiles обрабатывает запрос на удаление нескольких файлов
func (h *ServerHandler) DeleteFiles(c *gin.Context) {
	var req api.FileIDsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: "Invalid request body",
			Code:  &[]int{http.StatusBadRequest}[0],
		})
		return
	}

	// Конвертируем UUID в строки
	objectIDs := make([]string, len(req.FileIDs))
	for i, id := range req.FileIDs {
		objectIDs[i] = id.String()
	}

	// Удаляем файлы
	err := h.minioService.DeleteMany(objectIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Error: "Unable to delete files",
			Code:  &[]int{http.StatusInternalServerError}[0],
		})
		return
	}

	// Возвращаем успешный ответ
	c.JSON(http.StatusOK, api.FilesDeleteResponse{
		Status:  http.StatusOK,
		Message: "Files deleted successfully",
	})
}

// GetFileMetadata обрабатывает запрос на получение метаданных файла
func (h *ServerHandler) GetFileMetadata(c *gin.Context, fileId openapi_types.UUID) {
	// Получаем метаданные файла
	metadata, err := h.minioService.GetMetadata(fileId.String())
	if err != nil {
		c.JSON(http.StatusNotFound, api.ErrorResponse{
			Error: "File not found",
			Code:  &[]int{http.StatusNotFound}[0],
		})
		return
	}

	// Конвертируем в API тип
	fileID, err := uuid.Parse(metadata.FileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Error: "Unable to parse file ID",
			Code:  &[]int{http.StatusInternalServerError}[0],
		})
		return
	}

	// Возвращаем метаданные
	c.JSON(http.StatusOK, api.FileMetadataResponse{
		Status:  http.StatusOK,
		Message: "File metadata retrieved successfully",
		Data: api.FileMetadata{
			FileId:      fileID,
			Filename:    metadata.Filename,
			Size:        metadata.Size,
			ContentType: metadata.ContentType,
			UploadDate:  metadata.UploadDate,
			OwnerId:     metadata.OwnerID,
			Metadata:    &metadata.Metadata,
		},
	})
}

// GenerateTemporaryUrl обрабатывает запрос на генерацию временной ссылки
func (h *ServerHandler) GenerateTemporaryUrl(c *gin.Context, fileId openapi_types.UUID, params api.GenerateTemporaryUrlParams) {
	// Получаем время истечения из параметров (по умолчанию 1 час)
	expiresInSeconds := 3600
	if params.ExpiresIn != nil {
		expiresInSeconds = *params.ExpiresIn
	}

	// Генерируем временную ссылку
	result, err := h.minioService.GenerateTemporaryUrl(fileId.String(), expiresInSeconds)
	if err != nil {
		c.JSON(http.StatusNotFound, api.ErrorResponse{
			Error: "File not found",
			Code:  &[]int{http.StatusNotFound}[0],
		})
		return
	}

	// Возвращаем временную ссылку
	c.JSON(http.StatusOK, api.TemporaryUrlResponse{
		Status:  http.StatusOK,
		Message: "Temporary URL generated successfully",
		Data: api.TemporaryUrlResponseData{
			Url:       result.URL,
			ExpiresAt: result.ExpiresAt,
		},
	})
}

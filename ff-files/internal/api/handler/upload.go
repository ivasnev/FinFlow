package handler

import (
	"bufio"
	"errors"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-files/internal/service"
	"github.com/ivasnev/FinFlow/ff-files/pkg/api"
)

const (
	uploadMemoryBytes  = 1 << 20
	maxUploadBodyBytes = service.MaxUploadBytes + (1 << 20)
)

// uploadError возвращает согласованный ответ без передачи внутренних ошибок клиенту.
func uploadError(c *gin.Context, status int, message string) {
	c.JSON(status, api.ErrorResponse{Error: message, Code: &status})
}

// parseUpload ограничивает тело до multipart-разбора, затем проверяет весь пакет до открытия файлов.
func parseUpload(c *gin.Context, field string, maxFiles int) ([]*multipart.FileHeader, bool) {
	if c.Request.ContentLength > maxUploadBodyBytes {
		uploadError(c, http.StatusRequestEntityTooLarge, "Upload exceeds request size limit")
		return nil, false
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadBodyBytes)
	if err := c.Request.ParseMultipartForm(uploadMemoryBytes); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			uploadError(c, http.StatusRequestEntityTooLarge, "Upload exceeds request size limit")
		} else {
			uploadError(c, http.StatusBadRequest, "Invalid multipart form")
		}
		return nil, false
	}
	files := c.Request.MultipartForm.File[field]
	if len(files) > maxFiles {
		uploadError(c, http.StatusRequestEntityTooLarge, "Too many files")
		return nil, false
	}
	if len(files) == 0 || len(c.Request.MultipartForm.File) != 1 {
		uploadError(c, http.StatusBadRequest, "Unexpected or missing file field")
		return nil, false
	}
	var total int64
	for _, file := range files {
		if file.Size < 0 || file.Size > service.MaxUploadBytes-total {
			uploadError(c, http.StatusRequestEntityTooLarge, "Upload exceeds total file size limit")
			return nil, false
		}
		total += file.Size
	}
	return files, true
}

// cleanupUpload удаляет временные multipart-файлы при любом исходе запроса.
func cleanupUpload(c *gin.Context) {
	if c.Request.MultipartForm != nil {
		_ = c.Request.MultipartForm.RemoveAll()
	}
}

// openUpload определяет тип по первым 512 байтам и передаёт файл потоком без полной копии в памяти.
func openUpload(header *multipart.FileHeader) (service.FileData, io.Closer, error) {
	file, err := header.Open()
	if err != nil {
		return service.FileData{}, nil, err
	}
	reader := bufio.NewReader(file)
	prefix, err := reader.Peek(512)
	if err != nil && !errors.Is(err, io.EOF) {
		_ = file.Close()
		return service.FileData{}, nil, err
	}
	return service.FileData{FileName: header.Filename, Reader: reader, Size: header.Size, ContentType: http.DetectContentType(prefix)}, file, nil
}

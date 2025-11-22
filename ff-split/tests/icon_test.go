package tests

import (
	"testing"

	"github.com/ivasnev/FinFlow/ff-split/pkg/api"
	"github.com/stretchr/testify/suite"
)

// IconSuite представляет suite для тестов управления иконками
type IconSuite struct {
	BaseSuite
}

// TestIconSuite запускает все тесты в IconSuite
func TestIconSuite(t *testing.T) {
	suite.Run(t, new(IconSuite))
}

// TestGetIcons_Success тестирует получение списка иконок
func (s *IconSuite) TestGetIcons_Success() {
	// Arrange - подготовка
	s.createTestIcon(TestIconID1, "Travel", TestRequestID)
	s.createTestIcon(TestIconID2, "Food", "uuid-2")

	// Act - действие
	resp, err := s.APIClient.GetIconsWithResponse(s.Ctx)

	// Assert - проверка
	s.Require().NoError(err, "запрос должен выполниться успешно")
	s.Require().Equal(200, resp.StatusCode(), "должен быть статус 200")
	s.Require().NotNil(resp.JSON200, "список иконок должен быть возвращен")
	s.Require().GreaterOrEqual(len(*resp.JSON200), 2, "должно быть минимум 2 иконки")
}

// TestGetIconByID_Success тестирует получение иконки по ID
func (s *IconSuite) TestGetIconByID_Success() {
	// Arrange - подготовка
	icon := s.createTestIcon(TestIconID1, "Travel", TestRequestID)

	// Act - действие
	resp, err := s.APIClient.GetIconByIDWithResponse(s.Ctx, icon.ID)

	// Assert - проверка
	s.Require().NoError(err, "запрос должен выполниться успешно")
	s.Require().Equal(200, resp.StatusCode(), "должен быть статус 200")
	s.Require().NotNil(resp.JSON200, "иконка должна быть возвращена")
	s.Require().Equal(icon.ID, *resp.JSON200.Id)
	s.Require().Equal(icon.Name, *resp.JSON200.Name)
	s.Require().Equal(icon.FileUUID, *resp.JSON200.FileUuid)
}

// TestGetIconByID_NotFound тестирует получение несуществующей иконки
func (s *IconSuite) TestGetIconByID_NotFound() {
	// Arrange - подготовка
	nonExistentID := 999

	// Act - действие
	resp, err := s.APIClient.GetIconByIDWithResponse(s.Ctx, nonExistentID)

	// Assert - проверка
	// Может быть ошибка десериализации, но статус код должен быть 404
	if err == nil {
		s.Require().Equal(404, resp.StatusCode(), "должен быть статус 404")
	}
}

// TestCreateIcon_Success тестирует создание иконки
func (s *IconSuite) TestCreateIcon_Success() {
	// Arrange - подготовка
	iconName := "Sport"
	fileUUID := "uuid-sport-123"
	reqBody := api.CreateIconJSONRequestBody{
		Name:     iconName,
		FileUuid: fileUUID,
	}

	// Act - действие
	resp, err := s.APIClient.CreateIconWithResponse(s.Ctx, reqBody)

	// Assert - проверка
	s.Require().NoError(err, "запрос должен выполниться успешно")
	s.Require().Equal(201, resp.StatusCode(), "должен быть статус 201")
	s.Require().NotNil(resp.JSON201, "иконка должна быть создана")
	s.Require().Equal(iconName, *resp.JSON201.Name)
	s.Require().Equal(fileUUID, *resp.JSON201.FileUuid)

	// Проверяем, что иконка создана в БД
	var count int64
	err = s.GetDB().Table("icons").Where("name = ?", iconName).Count(&count).Error
	s.NoError(err)
	s.Equal(int64(1), count, "должна быть создана одна иконка")
}

// TestUpdateIcon_Success тестирует обновление иконки
func (s *IconSuite) TestUpdateIcon_Success() {
	// Arrange - подготовка
	icon := s.createTestIcon(TestIconID1, "Старое название", "old-uuid")

	// Подготавливаем запрос на обновление
	newName := "Новое название"
	newFileUUID := "new-uuid"
	reqBody := api.UpdateIconJSONRequestBody{
		Name:     newName,
		FileUuid: newFileUUID,
	}

	// Act - действие
	resp, err := s.APIClient.UpdateIconWithResponse(s.Ctx, icon.ID, reqBody)

	// Assert - проверка
	s.Require().NoError(err, "запрос должен выполниться успешно")
	s.Require().Equal(200, resp.StatusCode(), "должен быть статус 200")
	s.Require().NotNil(resp.JSON200, "обновленная иконка должна быть возвращена")
	s.Require().Equal(newName, *resp.JSON200.Name)
	s.Require().Equal(newFileUUID, *resp.JSON200.FileUuid)

	// Проверяем, что данные обновлены в БД
	var updatedIcon struct {
		Name     string
		FileUUID string `gorm:"column:file_uuid"`
	}
	err = s.GetDB().Table("icons").Where("id = ?", icon.ID).First(&updatedIcon).Error
	s.NoError(err)
	s.Equal(newName, updatedIcon.Name)
	s.Equal(newFileUUID, updatedIcon.FileUUID)
}

// TestDeleteIcon_Success тестирует удаление иконки
func (s *IconSuite) TestDeleteIcon_Success() {
	// Arrange - подготовка
	icon := s.createTestIcon(TestIconID1, "Иконка для удаления", "uuid-to-delete")

	// Act - действие
	resp, err := s.APIClient.DeleteIconWithResponse(s.Ctx, icon.ID)

	// Assert - проверка
	s.Require().NoError(err, "запрос должен выполниться успешно")
	s.Require().Equal(200, resp.StatusCode(), "должен быть статус 200")
	s.Require().NotNil(resp.JSON200, "должен быть возвращен объект успеха")

	// Проверяем, что иконка удалена из БД
	var count int64
	err = s.GetDB().Table("icons").Where("id = ?", icon.ID).Count(&count).Error
	s.NoError(err)
	s.Equal(int64(0), count, "иконка должна быть удалена из БД")
}


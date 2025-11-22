package tests

import (
	"testing"
	"time"

	"github.com/ivasnev/FinFlow/ff-id/internal/models"
	"github.com/ivasnev/FinFlow/ff-id/internal/service"
	"github.com/stretchr/testify/suite"
)

// FriendSuite представляет suite для тестов системы друзей
type FriendSuite struct {
	BaseSuite
}

// TestFriendSuite запускает все тесты в FriendSuite
func TestFriendSuite(t *testing.T) {
	suite.Run(t, new(FriendSuite))
}

// createTestUser создает тестового пользователя в БД с использованием константного времени
func (s *FriendSuite) createTestUser(id int64, email, nickname, name string) *models.User {
	now := s.TimeProvider.Now()

	user := &models.User{
		ID:        id,
		Email:     email,
		Nickname:  nickname,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if name != "" {
		user.Name.String = name
		user.Name.Valid = true
	}

	// Создаем пользователя напрямую в БД
	err := s.GetDB().Exec(`
		INSERT INTO users (id, email, nickname, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, user.ID, user.Email, user.Nickname, user.Name, user.CreatedAt, user.UpdatedAt).Error
	s.NoError(err, "не удалось создать тестового пользователя")

	return user
}

// TestAddFriend_Success тестирует успешную отправку заявки в друзья
func (s *FriendSuite) TestAddFriend_Success() {
	// Arrange - подготовка
	user1 := s.createTestUser(TestUserID1, TestEmail1, TestNickname1, TestName1)
	user2 := s.createTestUser(TestUserID2, TestEmail2, TestNickname2, TestName2)

	// Act - действие
	err := s.Container.FriendService.AddFriend(s.Ctx, user1.ID, service.AddFriendRequest{
		FriendNickname: user2.Nickname,
	})

	// Assert - проверка
	s.Require().NoError(err, "отправка заявки должна пройти успешно")

	// Проверяем, что заявка создана в БД
	var count int64
	err = s.GetDB().Table("user_friends").
		Where("user_id = ? AND friend_id = ? AND status = ?", user1.ID, user2.ID, "pending").
		Count(&count).Error
	s.Require().NoError(err, "запрос должен выполниться успешно")
	s.Require().Equal(int64(1), count, "должна быть создана одна заявка")
}

// TestAddFriend_AlreadyExists тестирует отправку повторной заявки в друзья
func (s *FriendSuite) TestAddFriend_AlreadyExists() {
	// Arrange - подготовка
	user1 := s.createTestUser(TestUserID1, TestEmail1, TestNickname1, TestName1)
	user2 := s.createTestUser(TestUserID2, TestEmail2, TestNickname2, TestName2)

	// Отправляем первую заявку
	err := s.Container.FriendService.AddFriend(s.Ctx, user1.ID, service.AddFriendRequest{
		FriendNickname: user2.Nickname,
	})
	s.NoError(err, "первая заявка должна пройти успешно")

	// Act - действие (повторная заявка)
	err = s.Container.FriendService.AddFriend(s.Ctx, user1.ID, service.AddFriendRequest{
		FriendNickname: user2.Nickname,
	})

	// Assert - проверка
	s.Require().Error(err, "должна быть ошибка при повторной заявке")
	s.Require().Contains(err.Error(), "уже отправлена", "ошибка должна содержать информацию о существующей заявке")
}

// TestAddFriend_ToSelf тестирует попытку добавить себя в друзья
func (s *FriendSuite) TestAddFriend_ToSelf() {
	// Используем s.Ctx из BaseSuite

	// Создаем пользователя
	user1 := s.createTestUser(1, "user1@example.com", "user1", "User One")

	// Пытаемся добавить себя в друзья
	err := s.Container.FriendService.AddFriend(s.Ctx, user1.ID, service.AddFriendRequest{
		FriendNickname: user1.Nickname,
	})

	s.Error(err, "должна быть ошибка при попытке добавить себя в друзья")
}

// TestAddFriend_NonExistentUser тестирует отправку заявки несуществующему пользователю
func (s *FriendSuite) TestAddFriend_NonExistentUser() {
	// Используем s.Ctx из BaseSuite

	// Создаем пользователя
	user1 := s.createTestUser(1, "user1@example.com", "user1", "User One")

	// Пытаемся отправить заявку несуществующему пользователю
	err := s.Container.FriendService.AddFriend(s.Ctx, user1.ID, service.AddFriendRequest{
		FriendNickname: "nonexistent",
	})

	s.Error(err, "должна быть ошибка при отправке заявки несуществующему пользователю")
	s.Contains(err.Error(), "не найден", "ошибка должна содержать информацию о ненайденном пользователе")
}

// TestAcceptFriendRequest_Success тестирует успешное принятие заявки в друзья
func (s *FriendSuite) TestAcceptFriendRequest_Success() {
	// Используем s.Ctx из BaseSuite

	// Создаем двух пользователей
	user1 := s.createTestUser(1, "user1@example.com", "user1", "User One")
	user2 := s.createTestUser(2, "user2@example.com", "user2", "User Two")

	// Создаем заявку в друзья от user1 к user2
	err := s.GetDB().Exec(`
		INSERT INTO user_friends (user_id, friend_id, status, created_at)
		VALUES ($1, $2, 'pending', $3)
	`, user1.ID, user2.ID, time.Now()).Error
	s.NoError(err, "не удалось создать заявку")

	// user2 принимает заявку
	err = s.Container.FriendService.AcceptFriendRequest(s.Ctx, user2.ID, service.FriendActionRequest{
		UserID: user1.ID,
		Action: "accept",
	})

	s.NoError(err, "принятие заявки должно пройти успешно")

	// Проверяем, что статус изменен на accepted
	var status string
	err = s.GetDB().Table("user_friends").
		Select("status").
		Where("user_id = ? AND friend_id = ?", user1.ID, user2.ID).
		Scan(&status).Error
	s.NoError(err, "запрос должен выполниться успешно")
	s.Equal("accepted", status, "статус должен быть accepted")

	// Проверяем, что создана обратная связь
	var count int64
	err = s.GetDB().Table("user_friends").
		Where("user_id = ? AND friend_id = ? AND status = ?", user2.ID, user1.ID, "accepted").
		Count(&count).Error
	s.NoError(err, "запрос должен выполниться успешно")
	s.Equal(int64(1), count, "должна быть создана обратная связь")
}

// TestRejectFriendRequest_Success тестирует успешное отклонение заявки в друзья
func (s *FriendSuite) TestRejectFriendRequest_Success() {
	// Используем s.Ctx из BaseSuite

	// Создаем двух пользователей
	user1 := s.createTestUser(1, "user1@example.com", "user1", "User One")
	user2 := s.createTestUser(2, "user2@example.com", "user2", "User Two")

	// Создаем заявку в друзья от user1 к user2
	err := s.GetDB().Exec(`
		INSERT INTO user_friends (user_id, friend_id, status, created_at)
		VALUES ($1, $2, 'pending', $3)
	`, user1.ID, user2.ID, time.Now()).Error
	s.NoError(err, "не удалось создать заявку")

	// user2 отклоняет заявку
	err = s.Container.FriendService.RejectFriendRequest(s.Ctx, user2.ID, service.FriendActionRequest{
		UserID: user1.ID,
		Action: "reject",
	})

	s.NoError(err, "отклонение заявки должно пройти успешно")

	// Проверяем, что статус изменен на rejected
	var status string
	err = s.GetDB().Table("user_friends").
		Select("status").
		Where("user_id = ? AND friend_id = ?", user1.ID, user2.ID).
		Scan(&status).Error
	s.NoError(err, "запрос должен выполниться успешно")
	s.Equal("rejected", status, "статус должен быть rejected")
}

// TestBlockUser_Success тестирует успешную блокировку пользователя
func (s *FriendSuite) TestBlockUser_Success() {
	// Используем s.Ctx из BaseSuite

	// Создаем двух пользователей
	user1 := s.createTestUser(1, "user1@example.com", "user1", "User One")
	user2 := s.createTestUser(2, "user2@example.com", "user2", "User Two")

	// Создаем заявку в друзья от user1 к user2
	err := s.GetDB().Exec(`
		INSERT INTO user_friends (user_id, friend_id, status, created_at)
		VALUES ($1, $2, 'pending', $3)
	`, user1.ID, user2.ID, time.Now()).Error
	s.NoError(err, "не удалось создать заявку")

	// user2 блокирует user1
	err = s.Container.FriendService.BlockUser(s.Ctx, user2.ID, service.FriendActionRequest{
		UserID: user1.ID,
		Action: "block",
	})

	s.NoError(err, "блокировка должна пройти успешно")

	// Проверяем, что создана связь с блокировкой от user2 к user1
	var status string
	err = s.GetDB().Table("user_friends").
		Select("status").
		Where("user_id = ? AND friend_id = ?", user2.ID, user1.ID).
		Scan(&status).Error
	s.NoError(err, "запрос должен выполниться успешно")
	s.Equal("blocked", status, "статус должен быть blocked")
}

// TestRemoveFriend_Success тестирует успешное удаление друга
func (s *FriendSuite) TestRemoveFriend_Success() {
	// Используем s.Ctx из BaseSuite

	// Создаем двух пользователей
	user1 := s.createTestUser(1, "user1@example.com", "user1", "User One")
	user2 := s.createTestUser(2, "user2@example.com", "user2", "User Two")

	// Создаем связь дружбы (accepted)
	err := s.GetDB().Exec(`
		INSERT INTO user_friends (user_id, friend_id, status, created_at)
		VALUES ($1, $2, 'accepted', $3), ($4, $5, 'accepted', $6)
	`, user1.ID, user2.ID, time.Now(), user2.ID, user1.ID, time.Now()).Error
	s.NoError(err, "не удалось создать связь дружбы")

	// user1 удаляет user2 из друзей
	err = s.Container.FriendService.RemoveFriend(s.Ctx, user1.ID, user2.ID)

	s.NoError(err, "удаление друга должно пройти успешно")

	// Проверяем, что обе связи удалены
	var count int64
	err = s.GetDB().Table("user_friends").
		Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
			user1.ID, user2.ID, user2.ID, user1.ID).
		Count(&count).Error
	s.NoError(err, "запрос должен выполниться успешно")
	s.Equal(int64(0), count, "обе связи должны быть удалены")
}

// TestRemoveFriend_NotFriends тестирует удаление несуществующей дружбы
func (s *FriendSuite) TestRemoveFriend_NotFriends() {
	// Используем s.Ctx из BaseSuite

	// Создаем двух пользователей
	user1 := s.createTestUser(1, "user1@example.com", "user1", "User One")
	user2 := s.createTestUser(2, "user2@example.com", "user2", "User Two")

	// Пытаемся удалить друга, которого нет
	err := s.Container.FriendService.RemoveFriend(s.Ctx, user1.ID, user2.ID)

	s.Error(err, "должна быть ошибка при удалении несуществующей дружбы")
	s.Contains(err.Error(), "not found", "ошибка должна содержать информацию о ненайденной дружбе")
}

// TestGetFriends_Success тестирует успешное получение списка друзей
func (s *FriendSuite) TestGetFriends_Success() {
	// Используем s.Ctx из BaseSuite

	// Создаем пользователей
	user1 := s.createTestUser(1, "user1@example.com", "user1", "User One")
	user2 := s.createTestUser(2, "user2@example.com", "user2", "User Two")
	user3 := s.createTestUser(3, "user3@example.com", "user3", "User Three")

	// Создаем связи дружбы
	err := s.GetDB().Exec(`
		INSERT INTO user_friends (user_id, friend_id, status, created_at)
		VALUES 
			($1, $2, 'accepted', $3),
			($4, $5, 'accepted', $6),
			($7, $8, 'accepted', $9),
			($10, $11, 'accepted', $12)
	`, user1.ID, user2.ID, time.Now(),
		user2.ID, user1.ID, time.Now(),
		user1.ID, user3.ID, time.Now(),
		user3.ID, user1.ID, time.Now()).Error
	s.NoError(err, "не удалось создать связи дружбы")

	// Получаем список друзей user1
	friendsList, err := s.Container.FriendService.GetFriends(s.Ctx, user1.Nickname, service.FriendsQueryParams{
		Page:     1,
		PageSize: 20,
		Status:   "accepted",
	})

	s.NoError(err, "получение списка друзей должно пройти успешно")
	s.NotNil(friendsList, "список друзей должен быть возвращен")
	s.Equal(2, len(friendsList.Objects), "должно быть 2 друга")
	s.Equal(int64(2), friendsList.Total, "общее количество должно быть 2")

	// Проверяем, что оба друга присутствуют
	friendIDs := make(map[int64]bool)
	for _, friend := range friendsList.Objects {
		friendIDs[friend.UserID] = true
	}
	s.True(friendIDs[user2.ID], "user2 должен быть в списке друзей")
	s.True(friendIDs[user3.ID], "user3 должен быть в списке друзей")
}

// TestGetFriends_WithPagination тестирует получение списка друзей с пагинацией
func (s *FriendSuite) TestGetFriends_WithPagination() {
	// Используем s.Ctx из BaseSuite

	// Создаем пользователя и несколько друзей
	user1 := s.createTestUser(1, "user1@example.com", "user1", "User One")

	// Создаем 5 друзей
	for i := 2; i <= 6; i++ {
		friend := s.createTestUser(int64(i),
			"friend"+string(rune('0'+i))+"@example.com",
			"friend"+string(rune('0'+i)),
			"Friend "+string(rune('0'+i)))

		// Создаем связь дружбы
		err := s.GetDB().Exec(`
			INSERT INTO user_friends (user_id, friend_id, status, created_at)
			VALUES ($1, $2, 'accepted', $3), ($4, $5, 'accepted', $6)
		`, user1.ID, friend.ID, time.Now(), friend.ID, user1.ID, time.Now()).Error
		s.NoError(err, "не удалось создать связь дружбы")
	}

	// Получаем первую страницу (3 записи)
	friendsList, err := s.Container.FriendService.GetFriends(s.Ctx, user1.Nickname, service.FriendsQueryParams{
		Page:     1,
		PageSize: 3,
		Status:   "accepted",
	})

	s.NoError(err, "получение списка друзей должно пройти успешно")
	s.Equal(3, len(friendsList.Objects), "должно быть 3 друга на первой странице")
	s.Equal(int64(5), friendsList.Total, "общее количество должно быть 5")

	// Получаем вторую страницу
	friendsList2, err := s.Container.FriendService.GetFriends(s.Ctx, user1.Nickname, service.FriendsQueryParams{
		Page:     2,
		PageSize: 3,
		Status:   "accepted",
	})

	s.NoError(err, "получение списка друзей должно пройти успешно")
	s.Equal(2, len(friendsList2.Objects), "должно быть 2 друга на второй странице")
	s.Equal(int64(5), friendsList2.Total, "общее количество должно быть 5")
}

// TestGetFriendRequests_Incoming тестирует получение входящих заявок в друзья
func (s *FriendSuite) TestGetFriendRequests_Incoming() {
	// Используем s.Ctx из BaseSuite

	// Создаем пользователей
	user1 := s.createTestUser(1, "user1@example.com", "user1", "User One")
	user2 := s.createTestUser(2, "user2@example.com", "user2", "User Two")
	user3 := s.createTestUser(3, "user3@example.com", "user3", "User Three")

	// Создаем входящие заявки для user1
	err := s.GetDB().Exec(`
		INSERT INTO user_friends (user_id, friend_id, status, created_at)
		VALUES ($1, $2, 'pending', $3), ($4, $5, 'pending', $6)
	`, user2.ID, user1.ID, time.Now(), user3.ID, user1.ID, time.Now()).Error
	s.NoError(err, "не удалось создать заявки")

	// Получаем входящие заявки для user1
	requestsList, err := s.Container.FriendService.GetFriendRequests(s.Ctx, user1.ID, 1, 20, true)

	s.NoError(err, "получение заявок должно пройти успешно")
	s.NotNil(requestsList, "список заявок должен быть возвращен")
	s.Equal(2, len(requestsList.Objects), "должно быть 2 входящие заявки")
	s.Equal(int64(2), requestsList.Total, "общее количество должно быть 2")
}

// TestGetFriendRequests_Outgoing тестирует получение исходящих заявок в друзья
func (s *FriendSuite) TestGetFriendRequests_Outgoing() {
	// Используем s.Ctx из BaseSuite

	// Создаем пользователей
	user1 := s.createTestUser(1, "user1@example.com", "user1", "User One")
	user2 := s.createTestUser(2, "user2@example.com", "user2", "User Two")
	user3 := s.createTestUser(3, "user3@example.com", "user3", "User Three")

	// Создаем исходящие заявки от user1
	err := s.GetDB().Exec(`
		INSERT INTO user_friends (user_id, friend_id, status, created_at)
		VALUES ($1, $2, 'pending', $3), ($4, $5, 'pending', $6)
	`, user1.ID, user2.ID, time.Now(), user1.ID, user3.ID, time.Now()).Error
	s.NoError(err, "не удалось создать заявки")

	// Получаем исходящие заявки для user1
	requestsList, err := s.Container.FriendService.GetFriendRequests(s.Ctx, user1.ID, 1, 20, false)

	s.NoError(err, "получение заявок должно пройти успешно")
	s.NotNil(requestsList, "список заявок должен быть возвращен")
	s.Equal(2, len(requestsList.Objects), "должно быть 2 исходящие заявки")
	s.Equal(int64(2), requestsList.Total, "общее количество должно быть 2")
}

// TestGetFriends_EmptyList тестирует получение пустого списка друзей
func (s *FriendSuite) TestGetFriends_EmptyList() {
	// Используем s.Ctx из BaseSuite

	// Создаем пользователя без друзей
	user1 := s.createTestUser(1, "user1@example.com", "user1", "User One")

	// Получаем список друзей
	friendsList, err := s.Container.FriendService.GetFriends(s.Ctx, user1.Nickname, service.FriendsQueryParams{
		Page:     1,
		PageSize: 20,
		Status:   "accepted",
	})

	s.NoError(err, "получение списка друзей должно пройти успешно")
	s.NotNil(friendsList, "список друзей должен быть возвращен")
	s.Equal(0, len(friendsList.Objects), "список должен быть пустым")
	s.Equal(int64(0), friendsList.Total, "общее количество должно быть 0")
}

package tests

import (
	"time"

	"github.com/ivasnev/FinFlow/ff-id/internal/service"
)

// TestAddFriend_ResendRejected проверяет повторную отправку без дублирования записи и последующее принятие.
func (s *FriendSuite) TestAddFriend_ResendRejected() {
	first := s.createTestUser(TestUserID1, TestEmail1, TestNickname1, TestName1)
	second := s.createTestUser(TestUserID2, TestEmail2, TestNickname2, TestName2)
	request := service.AddFriendRequest{FriendNickname: second.Nickname}
	s.Require().NoError(s.Container.FriendService.AddFriend(s.Ctx, first.ID, request))
	var original struct{ ID int64 }
	s.Require().NoError(s.GetDB().Table("user_friends").Take(&original).Error)
	s.Require().NoError(s.GetDB().Exec("UPDATE user_friends SET created_at='2000-01-01'").Error)
	s.Require().NoError(s.Container.FriendService.RejectFriendRequest(s.Ctx, second.ID, service.FriendActionRequest{UserID: first.ID}))
	s.Require().NoError(s.Container.FriendService.AddFriend(s.Ctx, first.ID, request))
	var rows []struct {
		ID        int64
		Status    string
		CreatedAt time.Time
	}
	s.Require().NoError(s.GetDB().Table("user_friends").Find(&rows).Error)
	s.Require().Len(rows, 1)
	s.Equal(original.ID, rows[0].ID)
	s.Equal(service.FriendStatusPending, rows[0].Status)
	s.True(rows[0].CreatedAt.After(time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)))
	s.Require().NoError(s.Container.FriendService.AcceptFriendRequest(s.Ctx, second.ID, service.FriendActionRequest{UserID: first.ID}))
	var count int64
	s.Require().NoError(s.GetDB().Table("user_friends").Where("status = ?", service.FriendStatusAccepted).Count(&count).Error)
	s.Equal(int64(2), count)
}

// TestAddFriend_ResendPreservesBlock проверяет отказ при чужой блокировке и защиту статусов на уровне репозитория.
func (s *FriendSuite) TestAddFriend_ResendPreservesBlock() {
	first := s.createTestUser(TestUserID1, TestEmail1, TestNickname1, TestName1)
	second := s.createTestUser(TestUserID2, TestEmail2, TestNickname2, TestName2)
	s.Require().NoError(s.Container.FriendService.AddFriend(s.Ctx, first.ID, service.AddFriendRequest{FriendNickname: second.Nickname}))
	s.Require().NoError(s.Container.FriendService.RejectFriendRequest(s.Ctx, second.ID, service.FriendActionRequest{UserID: first.ID}))
	s.Require().NoError(s.Container.FriendService.BlockUser(s.Ctx, second.ID, service.FriendActionRequest{UserID: first.ID}))
	s.Require().Error(s.Container.FriendService.AddFriend(s.Ctx, first.ID, service.AddFriendRequest{FriendNickname: second.Nickname}))
	for _, status := range []string{service.FriendStatusBlocked, service.FriendStatusAccepted, service.FriendStatusPending} {
		s.Require().NoError(s.GetDB().Exec("UPDATE user_friends SET status = ? WHERE user_id = ?", status, first.ID).Error)
		s.Require().Error(s.Container.FriendRepository.AddFriend(s.Ctx, first.ID, second.ID))
		var current string
		s.Require().NoError(s.GetDB().Table("user_friends").Where("user_id = ?", first.ID).Pluck("status", &current).Error)
		s.Equal(status, current)
	}
}

package tests

import "github.com/ivasnev/FinFlow/ff-id/internal/service"

// TestRemoveFriend_PreservesForeignBlock проверяет, что удаление дружбы не позволяет обойти чужую блокировку.
func (s *FriendSuite) TestRemoveFriend_PreservesForeignBlock() {
	first := s.createTestUser(TestUserID1, TestEmail1, TestNickname1, TestName1)
	second := s.createTestUser(TestUserID2, TestEmail2, TestNickname2, TestName2)
	s.Require().NoError(s.Container.FriendService.AddFriend(s.Ctx, first.ID, service.AddFriendRequest{FriendNickname: second.Nickname}))
	s.Require().NoError(s.Container.FriendService.BlockUser(s.Ctx, second.ID, service.FriendActionRequest{UserID: first.ID}))
	s.Require().NoError(s.Container.FriendService.RemoveFriend(s.Ctx, first.ID, second.ID))
	for i := 0; i < 2; i++ {
		s.Require().Error(s.Container.FriendService.AddFriend(s.Ctx, first.ID, service.AddFriendRequest{FriendNickname: second.Nickname}))
		var relations []struct {
			UserID, FriendID int64
			Status           string
		}
		s.Require().NoError(s.GetDB().Table("user_friends").Find(&relations).Error)
		s.Require().Len(relations, 1)
		s.Equal(second.ID, relations[0].UserID)
		s.Equal(first.ID, relations[0].FriendID)
		s.Equal(service.FriendStatusBlocked, relations[0].Status)
		_ = s.Container.FriendService.RemoveFriend(s.Ctx, first.ID, second.ID)
	}
	// Сценарий: снять собственную блокировку может только установивший её пользователь.
	s.Require().NoError(s.Container.FriendService.RemoveFriend(s.Ctx, second.ID, first.ID))
	s.Require().NoError(s.Container.FriendService.AddFriend(s.Ctx, first.ID, service.AddFriendRequest{FriendNickname: second.Nickname}))
}

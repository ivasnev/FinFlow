package tests

import (
	"context"
	"sync"
	"time"

	"github.com/ivasnev/FinFlow/ff-id/internal/service"
)

// TestAcceptFriendRequest_Crossed проверяет принятие уже существующей встречной заявки.
func (s *FriendSuite) TestAcceptFriendRequest_Crossed() {
	first := s.createTestUser(TestUserID1, TestEmail1, TestNickname1, TestName1)
	second := s.createTestUser(TestUserID2, TestEmail2, TestNickname2, TestName2)
	s.Require().NoError(s.Container.FriendService.AddFriend(s.Ctx, first.ID, service.AddFriendRequest{FriendNickname: second.Nickname}))
	s.Require().NoError(s.Container.FriendService.AddFriend(s.Ctx, second.ID, service.AddFriendRequest{FriendNickname: first.Nickname}))
	s.Require().NoError(s.Container.FriendService.AcceptFriendRequest(s.Ctx, second.ID, service.FriendActionRequest{UserID: first.ID}))
	var rows []struct{ Status string }
	s.Require().NoError(s.GetDB().Table("user_friends").Find(&rows).Error)
	s.Require().Len(rows, 2)
	for _, row := range rows {
		s.Equal(service.FriendStatusAccepted, row.Status)
	}
}

// TestAcceptFriendRequest_ConcurrentCrossed проверяет одинаковый порядок блокировки встречных строк.
func (s *FriendSuite) TestAcceptFriendRequest_ConcurrentCrossed() {
	first := s.createTestUser(TestUserID1, TestEmail1, TestNickname1, TestName1)
	second := s.createTestUser(TestUserID2, TestEmail2, TestNickname2, TestName2)
	s.Require().NoError(s.Container.FriendService.AddFriend(s.Ctx, first.ID, service.AddFriendRequest{FriendNickname: second.Nickname}))
	s.Require().NoError(s.Container.FriendService.AddFriend(s.Ctx, second.ID, service.AddFriendRequest{FriendNickname: first.Nickname}))
	ctx, cancel := context.WithTimeout(s.Ctx, 5*time.Second)
	defer cancel()
	start := make(chan struct{})
	errors := make(chan error, 12)
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			from, to := first.ID, second.ID
			if i%2 == 1 {
				from, to = to, from
			}
			errors <- s.Container.FriendRepository.CreateMutualFriendship(ctx, from, to)
		}(i)
	}
	close(start)
	wg.Wait()
	close(errors)
	for err := range errors {
		s.Require().NoError(err)
	}
	var count int64
	s.Require().NoError(s.GetDB().Table("user_friends").Where("status = ?", service.FriendStatusAccepted).Count(&count).Error)
	s.Equal(int64(2), count)
}

// TestAcceptFriendRequest_PreservesBlock проверяет отказ без частичной записи при блокировке любого направления.
func (s *FriendSuite) TestAcceptFriendRequest_PreservesBlock() {
	first := s.createTestUser(TestUserID1, TestEmail1, TestNickname1, TestName1)
	second := s.createTestUser(TestUserID2, TestEmail2, TestNickname2, TestName2)
	for _, blocked := range []int64{first.ID, second.ID} {
		s.Require().NoError(s.GetDB().Exec("DELETE FROM user_friends").Error)
		s.Require().NoError(s.GetDB().Exec(`INSERT INTO user_friends(user_id,friend_id,status) VALUES(?,?,'pending'),(?,?,'pending')`, first.ID, second.ID, second.ID, first.ID).Error)
		s.Require().NoError(s.GetDB().Exec("UPDATE user_friends SET status='blocked' WHERE user_id = ?", blocked).Error)
		s.Require().Error(s.Container.FriendRepository.CreateMutualFriendship(s.Ctx, second.ID, first.ID))
		var statuses []string
		s.Require().NoError(s.GetDB().Table("user_friends").Order("status").Pluck("status", &statuses).Error)
		s.Equal([]string{"blocked", "pending"}, statuses)
	}
}

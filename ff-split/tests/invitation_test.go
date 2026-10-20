package tests

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ivasnev/FinFlow/ff-split/internal/common/actor"
	domainerrors "github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	invitationrepo "github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/invitation"
	invitationservice "github.com/ivasnev/FinFlow/ff-split/internal/service/invitation"
)

// TestInvitationLifecycle проверяет отсутствие автоматического вступления, срок, хеш, отзыв и безопасный повтор.
func TestInvitationLifecycle(t *testing.T) {
	database, c := settlementFixture(t)
	ctx := actor.WithID(context.Background(), 101)
	s := invitationservice.NewService(invitationrepo.NewRepository(database.DB), c.UserService)
	if err := database.DB.Exec("INSERT INTO users(id,user_id,is_dummy,name_cashed) VALUES(4,404,false,'Новый участник')").Error; err != nil {
		t.Fatal(err)
	}
	row, token, err := s.Create(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(token) != 43 || row.CreatedBy != 101 || row.EventID != 1 || row.ExpiresAt.Before(time.Now().Add(6*24*time.Hour)) {
		t.Fatalf("invalid invitation: %+v", row)
	}
	var stored models.Invitation
	if err = database.DB.Take(&stored, row.ID).Error; err != nil {
		t.Fatal(err)
	}
	if len(stored.TokenHash) != 64 || strings.Contains(stored.TokenHash, token) {
		t.Fatal("raw secret stored")
	}
	outsider := actor.WithID(context.Background(), 404)
	preview, err := s.Preview(outsider, token)
	if err != nil || preview.IsMember || preview.EventName != "trip" {
		t.Fatalf("preview=%+v err=%v", preview, err)
	}
	var count int64
	database.DB.Table("user_event").Where("event_id=1 AND user_id=4").Count(&count)
	if count != 0 {
		t.Fatal("preview joined automatically")
	}
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id, err := s.Accept(outsider, token)
			if err == nil && id != 1 {
				err = errors.New("wrong event")
			}
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	database.DB.Table("user_event").Where("event_id=1 AND user_id=4").Count(&count)
	if count != 1 {
		t.Fatalf("membership count=%d", count)
	}
	database.DB.Table("event_history").Where("event_id=1 AND kind='member.added' AND actor_id=404").Count(&count)
	if count != 1 {
		t.Fatalf("join history count=%d", count)
	}
	if err = s.Revoke(actor.WithID(context.Background(), 202), 1, row.ID); !errors.Is(err, domainerrors.ErrForbidden) {
		t.Fatalf("foreign revocation=%v", err)
	}
	for i := 0; i < 2; i++ {
		if err = s.Revoke(ctx, 1, row.ID); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = s.Preview(outsider, token); !errors.Is(err, domainerrors.ErrInvitationUnavailable) {
		t.Fatalf("revoked preview=%v", err)
	}
	if _, err = s.Accept(outsider, token); !errors.Is(err, domainerrors.ErrInvitationUnavailable) {
		t.Fatalf("revoked accept=%v", err)
	}
	database.DB.Table("event_history").Where("event_id=1 AND kind='invitation.revoked'").Count(&count)
	if count != 1 {
		t.Fatalf("duplicate revocation history=%d", count)
	}
	expired, expiredToken, err := s.Create(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err = database.DB.Model(&models.Invitation{}).Where("id=?", expired.ID).UpdateColumn("expires_at", time.Now().Add(-time.Second)).Error; err != nil {
		t.Fatal(err)
	}
	if _, err = s.Accept(outsider, expiredToken); !errors.Is(err, domainerrors.ErrInvitationUnavailable) {
		t.Fatalf("expired accept=%v", err)
	}
	if _, err = s.Preview(outsider, "invalid"); !errors.Is(err, domainerrors.ErrInvitationUnavailable) {
		t.Fatalf("invalid preview=%v", err)
	}
}

// TestInvitationLimitAndRollback проверяет лимит при гонке и атомарность вступления при отказе журнала.
func TestInvitationLimitAndRollback(t *testing.T) {
	database, c := settlementFixture(t)
	ctx := actor.WithID(context.Background(), 101)
	s := invitationservice.NewService(invitationrepo.NewRepository(database.DB), c.UserService)
	var wg sync.WaitGroup
	results := make(chan error, 12)
	tokens := make(chan string, 12)
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, token, err := s.Create(ctx, 1)
			results <- err
			if err == nil {
				tokens <- token
			}
		}()
	}
	wg.Wait()
	close(results)
	close(tokens)
	success := 0
	for err := range results {
		if err == nil {
			success++
		} else {
			var validation *domainerrors.ValidationError
			if !errors.As(err, &validation) {
				t.Fatal(err)
			}
		}
	}
	if success != 5 {
		t.Fatalf("limit=%d", success)
	}
	if err := database.DB.Exec("INSERT INTO users(id,user_id,is_dummy) VALUES(4,404,false)").Error; err != nil {
		t.Fatal(err)
	}
	if err := database.DB.Exec(`CREATE FUNCTION fail_history() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'history unavailable'; END $$; CREATE TRIGGER fail_history BEFORE INSERT ON event_history FOR EACH ROW EXECUTE FUNCTION fail_history();`).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := s.Accept(actor.WithID(context.Background(), 404), <-tokens); err == nil {
		t.Fatal("joined without history")
	}
	var count int64
	database.DB.Table("user_event").Where("event_id=1 AND user_id=4").Count(&count)
	if count != 0 {
		t.Fatal("membership survived failed history")
	}
}

// TestInvitationCreatorRemoval не позволяет старой ссылке вернуть удалённого создателя или ожить при повторном добавлении.
func TestInvitationCreatorRemoval(t *testing.T) {
	database, c := settlementFixture(t)
	ctx := actor.WithID(context.Background(), 101)
	s := invitationservice.NewService(invitationrepo.NewRepository(database.DB), c.UserService)
	_, token, err := s.Create(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err = c.UserService.RemoveUserFromEvent(actor.WithID(context.Background(), 202), 1, 1); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Accept(ctx, token); !errors.Is(err, domainerrors.ErrInvitationUnavailable) {
		t.Fatalf("removed creator regained access: %v", err)
	}
	if err = c.UserService.AddUsersToEvent(actor.WithID(context.Background(), 202), []int64{1}, 1); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Preview(ctx, token); !errors.Is(err, domainerrors.ErrInvitationUnavailable) {
		t.Fatalf("old link revived: %v", err)
	}
}

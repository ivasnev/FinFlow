package tests

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/actor"
	domainerrors "github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	archiverepo "github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/archive"
	invitationrepo "github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/invitation"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
	archiveservice "github.com/ivasnev/FinFlow/ff-split/internal/service/archive"
	invitationservice "github.com/ivasnev/FinFlow/ff-split/internal/service/invitation"
	"gorm.io/gorm"
)

// purgeArchiveBatch выполняет тот же SQL, который запускает ночная задача, в одноразовом тестовом контейнере.
func purgeArchiveBatch(t *testing.T, database *TestDBContainer, wantSuccess bool) {
	t.Helper()
	ctx := context.Background()
	sql, err := os.ReadFile("../../scripts/stack/purge-events.sql")
	if err != nil {
		t.Fatal(err)
	}
	if err = database.Container.CopyToContainer(ctx, sql, "/tmp/purge-events.sql", 0600); err != nil {
		t.Fatal(err)
	}
	code, reader, err := database.Container.Exec(ctx, []string{"psql", "-X", "-qAt", "-U", "postgres", "-d", "ff_split_test", "-f", "/tmp/purge-events.sql"})
	if err != nil {
		t.Fatal(err)
	}
	output, _ := io.ReadAll(reader)
	if (code == 0) != wantSuccess {
		t.Fatalf("purge code=%d output=%s", code, output)
	}
}

// TestArchiveLifecycle проверяет сохранение расчётов, обратимость, запрет изменений и отзыв ссылок.
func TestArchiveLifecycle(t *testing.T) {
	database, c := settlementFixture(t)
	ctx := actor.WithID(context.Background(), 101)
	s := archiveservice.NewService(archiverepo.NewRepository(database.DB))
	invites := invitationservice.NewService(invitationrepo.NewRepository(database.DB), c.UserService)
	expense := settlementExpense(t, c, 1, 2, 100.03)
	_, token, err := invites.Create(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Archive(actor.WithID(ctx, 404), 1); !errors.Is(err, domainerrors.ErrForbidden) {
		t.Fatalf("foreign archive=%v", err)
	}
	if err = s.Archive(ctx, 1); err != nil {
		t.Fatal(err)
	}
	row, err := c.EventService.GetEventByID(ctx, 1)
	if err != nil || row.DeletedAt == nil || row.PurgeAt == nil {
		t.Fatalf("archive=%+v err=%v", row, err)
	}
	first := *row.DeletedAt
	if err = s.Archive(ctx, 1); err != nil {
		t.Fatal(err)
	}
	row, _ = c.EventService.GetEventByID(ctx, 1)
	if !first.Equal(*row.DeletedAt) {
		t.Fatal("retry extended deadline")
	}
	active, err := c.EventService.GetEventsByUserID(ctx, 1)
	if err != nil || len(active) != 0 {
		t.Fatalf("active list=%+v %v", active, err)
	}
	archived, err := s.List(ctx)
	if err != nil || len(archived) != 1 {
		t.Fatalf("archive list=%+v %v", archived, err)
	}
	foreign, err := s.List(actor.WithID(ctx, 404))
	if err != nil || len(foreign) != 0 {
		t.Fatalf("foreign list=%+v %v", foreign, err)
	}
	assertSettlementBalance(t, c, 1, 100.03)
	if _, err = c.TransactionService.UpdateTransaction(ctx, expense.ID, optimizationRequest(120)); !errors.Is(err, domainerrors.ErrEventArchived) {
		t.Fatalf("expense update in archive=%v", err)
	}
	if err = c.TransactionService.DeleteTransaction(ctx, expense.ID); !errors.Is(err, domainerrors.ErrEventArchived) {
		t.Fatalf("expense delete in archive=%v", err)
	}
	if _, err = c.SettlementService.Receive(ctx, 1, 101, service.SettlementRequest{FromUserID: 2, AmountCents: 100, RequestID: uuid.NewString()}); !errors.Is(err, domainerrors.ErrEventArchived) {
		t.Fatalf("settlement in archive=%v", err)
	}
	if _, err = invites.Preview(ctx, token); !errors.Is(err, domainerrors.ErrInvitationUnavailable) {
		t.Fatalf("invitation in archive=%v", err)
	}
	var revisionBefore, revisionAfter *int64
	database.DB.Table("events").Select("optimized_debt_revision").Where("id=1").Scan(&revisionBefore)
	if _, err = c.TransactionService.GetOptimizedDebtsByEventID(ctx, 1, nil); err != nil {
		t.Fatal(err)
	}
	database.DB.Table("events").Select("optimized_debt_revision").Where("id=1").Scan(&revisionAfter)
	if revisionBefore != nil || revisionAfter != nil {
		t.Fatal("archived GET wrote optimization cache")
	}
	if err = s.Restore(actor.WithID(ctx, 202), 1); err != nil {
		t.Fatal(err)
	}
	row, _ = c.EventService.GetEventByID(ctx, 1)
	if row.DeletedAt != nil || row.PurgeAt != nil {
		t.Fatal("restore retained deletion marker")
	}
	if _, err = invites.Preview(ctx, token); !errors.Is(err, domainerrors.ErrInvitationUnavailable) {
		t.Fatal("restore revived invitation")
	}
	if _, err = c.TransactionService.GetTransactionByID(ctx, expense.ID); err != nil {
		t.Fatal("expense disappeared on restore")
	}
}

// TestArchiveCalendarCutoff использует календарные месяцы и одинаковые результаты при разных timezone соединения.
func TestArchiveCalendarCutoff(t *testing.T) {
	database, _ := settlementFixture(t)
	for _, zone := range []string{"UTC", "Pacific/Auckland", "America/Los_Angeles"} {
		err := database.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec("SELECT set_config('TimeZone',?,true)", zone).Error; err != nil {
				return err
			}
			for _, tc := range []struct{ start, want string }{{"2024-08-31T10:30:00+03:00", "2025-02-28T10:30:00+03:00"}, {"2023-08-31T10:30:00+03:00", "2024-02-29T10:30:00+03:00"}} {
				var got time.Time
				if err := tx.Raw("SELECT event_purge_at(?::timestamptz)", tc.start).Scan(&got).Error; err != nil {
					return err
				}
				want, _ := time.Parse(time.RFC3339, tc.want)
				if !got.Equal(want) {
					t.Fatalf("zone=%s got=%v want=%v", zone, got, want)
				}
				var boundary bool
				if err := tx.Raw("SELECT event_purge_at(?::timestamptz) <= ?::timestamptz", tc.start, tc.want).Scan(&boundary).Error; err != nil {
					return err
				}
				if !boundary {
					t.Fatal("exact cutoff excluded")
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

// TestArchivePurgeAtomicity сохраняет молодые и восстановленные события, откатывает сбой и удаляет полную FK-цепочку.
func TestArchivePurgeAtomicity(t *testing.T) {
	database, c := settlementFixture(t)
	ctx := actor.WithID(context.Background(), 101)
	s := archiveservice.NewService(archiverepo.NewRepository(database.DB))
	if err := database.DB.Exec("INSERT INTO users(id,is_dummy,name_cashed) VALUES(4,true,'Гость'); INSERT INTO user_event(user_id,event_id) VALUES(4,1); INSERT INTO icons(id,name,file_uuid) VALUES(1,'Общая иконка','shared'); INSERT INTO activities(event_id,user_id,icon_id,description) VALUES(1,4,1,'Комментарий'); INSERT INTO tasks(event_id,user_id,title) VALUES(1,4,'Задача');").Error; err != nil {
		t.Fatal(err)
	}
	settlementExpense(t, c, 1, 4, 100)
	if _, err := c.SettlementService.Receive(ctx, 1, 101, service.SettlementRequest{FromUserID: 4, AmountCents: 1000, RequestID: uuid.NewString()}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.TransactionService.OptimizeDebts(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if err := s.Archive(ctx, 1); err != nil {
		t.Fatal(err)
	}
	purgeArchiveBatch(t, database, true)
	var count int64
	database.DB.Table("events").Where("id=1").Count(&count)
	if count != 1 {
		t.Fatal("young archive purged")
	}
	if err := database.DB.Exec("UPDATE events SET deleted_at=clock_timestamp()-INTERVAL '7 months' WHERE id=1").Error; err != nil {
		t.Fatal(err)
	}
	tx := database.DB.Begin()
	if err := tx.Exec("SELECT id FROM events WHERE id=1 FOR UPDATE").Error; err != nil {
		t.Fatal(err)
	}
	purgeArchiveBatch(t, database, true)
	tx.Rollback()
	database.DB.Table("events").Where("id=1").Count(&count)
	if count != 1 {
		t.Fatal("locked event purged")
	}
	if err := database.DB.Exec(`CREATE FUNCTION fail_purge() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'purge unavailable'; END $$; CREATE TRIGGER fail_purge BEFORE DELETE ON tasks FOR EACH ROW EXECUTE FUNCTION fail_purge();`).Error; err != nil {
		t.Fatal(err)
	}
	purgeArchiveBatch(t, database, false)
	database.DB.Table("transactions").Where("event_id=1").Count(&count)
	if count != 1 {
		t.Fatal("partial purge committed")
	}
	if err := database.DB.Exec("DROP TRIGGER fail_purge ON tasks; DROP FUNCTION fail_purge()").Error; err != nil {
		t.Fatal(err)
	}
	purgeArchiveBatch(t, database, true)
	for _, table := range []string{"transactions", "transaction_shares", "debts", "optimized_debts", "activities", "tasks", "user_event", "settlements", "event_history"} {
		database.DB.Table(table).Count(&count)
		if count != 0 {
			t.Fatalf("child rows remain in %s", table)
		}
	}
	database.DB.Table("users").Where("id=4").Count(&count)
	if count != 0 {
		t.Fatal("unreferenced dummy remained")
	}
	database.DB.Table("users").Where("id IN (1,2,3)").Count(&count)
	if count != 3 {
		t.Fatal("real user purged")
	}
	database.DB.Table("events").Where("id=2").Count(&count)
	if count != 1 {
		t.Fatal("unrelated event purged")
	}
	database.DB.Table("icons").Count(&count)
	if count != 1 {
		t.Fatal("shared icon purged")
	}
}

// TestArchiveFormerGuests очищает бывших гостей по финансовым ссылкам и истории, сохраняя гостя другого события.
func TestArchiveFormerGuests(t *testing.T) {
	database, c := settlementFixture(t)
	ctx := actor.WithID(context.Background(), 101)
	s := archiveservice.NewService(archiverepo.NewRepository(database.DB))
	if err := database.DB.Exec("INSERT INTO users(id,is_dummy,name_cashed) VALUES(4,true,'Гость с долгом'),(5,true,'Общий гость'),(6,true,'Гость без расходов'); INSERT INTO user_event(user_id,event_id) VALUES(4,1),(5,1),(6,1); INSERT INTO tasks(event_id,user_id,title) VALUES(2,5,'Задача другого события')").Error; err != nil {
		t.Fatal(err)
	}
	settlementExpense(t, c, 1, 4, 100)
	if err := c.UserService.RemoveUserFromEvent(context.Background(), 4, 1); err != nil {
		t.Fatal(err)
	}
	if err := c.UserService.RemoveUserFromEvent(ctx, 6, 1); err != nil {
		t.Fatal(err)
	}
	if err := s.Archive(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if err := database.DB.Exec("UPDATE events SET deleted_at=clock_timestamp()-INTERVAL '7 months' WHERE id=1").Error; err != nil {
		t.Fatal(err)
	}
	purgeArchiveBatch(t, database, true)
	var count int64
	database.DB.Table("users").Where("id IN (4,6)").Count(&count)
	if count != 0 {
		t.Fatal("former orphan guests retained")
	}
	database.DB.Table("users").Where("id=5").Count(&count)
	if count != 1 {
		t.Fatal("shared guest purged")
	}
}

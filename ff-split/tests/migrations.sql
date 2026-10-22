-- Таблица пользователей
create table users
(
    id                bigserial primary key, -- Внутренний ID пользователя
    user_id           bigint unique null,    -- Внешний ID пользователя из внешней системы
    nickname_cashed   varchar(255),          -- Кэшированный никнейм
    name_cashed       varchar(255),          -- Кэшированное имя
    photo_uuid_cashed varchar(255),          -- UUID аватарки
    is_dummy          boolean default false  -- Временный/фиктивный пользователь
);

-- Иконки для типа транзакций (визуальные теги)
create table icons
(
    id        serial primary key,    -- UUID иконки
    name      varchar(255) not null, -- Название иконки
    file_uuid varchar(255) not null  -- UUID иконки (в S3, например)
);

-- Категории (например, еда, поездка и т.д.)
create table event_categories
(
    id      serial primary key,           -- ID категории
    name    varchar(255) not null,        -- Название категории
    icon_id integer references icons (id) -- Иконка категории
);

-- Категории транзакций (Еда, напитки и т.д.)
create table transaction_categories
(
    id      serial primary key,           -- ID категории транзакции
    name    varchar(255) not null,        -- Название категории
    icon_id integer references icons (id) -- Иконка категории
);

-- События (мероприятия, к которым привязаны транзакции)
create table events
(
    id          bigserial primary key,               -- ID события
    name        varchar(255) not null,               -- Название события
    description text,                                -- Описание события
    category_id integer references event_categories, -- Категория события
    image_id    varchar(255),                        -- Иконка события
    status      varchar(20) default 'active'         -- Статус: active | archive
        check (status in ('active', 'archive'))
);

create index idx_events_category_id on events (category_id);

-- Связка пользователей с событиями
create table user_event
(
    user_id  bigint not null references users (id), -- Пользователь
    event_id bigint not null references events,     -- Событие
    primary key (user_id, event_id)                 -- Уникальность участия
);

-- Действия (например, комменты, лог активности)
create table activities
(
    id          serial primary key,                 -- ID действия
    event_id    bigint references events,           -- Событие
    user_id     bigint references users (id),       -- Автор действия
    icon_id     int references icons (id),          -- ID иконки
    description text,                               -- Текст действия
    created_at  timestamp default CURRENT_TIMESTAMP -- Время создания
);

create index idx_activities_event_id on activities (event_id);

-- Транзакции
create table transactions
(
    id                      serial primary key,                                -- ID транзакции
    event_id                bigint references events,                          -- Привязка к событию
    name                    varchar(255)   not null,                           -- Название/описание покупки
    transaction_category_id integer references transaction_categories,         -- Категория (для визуала)
    datetime                timestamp               default CURRENT_TIMESTAMP, -- Дата и время
    total_paid              numeric(10, 2) not null,                           -- Сумма потраченного
    payer_id                bigint references users (id),                      -- Кто заплатил
    split_type              smallint       not null default 0                  -- Тип деления: 0 — поровну, 1 — проценты, 2 — по частям
);

create index idx_transactions_event_id on transactions (event_id);

-- Участие пользователей в транзакции
create table transaction_shares
(
    id             serial primary key,                               -- ID записи
    transaction_id integer references transactions on delete cascade,-- Транзакция
    user_id        bigint references users (id),                     -- Пользователь-участник
    value          numeric(10, 2) not null,                          -- Значение (в зависимости от split_type: сумма/процент/часть)
    constraint uniq_tx_user unique (transaction_id, user_id)         -- Один пользователь — одна доля в одной транзакции
);

create index idx_transaction_shares_tx_id on transaction_shares (transaction_id);

-- Итоговая таблица долгов (кто кому сколько должен)
create table debts
(
    id             serial primary key,                               -- ID долга
    transaction_id integer references transactions on delete cascade,-- Транзакция
    from_user_id   bigint references users (id),                     -- Кто должен
    to_user_id     bigint references users (id),                     -- Кому должен
    amount         numeric(10, 2) not null,                          -- Сумма долга
    constraint uniq_debt unique (transaction_id, from_user_id, to_user_id)
);

create index idx_debts_transaction_id on debts (transaction_id);

-- Таблица заданий (например, задачи в рамках мероприятия)
create table tasks
(
    id          serial primary key,                 -- ID задачи
    user_id     bigint references users (id),       -- Ответственный
    event_id    bigint references events,           -- Событие
    title       varchar(255) not null,              -- Название задачи
    description text,                               -- Описание
    priority    integer   default 0,                -- Приоритет (чем больше — тем важнее)
    created_at  timestamp default CURRENT_TIMESTAMP -- Время создания
);

create index idx_tasks_event_id on tasks (event_id);
create index idx_tasks_user_id on tasks (user_id);

-- Оптимизированные долги
create table optimized_debts
(
    id           serial primary key,                  -- ID оптимизированного долга
    event_id     bigint references events,            -- Событие
    from_user_id bigint references users (id),        -- Кто должен
    to_user_id   bigint references users (id),        -- Кому должен
    amount       numeric(10, 2) not null,             -- Сумма долга
    created_at   timestamp default CURRENT_TIMESTAMP, -- Время создания
    updated_at   timestamp default CURRENT_TIMESTAMP  -- Время обновления
);

create index idx_optimized_debts_event_id on optimized_debts (event_id);
create index idx_optimized_debts_from_user_id on optimized_debts (from_user_id);
create index idx_optimized_debts_to_user_id on optimized_debts (to_user_id);


ALTER TABLE events ADD COLUMN debt_revision BIGINT NOT NULL DEFAULT 0;
ALTER TABLE events ADD COLUMN optimized_debt_revision BIGINT;

CREATE TABLE settlements (
    id BIGSERIAL PRIMARY KEY,
    event_id BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    from_user_id BIGINT NOT NULL REFERENCES users(id),
    to_user_id BIGINT NOT NULL REFERENCES users(id),
    created_by BIGINT NOT NULL REFERENCES users(id),
    amount_cents BIGINT NOT NULL CHECK (amount_cents > 0 AND amount_cents <= 9999999999),
    request_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    canceled_at TIMESTAMPTZ,
    CHECK (from_user_id <> to_user_id),
    CHECK (created_by = to_user_id),
    UNIQUE(event_id, created_by, request_id)
);
CREATE INDEX idx_settlements_event ON settlements(event_id, id DESC);
CREATE INDEX idx_settlements_active ON settlements(event_id) WHERE canceled_at IS NULL;

CREATE TABLE event_history (
 id BIGSERIAL PRIMARY KEY,
 event_id BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
 actor_id BIGINT NOT NULL,
 actor_name TEXT NOT NULL,
 kind TEXT NOT NULL,
 entity_id BIGINT NOT NULL,
 before_values JSONB NOT NULL,
 after_values JSONB NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX event_history_page ON event_history(event_id,id DESC);

CREATE TABLE event_invitations (
 id BIGSERIAL PRIMARY KEY,
 event_id BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
 token_hash CHAR(64) NOT NULL UNIQUE,
 created_by BIGINT NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
 expires_at TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP + INTERVAL '7 days'),
 revoked_at TIMESTAMPTZ
);
CREATE INDEX event_invitations_active ON event_invitations(event_id,created_by,expires_at) WHERE revoked_at IS NULL;

ALTER TABLE events ADD COLUMN deleted_at TIMESTAMPTZ;
CREATE INDEX events_archive ON events(deleted_at,id) WHERE deleted_at IS NOT NULL;
CREATE FUNCTION event_purge_at(archived TIMESTAMPTZ) RETURNS TIMESTAMPTZ
 LANGUAGE SQL IMMUTABLE STRICT PARALLEL SAFE AS $$
 SELECT ((archived AT TIME ZONE 'Europe/Moscow') + INTERVAL '6 months') AT TIME ZONE 'Europe/Moscow';
$$;

CREATE TABLE receipts (
    sequence bigserial UNIQUE NOT NULL,
    id uuid PRIMARY KEY,
    event_id bigint NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    created_by bigint NOT NULL,
    image_key text NOT NULL UNIQUE,
    image_sha256 char(64) NOT NULL,
    content_type text NOT NULL,
    status text NOT NULL DEFAULT 'queued' CHECK(status IN ('uploading','queued','processing','ready','failed','confirmed')),
    revision integer NOT NULL DEFAULT 1,
    attempts integer NOT NULL DEFAULT 0,
    lease_token uuid,
    lease_until timestamptz,
    available_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    result jsonb NOT NULL DEFAULT '{}',
    draft jsonb NOT NULL DEFAULT '{}',
    transaction_ids jsonb NOT NULL DEFAULT '[]',
    error text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    UNIQUE(event_id,image_sha256)
);
CREATE INDEX receipts_queue ON receipts(available_at,sequence) WHERE status IN ('queued','processing');
CREATE INDEX receipts_event ON receipts(event_id,sequence DESC);
CREATE INDEX receipts_quota ON receipts(created_by,created_at DESC);
CREATE INDEX receipts_pending ON receipts(created_by,event_id) WHERE status IN ('uploading','queued','processing');
CREATE INDEX receipts_created ON receipts(created_at);
CREATE TABLE receipt_image_gc (image_key text PRIMARY KEY, created_at timestamptz NOT NULL DEFAULT clock_timestamp());
CREATE FUNCTION receipt_image_deleted() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  INSERT INTO receipt_image_gc(image_key) VALUES(OLD.image_key) ON CONFLICT DO NOTHING;
  RETURN OLD;
END; $$;
CREATE TRIGGER receipt_image_deleted BEFORE DELETE ON receipts FOR EACH ROW EXECUTE FUNCTION receipt_image_deleted();

ALTER TABLE receipts ADD COLUMN deleted_at timestamptz;
ALTER TABLE receipts ADD COLUMN image_available boolean NOT NULL DEFAULT true;
UPDATE receipts SET image_available=false WHERE image_key IN (SELECT image_key FROM receipt_image_gc);
ALTER TABLE receipts ADD COLUMN queued_by bigint;
ALTER TABLE receipts ADD COLUMN queued_at timestamptz;
ALTER TABLE receipts ADD COLUMN retry_revision integer NOT NULL DEFAULT 0;
UPDATE receipts SET queued_by=created_by,queued_at=created_at;
ALTER TABLE receipts ALTER COLUMN queued_by SET NOT NULL;
ALTER TABLE receipts ALTER COLUMN queued_at SET NOT NULL;
ALTER TABLE receipts ALTER COLUMN queued_at SET DEFAULT clock_timestamp();
ALTER TABLE receipts DROP CONSTRAINT receipts_event_id_image_sha256_key;
CREATE UNIQUE INDEX receipts_active_hash ON receipts(event_id,image_sha256) WHERE deleted_at IS NULL;
DROP INDEX receipts_pending;
CREATE INDEX receipts_pending ON receipts(queued_by,event_id) WHERE deleted_at IS NULL AND status IN ('uploading','queued','processing');
CREATE INDEX receipts_activity ON receipts(updated_at) WHERE status!='confirmed';
CREATE TABLE receipt_admissions (
    id bigserial PRIMARY KEY,
    receipt_id uuid NOT NULL,
    actor_id bigint NOT NULL,
    created_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX receipt_admissions_actor ON receipt_admissions(actor_id,created_at DESC);
CREATE INDEX receipt_admissions_age ON receipt_admissions(created_at);
INSERT INTO receipt_admissions(receipt_id,actor_id,created_at) SELECT id,created_by,created_at FROM receipts WHERE created_at>clock_timestamp()-interval '7 days';
-- Журнал допуска переживает удаление чека и события, чтобы очистка не обнуляла квоту.
CREATE FUNCTION receipt_admit_upload() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  IF NEW.queued_by IS NULL THEN NEW.queued_by:=NEW.created_by; END IF;
  INSERT INTO receipt_admissions(receipt_id,actor_id,created_at) VALUES(NEW.id,NEW.created_by,NEW.created_at);
  RETURN NEW;
END; $$;
CREATE TRIGGER receipt_admit_upload BEFORE INSERT ON receipts FOR EACH ROW EXECUTE FUNCTION receipt_admit_upload();

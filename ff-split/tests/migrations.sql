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


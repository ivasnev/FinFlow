-- Создание таблицы пользователей
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    id_user BIGINT UNIQUE,
    nickname_cashed VARCHAR(255),
    name_cashed VARCHAR(255),
    photo_uuid_cashed VARCHAR(255),
    is_dummy BOOLEAN DEFAULT FALSE
);

-- Создание таблицы категорий
CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    image_id VARCHAR(255)
);

-- Создание таблицы мероприятий
CREATE TABLE IF NOT EXISTS events (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category_id INT REFERENCES categories(id),
    image_id VARCHAR(255),
    status VARCHAR(20) CHECK (status IN ('active', 'archive')) DEFAULT 'active'
);

-- Создание таблицы связи пользователей и мероприятий
CREATE TABLE IF NOT EXISTS user_event (
    id_user BIGINT REFERENCES users(id_user),
    id_event BIGINT REFERENCES events(id),
    PRIMARY KEY (id_user, id_event)
);

-- Создание таблицы активностей в мероприятии
CREATE TABLE IF NOT EXISTS activities (
    id SERIAL PRIMARY KEY,
    id_event BIGINT REFERENCES events(id),
    id_user BIGINT REFERENCES users(id_user),
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы для иконок
CREATE TABLE IF NOT EXISTS icons (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    file_uuid VARCHAR(255) NOT NULL
);

-- Создание таблицы типов транзакций
CREATE TABLE IF NOT EXISTS transaction_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    icon_id VARCHAR(255) REFERENCES icons(id)
);

-- Создание таблицы транзакций
CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    event_id BIGINT REFERENCES events(id),
    name VARCHAR(255) NOT NULL,
    transaction_type_id INT REFERENCES transaction_types(id),
    datetime TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    total_paid DECIMAL(10, 2) NOT NULL,
    payer_id BIGINT REFERENCES users(id_user)
);

-- Создание таблицы связи пользователей и транзакций
CREATE TABLE IF NOT EXISTS user_transaction (
    id SERIAL PRIMARY KEY,
    transaction_id INT REFERENCES transactions(id),
    user_id BIGINT REFERENCES users(id_user),
    user_part DECIMAL(10, 2) NOT NULL
);

-- Создание таблицы задач
CREATE TABLE IF NOT EXISTS tasks (
    id SERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id_user),
    event_id BIGINT REFERENCES events(id),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    priority INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Добавление индексов для оптимизации запросов
CREATE INDEX IF NOT EXISTS idx_events_category_id ON events(category_id);
CREATE INDEX IF NOT EXISTS idx_activities_id_event ON activities(id_event);
CREATE INDEX IF NOT EXISTS idx_transactions_event_id ON transactions(event_id);
CREATE INDEX IF NOT EXISTS idx_user_transaction_transaction_id ON user_transaction(transaction_id);
CREATE INDEX IF NOT EXISTS idx_tasks_event_id ON tasks(event_id);
CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks(user_id); 
-- Создаем таблицу для хранения связей между пользователями (друзья)
CREATE TABLE IF NOT EXISTS user_friends (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    friend_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Добавляем уникальный индекс, чтобы предотвратить дублирование связей
    CONSTRAINT unique_user_friend UNIQUE (user_id, friend_id),
    
    -- Проверка, чтобы пользователь не мог добавить сам себя в друзья
    CONSTRAINT user_not_friend_self CHECK (user_id <> friend_id)
);

-- Создаем индекс для быстрого поиска друзей пользователя
CREATE INDEX idx_user_friends_user_id ON user_friends(user_id);

-- Создаем индекс для быстрого поиска пользователей, у которых в друзьях указанный пользователь
CREATE INDEX idx_user_friends_friend_id ON user_friends(friend_id);

-- Создаем индекс для быстрого поиска по статусу
CREATE INDEX idx_user_friends_status ON user_friends(status); 
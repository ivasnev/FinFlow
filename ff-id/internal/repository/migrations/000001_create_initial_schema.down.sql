-- Откат начальной схемы базы данных

-- Сначала удаляем ограничение внешнего ключа
ALTER TABLE IF EXISTS users
    DROP CONSTRAINT IF EXISTS fk_users_avatar;

-- Удаляем таблицы в обратном порядке от их создания
DROP TABLE IF EXISTS user_avatars;
DROP TABLE IF EXISTS users; 
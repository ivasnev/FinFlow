-- Откат начальной схемы базы данных

-- Удаляем таблицы в обратном порядке от их создания
DROP TABLE IF EXISTS devices;
DROP TABLE IF EXISTS login_history;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS users; 
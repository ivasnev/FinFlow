-- Откат требует отсутствия совпадающих исторических копий; при конфликте миграция завершится ошибкой без удаления данных.
ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE (email);
ALTER TABLE users ADD CONSTRAINT users_nickname_key UNIQUE (nickname);

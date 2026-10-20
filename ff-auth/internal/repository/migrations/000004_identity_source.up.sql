-- Актуальные email и nickname принадлежат ff-id. Исторические копии не резервируют освобождённые значения.
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_nickname_key;

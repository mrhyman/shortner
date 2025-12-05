-- Добавляем колонку user_id в таблицу links
ALTER TABLE links ADD COLUMN IF NOT EXISTS user_id UUID NULL DEFAULT NULL;

-- Добавляем колонку is_deleted в таблицу links
ALTER TABLE links ADD COLUMN IF NOT EXISTS is_deleted boolean DEFAULT false;

-- Удаляем колонку user_id из таблицы links
ALTER TABLE links DROP COLUMN IF EXISTS user_id;

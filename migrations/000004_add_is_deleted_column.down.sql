-- Удаляем колонку is_deleted из таблицы links
ALTER TABLE links DROP COLUMN IF EXISTS is_deleted;

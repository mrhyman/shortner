-- Удаляем колонку correlation_id из таблицы links
ALTER TABLE links DROP COLUMN IF EXISTS correlation_id;

-- Добавляем колонку correlation_id в таблицу links
ALTER TABLE links ADD COLUMN IF NOT EXISTS correlation_id TEXT NOT NULL DEFAULT '';

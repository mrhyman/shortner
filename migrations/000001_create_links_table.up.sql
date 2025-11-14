CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Создание таблицы links
CREATE TABLE IF NOT EXISTS links (
    uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    short_url TEXT UNIQUE,
    original_url TEXT NOT NULL
);

-- Индекс для быстрого поиска по оригинальной ссылке
CREATE UNIQUE INDEX IF NOT EXISTS idx_links_original_url
    ON links(original_url);


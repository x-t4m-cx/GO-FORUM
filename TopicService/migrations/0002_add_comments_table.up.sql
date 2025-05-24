-- Создание таблицы комментариев
CREATE TABLE comments (
                          id SERIAL PRIMARY KEY,
                          topic_id INTEGER NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
                          username VARCHAR(100) NOT NULL,
                          content TEXT NOT NULL,
                          created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
                          updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Триггер для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_comment_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER comment_updated_at_trigger
    BEFORE UPDATE ON comments
    FOR EACH ROW
    EXECUTE FUNCTION update_comment_updated_at();
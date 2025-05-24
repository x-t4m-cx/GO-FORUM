-- Создание таблицы тем (topics)
CREATE TABLE topics (
                        id SERIAL PRIMARY KEY,
                        title VARCHAR(255) NOT NULL,
                        content TEXT NOT NULL,
                        username VARCHAR(100) NOT NULL,
                        created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Триггер для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_topic_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER topic_updated_at_trigger
    BEFORE UPDATE ON topics
    FOR EACH ROW
    EXECUTE FUNCTION update_topic_updated_at();
-- Индексы для ускорения поиска
CREATE INDEX idx_topics_username ON topics(username);
CREATE INDEX idx_topics_created_at ON topics(created_at);
CREATE INDEX idx_comments_topic_id ON comments(topic_id);
CREATE INDEX idx_comments_username ON comments(username);
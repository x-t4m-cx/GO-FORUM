DROP TRIGGER IF EXISTS comment_updated_at_trigger ON comments;
DROP FUNCTION IF EXISTS update_comment_updated_at;
DROP TABLE IF EXISTS comments;
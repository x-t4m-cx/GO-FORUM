package models

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type UpdateRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}
type TopicRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}
type CreateCommentRequest struct {
	TopicId string `json:"topic_id"`
	Content string `json:"content"`
}
type UpdateCommentRequest struct {
	Content string `json:"content"`
}

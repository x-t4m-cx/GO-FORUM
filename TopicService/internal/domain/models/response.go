package models

type CommentResponse struct {
	Data Comment `json:"data"`
}

type CommentsListResponse struct {
	Data []*Comment `json:"data"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type TopicResponse struct {
	Message string `json:"message,omitempty"`
	Topic   Topic  `json:"topic,omitempty"`
}

type TopicsListResponse struct {
	Data []*Topic `json:"data"`
}

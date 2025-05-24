package http

import (
	"TopicService/internal/domain/models"
	"TopicService/internal/usecases"
	"errors"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type CommentHandler struct {
	commentService usecases.CommentUseCasesInterface
	l              slog.Logger
}

func NewCommentHandler(commentService usecases.CommentUseCasesInterface, l slog.Logger) *CommentHandler {
	return &CommentHandler{commentService: commentService, l: l}
}

// CreateComment godoc
// @Summary Create a new comment
// @Description Create a new comment for a topic
// @Tags comments
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param input body models.CreateCommentRequest true "Comment data"
// @Success 200 {object} models.CommentResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /comments/ [post]
func (h *CommentHandler) CreateComment(c *gin.Context) {
	h.l.Info("CreateComment handler started")

	username, exists := c.Get("username")
	if !exists {
		h.l.Warn("CreateComment: unauthorized access attempt")
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "unauthorized"})
		return
	}

	var req models.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Error("CreateComment: invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	topicID, err := strconv.Atoi(req.TopicId)
	if err != nil {
		h.l.Error("CreateComment: invalid topic ID format", "topic_id", req.TopicId, "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	newComment := &models.Comment{
		TopicID:   topicID,
		Username:  username.(string),
		Content:   req.Content,
		CreatedAt: time.Now(),
	}

	h.l.Debug("CreateComment: creating new comment",
		"topic_id", topicID,
		"username", username.(string),
		"content_length", len(req.Content))

	if err := h.commentService.CreateComment(c.Request.Context(), newComment); err != nil {
		h.l.Error("CreateComment: failed to create comment",
			"topic_id", topicID,
			"error", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	h.l.Info("CreateComment: comment created successfully",
		"comment_id", newComment.Id,
		"topic_id", topicID)
	c.JSON(http.StatusOK, models.CommentResponse{Data: *newComment})
}

// GetAll godoc
// @Summary Get all comments for a topic
// @Description Get list of all comments for a specific topic
// @Tags comments
// @Accept  json
// @Produce  json
// @Param topic_id path int true "Topic ID"
// @Success 200 {object} models.CommentsListResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /topics/comments/{topic_id} [get]
func (h *CommentHandler) GetAll(c *gin.Context) {
	topicId, err := strconv.Atoi(c.Param("topic_id"))
	if err != nil {
		h.l.Error("GetAllComments: invalid topic ID format",
			"topic_id_param", c.Param("topic_id"),
			"error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	h.l.Info("GetAllComments handler started", "topic_id", topicId)

	comments, err := h.commentService.GetAllComments(c.Request.Context(), topicId)
	if err != nil {
		h.l.Error("GetAllComments: failed to get comments",
			"topic_id", topicId,
			"error", err)
		if errors.As(err, &models.ErrNotFound{}){
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	h.l.Info("GetAllComments: successfully retrieved comments",
		"topic_id", topicId,
		"count", len(comments))
	c.JSON(http.StatusOK, models.CommentsListResponse{Data: comments})
}

// GetComment godoc
// @Summary Get a comment by ID
// @Description Get a comment by its ID
// @Tags comments
// @Accept  json
// @Produce  json
// @Param id path int true "Comment ID"
// @Success 200 {object} models.Comment
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /comments/{id} [get]
func (h *CommentHandler) GetComment(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.l.Error("GetComment: invalid comment ID format",
			"id_param", c.Param("id"),
			"error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	h.l.Info("GetComment handler started", "comment_id", id)

	comment, err := h.commentService.GetComment(c.Request.Context(), id)
	if err != nil {
		h.l.Error("GetComment: failed to get comment",
			"comment_id", id,
			"error", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	h.l.Debug("GetComment: successfully retrieved comment",
		"comment_id", id,
		"topic_id", comment.TopicID,
		"username", comment.Username)
	c.JSON(http.StatusOK, comment)
}

// UpdateComment godoc
// @Summary Update a comment
// @Description Update a comment by ID
// @Tags comments
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Comment ID"
// @Param input body models.UpdateCommentRequest true "Comment data"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /comments/{id} [put]
func (h *CommentHandler) UpdateComment(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.l.Error("UpdateComment: invalid comment ID format",
			"id_param", c.Param("id"),
			"error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid comment ID"})
		return
	}

	h.l.Info("UpdateComment handler started", "comment_id", id)

	var req models.UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Error("UpdateComment: invalid request body",
			"comment_id", id,
			"error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	username, exists := c.Get("username")
	if !exists {
		h.l.Warn("UpdateComment: unauthorized access attempt",
			"comment_id", id)
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "unauthorized"})
		return
	}

	existingComment, err := h.commentService.GetComment(c.Request.Context(), id)
	if err != nil {
		h.l.Error("UpdateComment: failed to get existing comment",
			"comment_id", id,
			"error", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	if existingComment.Username != username.(string) && username.(string) != "admin" {
		h.l.Warn("UpdateComment: forbidden access attempt",
			"comment_id", id,
			"request_user", username.(string),
			"comment_owner", existingComment.Username)
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "forbidden"})
		return
	}

	updatedComment := &models.Comment{
		Id:        id,
		Content:   req.Content,
		UpdatedAt: time.Now(),
	}

	h.l.Debug("UpdateComment: updating comment",
		"comment_id", id,
		"content_length", len(req.Content))

	if err := h.commentService.UpdateComment(c.Request.Context(), updatedComment); err != nil {
		h.l.Error("UpdateComment: failed to update comment",
			"comment_id", id,
			"error", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	h.l.Info("UpdateComment: comment updated successfully", "comment_id", id)
	c.JSON(http.StatusOK, models.MessageResponse{Message: "Comment updated successfully"})
}

// DeleteComment godoc
// @Summary Delete a comment
// @Description Delete a comment by ID
// @Tags comments
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Comment ID"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /comments/{id} [delete]
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.l.Error("DeleteComment: invalid comment ID format",
			"id_param", c.Param("id"),
			"error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	h.l.Info("DeleteComment handler started", "comment_id", id)

	if err := h.commentService.DeleteComment(c.Request.Context(), id); err != nil {
		h.l.Error("DeleteComment: failed to delete comment",
			"comment_id", id,
			"error", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	h.l.Info("DeleteComment: comment deleted successfully", "comment_id", id)
	c.JSON(http.StatusOK, models.MessageResponse{Message: "Comment deleted successfully!"})
}

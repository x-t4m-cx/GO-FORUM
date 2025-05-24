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

type TopicHandler struct {
	topicService usecases.TopicUseCasesInterface
	l            slog.Logger
}

func NewTopicHandler(topicService usecases.TopicUseCasesInterface, l slog.Logger) *TopicHandler {
	return &TopicHandler{topicService: topicService, l: l}
}

// CreateTopic godoc
// @Summary Create a new topic
// @Description Create a new topic with the input payload
// @Tags topics
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param input body models.TopicRequest true "Topic data"
// @Success 201 {object} models.TopicResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /topics/ [post]
func (h *TopicHandler) CreateTopic(c *gin.Context) {
	h.l.Info("CreateTopic handler started")

	username, exists := c.Get("username")
	if !exists {
		h.l.Warn("CreateTopic: unauthorized access attempt")
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "unauthorized"})
		return
	}

	var req models.TopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Error("CreateTopic: invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	newTopic := &models.Topic{
		Title:     req.Title,
		Content:   req.Content,
		Username:  username.(string),
		CreatedAt: time.Now(),
	}

	h.l.Debug("CreateTopic: creating new topic", "topic", newTopic)

	if err := h.topicService.CreateTopic(c.Request.Context(), newTopic); err != nil {
		h.l.Error("CreateTopic: failed to create topic", "error", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	h.l.Info("CreateTopic: topic created successfully", "topicID", newTopic.Id)
	c.JSON(http.StatusCreated, models.TopicResponse{
		Message: "Topic created successfully!",
		Topic:   *newTopic,
	})
}

// GetAll godoc
// @Summary Get all topics
// @Description Get list of all topics
// @Tags topics
// @Accept  json
// @Produce  json
// @Success 200 {object} models.TopicsListResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /topics/ [get]
func (h *TopicHandler) GetAll(c *gin.Context) {
	h.l.Info("GetAll handler started")

	topics, err := h.topicService.GetAllTopics(c.Request.Context())
	if err != nil {
		h.l.Error("GetAll: failed to get topics", "error", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	h.l.Info("GetAll: successfully retrieved topics", "count", len(topics))
	c.JSON(http.StatusOK, models.TopicsListResponse{Data: topics})
}

// GetTopic godoc
// @Summary Get a topic by ID
// @Description Get a topic by its ID
// @Tags topics
// @Accept  json
// @Produce  json
// @Param id path int true "Topic ID"
// @Success 200 {object} models.Topic
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /topics/{id} [get]
func (h *TopicHandler) GetTopic(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.l.Error("GetTopic: invalid topic ID", "id", c.Param("id"), "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	h.l.Info("GetTopic handler started", "topicID", id)

	topic, err := h.topicService.GetTopic(c.Request.Context(), id)
	if err != nil {
		if errors.As(err, &models.ErrNotFound{}) {
			h.l.Error("GetTopic: topic not found", "topic", id, "error", err)
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: err.Error()})
			return
		}
		h.l.Error("GetTopic: failed to get topic", "topicID", id, "error", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	h.l.Debug("GetTopic: successfully retrieved topic", "topic", topic)
	c.JSON(http.StatusOK, topic)
}

// UpdateTopic godoc
// @Summary Update a topic
// @Description Update a topic by ID
// @Tags topics
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Topic ID"
// @Param input body models.UpdateRequest true "Topic data"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /topics/{id} [put]
func (h *TopicHandler) UpdateTopic(c *gin.Context) {
	h.l.Info("UpdateTopic handler started")

	var req models.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Error("UpdateTopic: invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	topicId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.l.Error("UpdateTopic: invalid topic ID", "id", c.Param("id"), "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	updateTopic := &models.Topic{
		Id:        topicId,
		Title:     req.Title,
		Content:   req.Content,
		UpdatedAt: time.Now(),
	}

	h.l.Debug("UpdateTopic: updating topic", "topic", updateTopic)

	err = h.topicService.UpdateTopic(c.Request.Context(), updateTopic)
	if err != nil {
		h.l.Error("UpdateTopic: failed to update topic", "topicID", topicId, "error", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	h.l.Info("UpdateTopic: topic updated successfully", "topicID", topicId)
	c.JSON(http.StatusOK, models.MessageResponse{Message: "Topic updated successfully!"})
}

// DeleteTopic godoc
// @Summary Delete a topic
// @Description Delete a topic by ID
// @Tags topics
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Topic ID"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /topics/{id} [delete]
func (h *TopicHandler) DeleteTopic(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.l.Error("DeleteTopic: invalid topic ID", "id", c.Param("id"), "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	h.l.Info("DeleteTopic handler started", "topicID", id)

	if err := h.topicService.DeleteTopic(c.Request.Context(), id); err != nil {
		h.l.Error("DeleteTopic: failed to delete topic", "topicID", id, "error", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	h.l.Info("DeleteTopic: topic deleted successfully", "topicID", id)
	c.JSON(http.StatusOK, models.MessageResponse{Message: "Topic deleted successfully!"})
}

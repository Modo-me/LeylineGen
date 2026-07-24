package world_graph

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GetResult handles GET /api/quest/result/:id.
// It reads the quest from the relational store, enriches each step's NPC
// with graph data (name, village position), and returns the final Result.
func (h *Handler) GetResult(c *gin.Context) {
	result, err := h.svc.BuildResult(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// CreateVillage handles POST /api/village.
func (h *Handler) CreateVillage(c *gin.Context) {
	var req VillageCreationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.CreateVillage(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "ok"})
}

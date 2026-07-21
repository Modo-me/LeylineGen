package world_graph

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *service
}

func NewHandler(svc *service) *Handler {
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

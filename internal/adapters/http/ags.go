package http

import (
	"LTICore/internal/core/domain"
	"net/http"

	"LTICore/internal/core/service"
	"github.com/gin-gonic/gin"
)

type AGSHandler struct {
	svc *service.AGSService
}

func NewAGSHandler(svc *service.AGSService) *AGSHandler {
	return &AGSHandler{svc: svc}
}

func (h *AGSHandler) GetLineItems(c *gin.Context) {
	items, err := h.svc.GetLineItems(c.Request.Context(), c.Query("context_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"lineitems": items})
}

func (h *AGSHandler) CreateLineItem(c *gin.Context) {
	var li domain.LineItem
	if err := c.ShouldBindJSON(&li); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.CreateLineItem(c.Request.Context(), &li); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, li)
}

func (h *AGSHandler) GetScore(c *gin.Context) {
	score, err := h.svc.GetScore(c.Request.Context(), c.Query("lineitem_id"), c.Query("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, score)
}

func (h *AGSHandler) PostScore(c *gin.Context) {
	var score domain.Score
	if err := c.ShouldBindJSON(&score); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.SaveScore(c.Request.Context(), &score); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, score)
}

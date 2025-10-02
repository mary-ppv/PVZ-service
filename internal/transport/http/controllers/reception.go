package controllers

import (
	"PVZ/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

func CreateReceptionHandler(svc service.ReceptionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			PvzID string `json:"pvzId"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		userRole := c.GetString("userRole")
		reception, err := svc.CreateReception(c.Request.Context(), req.PvzID, userRole)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id":       reception.ID,
			"pvzId":    reception.PvzID,
			"status":   reception.Status,
			"dateTime": reception.DateTime,
		})
	}
}

func CloseReceptionHandler(svc service.ReceptionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			PvzID string `json:"pvzId"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		userRole := c.GetString("userRole")
		reception, err := svc.CloseReception(c.Request.Context(), req.PvzID, userRole)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":       reception.ID,
			"pvzId":    reception.PvzID,
			"status":   reception.Status,
			"dateTime": reception.DateTime,
		})
	}
}

package controllers

import (
	"PVZ/internal/service"
	"PVZ/pkg/helper"
	"PVZ/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateReceptionHandler(svc *service.ReceptionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			PvzID string `json:"pvzId"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		userRole := c.GetString("userRole")
		logger.Log.Printf("CreateReceptionHandler: userRole from context='%s'", userRole)
		reception, err := svc.CreateReception(req.PvzID, userRole)
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

func CloseReceptionHandler(svc *service.ReceptionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			PvzID string `json:"pvzId"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		userRole := c.GetString("userRole")
		reception, err := svc.CloseReception(req.PvzID, userRole)
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

func DeleteLastProductHandler(svc *service.ReceptionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			PvzID string `json:"pvzId"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		userRole := helper.GetUserRole(c)
		reception, err := svc.DeleteLastProduct(req.PvzID, userRole)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":         reception.ID,
			"pvzId":      reception.PvzID,
			"status":     reception.Status,
			"dateTime":   reception.DateTime,
			"productIDs": reception.ProductIDs,
		})
	}
}

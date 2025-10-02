package controllers

import (
	"PVZ/internal/models"
	"PVZ/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AddProductHandler(svc *service.ProductService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			PvzID string `json:"pvzId"`
			Type  string `json:"type"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		userRole := c.GetString("userRole")
		product, err := svc.AddProduct(req.PvzID, userRole, models.ProductType(req.Type))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id":          product.ID,
			"receptionId": product.ReceptionID,
			"type":        product.Type,
			"addedAt":     product.AddedAt,
		})
	}
}

package controllers

import (
	"PVZ/internal/service"
	"PVZ/pkg/helper"
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

		userRole := helper.GetUserRole(c)
		product, err := svc.AddProduct(c, req.PvzID, userRole, req.Type)
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

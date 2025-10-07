package controllers

import (
	"PVZ/internal/service"
	"PVZ/pkg/helper"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// AddProductHandler godoc
// @Summary Добавление товара
// @Description Добавление товара в активную приемку ПВЗ (только для employee и moderator)
// @Tags Products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AddProductRequest true "Данные товара"
// @Success 201 {object} ProductResponse
// @Failure 400 {object} ErrorResponse
// @Router /products/ [post]
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

// DTO структуры для Product
type (
	AddProductRequest struct {
		PvzID string `json:"pvzId" example:"1"`
		Type  string `json:"type" example:"electronics"`
	}

	ProductResponse struct {
		ID          string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
		ReceptionID string    `json:"receptionId" example:"550e8400-e29b-41d4-a716-446655440001"`
		Type        string    `json:"type" example:"electronics"`
		AddedAt     time.Time `json:"addedAt" example:"2023-10-01T12:00:00Z"`
	}
)

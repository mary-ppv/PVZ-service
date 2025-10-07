package controllers

import (
	"PVZ/internal/service"
	"PVZ/pkg/helper"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AddProductHandler godoc
// @Summary Добавление товара
// @Description Добавление товара в активную приемку ПВЗ (только для employee и moderator)
// @Tags Products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object true "Данные товара"
// @Success 201 {object} object
// @Failure 400 {object} object
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

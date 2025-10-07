package controllers

import (
	"PVZ/internal/service"
	"PVZ/pkg/helper"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateReceptionHandler godoc
// @Summary Создание приемки
// @Description Создание новой приемки товаров для ПВЗ (только для employee и moderator)
// @Tags Receptions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object true "Данные приемки"
// @Success 201 {object} object
// @Failure 400 {object} object
// @Router /receptions/ [post]
func CreateReceptionHandler(svc *service.ReceptionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			PvzID string `json:"pvzId"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		ctx := c.Request.Context()
		userRole := c.GetString("userRole")

		reception, err := svc.CreateReception(ctx, req.PvzID, userRole)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id":       reception.ID,
			"pvzId":    reception.PVZID,
			"status":   reception.Status,
			"dateTime": reception.DateTime,
		})
	}
}

// CloseReceptionHandler godoc
// @Summary Закрытие приемки
// @Description Закрытие активной приемки товаров (только для employee и moderator)
// @Tags Receptions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object true "Данные для закрытия приемки"
// @Success 200 {object} object
// @Failure 400 {object} object
// @Router /receptions/close [put]
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
		reception, err := svc.CloseReception(c, req.PvzID, userRole)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":       reception.ID,
			"pvzId":    reception.PVZID,
			"status":   reception.Status,
			"dateTime": reception.DateTime,
		})
	}
}

// DeleteLastProductHandler godoc
// @Summary Удаление последнего товара
// @Description Удаление последнего добавленного товара из приемки (только для employee и moderator)
// @Tags Receptions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object true "Данные для удаления товара"
// @Success 200 {object} object
// @Failure 400 {object} object
// @Router /receptions/last-product [delete]
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
		reception, err := svc.DeleteLastProduct(c, req.PvzID, userRole)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":         reception.ID,
			"pvzId":      reception.PVZID,
			"status":     reception.Status,
			"dateTime":   reception.DateTime,
			"productIDs": reception.ProductIds,
		})
	}
}

package controllers

import (
	"PVZ/internal/service"
	"PVZ/pkg/helper"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// CreatePVZHandler godoc
// @Summary Создание ПВЗ
// @Description Создание нового пункта выдачи заказов (только для admin и moderator)
// @Tags PVZ
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreatePVZRequest true "Данные ПВЗ"
// @Success 201 {object} PVZResponse
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /pvz/ [post]
func CreatePVZHandler(svc *service.PVZService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name string `json:"name"`
			City string `json:"city"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		userRole := helper.GetUserRole(c)
		pvz, err := svc.CreatePVZ(c, req.Name, req.City, userRole)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, pvz)
	}
}

// GetPVZListHandler godoc
// @Summary Получение списка ПВЗ
// @Description Получение списка пунктов выдачи заказов с пагинацией и фильтрацией по городу (только для admin и moderator)
// @Tags PVZ
// @Produce json
// @Security BearerAuth
// @Param page query int false "Номер страницы"
// @Param limit query int false "Количество записей на странице"
// @Param city query string false "Фильтр по городу"
// @Success 200 {object} PVZListResponse
// @Failure 403 {object} ErrorResponse
// @Router /pvz/ [get]
func GetPVZListHandler(svc *service.PVZService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString("userRole")

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		offset := (page - 1) * limit

		city := c.Query("city")

		pvzList, err := svc.GetPVZList(c, offset, limit, city, userRole)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, pvzList)
	}
}

// DTO структуры для PVZ
type (
	CreatePVZRequest struct {
		Name string `json:"name" example:"ПВЗ Центральный"`
		City string `json:"city" example:"Москва"`
	}

	PVZResponse struct {
		ID        int64     `json:"id" example:"1"`
		Name      string    `json:"name" example:"ПВЗ Центральный"`
		City      string    `json:"city" example:"Москва"`
		CreatedAt time.Time `json:"createdAt" example:"2023-10-01T12:00:00Z"`
	}

	PVZListResponse struct {
		PVZs []PVZResponse `json:"pvzs"`
		Page int           `json:"page" example:"1"`
	}
)

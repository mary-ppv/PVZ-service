package controllers

import (
	"PVZ/internal/models"
	"PVZ/internal/service"
	"PVZ/pkg/helper"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreatePVZHandler(svc *service.PVZService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name string      `json:"name"`
			City models.City `json:"city"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		userRole := helper.GetUserRole(c)
		pvz, err := svc.CreatePVZ(req.Name, req.City, userRole)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, pvz)
	}
}

func GetPVZListHandler(svc *service.PVZService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString("userRole")

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		offset := (page - 1) * limit

		cityQuery := c.Query("city")
		var city *models.City
		if cityQuery != "" {
			cityVal := models.City(cityQuery)
			city = &cityVal
		}

		pvzList, err := svc.GetPVZList(offset, limit, city, userRole)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, pvzList)
	}
}

package controllers

import (
	"PVZ/internal/service"
	"PVZ/pkg/helper"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

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

		ctx := c.Request.Context()
		userRole := helper.GetUserRole(c)
		pvz, err := svc.CreatePVZ(ctx, req.Name, req.City, userRole)
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

		city := c.Query("city")

		ctx := c.Request.Context()
		pvzList, err := svc.GetPVZList(ctx, offset, limit, city, userRole)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, pvzList)
	}
}

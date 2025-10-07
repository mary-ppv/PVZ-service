package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheck godoc
// @Summary Проверка здоровья сервиса
// @Description Проверка доступности сервиса
// @Tags Health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status: "ok",
	})
}

type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}

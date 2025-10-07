package controllers

import (
	"PVZ/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	svc *service.UserService
}

func NewAuthHandler(svc *service.UserService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Register godoc
// @Summary Регистрация пользователя
// @Description Создание нового пользователя в системе
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Данные для регистрации"
// @Success 201 {object} RegisterResponse
// @Failure 400 {object} ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	user, err := h.svc.Register(c, req.Email, req.Password, req.Role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
	})
}

// Login godoc
// @Summary Вход в систему
// @Description Аутентификация пользователя и получение токена
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Учетные данные"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	ctx := c.Request.Context()
	token, err := h.svc.Login(ctx, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// DummyLogin godoc
// @Summary Тестовый вход
// @Description Получение тестового токена для указанной роли (для тестирования)
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body DummyLoginRequest true "Роль для тестового входа"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Router /auth/dummy [post]
func (h *AuthHandler) DummyLogin(c *gin.Context) {
	var req struct {
		Role string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	token, err := h.svc.DummyLogin(req.Role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// DTO структуры для Swagger документации
type (
	RegisterRequest struct {
		Email    string `json:"email" example:"user@example.com"`
		Password string `json:"password" example:"password123"`
		Role     string `json:"role" example:"employee"`
	}

	RegisterResponse struct {
		ID    string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
		Email string `json:"email" example:"user@example.com"`
		Role  string `json:"role" example:"employee"`
	}

	LoginRequest struct {
		Email    string `json:"email" example:"user@example.com"`
		Password string `json:"password" example:"password123"`
	}

	LoginResponse struct {
		Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	}

	DummyLoginRequest struct {
		Role string `json:"role" example:"employee"`
	}

	ErrorResponse struct {
		Error string `json:"error" example:"error message"`
	}
)

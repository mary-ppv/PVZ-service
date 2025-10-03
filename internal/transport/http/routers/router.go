package routers

import (
	"PVZ/internal/service"
	"PVZ/internal/transport/http/controllers"
	"PVZ/internal/transport/http/middleware"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRouter(
	receptionService *service.ReceptionService,
	pvzService *service.PVZService,
	productService *service.ProductService,
	userService *service.UserService,
	jwtKey []byte,
	logger *log.Logger,
) *gin.Engine {

	r := gin.Default()

	r.Use(middleware.PrometheusMetricsMiddleware())

	r.Use(middleware.JWTMiddleware(jwtKey, logger))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	authHandler := controllers.NewAuthHandler(userService)
	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/dummy", authHandler.DummyLogin)
	}

	pvz := r.Group("/pvz")
	{
		pvz.POST("/", controllers.CreatePVZHandler(pvzService))
		pvz.GET("/", controllers.GetPVZListHandler(pvzService))
	}

	reception := r.Group("/receptions")
	{
		reception.POST("/", controllers.CreateReceptionHandler(receptionService))
		reception.PUT("/close", controllers.CloseReceptionHandler(receptionService))
		reception.DELETE("/last-product", controllers.DeleteLastProductHandler(receptionService))
	}

	product := r.Group("/products")
	{
		product.POST("/", controllers.AddProductHandler(productService))
	}

	r.GET("/metrics", gin.WrapH(controllers.MetricsHandler()))

	return r
}

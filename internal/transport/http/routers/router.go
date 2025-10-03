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

	r.GET("/metrics", gin.WrapH(controllers.MetricsHandler()))

	api := r.Group("/")
	api.Use(middleware.JWTMiddleware(jwtKey))
	{
		pvz := api.Group("/pvz")
		{
			pvz.POST("/", controllers.CreatePVZHandler(pvzService))
			pvz.GET("/", controllers.GetPVZListHandler(pvzService))
		}

		reception := api.Group("/receptions")
		{
			reception.POST("/", controllers.CreateReceptionHandler(receptionService))
			reception.PUT("/close", controllers.CloseReceptionHandler(receptionService))
			reception.DELETE("/last-product", controllers.DeleteLastProductHandler(receptionService))
		}

		product := api.Group("/products")
		{
			product.POST("/", controllers.AddProductHandler(productService))
		}
	}

	return r
}

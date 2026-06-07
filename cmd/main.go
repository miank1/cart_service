package main

import (
	"cart-service/internal/handler"
	"cart-service/internal/models"
	"cart-service/internal/repository"
	"cart-service/internal/service"
	"cart-service/pkg/config"
	"cart-service/pkg/db"
	"cart-service/pkg/logger"
	"cart-service/pkg/middleware"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	logger.Init()
	defer logger.Sync()

	if err := godotenv.Load(".env"); err != nil {
		log.Println("⚠️  No .env found, continuing with system environment variables cart service")
	}

	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		log.Fatal("❌ DATABASE_DSN environment variable not set")
	}

	dbConn, err := db.InitDB(dsn)
	if err != nil {
		log.Fatalf("❌ Failed to connect to DB: %v", err)
	}

	// Auto-migrate tables
	if err := dbConn.AutoMigrate(&models.Cart{}, &models.CartItem{}); err != nil {
		log.Fatalf("❌ AutoMigrate failed: %v", err)
	}

	repo := repository.NewCartRepository(dbConn)
	cartService := service.NewCartService(repo, config.GetEnv("ORDER_SERVICE_URL", "http://localhost:8084"))
	cartHandler := handler.NewCartHandler(cartService)

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "cartservice up"})
	})

	api := router.Group("cart")
	api.Use(middleware.JWTAuth())
	{
		api.POST("/items", cartHandler.AddItem)
		api.GET("", cartHandler.GetCart)
		api.PUT("/items/:id", cartHandler.UpdateItem)
		api.DELETE("/items/:id", cartHandler.DeleteItem)
		api.POST("/checkout", cartHandler.Checkout)
	}

	port := config.GetEnv("PORT", "8083")

	log.Printf("🚀 CartService running on port %s", port)
	router.Run(":" + port)
}

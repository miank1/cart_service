package main

import (
	"cart_service/internal/handler"
	"cart_service/internal/models"
	"cart_service/internal/repository"
	"cart_service/internal/service"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/miank1/ecommerce_backend/pkg/config"
	"github.com/miank1/ecommerce_backend/pkg/db"
	"github.com/miank1/ecommerce_backend/pkg/logger"
	"github.com/miank1/ecommerce_backend/pkg/middleware"
	"github.com/miank1/ecommerce_backend/pkg/rabbitmq"
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

	rabbit, err := rabbitmq.New(
		config.GetEnv("RABBITMQ_URL", ""),
	)
	if err != nil {
		log.Fatalf("❌ Failed to connect RabbitMQ: %v", err)
	}
	defer rabbit.Close()

	_, err = rabbit.DeclareQueue("checkout_requested")
	if err != nil {
		log.Fatalf("❌ Failed to create queue: %v", err)
	}

	log.Println("✅ Queue Created")

	err = rabbit.Publish(
		"checkout_requested",
		map[string]interface{}{
			"user_id": "123",
			"message": "checkout requested",
		},
	)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("✅ Message Published")

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

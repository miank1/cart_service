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

func LoadEnv() {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")
}

func main() {

	logger.Init()
	defer logger.Sync()

	LoadEnv()

	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		log.Fatal("❌ DATABASE_DSN environment variable not set")
	}

	// database
	dbConn, err := db.InitDB(dsn)
	if err != nil {
		log.Fatalf("❌ Failed to connect to DB: %v", err)
	}

	// Auto-migrate tables
	if err := dbConn.AutoMigrate(&models.Cart{}, &models.CartItem{}); err != nil {
		log.Fatalf("❌ AutoMigrate failed: %v", err)
	}

	rabbit, err := rabbitmq.New(
		config.GetEnv("RABBITMQ_URL", ""),
	)
	if err != nil {
		log.Fatal(err)
	}

	defer rabbit.Close()

	_, err = rabbit.DeclareQueue("checkout_requested")
	if err != nil {
		log.Fatalf("❌ Failed to create queue: %v", err)
	}

	log.Println("✅ Queue checkout_requested created")

	repo := repository.NewCartRepository(dbConn)
	cartService := service.NewCartService(
		repo,
		rabbit,
	)
	cartHandler := handler.NewCartHandler(cartService)

	// router
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
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

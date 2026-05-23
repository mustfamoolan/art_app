package main

import (
	"fmt"
	"log"

	"Continuous_b/bootstrap"
	"Continuous_b/config"
	"Continuous_b/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	// 1. Initialize Config
	bootstrap.InitializeConfig()

	// 1.1 Initialize Logger (Zap)
	bootstrap.InitializeLogger()

	// 1.2 Initialize Cache (Redis)
	bootstrap.InitializeCache()

	// 2. Initialize Database (PostgreSQL via Docker)
	bootstrap.InitializeDatabase()

	// 2. Initialize Fiber
	app := fiber.New(fiber.Config{
		AppName:   config.Global.App.Name + " - String Art API",
		BodyLimit: 15 * 1024 * 1024, // 15 MB — enough for any phone photo
	})

	// Allow requests from Flutter (Android emulator, iOS, web)
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,OPTIONS",
		AllowHeaders: "Content-Type,Accept",
	}))

	// 3. Register Routes
	routes.Register(app)

	port := config.Global.App.Port
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 %s started on port %s\n", config.Global.App.Name, port)
	log.Fatal(app.Listen(":" + port))
}


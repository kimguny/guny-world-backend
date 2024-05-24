package main

import (
	"fmt"
	"guny-world-backend/api"
	"guny-world-backend/api/database"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}
	
	serverIp := os.Getenv("SERVER_IP")
	database.InitDB()
	app := fiber.New()
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowOrigins:     fmt.Sprintf("http://%s:3000", serverIp),
		AllowMethods:     "GET, POST, PUT, DELETE",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
	}))
	api.Setting(app)
	log.Fatal(app.Listen(":7949"))
}
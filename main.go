package main

import (
	"guny-world-backend/api"
	"guny-world-backend/api/database"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}
	
	database.InitDB()
	app := fiber.New()
	app.Use(recover.New())
	app.Use(cors.New())
	api.Setting(app)
	log.Fatal(app.Listen(":8080"))
}
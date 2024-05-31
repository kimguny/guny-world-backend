package api

import (
	login "guny-world-backend/api/login"
	register "guny-world-backend/api/register"

	"github.com/gofiber/fiber/v2"
)

func Setting(app *fiber.App) {
	api := app.Group("/api")

	api.Post("/register", register.Register)
	api.Post("/login", login.Login)
}

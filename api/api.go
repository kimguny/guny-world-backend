package api

import (
	login "guny-world-backend/api/login"
	register "guny-world-backend/api/register"

	"github.com/gofiber/fiber/v2"
)

func Setting(app *fiber.App) {
	app.Post("/register", register.Register)
	app.Post("/login", login.Login)
}
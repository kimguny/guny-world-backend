package api

import (
	"guny-world-backend/api/handlers"
	login "guny-world-backend/api/login"
	register "guny-world-backend/api/register"
	reissue "guny-world-backend/api/reissue"

	"github.com/gofiber/fiber/v2"
)

func Setting(app *fiber.App) {
	api := app.Group("/api")

	api.Post("/register", register.Register)
	api.Post("/login", login.Login)
	api.Post("/reissue", reissue.Reissue)

	api.Get("/user_info", handlers.GetUserInfo)
}

package handlers

import (
	"guny-world-backend/api/database"
	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
)

// 사용자 닉네임 조회 핸들러
func GetUserInfo(c *fiber.Ctx) (err error) {
	db := database.DB

	tokenString := c.Get("Authorization")
	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing JWT"})
	}

    jwtSecret := os.Getenv("JWT_SECRET_TOKEN")

	token, err := validateToken(tokenString, jwtSecret)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired JWT"})
	}


	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired JWT"})
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid JWT claims"})
	}

	// 사용자 닉네임 조회

	var nickname string
	err = db.QueryRow("SELECT nickname FROM Users WHERE id = ?", userID).Scan(&nickname)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not fetch nickname"})
	}

	// 닉네임을 JSON 응답으로 반환
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"nickname": nickname})
}

// JWT 토큰 검증 함수
func validateToken(tokenString, secretKey string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.NewError(fiber.StatusUnauthorized, "Unexpected signing method")
		}
		return []byte(secretKey), nil
	})
	return token, err
}
package reissue

import (
	"log"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
)

func Reissue(c *fiber.Ctx) (err error) {
    type RequestQuery struct {
        RefreshToken string `json:"refreshToken"`
    }

    // Body 파싱
    var requestQuery RequestQuery
    if err := c.BodyParser(&requestQuery); err != nil {
        log.Println("Error : ", err)
        return c.Status(400).JSON(fiber.Map{"error": err.Error()})
    }

    // JWT 키 불러오기
    jwtSecret := os.Getenv("JWT_SECRET_TOKEN")

    // 리프레시 토큰 검증 및 파싱
    token, err := jwt.ParseWithClaims(requestQuery.RefreshToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(jwtSecret), nil
    })
    if err != nil {
        log.Println("Error : ", err)
        return c.Status(401).JSON(fiber.Map{"error": "Invalid refresh token"})
    }

    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        log.Println("Error : Invalid token claims")
        return c.Status(401).JSON(fiber.Map{"error": "Invalid refresh token"})
    }

    // 새로운 엑세스 토큰 생성
    accessToken, err := makeAccessToken(claims.UserId, jwtSecret)
    if err != nil {
        log.Println("Error : ", err)
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    // 새로운 리프레시 토큰 생성
    refreshToken, err := makeRefreshToken(claims.UserId, jwtSecret)
    if err != nil {
        log.Println("Error : ", err)
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    return c.Status(200).JSON(fiber.Map{"message": "Success reissue!", "accessToken": accessToken, "refreshToken": refreshToken})
}

type Claims struct {
    UserId string `json:"user_id"`
    jwt.StandardClaims
}

func makeAccessToken(userId string, jwtSecret string) (accessToken string, err error) {
    claims := Claims{
        UserId: userId,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: time.Now().Add(time.Hour * 1).Unix(),
            Issuer:    "flexible-quest",
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    accessToken, err = token.SignedString([]byte(jwtSecret))
    if err != nil {
        return "", err
    }

    return accessToken, nil
}

func makeRefreshToken(userId string, jwtSecret string) (refreshToken string, err error) {
    claims := Claims{
        UserId: userId,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: time.Now().Add(time.Hour * 24 * 3).Unix(),
            Issuer:    "flexible-quest",
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    refreshToken, err = token.SignedString([]byte(jwtSecret))
    if err != nil {
        return "", err
    }

    return refreshToken, nil
}

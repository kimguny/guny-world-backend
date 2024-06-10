// login/login.go
package login

import (
	"guny-world-backend/api/database"
	"log"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func Login(c *fiber.Ctx) (err error) {
    db := database.DB

    type RequestQuery struct {
        UserId   string `json:"user_id"`
        Password string `json:"password"`
    }

    // Body 파싱
    var requestQuery RequestQuery
    if err := c.BodyParser(&requestQuery); err != nil {
        log.Println("Error : ", err)
        return c.Status(400).JSON(fiber.Map{"error": err.Error()})
    }

    // 유저 아이디의 대한 비번 정보 가져오기
    var password string
    err = db.Get(&password, "SELECT password FROM Users WHERE user_id = ?", requestQuery.UserId)
    if err != nil {
        log.Println("Error : ", err)
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    // 비밀번호 검증
    if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(requestQuery.Password)); err != nil {
        log.Println("Error : ", err)
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    // JWT 키 불러오기
    jwtSecret := os.Getenv("JWT_SECRET_TOKEN")

    // 해당 유저의 id 가져오기
    var id string
    err = db.Get(&id, "SELECT id FROM Users WHERE user_id = ?", requestQuery.UserId)
    if err != nil {
        log.Println("Error : ", err)
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    // 엑세스 토큰 생성
    accessToken, err := makeAccessToken(id, jwtSecret)
    if err != nil {
        log.Println("Error : ", err)
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    // 리프레시 토큰 생성
    refreshToken, err := makeRefreshToken(id, jwtSecret)
    if err != nil {
        log.Println("Error : ", err)
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    return c.Status(200).JSON(fiber.Map{"message": "Success login!", "accessToken": accessToken, "refreshToken": refreshToken})
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

func RefreshAccessToken(c *fiber.Ctx) (err error) {
    refreshToken := c.Query("refreshToken")

    claims := &Claims{}
    token, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
        return []byte(os.Getenv("JWT_SECRET_TOKEN")), nil
    })

    if err != nil {
        log.Println("Error : ", err)
        return c.Status(401).JSON(fiber.Map{"error": err.Error()})
    }

    if !token.Valid {
        log.Println("Error : ", err)
        return c.Status(401).JSON(fiber.Map{"error": "invalid token"})
    }

    err = godotenv.Load(".env")
    if err != nil {
        log.Println("Error : ", err)
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    jwtSecret := os.Getenv("JWT_SECRET_TOKEN")

    // 새로운 엑세스 토큰 생성
    userId := claims.UserId
    newAccessToken, err := makeAccessToken(userId, jwtSecret)
    if err != nil {
        log.Println("Error : ", err)
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    return c.Status(200).JSON(fiber.Map{"accessToken": newAccessToken})
}

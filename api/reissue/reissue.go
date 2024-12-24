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
        log.Println("Body 파싱 에러: ", err)
        return c.Status(400).JSON(fiber.Map{"error": "값을 입력해 주세요."})
    }

    // JWT 키 불러오기
    jwtSecret := os.Getenv("JWT_SECRET_TOKEN")
    if jwtSecret == "" {
        log.Println("JWT 시크릿 키가 설정되지 않았습니다.")
        return c.Status(500).JSON(fiber.Map{"error": "서버 설정 오류입니다. 관리자에게 문의하세요."})
    }

    // 리프레시 토큰 검증 및 파싱
    token, err := jwt.ParseWithClaims(requestQuery.RefreshToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(jwtSecret), nil
    })
    if err != nil {
        log.Println("리프레시 토큰 검증 실패: ", err)
        return c.Status(401).JSON(fiber.Map{"error": "유효하지 않은 리프레시 토큰입니다."})
    }

    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        log.Println("토큰 클레임 검증 실패")
        return c.Status(401).JSON(fiber.Map{"error": "유효하지 않은 리프레시 토큰입니다."})
    }

    // 새로운 엑세스 토큰 생성
    accessToken, err := makeAccessToken(claims.UserId, jwtSecret)
    if err != nil {
        log.Println("엑세스 토큰 생성 실패: ", err)
        return c.Status(500).JSON(fiber.Map{"error": "토큰 생성에 실패했습니다."})
    }

    // 새로운 리프레시 토큰 생성
    refreshToken, err := makeRefreshToken(claims.UserId, jwtSecret)
    if err != nil {
        log.Println("리프레시 토큰 생성 실패: ", err)
        return c.Status(500).JSON(fiber.Map{"error": "토큰 생성에 실패했습니다."})
    }

    return c.Status(200).JSON(fiber.Map{"message": "토큰 재발급 성공!", "accessToken": accessToken, "refreshToken": refreshToken})
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

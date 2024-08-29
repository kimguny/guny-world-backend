// register/register.go
package register

import (
	"guny-world-backend/api/database"
	"log"
	"regexp"
	"unicode/utf8"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *fiber.Ctx) (err error) {
    db := database.DB
    
    type RequestQuery struct {
        UserId   string `json:"user_id"`
        Password string `json:"password"`
        Nickname string `json:"nickname"`
    }

    // Body 파싱
    var requestQuery RequestQuery
    if err := c.BodyParser(&requestQuery); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": err.Error()})
    }

    // 각 필드가 비어 있는지 확인
    if requestQuery.UserId == "" {
        return c.Status(400).JSON(fiber.Map{"error": "user_id is required"})
    }
    if requestQuery.Password == "" {
        return c.Status(400).JSON(fiber.Map{"error": "password is required"})
    }
    if requestQuery.Nickname == "" {
        return c.Status(400).JSON(fiber.Map{"error": "nickname is required"})
    }

    // 사용자 이름 중복 확인
    var count int
    err = db.Get(&count, "SELECT COUNT(*) FROM users WHERE user_id = ?", requestQuery.UserId)
    if err != nil {
        log.Println("Error : ", err)
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    if count > 0 {
        log.Println("Error : user already exists")
        return c.Status(400).JSON(fiber.Map{"error": "user already exists"})
    }

    // UserId가 이메일 형태인지 확인
    emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
    if !regexp.MustCompile(emailRegex).MatchString(requestQuery.UserId) {
        return c.Status(400).JSON(fiber.Map{"error": "invalid email format"})
    }

    // 비밀번호가 8자리 이상인지 확인
    if len(requestQuery.Password) < 8 {
        return c.Status(400).JSON(fiber.Map{"error": "password must be at least 8 characters long"})
    }

    // 닉네임 길이 확인 (한국어 기준 6글자, 영어 기준 12글자)
    if utf8.RuneCountInString(requestQuery.Nickname) > 8 || len(requestQuery.Nickname) > 16 {
        return c.Status(400).JSON(fiber.Map{"error": "nickname must be at most 6 characters long in Korean or 12 characters long in English"})
    }

    // 비번 해쉬
    hashedPassword, err := hashPassword(requestQuery.Password)
    if err != nil {
        log.Println("Error : ", err)
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    // 사용자 정보 저장
    _, err = db.Exec("INSERT INTO users (user_id, password, nickname) VALUES (?, ?, ?)", requestQuery.UserId, hashedPassword, requestQuery.Nickname)
    if err != nil {
        log.Println("Error : ", err)
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    return c.Status(200).JSON(fiber.Map{"message": "Success register!"})
}

// 해쉬 함수
func hashPassword(password string) (hashedPassword string, err error) {
    passwordBytes := []byte(password)

    bytePass, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)
    if err != nil {
        return "", err
    }

    hashedPassword = string(bytePass)
    return hashedPassword, nil
}

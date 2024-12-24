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
        return c.Status(400).JSON(fiber.Map{"message": "요청 데이터를 파싱하는 데 실패했습니다."})
    }

    // 각 필드가 비어 있는지 확인
    if requestQuery.UserId == "" {
        return c.Status(400).JSON(fiber.Map{"message": "아이디 값이 존재하지 않습니다."})
    }
    if requestQuery.Password == "" {
        return c.Status(400).JSON(fiber.Map{"message": "비밀번호 값이 존재하지 않습니다."})
    }
    if requestQuery.Nickname == "" {
        return c.Status(400).JSON(fiber.Map{"message": "닉네임 값이 존재하지 않습니다."})
    }

    // 사용자 이름 중복 확인
    var count int
    err = db.Get(&count, "SELECT COUNT(*) FROM users WHERE user_id = ?", requestQuery.UserId)
    if err != nil {
        log.Println("Error : 서버 내부 오류")
        return c.Status(500).JSON(fiber.Map{"message": "서버 내부 오류입니다."})
    }

    if count > 0 {
        log.Println("Error : 중복된 아이디 입니다.")
        return c.Status(400).JSON(fiber.Map{"message": "중복된 아이디 입니다."})
    }

    // UserId가 이메일 형태인지 확인
    emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
    if !regexp.MustCompile(emailRegex).MatchString(requestQuery.UserId) {
        return c.Status(400).JSON(fiber.Map{"error": "아이디는 올바른 이메일 형식이어야 합니다."})
    }

    // 비밀번호가 8자리 이상인지 확인
    if len(requestQuery.Password) < 8 {
        return c.Status(400).JSON(fiber.Map{"error": "비밀번호는 최소 8자 이상이어야 합니다."})
    }

    // 닉네임 길이 확인 (한국어 기준 6글자, 영어 기준 12글자)
    if utf8.RuneCountInString(requestQuery.Nickname) > 8 || len(requestQuery.Nickname) > 16 {
        return c.Status(400).JSON(fiber.Map{"error": "닉네임은 한국어 최대 6글자 또는 영어 최대 12글자 이내여야 합니다."})
    }

    // 비번 해쉬
    hashedPassword, err := hashPassword(requestQuery.Password)
    if err != nil {
        log.Println("Error : 비밀번호 해시 생성 실패", err)
        return c.Status(500).JSON(fiber.Map{"error": "비밀번호를 처리하는 데 실패했습니다."})
    }

    // 사용자 정보 저장
    _, err = db.Exec("INSERT INTO users (user_id, password, nickname) VALUES (?, ?, ?)", requestQuery.UserId, hashedPassword, requestQuery.Nickname)
    if err != nil {
        log.Println("Error : 사용자 정보 저장 실패", err)
        return c.Status(500).JSON(fiber.Map{"error": "사용자 정보를 저장하는 데 실패했습니다."})
    }

    return c.Status(200).JSON(fiber.Map{"message": "회원가입이 성공!"})
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

// login/login.go
package login

import (
	"database/sql"
	"encoding/json"
	"guny-world-backend/api/database"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
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
    err = db.Get(&password, "SELECT password FROM users WHERE user_id = ?", requestQuery.UserId)
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
    err = db.Get(&id, "SELECT id FROM users WHERE user_id = ?", requestQuery.UserId)
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

func NaverLogin(c *fiber.Ctx) error {
	db := database.DB

	// 네이버에서 전달된 code와 state 파라미터를 가져옵니다.
	code := c.Query("code")
	state := c.Query("state")

	// 네이버 Client ID와 Client Secret을 환경변수에서 가져옵니다.
	clientID := os.Getenv("NAVER_CLIENT_ID")
	clientSecret := os.Getenv("NAVER_CLIENT_SECRET")
	redirectURI := "https://game.gunynote.com/naver/callback"

	// 액세스 토큰을 요청하기 위한 URL을 만듭니다.
	tokenURL := "https://nid.naver.com/oauth2.0/token"
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("code", code)
	data.Set("state", state)
	data.Set("redirect_uri", redirectURI)

	// 네이버에 액세스 토큰 요청
	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		log.Println("Error during token request:", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to request token from Naver"})
	}
	defer resp.Body.Close()

	// 응답을 파싱하여 액세스 토큰을 가져옵니다.
	var tokenResponse struct {
		AccessToken     string `json:"access_token"`
		TokenType       string `json:"token_type"`
		ExpiresIn       string `json:"expires_in"`
		RefreshToken    string `json:"refresh_token"`
		Error           string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		log.Println("Error parsing token response:", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to parse token response from Naver"})
	}

	if tokenResponse.Error != "" {
		log.Println("Naver token error:", tokenResponse.Error, tokenResponse.ErrorDescription)
		return c.Status(500).JSON(fiber.Map{"error": tokenResponse.ErrorDescription})
	}

	// 네이버 사용자 정보를 가져오기 위한 요청 설정
	userInfoURL := "https://openapi.naver.com/v1/nid/me"
	req, err := http.NewRequest("GET", userInfoURL, nil)
	if err != nil {
		log.Println("Error creating request for user info:", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create request for user info"})
	}
	req.Header.Add("Authorization", "Bearer "+tokenResponse.AccessToken)

	client := &http.Client{}
	userInfoResp, err := client.Do(req)
	if err != nil {
		log.Println("Error requesting user info:", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to request user info from Naver"})
	}
	defer userInfoResp.Body.Close()

	var userInfo struct {
		Response struct {
			Id          string `json:"id"`
			Nickname    string `json:"nickname"`
			Email       string `json:"email"`
			Name        string `json:"name"`
			ProfileImage string `json:"profile_image"`
		} `json:"response"`
	}

	if err := json.NewDecoder(userInfoResp.Body).Decode(&userInfo); err != nil {
		log.Println("Error parsing user info response:", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to parse user info response from Naver"})
	}

	// 이메일 주소를 user_id로 사용하여 네이버 사용자 정보를 데이터베이스에 저장하거나 갱신
	var existingUserID string
	err = db.Get(&existingUserID, "SELECT user_id FROM naver_user_info WHERE user_id = ?", userInfo.Response.Email)
	if err != nil && err != sql.ErrNoRows {
		// 데이터베이스 오류 처리
		log.Println("Error fetching user:", err)
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	if existingUserID == "" {
		// 새 사용자라면 정보를 데이터베이스에 추가
		_, err = db.Exec("INSERT INTO naver_user_info (user_id, nickname, profile_image, name, created_at) VALUES (?, ?, ?, ?, ?)",
			userInfo.Response.Email, userInfo.Response.Nickname, userInfo.Response.ProfileImage, userInfo.Response.Name, time.Now())
		if err != nil {
			log.Println("Error inserting new user:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Failed to insert new user into database"})
		}
	} else {
		// 기존 사용자라면 정보를 업데이트
		_, err = db.Exec("UPDATE naver_user_info SET nickname=?, profile_image=?, name=?, updated_at=? WHERE user_id=?",
			userInfo.Response.Nickname, userInfo.Response.ProfileImage, userInfo.Response.Name, time.Now(), userInfo.Response.Email)
		if err != nil {
			log.Println("Error updating existing user:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Failed to update user info in database"})
		}
	}

	// 사용자에게 JWT 발급
	jwtSecret := os.Getenv("JWT_SECRET_TOKEN")
	accessToken, err := makeAccessToken(userInfo.Response.Email, jwtSecret)
	if err != nil {
		log.Println("Error creating access token:", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create access token"})
	}

	refreshToken, err := makeRefreshToken(userInfo.Response.Email, jwtSecret)
	if err != nil {
		log.Println("Error creating refresh token:", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create refresh token"})
	}

	// 클라이언트에게 JWT 토큰 반환
	return c.Status(200).JSON(fiber.Map{"accessToken": accessToken, "refreshToken": refreshToken})
}

package chzzk

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func Chzzk(c *fiber.Ctx) (err error) {
	type RequestQuery struct {
		NID_AUT string `json:"NID_AUT"`
		NID_SES string `json:"NID_SES"`
		Id      string `json:"id"`
	}

	query := new(RequestQuery)
	if err := c.BodyParser(query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "JSON 데이터를 파싱할 수 없습니다.",
		})
	}

	client := &http.Client{}
	cookie := fmt.Sprintf("NID_AUT=%s; NID_SES=%s", query.NID_AUT, query.NID_SES)
	headers := map[string]string{
		"Cookie":     cookie,
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3",
	}

	// Fetch followers
	var followers []string
	for page := 0; page < 5; page++ {
		followersURL := fmt.Sprintf("https://api.chzzk.naver.com/manage/v1/channels/%s/followers?page=%d&size=10000&userNickname=", query.Id, page)
		req, err := http.NewRequest("GET", followersURL, nil)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "요청 생성에 실패했습니다.",
			})
		}

		req.Header.Set("Cookie", headers["Cookie"])
		req.Header.Set("User-Agent", headers["User-Agent"])

		respFollowers, err := client.Do(req)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "요청에 실패했습니다.",
			})
		}
		defer respFollowers.Body.Close()

		bodyFollowers, err := ioutil.ReadAll(respFollowers.Body)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "응답 본문을 읽는 데 실패했습니다.",
			})
		}

		var jsonResponseFollowers map[string]interface{}
		if err := json.Unmarshal(bodyFollowers, &jsonResponseFollowers); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "응답 본문을 파싱하는 데 실패했습니다.",
			})
		}

		contentFollowers, ok := jsonResponseFollowers["content"].(map[string]interface{})
		if !ok || len(contentFollowers) == 0 {
			break // 비어있는 경우 루프 중지
		}

		dataFollowers, ok := contentFollowers["data"].([]interface{})
		if !ok || len(dataFollowers) == 0 {
			break // 비어있는 경우 루프 중지
		}

		for _, item := range dataFollowers {
			user, ok := item.(map[string]interface{})["user"].(map[string]interface{})
			if ok {
				nickname, ok := user["nickname"].(string)
				if ok {
					followers = append(followers, nickname)
				}
			}
		}
	}

	// Fetch followings
	var followings []string
	for page := 0; page < 100; page++ {
		followingsURL := fmt.Sprintf("https://api.chzzk.naver.com/service/v1/channels/followings?size=500&page=%d", page)
		req, err := http.NewRequest("GET", followingsURL, nil)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "요청 생성에 실패했습니다.",
			})
		}

		req.Header.Set("Cookie", headers["Cookie"])
		req.Header.Set("User-Agent", headers["User-Agent"])

		respFollowings, err := client.Do(req)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "요청에 실패했습니다.",
			})
		}
		defer respFollowings.Body.Close()

		bodyFollowings, err := ioutil.ReadAll(respFollowings.Body)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "응답 본문을 읽는 데 실패했습니다.",
			})
		}

		var jsonResponseFollowings map[string]interface{}
		if err := json.Unmarshal(bodyFollowings, &jsonResponseFollowings); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "응답 본문을 파싱하는 데 실패했습니다.",
			})
		}

		contentFollowings, ok := jsonResponseFollowings["content"].(map[string]interface{})
		if !ok || len(contentFollowings) == 0 {
			break // 비어있는 경우 루프 중지
		}

		followingList, ok := contentFollowings["followingList"].([]interface{})
		if !ok || len(followingList) == 0 {
			break // 비어있는 경우 루프 중지
		}

		for _, item := range followingList {
			channel, ok := item.(map[string]interface{})["channel"].(map[string]interface{})
			if ok {
				channelName, ok := channel["channelName"].(string)
				if ok {
					followings = append(followings, channelName)
				}
			}
		}
	}

	// Separate lists into different categories
	followersSet := make(map[string]bool)
	followingsSet := make(map[string]bool)

	for _, follower := range followers {
		followersSet[follower] = true
	}

	for _, following := range followings {
		followingsSet[following] = true
	}

	var mutualFollows, onlyFollowers, onlyFollowing []string

	for _, follower := range followers {
		if followingsSet[follower] {
			mutualFollows = append(mutualFollows, follower)
		} else {
			onlyFollowers = append(onlyFollowers, follower)
		}
	}

	for _, following := range followings {
		if !followersSet[following] {
			onlyFollowing = append(onlyFollowing, following)
		}
	}

	return c.JSON(fiber.Map{
		"followers":     followers,
		"followings":    followings,
		"mutualFollows": mutualFollows,
		"onlyFollowers": onlyFollowers,
		"onlyFollowing": onlyFollowing,
	})
}

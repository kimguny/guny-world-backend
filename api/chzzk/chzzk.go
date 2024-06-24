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
			"error": "cannot parse JSON",
		})
	}

	// Fetch followers
	followersURL := fmt.Sprintf("https://api.chzzk.naver.com/manage/v1/channels/%s/followers?page=0&size=500&userNickname=", query.Id)
	client := &http.Client{}
	req, err := http.NewRequest("GET", followersURL, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create request",
		})
	}

	cookie := fmt.Sprintf("NID_AUT=%s; NID_SES=%s", query.NID_AUT, query.NID_SES)
	req.Header.Add("Cookie", cookie)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	respFollowers, err := client.Do(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "request failed",
		})
	}
	defer respFollowers.Body.Close()

	bodyFollowers, err := ioutil.ReadAll(respFollowers.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to read response body",
		})
	}

	var jsonResponseFollowers map[string]interface{}
	if err := json.Unmarshal(bodyFollowers, &jsonResponseFollowers); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to parse response body",
		})
	}

	contentFollowers, ok := jsonResponseFollowers["content"].(map[string]interface{})
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "unexpected response format",
		})
	}

	dataFollowers, ok := contentFollowers["data"].([]interface{})
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "unexpected response format",
		})
	}

	var followers []string
	for _, item := range dataFollowers {
		user, ok := item.(map[string]interface{})["user"].(map[string]interface{})
		if ok {
			nickname, ok := user["nickname"].(string)
			if ok {
				followers = append(followers, nickname)
			}
		}
	}

	// Fetch followings
	followingsURL := "https://api.chzzk.naver.com/service/v1/channels/followings?size=500&page=0"
	req, err = http.NewRequest("GET", followingsURL, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create request",
		})
	}

	req.Header.Set("Cookie", cookie)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	respFollowings, err := client.Do(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "request failed",
		})
	}
	defer respFollowings.Body.Close()

	bodyFollowings, err := ioutil.ReadAll(respFollowings.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to read response body",
		})
	}

	var jsonResponseFollowings map[string]interface{}
	if err := json.Unmarshal(bodyFollowings, &jsonResponseFollowings); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to parse response body",
		})
	}

	contentFollowings, ok := jsonResponseFollowings["content"].(map[string]interface{})
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "unexpected response format",
		})
	}

	followingList, ok := contentFollowings["followingList"].([]interface{})
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "unexpected response format",
		})
	}

	var followings []string
	for _, item := range followingList {
		channel, ok := item.(map[string]interface{})["channel"].(map[string]interface{})
		if ok {
			channelName, ok := channel["channelName"].(string)
			if ok {
				followings = append(followings, channelName)
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

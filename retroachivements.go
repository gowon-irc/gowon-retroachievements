package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	raAchievementsURL = "https://ra.hfc-essentials.com/user_by_range.php?user=%s&key=%s&member=%s&mode=json"
)

var (
	achievementStartTime = time.Date(2010, 10, 2, 6, 0, 0, 0, time.UTC)
)

type AchievementsByDateResp struct {
	AchievementList [][]Achievement `json:"achievement"`
}

type Achievement struct {
	Date          string `json:"Date"`
	HardcoreMode  int    `json:"HardcoreMode"`
	AchievementID int    `json:"AchievementID"`
	Title         string `json:"Title"`
	Description   string `json:"Description"`
	BadgeName     string `json:"BadgeName"`
	Points        int    `json:"Points"`
	Author        string `json:"Author"`
	GameTitle     string `json:"GameTitle"`
	GameIcon      string `json:"GameIcon"`
	GameID        int    `json:"GameID"`
	ConsoleName   string `json:"ConsoleName"`
	CumulScore    int    `json:"CumulScore"`
	BadgeURL      string `json:"BadgeURL"`
	GameURL       string `json:"GameURL"`
}

func formatAchievement(user string, a Achievement) string {
	return fmt.Sprintf("%s's last retro achievement: %s (%s) - %s (%s)", user, a.Title, a.Description, a.GameTitle, a.ConsoleName)
}

func ra(apiUser, apiKey, user string) (string, error) {
	url := fmt.Sprintf(raAchievementsURL, apiUser, apiKey, user)

	j := &AchievementsByDateResp{}

	res, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(body, &j)

	if err != nil {
		return "", err
	}

	if len(j.AchievementList[0]) == 0 {
		return fmt.Sprintf("No achievements found for user %s", user), nil
	}

	out := formatAchievement(user, j.AchievementList[0][len(j.AchievementList[0])-1])

	return out, nil
}

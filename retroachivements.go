package main

import (
	"fmt"

	"github.com/imroc/req/v3"
)

const (
	raAchievementsURL = "https://retroachievements.org/API/API_GetUserRecentAchievements.php"
)

type Achievement struct {
	HardcoreMode int    `json:"HardcoreMode"`
	Title        string `json:"Title"`
	Description  string `json:"Description"`
	Points       int    `json:"Points"`
	GameTitle    string `json:"GameTitle"`
	ConsoleName  string `json:"ConsoleName"`
}

func formatAchievement(user string, a Achievement) string {
	out := fmt.Sprintf("%s's last retro achievement: %s (%s) - %s (%s) - %d points", user, a.Title, a.Description, a.GameTitle, a.ConsoleName, a.Points)

	if a.HardcoreMode == 1 {
		out += " [Hardcore]"
	}

	return out
}

func raLastAchievement(client *req.Client, user string) (string, error) {
	var j []Achievement

	_, err := client.R().
		SetQueryParam("u", user).
		SetQueryParam("m", "131400").
		SetSuccessResult(&j).
		Get(raAchievementsURL)

	if err != nil {
		return "", err
	}

	if len(j) == 0 {
		return fmt.Sprintf("No achievements found for user %s", user), nil
	}

	out := formatAchievement(user, j[0])

	return out, nil
}

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/carlmjohnson/requests"
)

const (
	raAchievementsURL = "https://retroachievements.org/API/API_GetAchievementsEarnedBetween.php"
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

func ra(apiUser, apiKey, user string) (string, error) {
	t := fmt.Sprint(time.Now().Unix())

	var j []Achievement
	err := requests.
		URL(raAchievementsURL).
		Param("z", apiUser).
		Param("y", apiKey).
		Param("u", user).
		Param("f", "0").
		Param("t", t).
		Param("mode", "json").
		ToJSON(&j).
		Fetch(context.Background())

	if err != nil {
		return "", err
	}

	if len(j) == 0 {
		return fmt.Sprintf("No achievements found for user %s", user), nil
	}

	out := formatAchievement(user, j[len(j)-1])

	return out, nil
}

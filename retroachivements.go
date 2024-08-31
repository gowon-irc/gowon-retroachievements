package main

import (
	"fmt"
	"strings"

	"github.com/imroc/req/v3"
)

const (
	raAchievementsURL = "https://retroachievements.org/API/API_GetUserRecentAchievements.php"
)

func colourString(in, colour string) string {
	return fmt.Sprintf("{%s}%s{clear}", colour, in)
}

type Achievement struct {
	HardcoreMode int    `json:"HardcoreMode"`
	Title        string `json:"Title"`
	Description  string `json:"Description"`
	Points       int    `json:"Points"`
	GameTitle    string `json:"GameTitle"`
	ConsoleName  string `json:"ConsoleName"`
}

func formatAchievement(a Achievement) string {
	var sb strings.Builder

	w := func(in, colour string) {
		s := colourString(in, colour)
		sb.WriteString(s)
	}

	w(fmt.Sprintf("%s (%s)", a.Title, strings.TrimRight(a.Description, ".")), "cyan")

	sb.WriteString(" | ")

	w(fmt.Sprintf("%s (%s)", a.GameTitle, a.ConsoleName), "magenta")

	sb.WriteString(" | ")

	w(fmt.Sprintf("%d points", a.Points), "green")

	if a.HardcoreMode == 1 {
		w(" [Hardcore]", "yellow")
	}

	return sb.String()
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

	a := formatAchievement(j[0])
	out := fmt.Sprintf("%s's newest retroachievement: %s", user, a)

	return out, nil
}

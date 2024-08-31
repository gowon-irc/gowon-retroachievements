package main

import (
	"fmt"
	"strings"

	"github.com/imroc/req/v3"
)

const (
	raAchievementsURL = "https://retroachievements.org/API/API_GetUserRecentAchievements.php"
	raRecentGamesURL  = "https://retroachievements.org/API/API_GetUserRecentlyPlayedGames.php"

	achievementColour = "cyan"
	gameColour        = "magenta"
	pointsColour      = "green"
	hardcoreColour    = "yellow"
)

func colourString(in, colour string) string {
	return fmt.Sprintf("{%s}%s{clear}", colour, in)
}

func colourList(in []string) (out []string) {
	out = []string{}

	colours := []string{"green", "red", "blue", "orange", "magenta", "cyan", "yellow"}
	cl := len(colours)

	for n, i := range in {
		c := colours[n%cl]
		o := colourString(i, c)
		out = append(out, o)
	}

	return out
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

	w(fmt.Sprintf("%s (%s)", a.Title, strings.TrimRight(a.Description, ".")), achievementColour)

	sb.WriteString(" | ")

	w(fmt.Sprintf("%s (%s)", a.GameTitle, a.ConsoleName), gameColour)

	sb.WriteString(" | ")

	w(fmt.Sprintf("%d points", a.Points), pointsColour)

	if a.HardcoreMode == 1 {
		w(" [Hardcore]", hardcoreColour)
	}

	return sb.String()
}

type Game struct {
	Title string `json:"Title"`
}

func raNewestAchievement(client *req.Client, user string) (string, error) {
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

func raLastGames(client *req.Client, user string) (string, error) {
	var j []Game

	_, err := client.R().
		SetQueryParam("u", user).
		SetQueryParam("c", "10").
		SetSuccessResult(&j).
		Get(raRecentGamesURL)

	if err != nil {
		return "", err
	}

	if len(j) == 0 {
		return fmt.Sprintf("No played games found for user %s", user), nil
	}

	titles := []string{}
	for _, g := range j {
		titles = append(titles, g.Title)
	}

	cl := strings.Join(colourList(titles), ", ")

	return fmt.Sprintf("%s's last played retro games: %s", user, cl), nil
}

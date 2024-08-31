package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/imroc/req/v3"
)

const (
	timeDateFormat = "2006-01-02 15:04:05"

	raAchievementsURL = "https://retroachievements.org/API/API_GetUserRecentAchievements.php"
	raRecentGamesURL  = "https://retroachievements.org/API/API_GetUserRecentlyPlayedGames.php"
	raUserSummaryURL  = "https://retroachievements.org/API/API_GetUserSummary.php"

	achievementColour  = "cyan"
	gameColour         = "magenta"
	pointsColour       = "green"
	hardcoreColour     = "yellow"
	richPresenceColour = "yellow"
)

var (
	now = time.Now
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

type Game struct {
	Title string `json:"Title"`
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

type UserSummary struct {
	ID             int    `json:"ID"`
	Status         string `json:"Status"`
	RecentlyPlayed []struct {
		Title      string `json:"Title"`
		LastPlayed string `json:"LastPlayed"`
	} `json:"RecentlyPlayed"`
	RichPresenceMsg string `json:"RichPresenceMsg"`
}

func (us *UserSummary) IsOnline() bool {
	if len(us.RecentlyPlayed) == 0 {
		return false
	}

	t, _ := time.Parse(timeDateFormat, us.RecentlyPlayed[0].LastPlayed)

	return now().Unix() < t.Unix()+180
}

func raCurrentStatus(client *req.Client, user string) (string, error) {
	var j UserSummary

	_, err := client.R().
		SetQueryParam("u", user).
		SetQueryParam("g", "1").
		SetQueryParam("a", "1").
		SetSuccessResult(&j).
		Get(raUserSummaryURL)

	if err != nil {
		return "", err
	}

	if j.ID == 0 {
		return fmt.Sprintf("User %s not found", user), nil
	}

	var sb strings.Builder

	w := func(in, colour string) {
		s := colourString(in, colour)
		sb.WriteString(s)
	}

	sb.WriteString(fmt.Sprintf("%s | ", user))

	if !j.IsOnline() {
		w("Offline", "red")
		return sb.String(), nil
	}

	w("Online", "green")

	sb.WriteString(" | ")

	w(j.RecentlyPlayed[0].Title, gameColour)

	sb.WriteString(" | ")

	w(j.RichPresenceMsg, richPresenceColour)

	return sb.String(), nil
}

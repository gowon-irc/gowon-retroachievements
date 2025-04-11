package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/imroc/req/v3"
)

const (
	timeDateFormat = "2006-01-02 15:04:05"

	raRootURL         = "https://retroachievements.org/API/"
	raAchievementsURL = raRootURL + "API_GetUserRecentAchievements.php"
	raRecentGamesURL  = raRootURL + "API_GetUserRecentlyPlayedGames.php"
	raUserSummaryURL  = raRootURL + "API_GetUserSummary.php"
	raAwardsURL       = raRootURL + "API_GetUserAwards.php"
	raGameProgressURL = raRootURL + "API_GetGameInfoAndUserProgress.php"

	achievementColour       = "cyan"
	gameColour              = "magenta"
	pointsColour            = "green"
	relaxedPointsColour     = "magenta"
	hardcoreColour          = "yellow"
	richPresenceColour      = "yellow"
	rankColour              = "yellow"
	awardColour             = "yellow"
	beatenColour            = "red"
	completedColour         = "cyan"
	masteredColour          = "yellow"
	completionPercentColour = "blue"
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
	GameID       int    `json:"GameID"`
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
		SetQueryParam("m", "43200").
		SetSuccessResult(&j).
		Get(raAchievementsURL)

	if err != nil {
		return "", err
	}

	if len(j) == 0 {
		return fmt.Sprintf("No recent achievements found for user %s", user), nil
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
	RichPresenceMsg     string `json:"RichPresenceMsg"`
	TotalPoints         int    `json:"TotalPoints"`
	TotalTruePoints     int    `json:"TotalTruePoints"`
	TotalSoftcorePoints int    `json:"TotalSoftcorePoints"`
	Rank                int    `json:"Rank"`
	TotalRanked         int    `json:"TotalRanked"`
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

func raPoints(client *req.Client, user string) (string, error) {
	var j UserSummary

	_, err := client.R().
		SetQueryParam("u", user).
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

	w(fmt.Sprintf("Points: %d (%d)", j.TotalPoints, j.TotalTruePoints), pointsColour)

	sb.WriteString(" | ")

	w(fmt.Sprintf("Relaxed: %d", j.TotalSoftcorePoints), relaxedPointsColour)

	sb.WriteString(" | ")

	w(fmt.Sprintf("Rank: %d/%d", j.Rank, j.TotalRanked), rankColour)

	return sb.String(), nil
}

type Awards struct {
	BeatenHardcore int `json:"BeatenHardcoreAwardsCount"`
	BeatenSoftcore int `json:"BeatenSoftcoreAwardsCount"`
	Completed      int `json:"CompletionAwardsCount"`
	Mastered       int `json:"MasteryAwardsCount"`
}

func raAwards(client *req.Client, user string) (string, error) {
	var j Awards

	_, err := client.R().
		SetQueryParam("u", user).
		SetSuccessResult(&j).
		Get(raAwardsURL)

	if err != nil {
		return "", err
	}

	var sb strings.Builder

	w := func(in, colour string) {
		s := colourString(in, colour)
		sb.WriteString(s)
	}

	sb.WriteString(fmt.Sprintf("%s | ", user))

	w(fmt.Sprintf("Beaten: %d (Relaxed: %d)", j.BeatenHardcore, j.BeatenSoftcore), beatenColour)

	sb.WriteString(" | ")

	w(fmt.Sprintf("Completed: %d", j.Completed), completedColour)

	sb.WriteString(" | ")

	w(fmt.Sprintf("Mastered: %d", j.Mastered), masteredColour)

	return sb.String(), nil
}

type GameProgress struct {
	Title                string `json:"Title"`
	Console              string `json:"ConsoleName"`
	Completion           string `json:"UserCompletion"`
	CompletionHardcore   string `json:"UserCompletionHardcore"`
	NumAchievements      int    `json:"NumAchievements"`
	AchievementsRelaxed  int    `json:"NumAwardedToUser"`
	AchievementsHardcore int    `json:"NumAwardedToUserHardcore"`
	Achievements         map[string]struct {
		Points     int    `json:"Points"`
		DateEarned string `json:"DateEarned"`
	} `json:"Achievements"`
	PointsTotal  int    `json:"points_total"`
	HighestAward string `json:"HighestAwardKind"`
}

func (gp *GameProgress) PointsAwarded() string {
	points := 0
	pointsAwarded := 0

	for _, a := range gp.Achievements {
		points += a.Points

		if a.DateEarned != "" {
			pointsAwarded += a.Points
		}
	}

	return fmt.Sprintf("%d/%d", pointsAwarded, points)
}

func raGameProgress(client *req.Client, user string) (string, error) {
	var aj []Achievement

	_, err := client.R().
		SetQueryParam("u", user).
		SetQueryParam("m", "43200").
		SetSuccessResult(&aj).
		Get(raAchievementsURL)

	if err != nil {
		return "", err
	}

	if len(aj) == 0 {
		return fmt.Sprintf("No recent played games found for user %s", user), nil
	}

	gameID := strconv.Itoa(aj[0].GameID)

	var gj GameProgress

	_, err = client.R().
		SetQueryParam("u", user).
		SetQueryParam("g", string(gameID)).
		SetQueryParam("a", "1").
		SetSuccessResult(&gj).
		Get(raGameProgressURL)

	if err != nil {
		return "", err
	}

	var sb strings.Builder

	w := func(in, colour string) {
		s := colourString(in, colour)
		sb.WriteString(s)
	}

	sb.WriteString(fmt.Sprintf("%s | ", user))

	w(fmt.Sprintf("%s (%s)", gj.Title, gj.Console), gameColour)

	sb.WriteString(" | ")

	var cb strings.Builder
	cb.WriteString(fmt.Sprintf("Completion: %s", gj.CompletionHardcore))

	if gj.CompletionHardcore != gj.Completion {
		cb.WriteString(fmt.Sprintf(" (Relaxed: %s)", gj.Completion))
	}

	w(cb.String(), completionPercentColour)

	sb.WriteString(" | ")

	var ab strings.Builder
	ab.WriteString(fmt.Sprintf("Achievements: %d/%d", gj.AchievementsHardcore, gj.NumAchievements))

	if gj.AchievementsHardcore != gj.AchievementsRelaxed {
		ab.WriteString(fmt.Sprintf(" (Relaxed: %d)", gj.AchievementsRelaxed))
	}

	w(ab.String(), achievementColour)

	sb.WriteString(" | ")

	w(fmt.Sprintf("Points: %s", gj.PointsAwarded()), pointsColour)

	if gj.HighestAward != "" {
		awards := map[string]string{
			"beaten-softcore": "Beaten",
			"beaten-hardcore": "Beaten [Hardcore]",
			"completed":       "Completed",
			"mastered":        "Mastered",
		}

		sb.WriteString(" | ")
		w(awards[gj.HighestAward], awardColour)
	}

	return sb.String(), nil
}

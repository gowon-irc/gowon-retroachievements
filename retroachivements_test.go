package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/imroc/req/v3"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func openTestFile(t *testing.T, endpoint, filename string) []byte {
	fp := filepath.Join("testdata", endpoint, filename)
	out, err := os.ReadFile(fp)

	if err != nil {
		t.Fatalf("failed to read test file: %s", err)
	}

	return out
}

func TestColourList(t *testing.T) {
	cases := map[string]struct {
		in       []string
		expected []string
	}{
		"empty list": {
			in:       []string{},
			expected: []string{},
		},
		"single item": {
			in:       []string{"a"},
			expected: []string{"{green}a{clear}"},
		},
		"two items": {
			in:       []string{"a", "b"},
			expected: []string{"{green}a{clear}", "{red}b{clear}"},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			out := colourList(tc.in)

			assert.Equal(t, tc.expected, out)
		})
	}
}

func TestFormatAchievement(t *testing.T) {
	cases := map[string]struct {
		in       Achievement
		expected string
	}{
		"hardcore": {
			in: Achievement{
				HardcoreMode: 1,
				Title:        "achievement",
				Description:  "description",
				Points:       100,
				GameTitle:    "game",
				ConsoleName:  "console",
			},
			expected: "{cyan}achievement (description){clear} | {magenta}game (console){clear} | {green}100 points{clear}{yellow} [Hardcore]{clear}",
		},
		"softcore": {
			in: Achievement{
				HardcoreMode: 0,
				Title:        "achievement",
				Description:  "description",
				Points:       100,
				GameTitle:    "game",
				ConsoleName:  "console",
			},
			expected: "{cyan}achievement (description){clear} | {magenta}game (console){clear} | {green}100 points{clear}",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			out := formatAchievement(tc.in)

			assert.Equal(t, tc.expected, out)
		})
	}
}

func TestRaNewestAchievement(t *testing.T) {
	cases := map[string]struct {
		jsonfn   string
		expected string
		err      error
	}{
		"no achievements": {
			jsonfn:   "no_achievements.json",
			expected: "No achievements found for user user",
			err:      nil,
		},
		"one achievement": {
			jsonfn:   "one_achievement.json",
			expected: "user's newest retroachievement: {cyan}title 1 (description 1){clear} | {magenta}game 1 (console 1){clear} | {green}5 points{clear}{yellow} [Hardcore]{clear}",
			err:      nil,
		},
		"many achievements": {
			jsonfn:   "many_achievements.json",
			expected: "user's newest retroachievement: {cyan}title 1 (description 1){clear} | {magenta}game 1 (console 1){clear} | {green}5 points{clear}{yellow} [Hardcore]{clear}",
			err:      nil,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			json := openTestFile(t, "API_GetUserRecentAchievements", tc.jsonfn)

			client := req.C()
			httpmock.ActivateNonDefault(client.GetClient())
			httpmock.RegisterResponder("GET", raAchievementsURL, func(request *http.Request) (*http.Response, error) {
				resp := httpmock.NewBytesResponse(http.StatusOK, json)
				return resp, nil
			})

			out, err := raNewestAchievement(client, "user")

			assert.Equal(t, tc.expected, out)
			assert.ErrorIs(t, tc.err, err)
		})
	}
}

func TestRaRecentGames(t *testing.T) {
	cases := map[string]struct {
		jsonfn   string
		expected string
		err      error
	}{
		"no games": {
			jsonfn:   "no_games.json",
			expected: "No played games found for user user",
			err:      nil,
		},
		"one game": {
			jsonfn:   "one_game.json",
			expected: "user's last played retro games: {green}Game 1{clear}",
			err:      nil,
		},
		"many games": {
			jsonfn:   "many_games.json",
			expected: "user's last played retro games: {green}Game 1{clear}, {red}Game 2{clear}, {blue}Game 3{clear}",
			err:      nil,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			json := openTestFile(t, "API_GetUserRecentlyPlayedGames", tc.jsonfn)

			client := req.C()
			httpmock.ActivateNonDefault(client.GetClient())
			httpmock.RegisterResponder("GET", raRecentGamesURL, func(request *http.Request) (*http.Response, error) {
				resp := httpmock.NewBytesResponse(http.StatusOK, json)
				return resp, nil
			})

			out, err := raLastGames(client, "user")

			assert.Equal(t, tc.expected, out)
			assert.ErrorIs(t, tc.err, err)
		})
	}
}

func TestUserSummaryIsOnline(t *testing.T) {
	cases := map[string]struct {
		jsonfn   string
		now      string
		expected bool
	}{
		"online": {
			jsonfn:   "summary.json",
			now:      "2024-08-31 17:01:00",
			expected: true,
		},
		"offline": {
			jsonfn:   "summary.json",
			now:      "2024-08-31 17:04:00",
			expected: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			now = func() time.Time { n, _ := time.Parse(timeDateFormat, tc.now); return n }

			j := openTestFile(t, "API_GetUserSummary", "summary.json")
			us := UserSummary{}
			err := json.Unmarshal(j, &us)
			assert.Nil(t, err)

			assert.Equal(t, tc.expected, us.IsOnline())
		})
	}
}

func TestRaCurrentStatus(t *testing.T) {
	cases := map[string]struct {
		jsonfn   string
		now      string
		expected string
		err      error
	}{
		"online": {
			jsonfn:   "summary.json",
			now:      "2024-08-31 17:01:00",
			expected: "user | {green}Online{clear} | {magenta}game 1{clear} | {yellow}Titlescreen{clear}",
			err:      nil,
		},
		"offline": {
			jsonfn:   "summary.json",
			now:      "2024-08-31 17:04:00",
			expected: "user | {red}Offline{clear}",
			err:      nil,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			now = func() time.Time { n, _ := time.Parse(timeDateFormat, tc.now); return n }
			json := openTestFile(t, "API_GetUserSummary", "summary.json")

			client := req.C()
			httpmock.ActivateNonDefault(client.GetClient())
			httpmock.RegisterResponder("GET", raUserSummaryURL, func(request *http.Request) (*http.Response, error) {
				resp := httpmock.NewBytesResponse(http.StatusOK, json)
				return resp, nil
			})

			out, err := raCurrentStatus(client, "user")

			assert.Equal(t, tc.expected, out)
			assert.ErrorIs(t, tc.err, err)
		})
	}
}

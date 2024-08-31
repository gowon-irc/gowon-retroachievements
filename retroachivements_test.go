package main

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

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

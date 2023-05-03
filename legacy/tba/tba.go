package tba

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"net/http"
	"time"
)

type EventParams struct {
	Event  string
	Auth   string
	Secret string
}

type MatchCode struct {
	Level string `json:"comp_level"`
	Set   int    `json:"set_number"`
	Match int    `json:"match_number"`
}

func GetPlayoffCode(bracket_type, match_id int) MatchCode {
	return GetBracket(bracket_type)[match_id]
}

func SendRequest(tba_url string, url string, body []byte, params *EventParams) (*http.Response, error) {
	url = fmt.Sprintf("/api/trusted/v1/event/%s/%s", params.Event, url)
	sig := fmt.Sprintf("%x", md5.Sum(append([]byte(params.Secret+url), body...)))

	url = tba_url + url
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Add("X-TBA-Auth-Id", params.Auth)
	request.Header.Add("X-TBA-Auth-Sig", sig)
	client := http.Client{Timeout: 5 * time.Second}
	return client.Do(request)
}

type Bracket map[int]MatchCode

type playoffRoundInfo struct {
	level           string
	sets            int
	matches_per_set int
}

var playoffRounds = map[int][]playoffRoundInfo{
	// note: most finals rounds have 6 potential matches due to "up to 3" overtime matches
	BRACKET_TYPE_BRACKET_8_TEAM: {
		{level: "qf", sets: 4, matches_per_set: 3},
		{level: "sf", sets: 2, matches_per_set: 3},
		{level: "f", sets: 1, matches_per_set: 6},
	},
	BRACKET_TYPE_ROUND_ROBIN_6_TEAM: {
		{level: "sf", sets: 1, matches_per_set: 15},
		{level: "f", sets: 1, matches_per_set: 6},
	},
	BRACKET_TYPE_CUSTOM: {},
}

func generateBracket(bracket_type int) Bracket {
	rounds, ok := playoffRounds[bracket_type]
	if ok {
		codes := make(Bracket)
		i := 1
		for _, round := range rounds {
			for match := 1; match <= round.matches_per_set; match++ {
				for set := 1; set <= round.sets; set++ {
					codes[i] = MatchCode{Level: round.level, Set: set, Match: match}
					i++
				}
			}
		}
		return codes
	}
	return nil
}

var cachedBrackets = make(map[int]Bracket)

func GetBracket(bracket_type int) Bracket {
	bracket, ok := cachedBrackets[bracket_type]
	if !ok {
		bracket = generateBracket(bracket_type)
		cachedBrackets[bracket_type] = bracket
	}
	return bracket
}

package fms_parser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type fmsScoreInfo2019 struct {
	auto int64
	teleop int64
	fouls int64
	total int64

	baseRP int64  // win-loss-tie RP only
}

type extraMatchInfo2019 struct {
	Dqs []string `json:"dqs"`
	Surrogates []string `json:"surrogates"`
	InvertAuto bool `json:"invert_auto"`
}

func makeExtraMatchInfo2019() extraMatchInfo2019 {
	return extraMatchInfo2019{
		Dqs: make([]string, 0),
		Surrogates: make([]string, 0),
	}
}

func addManualFields2019(breakdown map[string]interface{}, info fmsScoreInfo2019, playoff bool) {
	rp := info.baseRP
	// adjust should be negative when total = 0
	breakdown["adjustPoints"] = info.total - info.auto - info.teleop - info.fouls

	breakdown["rp"] = rp
}

func parseHTMLtoJSON2019(filename string, playoff bool) (map[string]interface{}, error) {
	//////////////////////////////////////////////////
	// Parse html from FMS into TBA-compatible JSON //
	//////////////////////////////////////////////////

	// Open file
	r, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("Error opening file: %s", filename)
	}

	// Read from file
	dom, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("Error reading from file: %s", filename)
	}

	all_json := make(map[string]interface{})

	extra_info := make(map[string]extraMatchInfo2019)
	extra_info["blue"] = makeExtraMatchInfo2019()
	extra_info["red"] = makeExtraMatchInfo2019()
	extra_filename := filename[0:len(filename) - len(path.Ext(filename))] + ".extrajson"
	extra_raw, err := ioutil.ReadFile(extra_filename)
	if err == nil {
		err = json.Unmarshal(extra_raw, &extra_info)
		if err != nil {
			return nil, fmt.Errorf("Error reading JSON from %s: %s", extra_filename, err)
		}
	}

	alliances := map[string]map[string]interface{} {
		"blue": map[string]interface{} {
			"teams": make([]string, 3),
			"surrogates": extra_info["blue"].Surrogates,
			"dqs": extra_info["blue"].Dqs,
			"score": -1,
		},
		"red": map[string]interface{} {
			"teams": make([]string, 3),
			"surrogates": extra_info["red"].Surrogates,
			"dqs": extra_info["red"].Dqs,
			"score": -1,
		},
	}

	breakdown := map[string]map[string]interface{} {
		"blue": make(map[string]interface{}),
		"red": make(map[string]interface{}),
	}

	var scoreInfo struct {
		blue fmsScoreInfo2019
		red fmsScoreInfo2019
	}

	parse_error := ""
	dom.Find("tr").Each(func(i int, s *goquery.Selection){
		columns := s.Children()
		if columns.Length() == 3 {
			var infos [3]string
			columns.Each(func(ii int, column *goquery.Selection){
				infos[ii] = strings.TrimSpace(column.Text())
			})
			identifier := infos[1]

			// Handle each data row
			if identifier == "" {
				// Skip
			} else if identifier == "Final Score" {
				blue_score, err := strconv.ParseInt(infos[0], 10, 0)
				red_score, err := strconv.ParseInt(infos[2], 10, 0)
				if err != nil {
					parse_error = "final score failed"
				}
				breakdown["blue"]["totalPoints"] = blue_score
				breakdown["red"]["totalPoints"] = red_score
				alliances["blue"]["score"] = blue_score
				alliances["red"]["score"] = red_score
				scoreInfo.blue.total = blue_score
				scoreInfo.red.total = red_score
				if (blue_score == red_score) {
					scoreInfo.blue.baseRP = 1
					scoreInfo.red.baseRP = 1
				} else if (blue_score > red_score) {
					scoreInfo.blue.baseRP = 2
					scoreInfo.red.baseRP = 0
				} else {
					scoreInfo.blue.baseRP = 0
					scoreInfo.red.baseRP = 2
				}
			} else if identifier == "Teams" {
				blue_teams := split_and_strip(infos[0], "•")
				red_teams := split_and_strip(infos[2], "•")
				alliances["blue"]["teams"] = []string{
					"frc" + blue_teams[0],
					"frc" + blue_teams[1],
					"frc" + blue_teams[2],
				}
				alliances["red"]["teams"] = []string{
					"frc" + red_teams[0],
					"frc" + red_teams[1],
					"frc" + red_teams[2],
				}
			} else if identifier == "Sandstorm" {
				blue_auto_points, err := strconv.ParseInt(infos[0], 10, 0)
				red_auto_points, err := strconv.ParseInt(infos[2], 10, 0)
				if err != nil {
					parse_error = "sandstorm points failed"
				}
				scoreInfo.blue.auto = blue_auto_points
				breakdown["blue"]["autoPoints"] = blue_auto_points
				scoreInfo.red.auto = red_auto_points
				breakdown["red"]["autoPoints"] = red_auto_points
			} else if identifier == "Teleop" {
				blue_teleop_points, err := strconv.ParseInt(infos[0], 10, 0)
				red_teleop_points, err := strconv.ParseInt(infos[2], 10, 0)
				if err != nil {
					parse_error = "teleop points failed"
				}
				breakdown["blue"]["teleopPoints"] = blue_teleop_points
				breakdown["red"]["teleopPoints"] = red_teleop_points
				scoreInfo.blue.teleop = blue_teleop_points
				scoreInfo.red.teleop = red_teleop_points
			} else if identifier == "Fouls/Techs Committed" {
				blue_foul := split_and_strip(infos[0], "•")
				red_foul := split_and_strip(infos[2], "•")
				breakdown["blue"]["foulCount"], err = strconv.ParseInt(blue_foul[0], 10, 0)
				breakdown["blue"]["techFoulCount"], err = strconv.ParseInt(blue_foul[1], 10, 0)
				breakdown["red"]["foulCount"], err = strconv.ParseInt(red_foul[0], 10, 0)
				breakdown["red"]["techFoulCount"], err = strconv.ParseInt(red_foul[1], 10, 0)
				if err != nil {
					parse_error = "foul/tech count failed"
				}
			} else if identifier == "Foul Points" {
				blue_foul_points, err := strconv.ParseInt(infos[0], 10, 0)
				red_foul_points, err := strconv.ParseInt(infos[2], 10, 0)
				if err != nil {
					parse_error = "foul points failed"
				}
				breakdown["blue"]["foulPoints"] = blue_foul_points
				breakdown["red"]["foulPoints"] = red_foul_points
				scoreInfo.blue.fouls = blue_foul_points
				scoreInfo.red.fouls = red_foul_points
			} else {
				breakdown["blue"]["!" + identifier] = strings.TrimSpace(infos[0])
				breakdown["red"]["!" + identifier] = strings.TrimSpace(infos[2])
			}
		}
	})

	addManualFields2019(breakdown["blue"], scoreInfo.blue, playoff)
	addManualFields2019(breakdown["red"], scoreInfo.red, playoff)

	if parse_error != "" {
		return nil, fmt.Errorf("Parse error: %s", parse_error)
	}

	all_json["alliances"] = alliances
	all_json["score_breakdown"] = breakdown

	return all_json, nil
}

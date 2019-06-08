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
	rocketRP bool
	habRP bool

	fields map[string]int64  // indexed by values of simpleFields2019
}

func makeFmsScoreInfo2019() fmsScoreInfo2019 {
	return fmsScoreInfo2019{
		fields: make(map[string]int64),
	}
}

type extraMatchInfo2019 struct {
	Dqs []string `json:"dqs"`
	Surrogates []string `json:"surrogates"`
	AddRpRocket bool `json:"add_rp_rocket"`
	AddRpHabClimb bool `json:"add_rp_hab_climb"`
}

func makeExtraMatchInfo2019() extraMatchInfo2019 {
	return extraMatchInfo2019{
		Dqs: make([]string, 0),
		Surrogates: make([]string, 0),
	}
}

func addManualFields2019(breakdown map[string]interface{}, info fmsScoreInfo2019, extra extraMatchInfo2019, playoff bool) {
	rp := info.baseRP
	// adjust should be negative when total = 0
	breakdown["adjustPoints"] = info.total - info.auto - info.teleop - info.fouls

	rocket_rp := info.rocketRP || extra.AddRpRocket
	breakdown["completeRocketRankingPoint"] = rocket_rp
	if (rocket_rp) {
		rp++
	}

	hab_rp := info.habRP || extra.AddRpHabClimb
	breakdown["habDockingRankingPoint"] = hab_rp
	if (hab_rp) {
		rp++
	}

	breakdown["rp"] = rp
}

// map FMS names to API names of basic integer fields
var simpleFields2019 = map[string]string {
	"Cargo Points": "cargoPoints",
	"HAB Climb Points": "habClimbPoints",
	"Hatch Panel Points": "hatchPanelPoints",
	"Sandstorm Bonus Points": "sandStormBonusPoints",
}

const (
	K2019_BAY_NONE = "None"
	K2019_BAY_PANEL = "Panel"
	K2019_BAY_PANEL_AND_CARGO = "PanelAndCargo"
)

func parseRocketOrCargoShip2019(raw string) ([]string, error) {
	raw = strings.Replace(raw, "•", " ", -1)
	out := strings.Fields(raw)
	if len(out) != 6 && len(out) != 8 {
		return nil, fmt.Errorf("Invalid cargo/rocket ship: expected 6 or 8 bays, got %d", len(out))
	}
	for i := range out {
		if out[i] == "N" {
			out[i] = K2019_BAY_NONE
		} else if out[i] == "P" {
			out[i] = K2019_BAY_PANEL
		} else if out[i] == "B" {
			out[i] = K2019_BAY_PANEL_AND_CARGO
		} else {
			return nil, fmt.Errorf("Invalid cargo/rocket ship item: %s", out[i])
		}
	}
	return out, nil
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

	var scoreInfo = struct {
		blue fmsScoreInfo2019
		red fmsScoreInfo2019
	}{
		makeFmsScoreInfo2019(),
		makeFmsScoreInfo2019(),
	}

	parse_error := ""

	parseRocketOrCargoShipWrapper := func(raw string) []string {
		out, err := parseRocketOrCargoShip2019(raw)
		if err != nil {
			parse_error += err.Error() + "\n"
		}
		return out
	}

	assignRocket := func(alliance_breakdown map[string]interface{}, score_info *fmsScoreInfo2019, parsedRocket []string, loc string) {
		// modifies alliance_breakdown, score_info
		// loc: Near | Far
		alliance_breakdown["topLeftRocket" + loc]  = parsedRocket[0]
		alliance_breakdown["topRightRocket" + loc] = parsedRocket[1]
		alliance_breakdown["midLeftRocket" + loc]  = parsedRocket[2]
		alliance_breakdown["midRightRocket" + loc] = parsedRocket[3]
		alliance_breakdown["lowLeftRocket" + loc]  = parsedRocket[4]
		alliance_breakdown["lowRightRocket" + loc] = parsedRocket[5]

		complete := true
		for _, s := range parsedRocket {
			if s != K2019_BAY_PANEL_AND_CARGO {
				complete = false
			}
		}
		alliance_breakdown["completedRocket" + loc] = complete
		if complete {
			score_info.rocketRP = true
		}
	}

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
			} else if identifier == "Ranking Points" {
				// discard because it's always 0
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
			} else if identifier == "Pre-Match Robot Levels" {
				blue := split_and_strip(infos[0], "•")
				red := split_and_strip(infos[2], "•")
				breakdown["blue"]["preMatchLevelRobot1"] = blue[0]
				breakdown["blue"]["preMatchLevelRobot2"] = blue[1]
				breakdown["blue"]["preMatchLevelRobot3"] = blue[2]
				breakdown["red"]["preMatchLevelRobot1"] = red[0]
				breakdown["red"]["preMatchLevelRobot2"] = red[1]
				breakdown["red"]["preMatchLevelRobot3"] = red[2]
			} else if identifier == "HAB Line" {
				blue := split_and_strip(infos[0], "•")
				red := split_and_strip(infos[2], "•")
				process := func(arr []string) {
					for i := range arr {
						if (strings.Contains(arr[i], "Sandstorm")) {
							arr[i] = "CrossedHabLineInSandstorm"
						} else if (strings.Contains(arr[i], "Teleop")) {
							arr[i] = "CrossedHabLineInTeleop"
						} else {
							arr[i] = "None"
						}
					}
				}
				process(red)
				process(blue)
				breakdown["blue"]["habLineRobot1"] = blue[0]
				breakdown["blue"]["habLineRobot2"] = blue[1]
				breakdown["blue"]["habLineRobot3"] = blue[2]
				breakdown["red"]["habLineRobot1"] = red[0]
				breakdown["red"]["habLineRobot2"] = red[1]
				breakdown["red"]["habLineRobot3"] = red[2]
			} else if identifier == "HAB Line in Sandstorm" {
				// skip because provided by "Hab Line"
			} else if identifier == "HAB Endgame Climb" {
				blue := split_and_strip(infos[0], "•")
				red := split_and_strip(infos[2], "•")
				breakdown["blue"]["endgameRobot1"] = blue[0]
				breakdown["blue"]["endgameRobot2"] = blue[1]
				breakdown["blue"]["endgameRobot3"] = blue[2]
				breakdown["red"]["endgameRobot1"] = red[0]
				breakdown["red"]["endgameRobot2"] = red[1]
				breakdown["red"]["endgameRobot3"] = red[2]
			} else if identifier == "Cargoships" {
				blue := parseRocketOrCargoShipWrapper(infos[0])
				red := parseRocketOrCargoShipWrapper(infos[2])
				if blue == nil || red == nil {
					return;
				}

				breakdown["blue"]["bay1"] = blue[7]
				breakdown["blue"]["bay2"] = blue[6]
				breakdown["blue"]["bay3"] = blue[5]
				breakdown["blue"]["bay4"] = blue[4]
				breakdown["blue"]["bay5"] = blue[3]
				breakdown["blue"]["bay6"] = blue[0]
				breakdown["blue"]["bay7"] = blue[1]
				breakdown["blue"]["bay8"] = blue[2]

				breakdown["red"]["bay1"] = red[5]
				breakdown["red"]["bay2"] = red[6]
				breakdown["red"]["bay3"] = red[7]
				breakdown["red"]["bay4"] = red[4]
				breakdown["red"]["bay5"] = red[3]
				breakdown["red"]["bay6"] = red[2]
				breakdown["red"]["bay7"] = red[1]
				breakdown["red"]["bay8"] = red[0]
			} else if identifier == "Far SideRocket" {
				blue := parseRocketOrCargoShipWrapper(infos[0])
				red := parseRocketOrCargoShipWrapper(infos[2])
				if blue == nil || red == nil {
					return;
				}

				assignRocket(breakdown["blue"], &scoreInfo.blue, blue, "Far")
				assignRocket(breakdown["red"], &scoreInfo.red, red, "Far")
			} else if identifier == "Scoring Table SideRocket" {
				blue := parseRocketOrCargoShipWrapper(infos[0])
				red := parseRocketOrCargoShipWrapper(infos[2])
				if blue == nil || red == nil {
					return;
				}

				assignRocket(breakdown["blue"], &scoreInfo.blue, blue, "Near")
				assignRocket(breakdown["red"], &scoreInfo.red, red, "Near")
			} else if apiField, ok := simpleFields2019[identifier]; ok {
				blue_points, err := strconv.ParseInt(infos[0], 10, 0)
				red_points, err := strconv.ParseInt(infos[2], 10, 0)
				if err != nil {
					parse_error = "parse integer field \"" + apiField + "\" failed"
				}
				breakdown["blue"][apiField] = blue_points
				breakdown["red"][apiField] = red_points
				scoreInfo.blue.fields[apiField] = blue_points
				scoreInfo.red.fields[apiField] = red_points

				if apiField == "habClimbPoints" {
					scoreInfo.blue.habRP = (blue_points >= 15)
					scoreInfo.red.habRP = (red_points >= 15)
				}
			} else {
				breakdown["blue"]["!" + identifier] = strings.TrimSpace(infos[0])
				breakdown["red"]["!" + identifier] = strings.TrimSpace(infos[2])
			}
		}
	})

	addManualFields2019(breakdown["blue"], scoreInfo.blue, extra_info["blue"], playoff)
	addManualFields2019(breakdown["red"], scoreInfo.red, extra_info["red"], playoff)

	if parse_error != "" {
		return nil, fmt.Errorf("Parse error: %s", parse_error)
	}

	all_json["alliances"] = alliances
	all_json["score_breakdown"] = breakdown

	return all_json, nil
}

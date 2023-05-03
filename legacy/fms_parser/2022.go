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

type fmsScoreInfo2022 struct {
	auto   int
	teleop int
	fouls  int
	total  int
}

func makeFmsScoreInfo2022() fmsScoreInfo2022 {
	return fmsScoreInfo2022{}
}

type extraMatchAllianceInfo2022 struct {
	Dqs        []string `json:"dqs"`
	Surrogates []string `json:"surrogates"`
}

func makeExtraMatchAllianceInfo2022() extraMatchAllianceInfo2022 {
	return extraMatchAllianceInfo2022{
		Dqs:        make([]string, 0),
		Surrogates: make([]string, 0),
	}
}

func addManualFields2022(breakdown map[string]interface{}, info fmsScoreInfo2022, extra extraMatchAllianceInfo2022, playoff bool) {
	if _, ok := breakdown["adjustPoints"]; !ok {
		// adjust should be negative when total = 0
		breakdown["adjustPoints"] = info.total - info.auto - info.teleop - info.fouls
	}
}

// map FMS names to API names of basic integer fields
var simpleFields2022 = map[string]string{
	"auto cargo total scored":   "autoCargoTotal",
	"teleop cargo total scored": "teleopCargoTotal",
	"match cargo total scored":  "matchCargoTotal",
	"endgame points":            "endgamePoints",
	"taxi points":               "autoTaxiPoints",
	"adjustments":               "adjustPoints",
}

var RP_BADGE_NAMES_2022 = map[string]string{
	"Cargo Bonus Ranking Point Achieved":  "cargoBonusRankingPoint",
	"Hangar Bonus Ranking Point Achieved": "hangarBonusRankingPoint",
}

var DEFAULT_BREAKDOWN_VALUES_2022 = map[string]any{
	"adjustPoints":            0,
	"autoCargoLowerBlue":      0,
	"autoCargoLowerFar":       0,
	"autoCargoLowerNear":      0,
	"autoCargoLowerRed":       0,
	"autoCargoPoints":         0,
	"autoCargoTotal":          0,
	"autoCargoUpperBlue":      0,
	"autoCargoUpperFar":       0,
	"autoCargoUpperNear":      0,
	"autoCargoUpperRed":       0,
	"autoPoints":              0,
	"autoTaxiPoints":          0,
	"cargoBonusRankingPoint":  false,
	"endgamePoints":           0,
	"endgameRobot1":           "None",
	"endgameRobot2":           "None",
	"endgameRobot3":           "None",
	"foulCount":               0,
	"foulPoints":              0,
	"hangarBonusRankingPoint": false,
	"matchCargoTotal":         0,
	"quintetAchieved":         false,
	"rp":                      0,
	"taxiRobot1":              "No",
	"taxiRobot2":              "No",
	"taxiRobot3":              "No",
	"techFoulCount":           0,
	"teleopCargoLowerBlue":    0,
	"teleopCargoLowerFar":     0,
	"teleopCargoLowerNear":    0,
	"teleopCargoLowerRed":     0,
	"teleopCargoPoints":       0,
	"teleopCargoTotal":        0,
	"teleopCargoUpperBlue":    0,
	"teleopCargoUpperFar":     0,
	"teleopCargoUpperNear":    0,
	"teleopCargoUpperRed":     0,
	"teleopPoints":            0,
	"totalPoints":             0,
}

func parseHTMLtoJSON2022(filename string, playoff bool) (map[string]interface{}, error) {
	//////////////////////////////////////////////////
	// Parse html from FMS into TBA-compatible JSON //
	//////////////////////////////////////////////////

	// Open file
	r, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("Error opening file: %s: %s", filename, err)
	}
	defer r.Close()

	// Read from file
	dom, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("Error reading from file: %s: %s", filename, err)
	}

	all_json := make(map[string]interface{})

	extra_info := make(map[string]extraMatchAllianceInfo2022)
	extra_info["blue"] = makeExtraMatchAllianceInfo2022()
	extra_info["red"] = makeExtraMatchAllianceInfo2022()
	extra_filename := filename[0:len(filename)-len(path.Ext(filename))] + ".extrajson"
	extra_raw, err := ioutil.ReadFile(extra_filename)
	if err == nil {
		err = json.Unmarshal(extra_raw, &extra_info)
		if err != nil {
			return nil, fmt.Errorf("Error reading JSON from %s: %s", extra_filename, err)
		}
	}

	alliances := map[string]map[string]interface{}{
		"blue": {
			"teams":      make([]string, 3),
			"surrogates": extra_info["blue"].Surrogates,
			"dqs":        extra_info["blue"].Dqs,
			"score":      -1,
		},
		"red": {
			"teams":      make([]string, 3),
			"surrogates": extra_info["red"].Surrogates,
			"dqs":        extra_info["red"].Dqs,
			"score":      -1,
		},
	}

	breakdown := map[string]map[string]interface{}{
		"blue": make(map[string]interface{}),
		"red":  make(map[string]interface{}),
	}

	var scoreInfo = struct {
		blue fmsScoreInfo2022
		red  fmsScoreInfo2022
	}{
		makeFmsScoreInfo2022(),
		makeFmsScoreInfo2022(),
	}

	parse_errors := make([]string, 0)

	checkParseInt := func(s, desc string) int {
		n, err := strconv.ParseInt(s, 10, 0)
		if err != nil {
			panic(fmt.Sprintf("parse int %s failed: %s", desc, err))
		}
		return int(n)
	}

	match_phase := ""
	validateMatchPhase := func(desc string) {
		if match_phase == "" {
			panic(fmt.Sprintf("no active match phase: %s", desc))
		}
	}

	assignCargoScoredLocations := func(hub_name, alliance string, container *goquery.Selection) {
		desc := fmt.Sprintf("%s hub cargo scored for %s alliance", hub_name, alliance)
		validateMatchPhase(desc)
		cells := container.Find("div[title]")
		if cells.Length() != 4 {
			panic(fmt.Sprintf("invalid cell count: %d in %s", cells.Length(), desc))
		}
		cells.Each(func(_ int, cell *goquery.Selection) {
			title, _ := cell.Attr("title")
			exit_name := strings.Split(title, " ")[1]
			breakdown[alliance][fmt.Sprintf("%sCargo%s%s", match_phase, strings.Title(hub_name), exit_name)] =
				checkParseInt(cell.Text(), fmt.Sprintf("%s: %s", desc, title))
		})
	}

	dom.Find("tr").Each(func(i int, s *goquery.Selection) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Parse error in %s: %s\n", filename, r)
				parse_errors = append(parse_errors, fmt.Sprint(r))
			}
		}()

		columns := s.Children()
		if columns.Length() == 3 {
			row_name := strings.ToLower(strings.TrimSpace(columns.Eq(0).Text()))
			if row_name == "" || row_name == "match score item" {
				return // continue
			}
			if row_name == "taxi" {
				match_phase = "auto"
			}

			blue_cell := columns.Eq(1)
			red_cell := columns.Eq(2)
			blue_text := strings.TrimSpace(blue_cell.Text())
			red_text := strings.TrimSpace(red_cell.Text())

			parseIntWrapper := func(s, alliance string) int {
				return checkParseInt(s, alliance+" "+row_name)
			}

			// Handle each data row
			if api_field, ok := simpleFields2022[row_name]; ok {
				assignBreakdownAllianceFields(breakdown, api_field, identity_fn[int], breakdownAllianceFields[int]{
					blue: checkParseInt(blue_text, "blue "+api_field),
					red:  checkParseInt(red_text, "red "+api_field),
				})
			} else if row_name == "teams" {
				blue_teams := split_and_strip(blue_text, "\n")
				red_teams := split_and_strip(red_text, "\n")
				assignTbaTeams(alliances, breakdownRobotFields[string]{
					blue: blue_teams,
					red:  red_teams,
				})
			} else if row_name == "final score" {
				blue_score := checkParseInt(blue_text, "blue final score")
				red_score := checkParseInt(red_text, "red final score")
				breakdown["blue"]["totalPoints"] = blue_score
				breakdown["red"]["totalPoints"] = red_score
				alliances["blue"]["score"] = blue_score
				alliances["red"]["score"] = red_score
				scoreInfo.blue.total = blue_score
				scoreInfo.red.total = red_score
			} else if row_name == "ranking points" {
				blue_rp := checkParseInt(blue_text, "blue ranking points")
				red_rp := checkParseInt(red_text, "red ranking points")
				breakdown["blue"]["rp"] = blue_rp
				breakdown["red"]["rp"] = red_rp
			} else if row_name == "autonomous points" {
				blue_points := checkParseInt(blue_text, "blue "+row_name)
				red_points := checkParseInt(red_text, "red "+row_name)
				assignBreakdownAllianceFields(breakdown, "autoPoints", identity_fn[int], breakdownAllianceFields[int]{
					blue: blue_points,
					red:  red_points,
				})
				scoreInfo.blue.auto = blue_points
				scoreInfo.red.auto = red_points
				match_phase = "teleop"
			} else if row_name == "teleop points" {
				blue_points := checkParseInt(blue_text, "blue "+row_name)
				red_points := checkParseInt(red_text, "red "+row_name)
				assignBreakdownAllianceFields(breakdown, "teleopPoints", identity_fn[int], breakdownAllianceFields[int]{
					blue: blue_points,
					red:  red_points,
				})
				scoreInfo.blue.teleop = blue_points
				scoreInfo.red.teleop = red_points
				match_phase = ""
			} else if row_name == "foul points" {
				blue_points := checkParseInt(blue_text, "blue "+row_name)
				red_points := checkParseInt(red_text, "red "+row_name)
				assignBreakdownAllianceFields(breakdown, "foulPoints", identity_fn[int], breakdownAllianceFields[int]{
					blue: blue_points,
					red:  red_points,
				})
				scoreInfo.blue.fouls = blue_points
				scoreInfo.red.fouls = red_points
			} else if row_name == "fouls/techs committed" {
				assignBreakdownAllianceMultipleFields(breakdown, []string{"foulCount", "techFoulCount"}, parseIntWrapper, breakdownAllianceMultipleFields[string]{
					blue: split_and_strip(blue_text, "•"),
					red:  split_and_strip(red_text, "•"),
				})
			} else if row_name == "taxi" {
				assignBreakdownRobotFields(breakdown, "taxiRobot", boolToYesNo, breakdownRobotFields[bool]{
					blue: iconsToBools(blue_cell, 3, "fa-check", "fa-times"),
					red:  iconsToBools(red_cell, 3, "fa-check", "fa-times"),
				})
			} else if row_name == "cargo points" {
				validateMatchPhase(row_name)
				assignBreakdownAllianceFields(breakdown, match_phase+"CargoPoints", identity_fn[int], breakdownAllianceFields[int]{
					blue: checkParseInt(blue_text, "blue "+match_phase+" cargo points"),
					red:  checkParseInt(red_text, "red "+match_phase+" cargo points"),
				})
			} else if row_name == "quintet achieved?" {
				assignBreakdownAllianceFields(breakdown, "quintetAchieved", identity_fn[bool], breakdownAllianceFields[bool]{
					blue: iconsToBools(blue_cell, 1, "fa-check", "fa-times")[0],
					red:  iconsToBools(red_cell, 1, "fa-check", "fa-times")[0],
				})
			} else if row_name == "lower hub cargo scored" || row_name == "upper hub cargo scored" {
				hub_name := strings.Split(row_name, " ")[0]
				assignCargoScoredLocations(hub_name, "blue", blue_cell)
				assignCargoScoredLocations(hub_name, "red", red_cell)
			} else if row_name == "endgame" {
				assignBreakdownRobotFields(breakdown, "endgameRobot", identity_fn[string], breakdownRobotFields[string]{
					blue: split_and_strip(blue_text, "\n"),
					red:  split_and_strip(red_text, "\n"),
				})
			} else if row_name == "achievement badges" {
				assignBreakdownRpFromBadges(breakdown, RP_BADGE_NAMES_2022, breakdownAllianceFields[*goquery.Selection]{
					blue: blue_cell,
					red:  red_cell,
				})
			} else {
				breakdown["blue"]["!"+row_name] = blue_text
				breakdown["red"]["!"+row_name] = red_text
			}
		}
	})

	if playoff {
		// set bonus RPs to false since the row is absent
		assignBreakdownAllianceFieldsConst(breakdown, "rp", 0)
		for _, field := range RP_BADGE_NAMES_2022 {
			assignBreakdownAllianceFieldsConst(breakdown, field, false)
		}
	}

	addManualFields2022(breakdown["blue"], scoreInfo.blue, extra_info["blue"], playoff)
	addManualFields2022(breakdown["red"], scoreInfo.red, extra_info["red"], playoff)

	if len(parse_errors) > 0 {
		return nil, fmt.Errorf("Parse error (%d):\n%s", len(parse_errors), strings.Join(parse_errors, "\n"))
	}

	all_json["alliances"] = alliances
	all_json["score_breakdown"] = breakdown

	return all_json, nil
}

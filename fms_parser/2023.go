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

type fmsScoreInfo2023 struct {
	auto   int
	teleop int
	fouls  int
	total  int
	// year-specific:
	auto_charge_station   int
	teleop_charge_station int
	link                  int
}

func makeFmsScoreInfo2023() fmsScoreInfo2023 {
	return fmsScoreInfo2023{}
}

type extraMatchAllianceInfo2023 struct {
	Dqs        []string `json:"dqs"`
	Surrogates []string `json:"surrogates"`
}

func makeExtraMatchAllianceInfo2023() extraMatchAllianceInfo2023 {
	return extraMatchAllianceInfo2023{
		Dqs:        make([]string, 0),
		Surrogates: make([]string, 0),
	}
}

func addManualFields2023(breakdown map[string]interface{}, info fmsScoreInfo2023, extra extraMatchAllianceInfo2023, playoff bool) {
	breakdown["totalChargeStationPoints"] = info.auto_charge_station + info.teleop_charge_station

	if _, ok := breakdown["adjustPoints"]; !ok {
		// adjust should be negative when total = 0
		breakdown["adjustPoints"] = info.total - info.auto - info.teleop - info.fouls - info.link
	}
}

const (
	K2023_COMMUNITY_BOTTOM = "Bottom"
	K2023_COMMUNITY_MIDDLE = "Middle"
	K2023_COMMUNITY_TOP    = "Top"

	K2023_COMMUNITY_NONE = "None"
	K2023_COMMUNITY_CUBE = "Cube"
	K2023_COMMUNITY_CONE = "Cone"
)

// map FMS names (lowercase) to API names of basic integer fields
var simpleIntFields2023 = map[string]string{
	"coop game piece count":  "coopGamePieceCount",
	"mobility points":        "autoMobilityPoints",
	"endgame park points":    "endGameParkPoints",
	"extra game piece count": "extraGamePieceCount", // new for championship
	"adjustments":            "adjustPoints",
}

// Map FMS names (lowercase) to API name suffixes of basic integer fields.
// The match phase ("auto" or "teleop") will be prepended to the API names as appropriate.
var simpleIntMatchPhaseFields2023 = map[string]string{
	"game piece count":  "GamePieceCount",
	"game piece points": "GamePiecePoints",
}

var simpleStringFields2023 = map[string]string{
	"auto charge station":    "autoBridgeState",
	"endgame charge station": "endGameBridgeState",
}

var simpleIconFields2023 = map[string]string{
	"docked?":                    "autoDocked",
	"activation bonus?":          "activationBonusAchieved",
	"sustainability bonus?":      "sustainabilityBonusAchieved",
	"coopertition criteria met?": "coopertitionCriteriaMet",
}

var penaltyFields2023 = map[string]string{
	"G405": "g405Penalty",
	"H111": "h111Penalty",
}

var DEFAULT_BREAKDOWN_VALUES_2023 = map[string]any{}

const COMMUNITY_ROW_LENGTH = 9

type Community2023 struct {
	pieces             map[string][]string
	link_start_indexes map[string][]int
}

func makeCommunity2023() *Community2023 {
	return &Community2023{
		pieces:             make(map[string][]string),
		link_start_indexes: make(map[string][]int),
	}
}

func (self Community2023) isComplete() bool {
	for _, k := range []string{K2023_COMMUNITY_BOTTOM, K2023_COMMUNITY_MIDDLE, K2023_COMMUNITY_TOP} {
		row, ok := self.pieces[k]
		if !ok {
			return false
		}
		if len(row) != COMMUNITY_ROW_LENGTH {
			return false
		}
	}
	return true
}

func (self Community2023) parseCommunityRow(key string, cell *goquery.Selection) {
	icons := cell.Find("span.icon")
	if icons.Length() != COMMUNITY_ROW_LENGTH {
		panic(fmt.Sprintf("unexpected community row icon count: %d", icons.Length()))
	}

	pieces := make([]string, COMMUNITY_ROW_LENGTH)
	link_members := make([]bool, COMMUNITY_ROW_LENGTH)

	icons.Each(func(i int, icon *goquery.Selection) {
		svg := icon.Find("svg")
		piece := ""
		if svg.HasClass("bi-dot") {
			piece = K2023_COMMUNITY_NONE
		} else if svg.HasClass("bi-box") {
			piece = K2023_COMMUNITY_CUBE
		} else if svg.HasClass("bi-cone") {
			piece = K2023_COMMUNITY_CONE
		}
		if piece == "" {
			panic(fmt.Sprintf("unknown community icon: %s", svg.AttrOr("class", "")))
		}
		pieces[i] = piece
		link_members[i] = icon.HasClass("community-img-link")
	})

	for i := 0; i < COMMUNITY_ROW_LENGTH; i++ {
		if link_members[i] {
			self.link_start_indexes[key] = append(self.link_start_indexes[key], i)
			i += 2
		}
	}

	self.pieces[key] = pieces
}

func (self Community2023) assignPiecesToBreakdown(breakdown map[string]interface{}, field string) {
	community := make(map[string][]string)
	for key, pieces := range self.pieces {
		community[key[0:1]] = pieces
	}
	breakdown[field] = community
}

func (self Community2023) assignLinksToBreakdown(breakdown map[string]interface{}, field string) {
	links := make([]interface{}, 0)
	// match FMS order
	for _, row := range []string{K2023_COMMUNITY_BOTTOM, K2023_COMMUNITY_MIDDLE, K2023_COMMUNITY_TOP} {
		adjusted_row := row
		if row == K2023_COMMUNITY_MIDDLE {
			adjusted_row = "Mid" // why
		}

		for _, start_index := range self.link_start_indexes[row] {
			links = append(links, map[string]interface{}{
				"row":   adjusted_row,
				"nodes": []int{start_index, start_index + 1, start_index + 2},
			})
		}
	}
	breakdown[field] = links
}

func parseHTMLtoJSON2023(filename string, playoff bool) (map[string]interface{}, error) {
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

	extra_info := make(map[string]extraMatchAllianceInfo2023)
	extra_info["blue"] = makeExtraMatchAllianceInfo2023()
	extra_info["red"] = makeExtraMatchAllianceInfo2023()
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
		blue fmsScoreInfo2023
		red  fmsScoreInfo2023
	}{
		makeFmsScoreInfo2023(),
		makeFmsScoreInfo2023(),
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
	matchPhaseWithEndGame := func() string {
		validateMatchPhase(match_phase)
		if match_phase == "teleop" {
			return "endGame"
		}
		return match_phase
	}

	var cur_community struct {
		blue *Community2023
		red  *Community2023
	}
	communityRowToKey := func(row_name string) string {
		if row_name == "bottom" {
			return K2023_COMMUNITY_BOTTOM
		} else if row_name == "middle" {
			return K2023_COMMUNITY_MIDDLE
		} else if row_name == "top" {
			return K2023_COMMUNITY_TOP
		}
		panic("invalid community row name: " + row_name)
	}

	dom.Find("tr").Each(func(i int, s *goquery.Selection) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Parse error in %s: %s\n", filename, r)
				parse_errors = append(parse_errors, fmt.Sprint(r))
			}
		}()

		columns := s.Children()
		if columns.Length() < 1 {
			return // continue
		}

		row_name := strings.ToLower(strings.TrimSpace(columns.Eq(0).Text()))
		if row_name == "" || row_name == "match score item" {
			return // continue
		}

		if row_name == "community" {
			if cur_community.red != nil {
				panic("found community before end ")
			}
			cur_community.blue = makeCommunity2023()
			cur_community.red = makeCommunity2023()
			return // continue
		}

		if columns.Length() == 3 {
			if row_name == "mobility" {
				match_phase = "auto"
			}

			blue_cell := columns.Eq(1)
			red_cell := columns.Eq(2)
			blue_text := strings.TrimSpace(blue_cell.Text())
			red_text := strings.TrimSpace(red_cell.Text())

			parseIntWrapper := func(s, alliance string) int {
				return checkParseInt(s, alliance+" "+row_name)
			}

			if cur_community.red != nil {
				cur_community.blue.parseCommunityRow(communityRowToKey(row_name), blue_cell)
				cur_community.red.parseCommunityRow(communityRowToKey(row_name), red_cell)

				if cur_community.red.isComplete() {
					api_field := match_phase + "Community"
					cur_community.blue.assignPiecesToBreakdown(breakdown["blue"], api_field)
					cur_community.red.assignPiecesToBreakdown(breakdown["red"], api_field)
					if match_phase == "teleop" {
						cur_community.blue.assignLinksToBreakdown(breakdown["blue"], "links")
						cur_community.red.assignLinksToBreakdown(breakdown["red"], "links")
					}
					cur_community.blue = nil
					cur_community.red = nil
				}
				return // continue
			}

			// Handle each data row
			if api_field, ok := simpleStringFields2023[row_name]; ok {
				assignBreakdownAllianceFields(breakdown, api_field, identity_fn[string], breakdownAllianceFields[string]{
					blue: blue_text,
					red:  red_text,
				})
			} else if api_field, ok := simpleIntFields2023[row_name]; ok {
				assignBreakdownAllianceFields(breakdown, api_field, identity_fn[int], breakdownAllianceFields[int]{
					blue: checkParseInt(blue_text, "blue "+api_field),
					red:  checkParseInt(red_text, "red "+api_field),
				})
			} else if api_field_suffix, ok := simpleIntMatchPhaseFields2023[row_name]; ok {
				api_field := match_phase + api_field_suffix
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

				// begin year-specific
			} else if api_field, ok := simpleIconFields2023[row_name]; ok {
				assignBreakdownAllianceFields[bool](breakdown, api_field, identity_fn[bool], breakdownAllianceFields[bool]{
					blue: iconToBool(blue_cell.Find("i"), "fa-check", "fa-times"),
					red:  iconToBool(red_cell.Find("i"), "fa-check", "fa-times"),
				})
			} else if row_name == "mobility" {
				assignBreakdownRobotFields(breakdown, "mobilityRobot", boolToYesNo, breakdownRobotFields[bool]{
					blue: iconsToBools(blue_cell, 3, "fa-check", "fa-times"),
					red:  iconsToBools(red_cell, 3, "fa-check", "fa-times"),
				})
			} else if row_name == "charge station points" {
				api_field := matchPhaseWithEndGame() + "ChargeStationPoints"
				blue_points := checkParseInt(blue_text, "blue "+row_name)
				red_points := checkParseInt(red_text, "red "+row_name)
				assignBreakdownAllianceFields[int](breakdown, api_field, identity_fn[int], breakdownAllianceFields[int]{
					blue: blue_points,
					red:  red_points,
				})
				if match_phase == "auto" {
					scoreInfo.blue.auto_charge_station = blue_points
					scoreInfo.red.auto_charge_station = red_points
				} else {
					scoreInfo.blue.teleop_charge_station = blue_points
					scoreInfo.red.teleop_charge_station = red_points
				}
			} else if row_name == "charge station" {
				api_field_prefix := matchPhaseWithEndGame() + "ChargeStationRobot"
				assignBreakdownRobotFields(breakdown, api_field_prefix, identity_fn[string], breakdownRobotFields[string]{
					blue: split_and_strip(blue_text, "\n"),
					red:  split_and_strip(red_text, "\n"),
				})
			} else if row_name == "link points" {
				blue_points := checkParseInt(blue_text, "blue "+row_name)
				red_points := checkParseInt(red_text, "red "+row_name)
				assignBreakdownAllianceFields(breakdown, "linkPoints", identity_fn[int], breakdownAllianceFields[int]{
					blue: blue_points,
					red:  red_points,
				})
				scoreInfo.blue.link = blue_points
				scoreInfo.red.link = red_points

				// begin new championship fields
			} else if row_name == "penalties" {
				assignPenaltyFields(breakdown, penaltyFields2023, breakdownAllianceFields[*goquery.Selection]{
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
		// set "rp" to 0 since the row is absent
		assignBreakdownAllianceFieldsConst(breakdown, "rp", 0)
	}

	addManualFields2023(breakdown["blue"], scoreInfo.blue, extra_info["blue"], playoff)
	addManualFields2023(breakdown["red"], scoreInfo.red, extra_info["red"], playoff)

	if len(parse_errors) > 0 {
		return nil, fmt.Errorf("Parse error (%d):\n%s", len(parse_errors), strings.Join(parse_errors, "\n"))
	}

	all_json["alliances"] = alliances
	all_json["score_breakdown"] = breakdown

	return all_json, nil
}

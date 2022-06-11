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
	auto int64
	teleop int64
	fouls int64
	total int64

	baseRP int64  // win-loss-tie RP only
	rocketRP bool
	habRP bool

	hatchPanels int64
	scoredHatchPanels int64

	fields map[string]int64  // indexed by values of simpleFields2022
}

func makeFmsScoreInfo2022() fmsScoreInfo2022 {
	return fmsScoreInfo2022{
		fields: make(map[string]int64),
	}
}

type extraMatchAllianceInfo2022 struct {
	Dqs []string `json:"dqs"`
	Surrogates []string `json:"surrogates"`
	AddRpRocket bool `json:"add_rp_rocket"`
	AddRpHabClimb bool `json:"add_rp_hab_climb"`
}

func makeExtraMatchAllianceInfo2022() extraMatchAllianceInfo2022 {
	return extraMatchAllianceInfo2022{
		Dqs: make([]string, 0),
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
var simpleFields2022 = map[string]string {
	"endgame points": "endgamePoints",
	"taxi points": "taxiPoints",
	// "ranking points": "rp",
	"adjustments": "adjustPoints",
}

func parseHTMLtoJSON2022(filename string, playoff bool) (map[string]interface{}, error) {
	//////////////////////////////////////////////////
	// Parse html from FMS into TBA-compatible JSON //
	//////////////////////////////////////////////////

	// Open file
	r, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("Error opening file: %s", filename)
	}
	defer r.Close()

	// Read from file
	dom, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("Error reading from file: %s", filename)
	}

	all_json := make(map[string]interface{})

	extra_info := make(map[string]extraMatchAllianceInfo2022)
	extra_info["blue"] = makeExtraMatchAllianceInfo2022()
	extra_info["red"] = makeExtraMatchAllianceInfo2022()
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
		blue fmsScoreInfo2022
		red fmsScoreInfo2022
	}{
		makeFmsScoreInfo2022(),
		makeFmsScoreInfo2022(),
	}

	parse_errors := make([]string, 0)

	checkParseInt := func(s, desc string) int64 {
		n, err := strconv.ParseInt(s, 10, 0)
		if err != nil {
			panic(fmt.Sprintf("parse int %s failed: %s", desc, err))
		}
		return n
	}

	dom.Find("tr").Each(func(i int, s *goquery.Selection){
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Parse error in %s: %s\n", filename, r)
				parse_errors = append(parse_errors, fmt.Sprint(r))
			}
		}()

		columns := s.Children()
		if columns.Length() == 3 {
			row_name := strings.ToLower(strings.TrimSpace(columns.Eq(0).Text()))
			if (row_name == "") {
				return // continue
			}

			blue_cell := columns.Eq(1)
			red_cell  := columns.Eq(2)
			blue_text := strings.ToLower(strings.TrimSpace(blue_cell.Text()))
			red_text  := strings.ToLower(strings.TrimSpace(red_cell.Text()))

			// Handle each data row
			if row_name == "final score" {
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
				breakdown["blue"]["rankingPoints"] = blue_rp
				breakdown["red"]["rankingPoints"] = red_rp

			} else {
				breakdown["blue"]["!" + row_name] = blue_text
				breakdown["red"]["!" + row_name] = red_text
			}
		}
	})

	addManualFields2022(breakdown["blue"], scoreInfo.blue, extra_info["blue"], playoff)
	addManualFields2022(breakdown["red"], scoreInfo.red, extra_info["red"], playoff)

	if len(parse_errors) > 0 {
		return nil, fmt.Errorf("Parse error (%d):\n%s", len(parse_errors), strings.Join(parse_errors, "\n"))
	}

	all_json["alliances"] = alliances
	all_json["score_breakdown"] = breakdown

	return all_json, nil
}

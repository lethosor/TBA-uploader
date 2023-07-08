package fms_parser

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/lethosor/TBA-uploader/tba"
)

func ParseHTMLtoJSON(year int, filename string, playoff bool) (map[string]interface{}, error) {
	if year == 2018 {
		return parseHTMLtoJSON2018(filename, playoff)
	} else if year == 2019 {
		return parseHTMLtoJSON2019(filename, playoff)
	} else if year == 2022 {
		return parseHTMLtoJSON2022(filename, playoff)
	} else if year == 2023 {
		return parseHTMLtoJSON2023(filename, playoff)
	} else {
		return nil, fmt.Errorf("ParseHTMLtoJSON: unsupported year: %d", year)
	}
}

type ExtraMatchInfo struct {
	MatchCodeOverride *tba.MatchCode         `json:"match_code_override"`
	Red               ExtraMatchAllianceInfo `json:"red"`
	Blue              ExtraMatchAllianceInfo `json:"blue"`
}

type ExtraMatchAllianceInfo interface{}

type extraMatchAllianceInfoCommon struct {
	Dqs        []string `json:"dqs"`
	Surrogates []string `json:"surrogates"`
	ExtraRps   []bool   `json:"extra_rps"`
}

func makeExtraMatchAllianceInfoCommon() extraMatchAllianceInfoCommon {
	return extraMatchAllianceInfoCommon{
		Dqs:        make([]string, 0),
		Surrogates: make([]string, 0),
		ExtraRps:   make([]bool, 0),
	}
}

var extraAllianceInfoCtors = map[int]func() ExtraMatchAllianceInfo{
	2018: func() ExtraMatchAllianceInfo {
		return makeExtraMatchAllianceInfo2018()
	},
	2019: func() ExtraMatchAllianceInfo {
		return makeExtraMatchAllianceInfo2019()
	},
	2022: func() ExtraMatchAllianceInfo {
		return makeExtraMatchAllianceInfo2022()
	},
	2023: func() ExtraMatchAllianceInfo {
		return makeExtraMatchAllianceInfo2023()
	},
}

func MakeExtraMatchInfo(year int) (ExtraMatchInfo, error) {
	if ctor, ok := extraAllianceInfoCtors[year]; ok {
		return ExtraMatchInfo{
			MatchCodeOverride: nil,
			Red:               ctor(),
			Blue:              ctor(),
		}, nil
	} else {
		return ExtraMatchInfo{}, fmt.Errorf("unsupported year: %d", year)
	}
}

func GetDefaultBreakdowns(year int) map[string]any {
	if year == 2022 {
		return DEFAULT_BREAKDOWN_VALUES_2022
	} else {
		return nil
	}
}

func split_and_strip(text string, separator string) []string {
	// Split text into parts at separator character. Remove whitespace from parts.
	// "a • b • c" -> ["a", "b", "c"]
	parts := strings.Split(strings.TrimSpace(text), separator)
	var result []string
	for _, part := range parts {
		result = append(result, strings.TrimSpace(part))
	}
	return result
}

func iconToBool(node *goquery.Selection, true_class, false_class string) bool {
	if node.HasClass(true_class) {
		return true
	} else if node.HasClass(false_class) {
		return false
	} else {
		class, _ := node.Attr("class")
		panic(fmt.Sprintf("icon has unexpected classes: \"%s\"", class))
	}
}

func iconsToBools(node *goquery.Selection, count int, true_class, false_class string) (out []bool) {
	icons := node.Find("i")
	if icons.Length() != count {
		panic(fmt.Sprintf("could not find expected %d icons", count))
	}
	out = make([]bool, count)

	icons.Each(func(i int, s *goquery.Selection) {
		out[i] = iconToBool(s, true_class, false_class)
	})

	return out
}

func identity_fn[T any](in T) T {
	return in
}

func boolToYesNo(in bool) string {
	if in {
		return "Yes"
	}
	return "No"
}

type breakdownAllianceFields[T any] struct {
	blue T
	red  T
}

func assignBreakdownAllianceFields[T, T2 any](breakdowns map[string]map[string]interface{}, field string, callback func(T) T2, values breakdownAllianceFields[T]) {
	breakdowns["blue"][field] = callback(values.blue)
	breakdowns["red"][field] = callback(values.red)
}

func assignBreakdownAllianceFieldsConst[T any](breakdowns map[string]map[string]interface{}, field string, value T) {
	breakdowns["blue"][field] = value
	breakdowns["red"][field] = value
}

type breakdownRobotFields[T any] struct {
	blue []T
	red  []T
}

func assignBreakdownRobotFields[T, T2 any](breakdowns map[string]map[string]interface{}, prefix string, callback func(T) T2, values breakdownRobotFields[T]) {
	for i := 0; i < 3; i++ {
		breakdowns["blue"][fmt.Sprintf("%s%d", prefix, i+1)] = callback(values.blue[i])
		breakdowns["red"][fmt.Sprintf("%s%d", prefix, i+1)] = callback(values.red[i])
	}
}

type breakdownAllianceMultipleFields[T any] struct {
	blue []T
	red  []T
}

func assignBreakdownAllianceMultipleFields[T, T2 any](breakdowns map[string]map[string]interface{}, fields []string, callback func(value T, alliance string) T2, values breakdownAllianceMultipleFields[T]) {
	if len(values.blue) != len(fields) {
		panic(fmt.Sprintf("blue length mismatch: expected %d, got %d", len(fields), len(values.blue)))
	}
	if len(values.red) != len(fields) {
		panic(fmt.Sprintf("red length mismatch: expected %d, got %d", len(fields), len(values.red)))
	}
	for i, field := range fields {
		breakdowns["blue"][field] = callback(values.blue[i], "blue")
		breakdowns["red"][field] = callback(values.red[i], "red")
	}
}

func assignTbaTeams(alliances map[string]map[string]interface{}, teams breakdownRobotFields[string]) {
	groups := map[string][]string{
		"blue": teams.blue,
		"red":  teams.red,
	}
	for alliance, alliance_teams := range groups {
		tba_teams := make([]string, len(alliance_teams))
		for i := 0; i < 3; i++ {
			tba_teams[i] = "frc" + alliance_teams[i]
		}
		alliances[alliance]["teams"] = tba_teams
	}
}

func assignBreakdownRpFromBadges(breakdowns map[string]map[string]interface{}, rp_badge_names map[string]string, cells breakdownAllianceFields[*goquery.Selection]) {
	groups := map[string]*goquery.Selection{
		"blue": cells.blue,
		"red":  cells.red,
	}
	for alliance, cell := range groups {
		for _, field := range rp_badge_names {
			breakdowns[alliance][field] = false
		}
		cell.Find("img").Each(func(_ int, img *goquery.Selection) {
			title, _ := img.Attr("title")
			if field, ok := rp_badge_names[title]; ok {
				breakdowns[alliance][field] = true
			}
		})
	}
}

func assignPenaltyFields(breakdowns map[string]map[string]interface{}, penalty_fields map[string]string, cells breakdownAllianceFields[*goquery.Selection]) {
	groups := map[string]*goquery.Selection{
		"blue": cells.blue,
		"red":  cells.red,
	}
	for alliance, cell := range groups {
		for penalty_name, field := range penalty_fields {
			breakdowns[alliance][field] = strings.Contains(cell.Text(), penalty_name)
		}
	}
}

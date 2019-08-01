package fms_parser

import (
	"fmt"
	"strings"

	"../tba"
)

func split_and_strip(text string, separator string) ([]string) {
	// Split text into parts at separator character. Remove whitespace from parts.
	// "a • b • c" -> ["a", "b", "c"]
	parts := strings.Split(strings.TrimSpace(text), separator)
	var result []string
	for _, part := range parts {
		result = append(result, strings.TrimSpace(part))
	}
	return result
}

func ParseHTMLtoJSON(year int, filename string, playoff bool) (map[string]interface{}, error) {
	if (year == 2018) {
		return parseHTMLtoJSON2018(filename, playoff)
	} else if (year == 2019) {
		return parseHTMLtoJSON2019(filename, playoff)
	} else {
		return nil, fmt.Errorf("ParseHTMLtoJSON: unsupported year: %d", year)
	}
}

type ExtraMatchInfo struct {
	MatchCodeOverride *tba.MatchCode `json:"match_code_override"`
	Red ExtraMatchAllianceInfo `json:"red"`
	Blue ExtraMatchAllianceInfo `json:"blue"`
}

type ExtraMatchAllianceInfo interface {}

var extraAllianceInfoCtors = map[int]func() ExtraMatchAllianceInfo {
	2018: func() ExtraMatchAllianceInfo {
		return makeExtraMatchAllianceInfo2018()
	},
	2019: func() ExtraMatchAllianceInfo {
		return makeExtraMatchAllianceInfo2019()
	},
}

func MakeExtraMatchInfo(year int) (ExtraMatchInfo, error) {
	if ctor, ok := extraAllianceInfoCtors[year]; ok {
		return ExtraMatchInfo{
			MatchCodeOverride: nil,
			Red: ctor(),
			Blue: ctor(),
		}, nil
	} else {
		return ExtraMatchInfo{}, fmt.Errorf("unsupported year: %d", year)
	}
}

package fms_parser

import (
	"fmt"
	"strings"
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

func MakeExtraMatchInfo(year int) interface{} {
	if (year == 2018) {
		return makeExtraMatchInfo2018()
	} else if (year == 2019) {
		return makeExtraMatchInfo2019()
	} else {
		return nil
	}
}

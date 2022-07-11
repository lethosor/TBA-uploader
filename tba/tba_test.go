package tba

import (
	"testing"
)

func testBracket(t *testing.T, bracket Bracket, bracket_type int, name string) {
	for i, expected_code := range bracket {
		code := GetPlayoffCode(bracket_type, i)
		if code.Level != expected_code.Level {
			t.Errorf("playoff %d of %s: got level=%s, expected %s", i, name, code.Level, expected_code.Level)
		}
		if code.Set != expected_code.Set {
			t.Errorf("playoff %d of %s: got set=%d, expected %d", i, name, code.Set, expected_code.Set)
		}
		if code.Match != expected_code.Match {
			t.Errorf("playoff %d of %s: got match=%d, expected %d", i, name, code.Match, expected_code.Match)
		}
	}
}

var playoff_codes_8_bracket = Bracket{
	1:  {Level: "qf", Set: 1, Match: 1},
	2:  {Level: "qf", Set: 2, Match: 1},
	3:  {Level: "qf", Set: 3, Match: 1},
	4:  {Level: "qf", Set: 4, Match: 1},
	5:  {Level: "qf", Set: 1, Match: 2},
	6:  {Level: "qf", Set: 2, Match: 2},
	7:  {Level: "qf", Set: 3, Match: 2},
	8:  {Level: "qf", Set: 4, Match: 2},
	9:  {Level: "qf", Set: 1, Match: 3},
	10: {Level: "qf", Set: 2, Match: 3},
	11: {Level: "qf", Set: 3, Match: 3},
	12: {Level: "qf", Set: 4, Match: 3},

	13: {Level: "sf", Set: 1, Match: 1},
	14: {Level: "sf", Set: 2, Match: 1},
	15: {Level: "sf", Set: 1, Match: 2},
	16: {Level: "sf", Set: 2, Match: 2},
	17: {Level: "sf", Set: 1, Match: 3},
	18: {Level: "sf", Set: 2, Match: 3},

	19: {Level: "f", Set: 1, Match: 1},
	20: {Level: "f", Set: 1, Match: 2},
	21: {Level: "f", Set: 1, Match: 3},
	22: {Level: "f", Set: 1, Match: 4},
	23: {Level: "f", Set: 1, Match: 5},
	24: {Level: "f", Set: 1, Match: 6},
}

func TestPlayoffCodes8Bracket(t *testing.T) {
	testBracket(t, playoff_codes_8_bracket, BRACKET_TYPE_BRACKET_8_TEAM, BRACKET_NAME_BRACKET_8_TEAM)
}

var playoff_codes_6_round_robin = Bracket{
	1:  {Level: "sf", Set: 1, Match: 1},
	2:  {Level: "sf", Set: 1, Match: 2},
	3:  {Level: "sf", Set: 1, Match: 3},
	4:  {Level: "sf", Set: 1, Match: 4},
	5:  {Level: "sf", Set: 1, Match: 5},
	6:  {Level: "sf", Set: 1, Match: 6},
	7:  {Level: "sf", Set: 1, Match: 7},
	8:  {Level: "sf", Set: 1, Match: 8},
	9:  {Level: "sf", Set: 1, Match: 9},
	10: {Level: "sf", Set: 1, Match: 10},
	11: {Level: "sf", Set: 1, Match: 11},
	12: {Level: "sf", Set: 1, Match: 12},
	13: {Level: "sf", Set: 1, Match: 13},
	14: {Level: "sf", Set: 1, Match: 14},
	15: {Level: "sf", Set: 1, Match: 15},

	16: {Level: "f", Set: 1, Match: 1},
	17: {Level: "f", Set: 1, Match: 2},
	18: {Level: "f", Set: 1, Match: 3},
	19: {Level: "f", Set: 1, Match: 4},
	20: {Level: "f", Set: 1, Match: 5},
	21: {Level: "f", Set: 1, Match: 6},
}

func TestPlayoffCodes6RoundRobin(t *testing.T) {
	testBracket(t, playoff_codes_6_round_robin, BRACKET_TYPE_ROUND_ROBIN_6_TEAM, BRACKET_NAME_ROUND_ROBIN_6_TEAM)
}

var playoff_codes_custom = Bracket{
	// for now, these should return default-initialized structs
	1: {},
	2: {},
	3: {},
	4: {},
}

func TestPlayoffCodesCustom(t *testing.T) {
	testBracket(t, playoff_codes_custom, BRACKET_TYPE_CUSTOM, BRACKET_NAME_CUSTOM)
}

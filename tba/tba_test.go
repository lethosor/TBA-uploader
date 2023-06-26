package tba

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testBracket(t *testing.T, bracket Bracket, bracket_type int, name string) {
	for i, expected_code := range bracket {
		code := GetPlayoffCode(bracket_type, i)
		assert.Equalf(t, expected_code, code, "playoff %d of %s", i, name)
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
	25: {},
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
	22: {},
}

func TestPlayoffCodes6RoundRobin(t *testing.T) {
	testBracket(t, playoff_codes_6_round_robin, BRACKET_TYPE_ROUND_ROBIN_6_TEAM, BRACKET_NAME_ROUND_ROBIN_6_TEAM)
}

var playoff_codes_8_double_elim = Bracket{
	1:  {Level: "sf", Set: 1, Match: 1},
	2:  {Level: "sf", Set: 2, Match: 1},
	3:  {Level: "sf", Set: 3, Match: 1},
	4:  {Level: "sf", Set: 4, Match: 1},
	5:  {Level: "sf", Set: 5, Match: 1},
	6:  {Level: "sf", Set: 6, Match: 1},
	7:  {Level: "sf", Set: 7, Match: 1},
	8:  {Level: "sf", Set: 8, Match: 1},
	9:  {Level: "sf", Set: 9, Match: 1},
	10: {Level: "sf", Set: 10, Match: 1},
	11: {Level: "sf", Set: 11, Match: 1},
	12: {Level: "sf", Set: 12, Match: 1},
	13: {Level: "sf", Set: 13, Match: 1},

	14: {Level: "f", Set: 1, Match: 1},
	15: {Level: "f", Set: 1, Match: 2},
	16: {Level: "f", Set: 1, Match: 3},
	17: {Level: "f", Set: 1, Match: 4},
	18: {Level: "f", Set: 1, Match: 5},
	19: {Level: "f", Set: 1, Match: 6},
	20: {},
}

func TestPlayoffCodes8DoubleElim(t *testing.T) {
	testBracket(t, playoff_codes_8_double_elim, BRACKET_TYPE_DOUBLE_ELIM_8_TEAM, BRACKET_NAME_DOUBLE_ELIM_8_TEAM)
}

var playoff_codes_4_double_elim = Bracket{
	1: {Level: "sf", Set: 1, Match: 1},
	2: {Level: "sf", Set: 2, Match: 1},
	3: {Level: "sf", Set: 3, Match: 1},
	4: {Level: "sf", Set: 4, Match: 1},
	5: {Level: "sf", Set: 5, Match: 1},

	6:  {Level: "f", Set: 1, Match: 1},
	7:  {Level: "f", Set: 1, Match: 2},
	8:  {Level: "f", Set: 1, Match: 3},
	9:  {Level: "f", Set: 1, Match: 4},
	10: {Level: "f", Set: 1, Match: 5},
	11: {Level: "f", Set: 1, Match: 6},
	12: {},
}

func TestPlayoffCodes4DoubleElim(t *testing.T) {
	testBracket(t, playoff_codes_4_double_elim, BRACKET_TYPE_DOUBLE_ELIM_4_TEAM, BRACKET_NAME_DOUBLE_ELIM_4_TEAM)
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

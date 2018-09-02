package main

import (
    "testing"
)

var playoff_codes = map[int]tbaPlayoffCode {
    1:  tbaPlayoffCode{level: "qf", set: 1, match: 1},
    2:  tbaPlayoffCode{level: "qf", set: 2, match: 1},
    3:  tbaPlayoffCode{level: "qf", set: 3, match: 1},
    4:  tbaPlayoffCode{level: "qf", set: 4, match: 1},
    5:  tbaPlayoffCode{level: "qf", set: 1, match: 2},
    6:  tbaPlayoffCode{level: "qf", set: 2, match: 2},
    7:  tbaPlayoffCode{level: "qf", set: 3, match: 2},
    8:  tbaPlayoffCode{level: "qf", set: 4, match: 2},
    9:  tbaPlayoffCode{level: "qf", set: 1, match: 3},
    10: tbaPlayoffCode{level: "qf", set: 2, match: 3},
    11: tbaPlayoffCode{level: "qf", set: 3, match: 3},
    12: tbaPlayoffCode{level: "qf", set: 4, match: 3},

    13: tbaPlayoffCode{level: "sf", set: 1, match: 1},
    14: tbaPlayoffCode{level: "sf", set: 2, match: 1},
    15: tbaPlayoffCode{level: "sf", set: 1, match: 2},
    16: tbaPlayoffCode{level: "sf", set: 2, match: 2},
    17: tbaPlayoffCode{level: "sf", set: 1, match: 3},
    18: tbaPlayoffCode{level: "sf", set: 2, match: 3},

    19: tbaPlayoffCode{level: "f",  set: 1, match: 1},
    20: tbaPlayoffCode{level: "f",  set: 1, match: 2},
    21: tbaPlayoffCode{level: "f",  set: 1, match: 3},
}

func TestPlayoffCodes(t *testing.T) {
    for i := 1; i <= 21; i++ {
        code := getTBAPlayoffCode(i)
        if code.level != playoff_codes[i].level {
            t.Errorf("playoff %d: got level=%s, expected %s", i, code.level, playoff_codes[i].level)
        }
        if code.set != playoff_codes[i].set {
            t.Errorf("playoff %d: got set=%d, expected %d", i, code.set, playoff_codes[i].set)
        }
        if code.match != playoff_codes[i].match {
            t.Errorf("playoff %d: got match=%d, expected %d", i, code.match, playoff_codes[i].match)
        }
    }
}

package tba

import (
    "testing"
)

var playoff_codes = map[int]MatchCode {
    1:  MatchCode{Level: "qf", Set: 1, Match: 1},
    2:  MatchCode{Level: "qf", Set: 2, Match: 1},
    3:  MatchCode{Level: "qf", Set: 3, Match: 1},
    4:  MatchCode{Level: "qf", Set: 4, Match: 1},
    5:  MatchCode{Level: "qf", Set: 1, Match: 2},
    6:  MatchCode{Level: "qf", Set: 2, Match: 2},
    7:  MatchCode{Level: "qf", Set: 3, Match: 2},
    8:  MatchCode{Level: "qf", Set: 4, Match: 2},
    9:  MatchCode{Level: "qf", Set: 1, Match: 3},
    10: MatchCode{Level: "qf", Set: 2, Match: 3},
    11: MatchCode{Level: "qf", Set: 3, Match: 3},
    12: MatchCode{Level: "qf", Set: 4, Match: 3},

    13: MatchCode{Level: "sf", Set: 1, Match: 1},
    14: MatchCode{Level: "sf", Set: 2, Match: 1},
    15: MatchCode{Level: "sf", Set: 1, Match: 2},
    16: MatchCode{Level: "sf", Set: 2, Match: 2},
    17: MatchCode{Level: "sf", Set: 1, Match: 3},
    18: MatchCode{Level: "sf", Set: 2, Match: 3},

    19: MatchCode{Level: "f",  Set: 1, Match: 1},
    20: MatchCode{Level: "f",  Set: 1, Match: 2},
    21: MatchCode{Level: "f",  Set: 1, Match: 3},
    22: MatchCode{Level: "f",  Set: 1, Match: 4},
}

func TestPlayoffCodes(t *testing.T) {
    for i := 1; i <= 21; i++ {
        code := GetPlayoffCode(i)
        if code.Level != playoff_codes[i].Level {
            t.Errorf("playoff %d: got level=%s, expected %s", i, code.Level, playoff_codes[i].Level)
        }
        if code.Set != playoff_codes[i].Set {
            t.Errorf("playoff %d: got set=%d, expected %d", i, code.Set, playoff_codes[i].Set)
        }
        if code.Match != playoff_codes[i].Match {
            t.Errorf("playoff %d: got match=%d, expected %d", i, code.Match, playoff_codes[i].Match)
        }
    }
}

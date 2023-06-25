function toNumber(value) {
    if (typeof value == 'string' && !isNaN(Number(value))) {
        return Number(value);
    }
    return value;
}

/*
interface for ranking reducers:
{
    add: function(matchResult, alliance: 'red'|'blue', teamKey: 'frc####'),
    get: function() => ranking report cell value
}
*/

function RankingReducerAverage(breakdownFields, defaultValue=-1) {
    if (!Array.isArray(breakdownFields)) {
        breakdownFields = [breakdownFields];
    }
    return function() {
        const matchValues = [];
        return {
            add(match, alliance) {
                let matchValue = 0;
                for (let field of breakdownFields) {
                    let mult = 1;
                    if (field.startsWith('-')) {
                        mult = -1;
                        field = field.replace(/^-/, '');
                    }
                    const breakdownValue = match.score_breakdown[alliance][field];
                    if (breakdownValue !== undefined) {
                        matchValue += mult * breakdownValue;
                    }
                    else {
                        matchValue = defaultValue;
                        break;
                    }
                }
                matchValues.push(matchValue);
            },
            get() {
                return matchValues.reduce((a, b) => (a + b), 0) / matchValues.length;
            },
        };
    };
}

const rankingBreakdownSources = {
    2022: {
        'Ranking Score': RankingReducerAverage('rp'),
        'Avg Match': RankingReducerAverage(['totalPoints', '-foulPoints']),
        'Avg Hangar': RankingReducerAverage('endgamePoints'),
        'Avg Taxi + Auto Cargo': RankingReducerAverage(['autoTaxiPoints', 'autoCargoPoints']),
    },
    2023: {
        'Ranking Score': RankingReducerAverage('rp'),
        'Avg Match': RankingReducerAverage(['totalPoints', '-foulPoints']),
        'Avg Charge Station': RankingReducerAverage('totalChargeStationPoints'),
        'Avg Auto': RankingReducerAverage('autoPoints'),
    },
};

const tba = Object.freeze({
    isValidEventCode(event) {
        return event && Boolean(event.match(/^\d+/));
    },

    isValidYear(year) {
        year = parseInt(year);
        return !isNaN(year) && tba.RANKING_NAMES[year] != undefined;
    },

    convertToTBARankings: Object.freeze({
        common(r) {
            return {
                team_key: 'frc' + r.team,
                rank: toNumber(r.rank),
                played: toNumber(r.played),
                dqs: toNumber(r.dq),
                wins: toNumber(r.wins),
                losses: toNumber(r.losses),
                ties: toNumber(r.ties),
            };
        },
        // keys should match https://github.com/the-blue-alliance/the-blue-alliance/blob/py3/src/backend/common/consts/ranking_sort_orders.py
        2018: function(r) {
            return Object.assign(tba.convertToTBARankings.common(r), {
                "Ranking Score": r.sort1,
                "End Game": r.sort2,
                "Auto": r.sort3,
                "Ownership": r.sort4,
                "Vault": r.sort5,
            });
        },
        2019: function(r) {
            return Object.assign(tba.convertToTBARankings.common(r), {
                "Ranking Score": r.sort1,
                "Cargo": r.sort2,
                "Hatch Panel": r.sort3,
                "HAB Climb": r.sort4,
                "Sandstorm Bonus": r.sort5,
            });
        },
        2022: function(r) {
            return Object.assign(tba.convertToTBARankings.common(r), {
                "Ranking Score": r.sort1,
                "Avg Match": r.sort2,
                "Avg Hangar": r.sort3,
                "Avg Taxi + Auto Cargo": r.sort4,
            });
        },
        2023: function(r) {
            return Object.assign(tba.convertToTBARankings.common(r), {
                "Ranking Score": r.sort1,
                "Avg Match": r.sort2,
                "Avg Charge Station": r.sort3,
                "Avg Auto": r.sort4,
            });
        },
    }),

    RANKING_NAMES: Object.freeze({
        2018: [
            "Ranking Score",
            "End Game",
            "Auto",
            "Ownership",
            "Vault",
        ],
        2019: [
            "Ranking Score",
            "Cargo",
            "Hatch Panel",
            "HAB Climb",
            "Sandstorm Bonus",
        ],
        2022: [
            "Ranking Score",
            "Avg Match",
            "Avg Hangar",
            "Avg Taxi + Auto Cargo",
        ],
        2023: [
            "Ranking Score",
            "Avg Match",
            "Avg Charge Station",
            "Avg Auto",
        ],
    }),

    generateRankingsFromMatchResults: function(matchResults, year) {
        const getMatchTeams = function(match) {
            const teams = [];
            for (const alliance of ['red', 'blue']) {
                for (const team of match.alliances[alliance].team_keys) {
                    teams.push({
                        alliance,
                        team_key: team,
                        dq: match.alliances[alliance].dq_team_keys.includes(team),
                        surrogate: match.alliances[alliance].surrogate_team_keys.includes(team),
                    });
                }
            }
            return teams;
        };

        const makeRankings = function(teamKey) {
            return {
                team_key: teamKey,
                rank: undefined,
                played: 0,
                dqs: 0,
                wins: 0,
                losses: 0,
                ties: 0,
            };
        };

        const makeRankingReducers = function() {
            const out = {};
            for (const [key, func] of Object.entries(rankingBreakdownSources[year])) {
                out[key] = func();
            }
            return out;
        };

        const roundRankingValue = function(val) {
            // round down to 2 decimal places
            return Math.floor(val * 100) / 100;
        };

        const teamRankingReducers = {};
        const rankings = {};
        for (const match of matchResults) {
            if (match.comp_level != 'qm') {
                continue;
            }
            if (match.alliances.red.score == -1 || match.alliances.blue.score == -1) {
                continue;
            }

            const teams = getMatchTeams(match);
            for (const teamEntry of teams) {
                const teamKey = teamEntry.team_key;
                if (!rankings[teamKey]) {
                    rankings[teamKey] = makeRankings(teamKey);
                }
                if (!teamRankingReducers[teamKey]) {
                    teamRankingReducers[teamKey] = makeRankingReducers();
                }

                if (teamEntry.dq) {
                    rankings[teamKey].dqs++;
                    continue;
                }
                if (teamEntry.surrogate) {
                    continue;
                }

                rankings[teamKey].played++;

                const scoreDiff = match.alliances[teamEntry.alliance].score -
                    match.alliances[teamEntry.alliance == 'red' ? 'blue' : 'red'].score;
                if (scoreDiff > 0) {
                    rankings[teamKey].wins++;
                }
                else if (scoreDiff < 0) {
                    rankings[teamKey].losses++;
                }
                else {
                    rankings[teamKey].ties++;
                }

                for (const reducer of Object.values(teamRankingReducers[teamKey])) {
                    reducer.add(match, teamEntry.alliance, teamKey);
                }
            }
        }

        for (const [teamKey, reducers] of Object.entries(teamRankingReducers)) {
            for (const [name, r] of Object.entries(reducers)) {
                rankings[teamKey][name] = roundRankingValue(r.get());
            }
        }

        // sort references, then modify in place to add rank field
        const sortedRankings = Object.values(rankings).sort(function(a, b) {
            for (const field of tba.RANKING_NAMES[year]) {
                if (a[field] != b[field]) {
                    return b[field] - a[field];
                }
            }
            return 0;
        });
        for (let i = 0; i < sortedRankings.length; i++) {
            sortedRankings[i].rank = i + 1;
        }

        return sortedRankings;
    },
});

module.exports = tba;

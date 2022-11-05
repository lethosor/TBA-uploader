function toNumber(value) {
    if (typeof value == 'string' && !isNaN(Number(value))) {
        return Number(value);
    }
    return value;
}

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
        2018(r) {
            return Object.assign(tba.convertToTBARankings.common(r), {
                "Ranking Score": r.sort1,
                "End Game": r.sort2,
                "Auto": r.sort3,
                "Ownership": r.sort4,
                "Vault": r.sort5,
            });
        },
        2019(r) {
            return Object.assign(tba.convertToTBARankings.common(r), {
                "Ranking Score": r.sort1,
                "Cargo": r.sort2,
                "Hatch Panel": r.sort3,
                "HAB Climb": r.sort4,
                "Sandstorm Bonus": r.sort5,
            });
        },
        2022(r) {
            return Object.assign(tba.convertToTBARankings.common(r), {
                "Ranking Score": r.sort1,
                "Avg Match": r.sort2,
                "Avg Hangar": r.sort3,
                "Avg Taxi + Auto Cargo": r.sort4,
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
    }),
});

module.exports = tba;

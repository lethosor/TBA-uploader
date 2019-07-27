const tba = Object.freeze({
    isValidEventCode(event) {
        return event && Boolean(event.match(/^\d+/));
    },

    isValidYear(year) {
        year = parseInt(year);
        return year >= 2018 && year <= 2019;
    },

    convertToTBARankings: Object.freeze({
        common(r) {
            return {
                team_key: 'frc' + r.team,
                rank: r.rank,
                played: r.played,
                dqs: r.dq,
                "Record (W-L-T)": r.wins + '-' + r.losses + '-' + r.ties,
            };
        },
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
    }),

    RANKING_NAMES: Object.freeze({
        2018: [
            "Ranking Score",
            "End Game",
            "Auto",
            "Ownership",
            "Vault",
            "Record (W-L-T)",
        ],
        2019: [
            "Ranking Score",
            "Cargo",
            "Hatch Panel",
            "HAB Climb",
            "Sandstorm Bonus",
            "Record (W-L-T)",
        ],
    }),
});

export default tba;

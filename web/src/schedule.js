const Schedule = {};

/*
Example custom schedule CSV (TODO: document)

TBA Match Schedule,
level,set,match,Time,Description,Blue 1,Blue 2,Blue 3,Red 1,Red 2,Red 3,
sf,1,1 ,,Round-robin 1,5498,815,3175,9993,33,313,
sf,1,2 ,,Round-robin 2,862,3604,7174,5069,5530,280,
sf,1,3 ,,Round-robin 3,5907,5531,9996,6528,5050,7191,
sf,1,4 ,,Round-robin 4,3655,5090,6914,6618,9994,2620,
sf,1,5 ,,Round-robin 5,5498,815,3175,9993,33,313,
sf,1,6 ,,Round-robin 6,862,3604,7174,5069,5530,280,
sf,1,7 ,,Round-robin 7,5907,5531,9996,6528,5050,7191,
sf,1,8 ,,Round-robin 8,3655,5090,6914,6618,9994,2620,
sf,1,9 ,,Round-robin 9,5498,815,3175,9993,33,313,
sf,1,10,,Round-robin 10,862,3604,7174,5069,5530,280,
sf,1,11,,Round-robin 11,5907,5531,9996,6528,5050,7191,
sf,1,12,,Round-robin 12,3655,5090,6914,6618,9994,2620,
*/

Schedule.getTBAPlayoffCode = function(bracket_type, match_id) {
    let bracket = BRACKETS[bracket_type];
    if (!bracket) {
        throw 'Unsupported bracket type: ' + bracket_type;
    }
    return bracket[match_id];
};

Schedule.getTBAMatchKey = function(match) {
    if (match.comp_level == 'qm') {
        return 'qm' + match.match_number;
    }
    else {
        return match.comp_level + match.set_number + 'm' + match.match_number;
    }
};

Schedule.parse = function(rawCsv, playoffType) {
    var lines = rawCsv.split('\n').map(function(line) {
        return line.trim().toLowerCase().replace(/\s*/g, '').split(',');
    });
    if (lines[0][0].indexOf('matchschedule') < 0) {
        throw 'Wrong report type. You uploaded: ' + rawCsv.split(',')[0];
    }
    // find header
    var columnIndices = {};
    for (var i = 0; i < lines.length; i++) {
        if (lines[i].indexOf('red1') >= 0) {
            for (var j = 0; j < lines[i].length; j++) {
                if (lines[i][j]) {
                    columnIndices[lines[i][j]] = j;
                }
            }
            break;
        }
    }
    const requiredColumns = ['red1', 'red2', 'red3', 'blue1', 'blue2', 'blue3', 'time', 'description'];
    const hasTbaCodes = ['level', 'set', 'match'].every(k => columnIndices[k] !== undefined);
    var missingColumns = requiredColumns.filter(function(col) {
        return columnIndices[col] === undefined;
    });
    if (missingColumns.length) {
        throw 'Missing columns: ' + missingColumns.join(', ');
    }

    var genTeams = function(match, color) {
        return [1, 2, 3].map(function(i) {
            return 'frc' + match[color + i].match(/\d+/);
        });
    };
    var genSurrogates = function(match, color) {
        return [1, 2, 3].filter(function(i) {
            return match[color + i].indexOf('*') >= 0;
        }).map(function(i) {
            return 'frc' + match[color + i].match(/\d+/);
        });
    };
    var genAlliance = function(match, color) {
        return {
            dqs: [],
            score: -1,
            surrogates: genSurrogates(match, color),
            teams: genTeams(match, color),
        };
    };

    return lines.filter(function(line) {
        return ['red1', 'red2', 'red3', 'blue1', 'blue2', 'blue3'].every(function(k) {
            return line[columnIndices[k]] && line[columnIndices[k]] != k && line[columnIndices[k]].match(/\d+/);
        });
    }).map(function(line) {
        var match = {};
        Object.keys(columnIndices).forEach(function(k) {
            match[k] = line[columnIndices[k]];
        });
        return match;
    }).map(function(match) {
        let raw_id;
        let code;
        let tryParseMatchId = (pattern, index) => {
            let matches = match.description.match(pattern);
            if (!matches) {
                throw 'Failed to parse match ID from: ' + match.description;
            }
            return Number(matches[index]);
        };
        if (hasTbaCodes) {
            code = {
                comp_level: match.level,
                set_number: Number(match.set),
                match_number: Number(match.match),
            };
        }
        else if (match.description.startsWith('qual')) {
            raw_id = tryParseMatchId(/\d+/, 0);
            code = {
                comp_level: 'qm',
                set_number: 1,
                match_number: raw_id,
            };
        }
        else {
            raw_id = tryParseMatchId(/#(\d+)/, 1);
            code = Schedule.getTBAPlayoffCode(playoffType, raw_id);
            if (!code) {
                throw 'Playoff match ID out of range: ' + raw_id;
            }
        }

        return Object.assign(code, {
            _key: Schedule.getTBAMatchKey(code),
            _id: raw_id,
            alliances: {
                red: genAlliance(match, 'red'),
                blue: genAlliance(match, 'blue'),
            },
            time_string: match.time.replace(/(\d+:\d+)/, function(m) {
                return ' ' + m + ' ';
            }),
        });
    });
};

Schedule.findAllCompLevels = function(matches) {
    var levels = [];
    matches.forEach(function(m) {
        if (levels.indexOf(m.comp_level) < 0) {
            levels.push(m.comp_level);
        }
    });
    return levels;
};

export default Object.freeze(Schedule);

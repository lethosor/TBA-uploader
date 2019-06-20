Schedule = {};

Schedule.getTBAPlayoffCode = function(match_id) {
    if (match_id <= 12) {
        return {
            comp_level: "qf",
            set_number: ((match_id - 1) % 4) + 1,
            match_number: Math.floor((match_id - 1) / 4) + 1,
        }
    }
    else if (match_id <= 18) {
        return {
            comp_level: "sf",
            set_number: ((match_id - 1) % 2) + 1,
            match_number: Math.floor((match_id - 1) / 2) - 5,
        }
    }
    else {
        return {
            comp_level: "f",
            set_number: 1,
            match_number: match_id - 18,
        }
    }
};

Schedule.getTBAMatchKey = function(match) {
    if (match.comp_level == 'qm') {
        return 'qm' + match.match_number;
    }
    else {
        return match.comp_level + match.set_number + 'm' + match.match_number;
    }
};

Schedule.parse = function(rawCsv) {
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
    var missingColumns = ['red1', 'red2', 'red3', 'blue1', 'blue2', 'blue3', 'time', 'description'].filter(function(col) {
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
        var raw_id;
        var code;
        if (match.description.startsWith('qual')) {
            raw_id = Number(match.description.match(/\d+/)[0]);
            code = {
                comp_level: 'qm',
                set_number: 1,
                match_number: raw_id,
            };
        }
        else {
            raw_id = match.description.match(/#(\d+)/)[1];
            code = Schedule.getTBAPlayoffCode(raw_id);
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

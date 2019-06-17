function safeParseLocalStorageObject(key, allow_array) {
    var res;
    try {
        res = JSON.parse(localStorage.getItem(key));
        if (typeof res != 'object') {
            throw new TypeError();
        }
        if (!allow_array && Array.isArray(res)) {
            throw new TypeError();
        }
    }
    catch (e) { }
    return res || {};
}

STORED_EVENTS = safeParseLocalStorageObject('storedEvents');
STORED_AWARDS = safeParseLocalStorageObject('awards');

if (Array.isArray(STORED_AWARDS)) {
    // move old awards from array to object
    var new_awards = {};
    new_awards[localStorage.getItem('selectedEvent') || '?'] = STORED_AWARDS;
    STORED_AWARDS = new_awards;
}

function sendApiRequest(url, event, body) {
    return $.ajax({
        type: 'POST',
        url: url,
        contentType: 'application/json',
        data: JSON.stringify(body),
        headers: {
            'X-Event': event,
            'X-Auth': STORED_EVENTS[event].auth,
            'X-Secret': STORED_EVENTS[event].secret,
        },
    });
}

function tbaApiEventRequest(event, route) {
    var url = 'https://www.thebluealliance.com/api/v3/event/' + event;
    if (route) {
        url += '/' + route;
    }
    return $.ajax({
        type: 'GET',
        url: url,
        headers: {
            'X-TBA-Auth-Key': localStorage.getItem('readApiKey'),
        },
        cache: false,
    });
}

function parseTbaError(error) {
    if (error.responseJSON) {
        if (Array.isArray(error.responseJSON.Errors)) {
            return error.responseJSON.Errors.map(function(err) {
                return Object.values(err).join('\n');
            }).join('\n');
        }
        else if (typeof error.responseJSON.Error == 'string') {
            return error.responseJSON.Error;
        }
    }
    else {
        return error;
    }
}

function confirmPurge() {
    return confirm('Are you sure? This may replace old match results and re-send notifications when these match(es) are uploaded again.');
}

function makeAddEventUI() {
    return {
        event: '',
        auth: '',
        secret: '',
        showAuth: false,
    };
}

function makeAward(data) {
    return $.extend({}, {
        name: '',
        team: '',
        person: '',
    }, data || {});
}

function isValidEventCode(event) {
    return event && Boolean(event.match(/^\d+/));
}

function isValidYear(year) {
    year = parseInt(year);
    return year >= 2018 && year <= 2019;
}

convertToTBARankings = {
    common: function(r) {
        return {
            team_key: 'frc' + r.team,
            rank: r.rank,
            played: r.played,
            dqs: r.dq,
            "Record (W-L-T)": r.wins + '-' + r.losses + '-' + r.ties,
        };
    },
    2018: function(r) {
        return Object.assign(convertToTBARankings.common(r), {
            "Ranking Score": r.sort1,
            "End Game": r.sort2,
            "Auto": r.sort3,
            "Ownership": r.sort4,
            "Vault": r.sort5,
        });
    },
    2019: function(r) {
        return Object.assign(convertToTBARankings.common(r), {
            "Ranking Score": r.sort1,
            "Cargo": r.sort2,
            "Hatch Panel": r.sort3,
            "HAB Climb": r.sort4,
            "Sandstorm Bonus": r.sort5,
        });
    },
};

TBARankingNames = {
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
};

EXTRA_FIELDS = {
    2018: {
        invert_auto: false,
    },
    2019: {
        add_rp_rocket: false,
        add_rp_hab_climb: false,
    },
};

app = new Vue({
    el: '#main',
    data: {
        version: window.VERSION || 'missing version',
        helpHTML: '',
        fmsConfig: window.FMS_CONFIG || {},
        fmsConfigError: '',
        events: Object.keys(STORED_EVENTS).sort(),
        selectedEvent: '',
        addEventUI: makeAddEventUI(),
        readApiKey: localStorage.getItem('readApiKey') || '',
        tbaEventData: {},
        tbaReadError: '',

        uiOptions: $.extend({
            showAllLevels: false,
        }, safeParseLocalStorageObject('uiOptions')),
        eventExtras: safeParseLocalStorageObject('eventExtras'),
        remapError: '',

        matchLevel: 2,
        showAllLevels: false,
        inMatchRequest: false,
        matchError: '',
        // pendingMatches: [], // not set yet to avoid Vue binding to this
        matchSummaries: [],
        fetchedScorelessMatches: false,
        inMatchAdvanced: false,
        advSelectedMatch: '',
        advMatchError: '',

        inEditMatch: false,
        matchEditing: null,
        matchEditData: null,
        matchEditError: '',

        inUploadRankings: false,
        rankingsError: '',

        videos: {},
        inVideoRequest: false,
        videoError: '',
        showExistingVideos: false,

        awards: STORED_AWARDS,
        awardStatus: '',
        inAwardRequest: false,
    },
    computed: {
        eventSelected: function() {
            return !!this.selectedEvent && !this.inAddEvent;
        },
        inAddEvent: function() {
            return this.selectedEvent == '_add';
        },
        canAddEvent: function() {
            return this.addEventUI.event && this.addEventUI.auth && this.addEventUI.secret &&
                isValidYear(this.addEventUI.event);
        },
        addEventIsValidYear: function() {
            return isValidYear(this.addEventUI.event);
        },
        authInputType: function() {
            return this.addEventUI.showAuth ? 'text' : 'password';
        },
        isEventSelected: function() {
            return isValidEventCode(this.selectedEvent);
        },
        eventYear: function() {
            var year = parseInt(this.selectedEvent);
            if (isNaN(year)) {
                return new Date().getFullYear();
            }
            return year;
        },
    },
    methods: {
        saveFMSConfig: function() {
            this.fmsConfigError = '';
            $.ajax({
                type: 'POST',
                url: '/api/fms_config/set',
                contentType: 'application/json',
                data: JSON.stringify(this.fmsConfig),
            }).always(function(data) {
                data = JSON.parse(data);
                if (!data.ok) {
                    this.fmsConfigError = 'Failed to save options: ' + data.error;
                }
            }.bind(this));
        },
        resetFMSConfig: function() {
            $.getJSON('/api/fms_config/get', function(data) {
                this.fmsConfig = data;
            }.bind(this));
        },

        addEvent: function() {
            var event = this.addEventUI.event;
            STORED_EVENTS[event] = {
                auth: this.addEventUI.auth,
                secret: this.addEventUI.secret,
            };
            this.selectedEvent = event;
            if (this.events.indexOf(event) == -1) {
                this.events.push(event);
                this.events.sort();
                this.initAwards(event);
                this.initEventExtras(event);
            }
            localStorage.setItem('storedEvents', JSON.stringify(STORED_EVENTS));
            this.saveAwards();
            this.addEventUI = makeAddEventUI();
        },
        cancelAddEvent: function() {
            this.selectedEvent = '';
        },
        editSelectedEvent: function() {
            this.addEventUI.event = this.selectedEvent;
            this.addEventUI.auth = STORED_EVENTS[this.selectedEvent].auth;
            this.addEventUI.secret = STORED_EVENTS[this.selectedEvent].secret;
            this.selectedEvent = '_add';
        },
        deleteSelectedEvent: function() {
            var oldEvent = this.selectedEvent;
            if (!confirm('Are you sure you want to delete the event "' + oldEvent + '"?')) {
                return;
            }
            this.selectedEvent = '';
            STORED_EVENTS[oldEvent] = undefined;
            this.events = this.events.filter(function(event) {
                return event != oldEvent;
            }.bind(this));
            localStorage.setItem('storedEvents', JSON.stringify(STORED_EVENTS));
        },
        fetchEventData: function() {
            this.tbaReadError = '';
            this.$set(this, 'tbaEventData', {});
            if (!isValidEventCode(this.selectedEvent)) {
                return;
            }
            if (!this.readApiKey) {
                this.tbaReadError = 'No TBA Read API key is present, so event data cannot be retrieved from TBA.';
                return;
            }
            tbaApiEventRequest(this.selectedEvent).then(function(data) {
                this.$set(this, 'tbaEventData', data);
            }.bind(this))
            .fail(function(error) {
                this.tbaReadError = parseTbaError(error);
            }.bind(this));
        },

        initEventExtras: function(event) {
            if (!isValidEventCode(event)) {
                return;
            }
            this.$set(this.eventExtras, event, $.extend({}, {
                remap_teams: [],
            }, this.eventExtras[event]));
        },
        addTeamRemap: function() {
            this.eventExtras[this.selectedEvent].remap_teams.push({
                fms: '',
                tba: '',
            });
        },
        removeTeamRemap: function(i) {
            this.eventExtras[this.selectedEvent].remap_teams.splice(i, 1);
        },
        uploadTeamRemap: function() {
            this.remapError = '';
            var remapList = this.eventExtras[this.selectedEvent].remap_teams;
            var remapMap = {};
            var validate = function(team, isTba) {
                var match = team
                    .toUpperCase()
                    .trim()
                    .replace(/^FRC/, '')
                    .match(isTba ? /^\d+[B-Z]$/ : /^\d+$/);
                if (!match) {
                    this.remapError += 'Invalid ' + (isTba ? 'TBA' : 'FMS') + ' team number: ' + team +
                        ': expected format: 1234' + (isTba ? 'B (or 1234C, etc.)' : '') + '\n';
                }
                return 'frc' + match;
            }.bind(this);
            remapList.forEach(function(r) {
                var fms_team = validate(r.fms);
                var tba_team = validate(r.tba, true);
                if (fms_team && tba_team) {
                    remapMap[fms_team] = tba_team;
                }
            });
            if (this.remapError) {
                return;
            }
            sendApiRequest('/api/info/upload', this.selectedEvent, {
                remap_teams: remapMap,
            }).fail(function(error) {
                this.remapError = parseTbaError(error);
            }.bind(this));
        },

        fetchMatches: function(all) {
            if (all && !confirmPurge()) {
                return;
            }
            this.inMatchRequest = true;
            this.fetchedScorelessMatches = false;
            this.matchError = '';
            $.get('/api/matches/fetch', {
                event: this.selectedEvent,
                level: this.matchLevel,
                all: all ? '1' : '',
            }).always(function() {
                this.inMatchRequest = false;
            }.bind(this)).then(function(data) {
                this.pendingMatches = JSON.parse(data);
                this.pendingMatches.sort(function(a, b) {
                    return Number(a._fms_id.split('-')[0]) - Number(b._fms_id.split('-')[0]);
                });
                this.matchSummaries = this.generateMatchSummaries(this.pendingMatches);
                this.fetchedScorelessMatches = this.checkScorelessMatches(this.pendingMatches);
            }.bind(this)).fail(function(res) {
                this.matchError = res.responseText;
            }.bind(this));
        },
        refetchMatches: function() {
            var match_ids = this.pendingMatches.map(function(match) {
                return match._fms_id;
            });
            this.inMatchRequest = true;
            this.matchError = '';
            sendApiRequest('/api/matches/purge?level=' + this.matchLevel, this.selectedEvent, match_ids)
            .then(function() {
                this.fetchMatches(false);
            }.bind(this))
            .fail(function(res) {
                this.matchError = 'Purge: ' + res.responseText;
                this.inMatchRequest = false;
            }.bind(this));
        },
        generateMatchSummaries: function(matches) {
            var rmFRC = function(team) {
                return team.replace('frc', '');
            };
            var formatMatchCode = function(match) {
                if (match.comp_level == 'qm') {
                    return 'qm' + match.match_number;
                }
                else {
                    return match.comp_level + match.set_number + 'm' + match.match_number;
                }
            };
            var formatScoreSummary = function(match, breakdown, color) {
                var s = '' + match.alliances[color].score;
                if (match.comp_level == 'qm') {
                    s += ' (' + breakdown[color].rp + ')';
                }
                return s;
            };
            var genClasses = function(match, team_key, color) {
                var classes = [color];
                if (match.alliances[color].dqs.indexOf(team_key) != -1) {
                    classes.push('dq');
                }
                if (match.alliances[color].surrogates.indexOf(team_key) != -1) {
                    classes.push('surrogate');
                }
                return classes;
            };

            return matches.map(function(match) {
                classes = {};
                match.alliances.blue.teams.forEach(function(team_key) {
                    classes[rmFRC(team_key)] = genClasses(match, team_key, 'blue');
                });
                match.alliances.red.teams.forEach(function(team_key) {
                    classes[rmFRC(team_key)] = genClasses(match, team_key, 'red');
                });
                return {
                    id: match._fms_id,
                    code: formatMatchCode(match),
                    teams: {
                        blue: match.alliances.blue.teams.map(rmFRC),
                        red: match.alliances.red.teams.map(rmFRC),
                    },
                    score_summary: {
                        blue: formatScoreSummary(match, match.score_breakdown, 'blue'),
                        red: formatScoreSummary(match, match.score_breakdown, 'red'),
                    },
                    classes: classes,
                }
            });
        },
        cleanMatches: function(matches) {
            return matches.map(function(match) {
                var match = Object.assign({}, match);
                delete match._fms_id;
                return match;
            });
        },
        uploadMatches: function() {
            this.matchError = '';
            this.inMatchRequest = true;
            var matches = this.cleanMatches(this.pendingMatches);
            var match_ids = this.pendingMatches.map(function(match) {
                return match._fms_id;
            });
            sendApiRequest('/api/matches/upload', this.selectedEvent, matches).always(function() {
                this.inMatchRequest = false;
            }.bind(this)).then(function() {
                this.pendingMatches = [];
                this.matchSummaries = [];
                sendApiRequest('/api/matches/mark_uploaded?level=' + this.matchLevel,
                                this.selectedEvent, match_ids
                ).fail(function(res) {
                    this.matchError += '\nReceipt generation failed: ' + res.responseText;
                }.bind(this));
            }.bind(this)).fail(function(res) {
                this.matchError = res.responseText;
            }.bind(this));
        },
        checkScorelessMatches: function(matches) {
            return matches.filter(function(match) {
                return match.alliances.blue.score == -1 || match.alliancs.blue.teams[0] == '';
            }).length > 0;
        },
        _checkAdvSelectedMatch: function() {
            parts = this.advSelectedMatch.split('-');
            if (parts.length == 1) {
                parts.push('1');
            }
            this.advSelectedMatch = parts.join('-');
            this.advMatchError = '';
            if (!this.advSelectedMatch.match(/^\d+\-\d+$/)) {
                this.advMatchError = 'Invalid match ID format';
                return false;
            }
            return true;
        },
        purgeAdvSelectedMatch: function() {
            if (!this._checkAdvSelectedMatch() || !confirmPurge()) {
                return;
            }
            this.inMatchRequest = true;
            this.advMatchError = '';
            sendApiRequest('/api/matches/purge?level=' + this.matchLevel, this.selectedEvent, [this.advSelectedMatch])
            .always(function() {
                this.inMatchRequest = false;
            }.bind(this))
            .then(function() {
                this.fetchMatches(false);
            }.bind(this))
            .fail(function(res) {
                this.advMatchError = res.responseText;
            }.bind(this));
        },
        markAdvSelectedMatchUploaded: function() {
            if (!this._checkAdvSelectedMatch()) {
                return;
            }
            this.inMatchRequest = true;
            this.advMatchError = '';
            sendApiRequest('/api/matches/mark_uploaded?level=' + this.matchLevel,
                            this.selectedEvent, [this.advSelectedMatch])
            .always(function() {
                this.inMatchRequest = false;
            }.bind(this))
            .then(function() {
                this.fetchMatches(false);
            }.bind(this))
            .fail(function(res) {
                this.advMatchError = 'Receipt generation failed: ' + res.responseText;
            }.bind(this));
        },

        showEditMatch: function(match) {
            if (this.inMatchRequest)
                return;
            this.inMatchRequest = true;
            this.matchEditing = match;
            var score_breakdown = app.pendingMatches.filter(function(m) {
                return m._fms_id == match.id;
            })[0].score_breakdown;
            sendApiRequest('/api/matches/extra?id=' + this.matchEditing.id + '&level=' + this.matchLevel, this.selectedEvent)
            .then(function(raw) {
                this.inEditMatch = true;
                var data = JSON.parse(raw);
                this.matchEditData = {
                    teams: {},
                    flags: {},
                    text: {},
                };
                ['blue', 'red'].forEach(function(color) {
                    this.matchEditData.teams[color] = this.matchEditing.teams[color].map(function(team) {
                        return {
                            team: team,
                            dq: data[color].dqs.indexOf('frc' + team) != -1,
                            surrogate: data[color].surrogates.indexOf('frc' + team) != -1,
                        };
                    });
                    if (this.matchLevel == 3) {
                        this.matchEditData.flags[color] = {
                            dq: data[color].dqs.length > 0,
                        };
                    }
                    else {
                        var editData = {};
                        Object.keys(EXTRA_FIELDS[this.eventYear]).forEach(function(field) {
                            editData[field] = data[color][field];
                        });
                        this.matchEditData.flags[color] = editData;
                        if (this.eventYear == 2018) {
                            this.matchEditData.text[color] = {
                                auto_rp: score_breakdown[color].autoQuestRankingPoint ^ data[color].invert_auto ?
                                         'missed (FMS returned scored)' :
                                         'scored (FMS returned missed)',
                            };
                        }
                    }
                }.bind(this));
                $('#match-edit-modal').modal('show');
            }.bind(this))
            .always(function() {
                this.inMatchRequest = false;
            }.bind(this));
        },
        hideEditMatch: function() {
            this.inEditMatch = false;
            this.matchEditing = null;
            $('#match-edit-modal').modal('hide');
        },
        saveEditMatch: function() {
            if (this.inMatchRequest)
                return;
            this.matchEditError = '';
            this.inMatchRequest = true;

            var findTeamKeysByFlag = function(color, flag) {
                return this.matchEditData.teams[color].filter(function(t) {
                    return t[flag];
                }).map(function(t) {
                    return 'frc' + t.team;
                });
            }.bind(this);
            var genExtraData = function(color) {
                if (this.matchLevel == 3) {
                    return Object.assign({
                        dqs: this.matchEditData.flags[color].dq ?
                             this.matchEditData.teams[color].map(function(t) {
                                return 'frc' + t.team;
                             }) :
                             [],
                        surrogates: [],
                    }, EXTRA_FIELDS[this.eventYear]);
                }
                return Object.assign({
                    dqs: findTeamKeysByFlag(color, 'dq'),
                    surrogates: findTeamKeysByFlag(color, 'surrogate'),
                }, this.matchEditData.flags[color]);
            }.bind(this);
            var data = {
                blue: genExtraData('blue'),
                red: genExtraData('red'),
            };

            sendApiRequest('/api/matches/extra/save?id=' + this.matchEditing.id + '&level=' + this.matchLevel,
                           this.selectedEvent, data)
            .always(function() {
                this.inMatchRequest = false;
            }.bind(this))
            .then(this.hideEditMatch.bind(this))
            .then(this.refetchMatches.bind(this))
            .fail(function(res) {
                this.matchEditError = res.responseText;
            });
        },

        uploadRankings: function() {
            this.rankingsError = '';
            this.inUploadRankings = true;
            $.getJSON('/api/rankings/fetch', function(data) {
                var rankings = ((data && data.qualRanks) || []).map(convertToTBARankings[this.eventYear]);
                if (!rankings || !rankings.length) {
                    this.rankingsError = 'No rankings available from FMS';
                    this.inUploadRankings = false;
                    return;
                }

                sendApiRequest('/api/rankings/upload', this.selectedEvent, {
                    breakdowns: TBARankingNames[this.eventYear],
                    rankings: rankings,
                }).fail(function(res) {
                    this.rankingsError = res.responseText;
                }.bind(this)).always(function() {
                    this.inUploadRankings = false;
                }.bind(this));
            }.bind(this)).fail(function(res) {
                this.rankingsError = 'fetch failed: ' + res.responseText;
                this.inUploadRankings = false;
            }.bind(this));
        },

        fetchVideos: function() {
            this.inVideoRequest = true;
            this.videoError = '';
            tbaApiEventRequest(this.selectedEvent, 'matches')
            .always(function() {
                this.inVideoRequest = false;
            }.bind(this))
            .then(function(matches) {
                matches.forEach(function(match) {
                    var key = match.key.split('_')[1];
                    if (match.alliances && match.alliances.blue && match.alliances.blue.score != -1) {
                        var v = this.videos[key] || {};
                        v.tba = match.videos.filter(function(cv) {
                            return cv.type == 'youtube';
                        }).map(function(cv) {
                            return cv.key;
                        })[0] || '';
                        v.current = v.current || v.tba || '';
                        Vue.set(this.videos, key, v);
                    }
                }.bind(this));
            }.bind(this))
            .fail(function(error) {
                this.videoError = parseTbaError(error);
            }.bind(this));
        },
        uploadVideos: function() {
            this.cleanVideoUrls();
            var videos = this.getChangedVideos();
            if (!Object.keys(videos).length) {
                this.videoError = 'No videos have changed; not uploading anything.';
                return;
            }

            this.inVideoRequest = true;
            this.videoError = '';
            sendApiRequest('/api/videos/upload', this.selectedEvent, videos)
            .always(function() {
                this.inVideoRequest = false;
            }.bind(this))
            .then(function() {
                this.fetchVideos();
            }.bind(this))
            .fail(function(error) {
                this.videoError = parseTbaError(error);
            }.bind(this));
        },
        getSortedVideos: function() {
            return Object.entries(this.videos).sort(function(a, b) {
                return Number(a[0].replace(/[^\d]/g, '')) - Number(b[0].replace(/[^\d]/g, ''));
            }).filter(function(v) {
                return this.showExistingVideos || !v[1].tba;
            }.bind(this));
        },
        getChangedVideos: function() {
            var videos = {};
            Object.entries(this.videos).forEach(function(v) {
                if (v[1].current && v[1].current != v[1].tba) {
                    videos[v[0]] = v[1].current;
                }
            });
            return videos;
        },
        cleanVideoUrls: function() {
            Object.values(this.videos).forEach(function(v) {
                var match = v.current.match(/[?&]v=([A-Za-z0-9_-]+)/);
                if (match) {
                    v.current = match[1];
                }
            });
        },

        initAwards: function(event) {
            if (!isValidEventCode(event)) {
                return;
            }
            if (this.awards[event] && this.awards[event].length) {
                return;
            }
            this.$set(this.awards, event, [makeAward()]);
        },
        addAward: function() {
            this.awards[this.selectedEvent].push(makeAward());
            this.saveAwards();
        },
        duplicateAward: function(award) {
            var newAward = makeAward();
            newAward.name = award.name;
            this.awards[this.selectedEvent].splice(this.awards[this.selectedEvent].indexOf(award) + 1, 0, newAward);
            this.saveAwards();
        },
        clearAward: function(award) {
            award.name = award.team = award.person = '';
            this.saveAwards();
        },
        deleteAward: function(award) {
            var index = this.awards[this.selectedEvent].indexOf(award);
            if (index >= 0) {
                this.awards[this.selectedEvent].splice(index, 1);
            }
            this.saveAwards();
        },
        fetchAutomaticAwards: function() {
            var cleanedAwards = this.awards[this.selectedEvent].filter(function(award) {
                return ['winner', 'finalist'].indexOf(award.name.toLowerCase().trim()) == -1;
            });
            if (cleanedAwards.length != this.awards[this.selectedEvent].length &&
                    !confirm('This will remove all current Winner/Finalist awards from this event. Continue?')) {
                return;
            }
            this.awards[this.selectedEvent] = cleanedAwards;

            this.inAwardRequest = true;
            tbaApiEventRequest(this.selectedEvent, 'alliances')
            .always(function() {
                this.inAwardRequest = false;
            }.bind(this))
            .then(function(data) {
                var alliances = data || [];
                var winnerAwards = [];
                var finalistAwards = [];
                alliances.forEach(function(alliance) {
                    var status = alliance.status || {};
                    if (status.level == 'f') {
                        var awardName;
                        var awardList;
                        if (status.status == 'won') {
                            awardName = 'Winner';
                            awardList = winnerAwards;
                        }
                        else if (status.status == 'eliminated') {
                            awardName = 'Finalist';
                            awardList = finalistAwards;
                        }
                        if (awardName) {
                            alliance.picks.forEach(function(team) {
                                awardList.push(makeAward({
                                    team: team.replace('frc', ''),
                                    name: awardName,
                                }));
                            });
                        }
                    }
                });
                if (!winnerAwards.length && !finalistAwards.length) {
                    this.awardStatus = 'No winners or finalists were detected. Make sure finals have been uploaded ' +
                        'and the TBA event page is up to date.';
                }
                this.awards[this.selectedEvent] = [].concat(winnerAwards, finalistAwards, this.awards[this.selectedEvent]);
                this.saveAwards();
            }.bind(this))
            .fail(function(error) {
                this.awardStatus = parseTbaError(error);
            }.bind(this))
        },
        saveAwards: function() {
            if (typeof this.awards != 'object' || Array.isArray(this.awards)) {
                throw new TypeError('awards is not a map');
            }
            if (isValidEventCode(this.selectedEvent) && !Array.isArray(this.awards[this.selectedEvent])) {
                throw new TypeError('awards[' + this.selectedEvent + '] is not an array');
            }
            localStorage.setItem('awards', JSON.stringify(this.awards));
        },
        uploadAwards: function() {
            var json = this.awards[this.selectedEvent].map(function(award) {
                return {
                    name_str: award.name,
                    team_key: award.team ? 'frc' + award.team : null,
                    awardee: award.person || null,
                };
            });
            this.inAwardRequest = true;
            this.awardStatus = 'Uploading...';
            var request = sendApiRequest('/api/awards/upload', this.selectedEvent, json);
            request.always(function() {
                this.inAwardRequest = false;
            }.bind(this));
            request.then(function() {
                this.awardStatus = '';
            }.bind(this));
            request.fail(function(res) {
                this.awardStatus = 'Error: ' + res.responseText;
            }.bind(this));
        },
    },
    watch: {
        readApiKey: function(key) {
            localStorage.setItem('readApiKey', key);
        },
        selectedEvent: function(event) {
            localStorage.setItem('selectedEvent', event);
            if (!this.awards[event]) {
                this.awards[event] = [makeAward()];
                this.saveAwards();
            }
            this.initEventExtras(event);
            this.fetchEventData();
        },
        uiOptions: {
            handler: function() {
                localStorage.setItem('uiOptions', JSON.stringify(this.uiOptions));
            },
            deep: true,
        },
        eventExtras: {
            handler: function() {
                localStorage.setItem('eventExtras', JSON.stringify(this.eventExtras));
            },
            deep: true,
        },
    },
    mounted: function() {
        var event = localStorage.getItem('selectedEvent') || '';
        if (event) {
            this.initAwards(event);
            this.initEventExtras(event);
            this.saveAwards();
        }
        this.selectedEvent = event;
        $.get('/README.md', function(readme) {
            // remove first line (header)
            readme = readme.substr(readme.indexOf('\n'));
            this.helpHTML = new showdown.Converter().makeHtml(readme);
        }.bind(this));

        $(this.$refs.mainTabs).on('shown.bs.tab', 'a', function() {
            localStorage.setItem('lastTab', this.id);
        });

        $(function() {
            $(this.$el).removeClass('hidden');
            var lastTab = localStorage.getItem('lastTab');
            if (lastTab) {
                var tab = document.getElementById(lastTab);
                if (tab) {
                    tab.click();
                    $('.tab-pane').removeClass('show active');
                    $('.tab-pane[aria-labelledby=' + lastTab + ']').addClass('show active');
                }
            }
        }.bind(this));
    },
});

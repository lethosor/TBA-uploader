try {
    STORED_EVENTS = JSON.parse(localStorage.getItem('storedEvents')) || {};
}
catch (e) {
    STORED_EVENTS = {};
}
try {
    STORED_AWARDS = JSON.parse(localStorage.getItem('awards'));
}
catch (e) {
    STORED_AWARDS = [];
}
if (!Array.isArray(STORED_AWARDS) || STORED_AWARDS.length == 0) {
    STORED_AWARDS = [makeAward()];
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

function makeAward() {
    return {
        name: '',
        team: '',
        person: '',
    };
}

app = new Vue({
    el: '#main',
    data: {
        version: window.VERSION || 'missing version',
        helpHTML: '',
        fmsConfig: window.FMS_CONFIG || {},
        fmsConfigError: '',
        events: Object.keys(STORED_EVENTS),
        selectedEvent: '',
        addEventUI: makeAddEventUI(),

        matchLevel: 2,
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

        awards: STORED_AWARDS,
        awardStatus: '',
        uploadingAwards: false,
    },
    computed: {
        eventSelected: function() {
            return !!this.selectedEvent && !this.inAddEvent;
        },
        inAddEvent: function() {
            return this.selectedEvent == '_add';
        },
        canAddEvent: function() {
            return this.addEventUI.event && this.addEventUI.auth && this.addEventUI.secret;
        },
        authInputType: function() {
            return this.addEventUI.showAuth ? 'text' : 'password';
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
            }
            localStorage.setItem('storedEvents', JSON.stringify(STORED_EVENTS));
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
            var foundScoreless = false;
            matches.forEach(function(match) {
                var foundNonzero = false;
                ['red', 'blue'].forEach(function(color) {
                    Object.entries(match.score_breakdown[color]).forEach(function(k, v) {
                        if ((typeof v == 'number' || typeof v == 'boolean') && v) {
                            foundNonzero = true;
                        }
                    })
                });
                if (!foundNonzero) {
                    foundScoreless = true;
                }
            });
            return foundScoreless;
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
                this.matchEditData = {
                    teams: {},
                    flags: {},
                    text: {},
                };
                var data = JSON.parse(raw);
                ['blue', 'red'].forEach(function(color) {
                    this.matchEditData.teams[color] = this.matchEditing.teams[color].map(function(team) {
                        return {
                            team: team,
                            dq: data[color].dqs.indexOf('frc' + team) != -1,
                            surrogate: data[color].surrogates.indexOf('frc' + team) != -1,
                        };
                    });
                    this.matchEditData.flags[color] = {
                        invert_auto: data[color].invert_auto,
                    };
                    this.matchEditData.text[color] = {
                        auto_rp: score_breakdown[color].autoQuestRankingPoint ^ data[color].invert_auto ?
                                 'missed (FMS returned scored)' :
                                 'scored (FMS returned missed)',
                    };
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
                return {
                    dqs: findTeamKeysByFlag(color, 'dq'),
                    surrogates: findTeamKeysByFlag(color, 'surrogate'),
                    invert_auto: this.matchEditData.flags[color].invert_auto,
                };
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
                var rankings = data.qualRanks.map(function(r) {
                    return {
                        team_key: 'frc' + r.team,
                        rank: r.rank,
                        played: r.played,
                        dqs: r.dq,

                        "Record (W-L-T)": r.wins + '-' + r.losses + '-' + r.ties,

                        "Auto": r.sort3,
                        "End Game": r.sort2,
                        "Ownership": r.sort4,
                        "Ranking Score": r.sort1,
                        "Vault": r.sort5,
                    };
                });
                if (!rankings || !rankings.length) {
                    this.rankingsError = 'No rankings available from FMS';
                    this.inUploadRankings = false;
                    return;
                }

                sendApiRequest('/api/rankings/upload', this.selectedEvent, {
                    breakdowns: [
                        "Ranking Score",
                        "End Game",
                        "Auto",
                        "Ownership",
                        "Vault",
                        "Record (W-L-T)",
                    ],
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

        addAward: function() {
            this.awards.push(makeAward());
        },
        duplicateAward: function(award) {
            var newAward = makeAward();
            newAward.name = award.name;
            this.awards.splice(this.awards.indexOf(award) + 1, 0, newAward);
        },
        clearAward: function(award) {
            award.name = award.team = award.person = '';
        },
        deleteAward: function(award) {
            var index = this.awards.indexOf(award);
            if (index >= 0) {
                this.awards.splice(index, 1);
            }
        },
        saveAwards: function() {
            localStorage.setItem('awards', JSON.stringify(this.awards));
        },
        uploadAwards: function() {
            var json = this.awards.map(function(award) {
                return {
                    name_str: award.name,
                    team_key: award.team ? 'frc' + award.team : null,
                    awardee: award.person || null,
                };
            });
            this.uploadingAwards = true;
            this.awardStatus = 'Uploading...';
            var request = sendApiRequest('/api/awards/upload', this.selectedEvent, json);
            request.always(function() {
                this.uploadingAwards = false;
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
        selectedEvent: function(event) {
            localStorage.setItem('selectedEvent', event);
        },
    },
    mounted: function() {
        $(this.$el).removeClass('hidden');
        this.selectedEvent = localStorage.getItem('selectedEvent') || '';
        $.get('/README.md', function(readme) {
            // remove first line (header)
            readme = readme.substr(readme.indexOf('\n'));
            this.helpHTML = new showdown.Converter().makeHtml(readme);
        }.bind(this));
    },
});

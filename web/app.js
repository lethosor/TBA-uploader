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
        events: Object.keys(STORED_EVENTS),
        selectedEvent: '',
        addEventUI: makeAddEventUI(),

        matchLevel: 2,
        inMatchRequest: false,
        matchError: '',
        // pendingMatches: [], // not set yet to avoid Vue binding to this
        matchSummaries: [],

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
            this.inMatchRequest = true;
            this.matchError = '';
            $.get('/api/matches/fetch', {
                level: this.matchLevel,
                all: all ? '1' : '',
            }).always(function() {
                this.inMatchRequest = false;
            }.bind(this)).then(function(data) {
                this.pendingMatches = JSON.parse(data);
                this.matchSummaries = this.generateMatchSummaries(this.pendingMatches);
            }.bind(this)).fail(function(res) {
                this.matchError = res.responseText;
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
            return matches.map(function(match) {
                var breakdown = JSON.parse(match.score_breakdown);
                return {
                    id: match._fms_id,
                    code: formatMatchCode(match),
                    teams: {
                        blue: match.alliances.blue.teams.map(rmFRC),
                        red: match.alliances.red.teams.map(rmFRC),
                    },
                    score_summary: {
                        blue: formatScoreSummary(match, breakdown, 'blue'),
                        red: formatScoreSummary(match, breakdown, 'red'),
                    },
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
            sendApiRequest('/api/awards/upload', this.selectedEvent, matches).always(function() {
                this.inMatchRequest = false;
            }.bind(this)).then(function() {
                this.pendingMatches = [];
                this.matchSummaries = [];
            }.bind(this)).fail(function(res) {
                this.matchError = res.responseText;
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
    },
});

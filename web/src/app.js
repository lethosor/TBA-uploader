import 'regenerator-runtime';

import api from 'src/api.js';
import {
    BRACKET_NAME,
    BRACKET_TYPE,
    MATCH_LEVEL,
} from 'src/consts.js';
import Schedule from 'src/schedule.js';
import tba from 'src/tba.js';
import utils from 'src/utils.js';

import Alert from 'components/Alert.vue';
import Dropzone from 'components/Dropzone.vue';
import ScoreSummary from 'components/ScoreSummary.vue';

const STORED_EVENTS = utils.safeParseLocalStorageObject('storedEvents');
const STORED_AWARDS = utils.safeParseLocalStorageObject('awards');

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
window.sendApiRequest = sendApiRequest;

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
window.tbaApiEventRequest = tbaApiEventRequest;

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

const EXTRA_FIELDS = {
    2018: {
        invert_auto: false,
    },
    2019: {
        add_rp_rocket: false,
        add_rp_hab_climb: false,
    },
};

const app = new Vue({
    el: '#main',
    components: {
        Alert,
        Dropzone,
        ScoreSummary,
    },
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
        }, utils.safeParseLocalStorageObject('uiOptions')),
        eventExtras: utils.safeParseLocalStorageObject('eventExtras'),
        remapError: '',

        inScheduleRequest: false,
        scheduleUploaded: false,
        scheduleError: '',
        scheduleStats: [],
        schedulePendingMatches: [],

        matchLevel: MATCH_LEVEL.QUAL,
        showAllLevels: false,
        inMatchRequest: false,
        matchError: '',
        // pendingMatches: [], // not set yet to avoid Vue binding to this
        matchSummaries: [],
        fetchedScorelessMatches: false,
        unhandledBreakdowns: [],
        inMatchAdvanced: false,
        advSelectedMatch: '',
        advMatchError: '',

        inEditMatch: false,
        matchEditing: null,
        matchEditData: null,
        matchEditError: '',
        matchEditOverrideCode: false,

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
        BRACKET_TYPES: function() {
            return Object.fromEntries(Object.keys(BRACKET_TYPE).map((key) => [
                BRACKET_TYPE[key], BRACKET_NAME[key],
            ]));
        },
        eventSelected: function() {
            return !!this.selectedEvent && !this.inAddEvent;
        },
        inAddEvent: function() {
            return this.selectedEvent == '_add';
        },
        canAddEvent: function() {
            return this.addEventUI.event && this.addEventUI.auth && this.addEventUI.secret &&
                tba.isValidYear(this.addEventUI.event);
        },
        addEventIsValidYear: function() {
            return tba.isValidYear(this.addEventUI.event);
        },
        eventRequestHeaders: function() {
            if (!STORED_EVENTS[this.selectedEvent]) {
                return {};
            }
            return {
                'X-Event': this.selectedEvent,
                'X-Auth': STORED_EVENTS[this.selectedEvent].auth,
                'X-Secret': STORED_EVENTS[this.selectedEvent].secret,
            };
        },
        authInputType: function() {
            return this.addEventUI.showAuth ? 'text' : 'password';
        },
        isEventSelected: function() {
            return tba.isValidEventCode(this.selectedEvent);
        },
        eventYear: function() {
            var year = parseInt(this.selectedEvent);
            if (isNaN(year)) {
                return new Date().getFullYear();
            }
            return year;
        },
        isQual: function() {
            return this.matchLevel == MATCH_LEVEL.QUAL;
        },
        isPlayoff: function() {
            return this.matchLevel == MATCH_LEVEL.PLAYOFF;
        },
        schedulePendingMatchCells: function() {
            var addTeamCell = function(cells, match, color, i) {
                var cls = {};
                cls[color] = true;
                cls['surrogate'] = match.alliances[color].surrogates.indexOf(match.alliances[color].teams[i]) >= 0;
                cells.push({
                    text: match.alliances[color].teams[i].replace('frc', ''),
                    cls: cls,
                });
            };
            return this.schedulePendingMatches.map(function(match) {
                var cells = [
                    {text: match._id},
                    {text: match._key},
                    {text: match.time_string},
                ];
                ['red', 'blue'].forEach(function(color) {
                    for (var i = 0; i < 3; i++) {
                        addTeamCell(cells, match, color, i);
                    }
                });
                return cells;
            });
        },
    },
    watch: {
        readApiKey: function(key) {
            localStorage.setItem('readApiKey', key);
        },
        selectedEvent: function(event) {
            localStorage.setItem('selectedEvent', event);
            this.initEvent(event);
            this.fetchEventData();
            this.scheduleReset(false);
        },
        matchLevel: function() {
            localStorage.setItem('matchLevel', this.matchLevel);
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
            this.initEvent(event);
        }
        this.selectedEvent = event;
        this.matchLevel = localStorage.getItem('matchLevel') || this.matchLevel;
        $.get('/README.md', function(readme) {
            // remove first line (header)
            readme = readme.substr(readme.indexOf('\n'));
            this.helpHTML = new showdown.Converter().makeHtml(readme);
        }.bind(this));

        $(this.$refs.mainTabs).on('shown.bs.tab', 'a', function() {
            localStorage.setItem('lastTab', this.id);
            $('[data-accesskey]').each(function(_, e) {
                e = $(e);
                if (e.is(':visible')) {
                    e.attr('accesskey', e.attr('data-accesskey'));
                }
                else {
                    e.removeAttr('accesskey');
                }
                e.attr('title', '[' + e.attr('accesskey') + ']');
            });
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

        this.$refs.scheduleUpload.$on('upload', this.onScheduleUpload);
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
                this.initEvent(event);
            }
            localStorage.setItem('storedEvents', JSON.stringify(STORED_EVENTS));
            this.addEventUI = makeAddEventUI();
            this.syncEvents();
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
            if (!tba.isValidEventCode(this.selectedEvent)) {
                return;
            }
            if (!this.readApiKey) {
                this.tbaReadError = 'No TBA Read API key is present, so event data cannot be retrieved from TBA.';
                return;
            }
            tbaApiEventRequest(this.selectedEvent).then(function(data) {
                this.$set(this, 'tbaEventData', data);
                this.eventExtras[this.selectedEvent].playoff_type = data.playoff_type;
            }.bind(this))
            .fail(function(error) {
                this.tbaReadError = utils.parseErrorJSON(error);
            }.bind(this));
        },
        initEvent: function(event) {
            if (!tba.isValidEventCode(event)) {
                return;
            }

            this.$set(this.eventExtras, event, $.extend({}, {
                remap_teams: [],
            }, this.eventExtras[event]));

            if (!this.awards[event] || !this.awards[event].length) {
                this.$set(this.awards, event, [makeAward()]);
                this.saveAwards();
            }
        },
        syncEvents: async function() {
            var events = await api.postJson({url: '/api/keys/fetch'});
            try {
                events = JSON.parse(events);
            }
            catch (e) {
                events = {};
            }
            for (const k of Object.keys(events)) {
                if (!STORED_EVENTS[k]) {
                    STORED_EVENTS[k] = events[k];
                    this.events.push(k);
                }
            }
            localStorage.setItem('storedEvents', JSON.stringify(STORED_EVENTS));
            await api.postJson({
                url: '/api/keys/update',
                body: STORED_EVENTS,
            });
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
                this.remapError = utils.parseErrorText(error);
            }.bind(this));
        },

        updatePlayoffType: function() {
            const playoff_type = this.eventExtras[this.selectedEvent].playoff_type;
            sendApiRequest('/api/info/upload', this.selectedEvent, {
                playoff_type,
            }).then(function() {
                this.tbaEventData.playoff_type = playoff_type;
            }.bind(this)).fail(function(error) {
                this.remapError = utils.parseErrorText(error);
            }.bind(this));
        },

        scheduleReset: function(keepFile) {
            this.inScheduleRequest = false;
            this.scheduleUploaded = false;
            this.scheduleError = '';
            this.scheduleStats = [];
            this.schedulePendingMatches = [];

            if (!keepFile) {
                this.$refs.scheduleUpload.reset();
            }
        },
        onScheduleUpload: function(event) {
            this.scheduleReset(true);
            try {
                var schedule = Schedule.parse(event.body);
            }
            catch (error) {
                if (typeof error == 'string') {
                    this.scheduleError = error;
                    return;
                }
                else {
                    throw error;
                }
            }
            this.scheduleStats.push(schedule.length + ' match(es) found');
            var numSurrogates = schedule.map(function(match) {
                return match.alliances.red.surrogates.length + match.alliances.blue.surrogates.length;
            }).reduce(function(a, b) {
                return a + b;
            }, 0);
            this.scheduleStats.push(numSurrogates + ' surrogate team(s)');

            this.scheduleStats.push('Checking against TBA schedule...');
            this.inScheduleRequest = true;
            tbaApiEventRequest(this.selectedEvent, 'matches').always(function() {
                this.inScheduleRequest = false;
                this.scheduleStats.pop();
            }.bind(this)).then(function(tbaMatches) {
                if (!tbaMatches) {
                    tbaMatches = [];
                }
                var newLevels = Schedule.findAllCompLevels(schedule);
                var tbaLevels = Schedule.findAllCompLevels(tbaMatches);
                this.scheduleStats.push('TBA has level(s): ' + tbaLevels.join(', '));
                this.scheduleStats.push('The FMS report has level(s): ' + newLevels.join(', '));
                newLevels = newLevels.filter(function(level) {
                    return tbaLevels.indexOf(level) < 0;
                });
                if (!newLevels.length) {
                    this.scheduleStats.push('No new levels are present in the FMS report.');
                    return;
                }
                this.scheduleStats.push('Level(s) to be added from the FMS report: ' + newLevels.join(', '));
                this.schedulePendingMatches = schedule.filter(function(match) {
                    return newLevels.indexOf(match.comp_level) >= 0;
                });
            }.bind(this)).fail(function(error) {
                this.scheduleError = utils.parseErrorJSON(error);
            }.bind(this));
        },
        postSchedule: function() {
            this.scheduleError = '';
            this.inScheduleRequest = true;
            sendApiRequest('/api/matches/upload', this.selectedEvent, this.schedulePendingMatches).always(function() {
                this.inScheduleRequest = false;
            }.bind(this)).then(function() {
                this.scheduleReset(false);
                this.scheduleUploaded = true;
            }.bind(this)).fail(function(res) {
                this.scheduleError = res.responseText;
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
                this.unhandledBreakdowns = this.findUnhandledBreakdowns(this.pendingMatches);
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
                var classes = {};
                match.alliances.blue.teams.forEach(function(team_key) {
                    classes[rmFRC(team_key)] = genClasses(match, team_key, 'blue');
                });
                match.alliances.red.teams.forEach(function(team_key) {
                    classes[rmFRC(team_key)] = genClasses(match, team_key, 'red');
                });
                return {
                    id: match._fms_id,
                    key: Schedule.getTBAMatchKey(match),
                    code: {
                        comp_level: match.comp_level,
                        set_number: match.set_number,
                        match_number: match.match_number,
                    },
                    teams: {
                        blue: match.alliances.blue.teams.map(rmFRC),
                        red: match.alliances.red.teams.map(rmFRC),
                    },
                    score_summary: {
                        blue: formatScoreSummary(match, match.score_breakdown, 'blue'),
                        red: formatScoreSummary(match, match.score_breakdown, 'red'),
                    },
                    classes: classes,
                };
            });
        },
        cleanMatches: function(matches) {
            return matches.map(function(match) {
                match = Object.assign({}, match);
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
                if (this.isQual) {
                    this.uploadRankings();
                }
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
                return match.alliances.blue.score == -1 || match.alliances.blue.teams[0] == '';
            }).length > 0;
        },
        findUnhandledBreakdowns: function(matches) {
            var unhandled = new Set();
            for (const match of matches) {
                for (const breakdown of Object.values(match.score_breakdown)) {
                    for (const field of Object.keys(breakdown)) {
                        if (field.startsWith('!')) {
                            unhandled.add(field.replace(/!/g, ''));
                        }
                    }
                }
            }
            return [...unhandled];
        },
        _checkAdvSelectedMatch: function() {
            var parts = this.advSelectedMatch.split('-');
            if (parts.length == 1) {
                parts.push('1');
            }
            this.advSelectedMatch = parts.join('-');
            this.advMatchError = '';
            if (!this.advSelectedMatch.match(/^\d+-\d+$/)) {
                this.advMatchError = 'Invalid match ID format';
                return false;
            }
            return true;
        },
        purgeAdvSelectedMatch: async function() {
            if (!this._checkAdvSelectedMatch() || !confirmPurge()) {
                return;
            }
            this.inMatchRequest = true;
            this.advMatchError = '';
            try {
                await api.postJson({
                    url: '/api/matches/purge?level=' + this.matchLevel,
                    headers: this.eventRequestHeaders,
                    body: [this.advSelectedMatch],
                });
                this.fetchMatches(false);
            }
            catch (error) {
                this.advMatchError = error;
            }
            finally {
                this.inMatchRequest = false;
            }
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
                this.matchEditOverrideCode = Boolean(data.match_code_override);
                ['blue', 'red'].forEach(function(color) {
                    this.matchEditData.teams[color] = this.matchEditing.teams[color].map(function(team) {
                        return {
                            team: team,
                            dq: data[color].dqs.indexOf('frc' + team) != -1,
                            surrogate: data[color].surrogates.indexOf('frc' + team) != -1,
                        };
                    });
                    if (this.isPlayoff) {
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
                if (this.isPlayoff) {
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
            if (this.matchEditOverrideCode) {
                data.match_code_override = this.matchEditing.code;
            }

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
        editMatchMarkUploaded: function() {
            this.advSelectedMatch = this.matchEditing.id;
            this.markAdvSelectedMatchUploaded();
            this.hideEditMatch();
        },

        uploadRankings: function() {
            this.rankingsError = '';
            this.inUploadRankings = true;
            $.getJSON('/api/rankings/fetch', function(data) {
                var rankings = ((data && data.qualRanks) || []).map(tba.convertToTBARankings[this.eventYear]);
                if (!rankings || !rankings.length) {
                    this.rankingsError = 'No rankings available from FMS';
                    this.inUploadRankings = false;
                    return;
                }

                sendApiRequest('/api/rankings/upload', this.selectedEvent, {
                    breakdowns: tba.RANKING_NAMES[this.eventYear],
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
                        v.uploaded = Boolean(v.uploaded);
                        Vue.set(this.videos, key, v);
                    }
                }.bind(this));
            }.bind(this))
            .fail(function(error) {
                this.videoError = utils.parseErrorJSON(error);
            }.bind(this));
        },
        uploadVideos: function() {
            this.cleanVideoUrls();
            var videos = this.getChangedVideos();
            if (!Object.keys(videos).length) {
                this.videoError = 'No videos have changed; not uploading anything.';
                return;
            }
            var invalidVideos = Object.values(videos).filter(function(value) {
                return !value.match(/^[A-Za-z0-9_-]{11}$/);
            });
            if (invalidVideos.length) {
                this.videoError = 'The following IDs are not valid Youtube video IDs. Please submit only the 11-character video ID.\n' +
                    invalidVideos.join('\n');
                return;
            }

            this.inVideoRequest = true;
            this.videoError = '';
            sendApiRequest('/api/videos/upload', this.selectedEvent, videos)
            .always(function() {
                this.inVideoRequest = false;
            }.bind(this))
            .then(function() {
                Object.keys(videos).forEach(function(key) {
                    this.videos[key].uploaded = true;
                }.bind(this));
                this.fetchVideos();
            }.bind(this))
            .fail(function(error) {
                this.videoError = utils.parseErrorJSON(error);
            }.bind(this));
        },
        getSortedVideos: function() {
            return Object.entries(this.videos).sort(function(a, b) {
                return Number(a[0].replace(/[^\d]/g, '')) - Number(b[0].replace(/[^\d]/g, ''));
            }).filter(function(v) {
                return this.showExistingVideos || (!v[1].uploaded && !v[1].tba);
            }.bind(this));
        },
        getChangedVideos: function() {
            var videos = {};
            Object.entries(this.videos).forEach(function(v) {
                if (v[1].current && utils.cleanYoutubeUrl(v[1].current) != utils.cleanYoutubeUrl(v[1].tba)) {
                    videos[v[0]] = v[1].current;
                }
            });
            return videos;
        },
        cleanVideoUrls: function() {
            Object.values(this.videos).forEach(function(v) {
                v.current = utils.cleanYoutubeUrl(v.current);
            });
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
                this.awardStatus = utils.parseErrorJSON(error);
            }.bind(this));
        },
        saveAwards: function() {
            if (typeof this.awards != 'object' || Array.isArray(this.awards)) {
                throw new TypeError('awards is not a map');
            }
            if (tba.isValidEventCode(this.selectedEvent) && !Array.isArray(this.awards[this.selectedEvent])) {
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
            if (json.filter(function(award) { return !award.name_str; }).length) {
                this.awardStatus = 'One or more awards have an empty name. Please correct this and try again.';
                return;
            }
            this.inAwardRequest = true;
            this.awardStatus = 'Uploading...';
            var request = sendApiRequest('/api/awards/upload', this.selectedEvent, json);
            request.always(function() {
                this.inAwardRequest = false;
            }.bind(this));
            request.then(function() {
                this.awardStatus = 'Upload succeeded.';
            }.bind(this));
            request.fail(function(res) {
                this.awardStatus = 'Error: ' + res.responseText;
            }.bind(this));
        },
    },
});

window.app = app;

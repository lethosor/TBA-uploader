import { FIELD_STATE, MATCH_LEVEL } from './consts';

function parseGenericResponseError(res) {
    if (parseInt(res.status) >= 400) {
        return 'Connection failed with code ' + res.status;
    }
    if (res.readyState === 0 || (typeof res.statusText == 'string' && res.statusText.toLowerCase() == 'error')) {
        return 'Connection failed with unspecified error';
    }
    return String(res);
}

const utils = Object.freeze({
    safeParseLocalStorageObject(key) {
        var res;
        try {
            res = JSON.parse(localStorage.getItem(key));
            if (typeof res != 'object') {
                throw new TypeError();
            }
            if (Array.isArray(res)) {
                throw new TypeError();
            }
        }
        catch (e) {
            // ignore
        }
        return res || {};
    },

    safeParseLocalStorageInteger(key, defaultValue) {
        var res = parseInt(localStorage.getItem(key));
        if (isNaN(res)) {
            return defaultValue;
        }
        return res;
    },

    parseCSVRaw(raw) {
        return raw.split('\n').map(function(line) {
            return line.trim().split(',');
        });
    },

    parseCSVObjects(cells, headerRowIndex=0) {
        const headerRow = cells[headerRowIndex];
        return cells.slice(headerRowIndex + 1).map(row => {
            let obj = {};
            for (let i = 0; i < row.length; i++) {
                if (headerRow[i]) {
                    obj[headerRow[i]] = row[i];
                }
            }
            return obj;
        });
    },

    parseErrorText(res) {
        if (res.responseText) {
            return res.responseText;
        }
        return parseGenericResponseError(res);
    },

    parseErrorJSON(res) {
        if (res.responseJSON) {
            if (Array.isArray(res.responseJSON.Errors)) {
                return res.responseJSON.Errors.map(function(err) {
                    return Object.values(err).join('\n');
                }).join('\n');
            }
            else if (typeof res.responseJSON.Error == 'string') {
                return res.responseJSON.Error;
            }
        }
        else {
            return 'JSON fetch failed: ' + this.parseErrorText(res);
        }
    },

    cleanYoutubeUrl(url) {
        var match = url.match(/(youtu.be\/|\/video\/|[?&]v=)([A-Za-z0-9_-]+)/);
        if (match) {
            url = match[2];
        }
        return url;
    },

    makeProxiedAjaxRequest(options, proxyUrl) {
        options = Object.assign({}, options);  // copy
        if (proxyUrl) {
            if (String(options.method).toUpperCase() == 'GET' || String(options.type).toUpperCase() == 'GET') {
                if (options.data) {
                    // convert query string manually
                    options.url += '?' + $.param(options.data);
                    delete options.data;
                }
            }
            options.url = proxyUrl + '?' + $.param({url: options.url});
        }
        return $.ajax(options);
    },

    FIELD_STATE_MESSAGES: Object.freeze({
        [FIELD_STATE.WaitingForPrestart]:   "Ready to Pre-Start",
        [FIELD_STATE.WaitingForPrestartTO]: "Ready to Pre-Start?",
        [FIELD_STATE.Prestarting]: "Pre-Starting",
        [FIELD_STATE.PrestartingTO]: "Pre-Starting?",
        [FIELD_STATE.WaitingForSetAudience]: "Pre-Start Complete. Waiting for Audience Display",
        [FIELD_STATE.WaitingForSetAudienceTO]: "Pre-Start Complete? Waiting for Audience Display",
        [FIELD_STATE.WaitingForMatchPreview]: "Pre-Start Complete. Waiting for Match Preview",
        [FIELD_STATE.WaitingForMatchPreviewTO]: "Pre-Start Complete? Waiting for Match Preview",
        [FIELD_STATE.WaitingForMatchReady]: "Waiting for Teams",
        [FIELD_STATE.WaitingForMatchStart]: "Match Ready",
        [FIELD_STATE.GameSpecific]: "Sending Game-Specific Data",
        [FIELD_STATE.MatchAuto]: "Match Running (Auto)",
        [FIELD_STATE.MatchTransition]: "Match Transitioning",
        [FIELD_STATE.MatchTeleop]: "Match Running (Teleop)",
        [FIELD_STATE.WaitingForCommit]: "Match Over. Waiting for Commit",
        [FIELD_STATE.WaitingForPostResults]: "Match Over. Waiting for Post-Result",
        [FIELD_STATE.TournamentLevelComplete]: "Tournament Level Complete",
        [FIELD_STATE.MatchCancelled]: "Match Aborted",
    }),

    describeFieldState(fieldState) {
        if (fieldState === undefined) {
            return "Unknown";
        }
        return utils.FIELD_STATE_MESSAGES[fieldState] || ("Unknown: " + fieldState);
    },

    isFieldStateInMatch(fieldState) {
        return Boolean({
            [FIELD_STATE.GameSpecific]: true,
            [FIELD_STATE.MatchAuto]: true,
            [FIELD_STATE.MatchTransition]: true,
            [FIELD_STATE.MatchTeleop]: true,
        }[fieldState]);
    },

    isFieldStateInMatchLoaded(fieldState) {
        return Boolean({
            [FIELD_STATE.WaitingForSetAudience]: true,
            [FIELD_STATE.WaitingForMatchPreview]: true,
            [FIELD_STATE.WaitingForMatchReady]: true,
            [FIELD_STATE.WaitingForMatchStart]: true,
            [FIELD_STATE.GameSpecific]: true,
            [FIELD_STATE.MatchAuto]: true,
            [FIELD_STATE.MatchTransition]: true,
            [FIELD_STATE.MatchTeleop]: true,
            [FIELD_STATE.WaitingForCommit]: true,
            [FIELD_STATE.WaitingForPostResults]: true,
            [FIELD_STATE.MatchCancelled]: true,
        }[fieldState]);
    },

    describeMatchLevel(matchLevel, defaultValue) {
        return ({
            [MATCH_LEVEL.TEST]: "Test",
            [MATCH_LEVEL.PRACTICE]: "Practice",
            [MATCH_LEVEL.QUAL]: "Qualification",
            [MATCH_LEVEL.PLAYOFF]: "Playoff",
            [MATCH_LEVEL.MANUAL]: "Manual",
        })[matchLevel] || defaultValue || "Invalid";
    },
});

export default utils;

function parseGenericResponseError(res) {
    if (parseInt(res.status) >= 400) {
        return 'Connection failed with code ' + res.status;
    }
    if (res.readyState === 0 || res.statusText.toLowerCase() == 'error') {
        return 'Connection failed with unspecified error';
    }
    return String(res);
}

export default Object.freeze({
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

    parseCSV(raw) {
        return raw.split('\n').map(function(line) {
            return line.trim().split(',');
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

});

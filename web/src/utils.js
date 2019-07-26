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
    }
});

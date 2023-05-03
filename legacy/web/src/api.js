import utils from 'src/utils.js';

export default Object.freeze({
    async postJson({url, headers, body}) {
        return new Promise((resolve, reject) => {
            $.ajax(url, {
                type: 'POST',
                contentType: 'application/json',
                url,
                headers,
                data: JSON.stringify(body),
            })
            .then((data) => resolve(data))
            .fail((error) => reject(utils.parseErrorText(error)));
        });
    },

    async getTbaJson({event, route, key}) {
        var url = FMS_CONFIG.tba_url + '/api/v3/event/' + event;
        if (route) {
            url += '/' + route;
        }
        return new Promise((resolve, reject) => {
            $.ajax({
                type: 'GET',
                url,
                headers: {
                    'X-TBA-Auth-Key': key,
                },
                cache: false,
            })
            .then((data) => resolve(data))
            .fail((error) => reject(error));
        });
    },

});

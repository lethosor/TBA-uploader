export default function SocketConnection(url) {
    const state = {
        url,
        socket: null,
        reconnectTimeout: 1000,
        eventHandlers: {
            open: [],
            close: [],
            message: [],
        },
    };

    function triggerEvent(type, data) {
        console.log(type, data);
        for (const handler of state.eventHandlers[type]) {
            handler(data);
        }
    }

    function connect() {
        if (state.socket) {
            console.warn('already have a socket');
            return;
        }
        state.socket = new WebSocket(state.url);
        state.socket.addEventListener('open', function(e) {
            triggerEvent('open', e);
        });
        state.socket.addEventListener('close', function(e) {
            state.socket = null;
            setTimeout(connect, state.reconnectTimeout);
            triggerEvent('close', e);
        });
        state.socket.addEventListener('message', function(e) {
            triggerEvent('message', e);
        });
    }

    function reconnect() {
        if (state.socket) {
            state.socket.close();
        }
        connect();
    }

    if (state.url) {
        connect();
    }

    return {
        on: function(type, handler) {
            if (!state.eventHandlers[type]) {
                throw 'invalid message type: ' + type;
            }
            state.eventHandlers[type].push(handler);
        },
        setUrl: function(url) {
            console.log('setUrl', url);
            if (url != state.url) {
                state.url = url;
                reconnect();
            }
        },
        _state: state,
    };
}

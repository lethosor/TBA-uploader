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
        events: Object.keys(STORED_EVENTS),
        selectedEvent: '',
        addEventUI: makeAddEventUI(),
        awards: STORED_AWARDS,
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

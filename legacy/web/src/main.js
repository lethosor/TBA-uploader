import Vue from 'vue';

import App from 'components/App.vue';

const app = new Vue({
    el: '#main',
    render: (createElement) => createElement(App),
});

window.app = app.$children[0];

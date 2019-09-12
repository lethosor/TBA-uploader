<template>
    <b-alert
        :variant="type"
        :show="text.length > 0"
        dismissible
        @dismissed="$emit('input', '')"
    >
        <span class="raw-message">{{ text }}</span>
    </b-alert>
</template>

<script>
import {BAlert} from 'bootstrap-vue';

export default {
    name: 'Alert',
    components: {
        BAlert,
    },
    props: {
        type: {
            type: String,
            required: true,
        },
        value: {
            type: String,
            required: true,
        },
        prefix: {
            type: String,
            default: '',
        },
    },
    computed: {
        text() {
            var text = this.value;
            if (!text) {
                return '';
            }
            if (this.prefix) {
                text = `${this.prefix} ${text}`;
            }
            return text;
        },
    },
};
</script>

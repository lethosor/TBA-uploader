<template>
    <div
        v-if="value"
        :class="{alert: true, ['alert-' + type]: true}"
    >
        <button
            type="button"
            class="close"
            aria-label="Close"
            @click="$emit('input', '')"
        >
            <span aria-hidden="true">&times;</span>
        </button>
        <span class="raw-message">{{ text }}</span>
    </div>
</template>

<script>
export default {
    name: 'Alert',
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

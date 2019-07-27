<template>
    <div v-if="value" :class="{alert: true, ['alert-' + type]: true}">
        <button @click="$emit('input', '')" type="button" class="close" aria-label="Close">
            <span aria-hidden="true">&times;</span>
        </button>
        <span class="raw-message">{{text}}</span>
    </div>
</template>

<script>
export default {
    name: 'Alert',
    props: {
        type: String,
        value: String,
        prefix: String,
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

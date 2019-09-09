<template>
    <div
        :id="id"
        :class="{'tab-pane': true, fade: true, show: selected, active: selected}"
        role="tabpanel"
        :aria-labelledby="id"
    >
        <slot />
    </div>
</template>

<script>
export default {
    name: 'TabContent',
    props: {
        id: {
            type: String,
            required: true,
        },
        selected: {
            type: Boolean,
            default: false,
        },
    },
};
</script>

<template>
    <div :class="{'tab-pane': true, fade: true, show: selected, active: selected}" :id="id" blah role="tabpanel" :aria-labelledby="id">
        <slot></slot>
    </div>
</template>

<script>
export default {
    name: 'TabContent',
    props: {
        id: String,
        selected: Boolean,
    },
};
</script>

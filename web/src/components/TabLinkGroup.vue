<template>
    <ul class="nav nav-tabs" role="tablist">
        <slot></slot>
    </ul>
</template>

<script>
export default {
    props: {
        id: String,
        href: String,
        name: String,
        selected: Boolean,
    },
};
</script>

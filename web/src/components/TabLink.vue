<template>
    <li class="nav-item">
        <a :class="{'nav-link': true, active: selected}" data-toggle="tab" :href="'#' + id" role="tab" :aria-controls="id" :aria-selected="selected">{{name}}</a>
    </li>
</template>

<script>
export default {
    name: 'TabLink',
    props: {
        id: String,
        name: String,
        selected: Boolean,
    }
}
</script>

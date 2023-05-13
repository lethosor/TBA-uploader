<script lang="ts">
export let href = "";

$: target = (href.includes("://") ? "_blank" : "");
</script>

<a {href} {target}><slot>{href}</slot></a>

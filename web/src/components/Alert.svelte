<script lang="ts">
    import { getFormContext } from '$lib/FormContext';

    export let text = "";
    let classProp = "";
    export {classProp as class};

    $: classes = ["alert", classProp];

    function clear() {
        text = "";
    }
</script>

{#if text}
<aside class={classes.filter(Boolean).join(" ")}>
    <div class="alert-message">
        {text}
    </div>
    <div class="alert-buttons">
        <button class="btn-icon variant-filled text-xl" on:click={clear}>&times;</button>
    </div>
</aside>
{/if}

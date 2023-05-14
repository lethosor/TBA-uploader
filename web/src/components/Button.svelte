<script lang="ts">
    export let color = "";
    export let variant = "filled";
    export let disabled = false;
    let classProp = "";
    export {classProp as class};

    $: classes = ["btn", `variant-${variant}${color ? '-' + color : ''}`, classProp];
</script>

<button on:click type="button" class={classes.filter(Boolean).join(" ")} {disabled}><slot/></button>

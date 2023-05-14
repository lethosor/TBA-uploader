<script lang="ts">
    import { getFormContext } from '$lib/FormContext';

    export let color = "";
    export let variant = "filled";
    export let disabled = false;
    let classProp = "";
    export {classProp as class};

    $: classes = ["btn", `variant-${variant}${color ? '-' + color : ''}`, classProp];

    $: disabledEffective = disabled;
    $: {
        let formContext = getFormContext();
        if (formContext) {
            formContext.subscribe(state => {
                disabledEffective = disabled || (state.inSubmit);
            });
        }
    }

</script>

<button on:click type="button" class={classes.filter(Boolean).join(" ")} disabled={disabledEffective}><slot/></button>

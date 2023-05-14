<script lang="ts">
    import { setContext } from 'svelte';
    import { FORM_CONTEXT_KEY, makeFormContext } from './FormContext';

    export let context = makeFormContext();
    setContext(FORM_CONTEXT_KEY, context);
</script>

<slot/>

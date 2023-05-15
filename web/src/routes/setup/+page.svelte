<script lang="ts">
    import { getContext, setContext, onMount } from 'svelte';
    import { TabGroup, Tab } from '@skeletonlabs/skeleton';

    import * as state from '$routes/state';

    import { tbaRead } from '$lib/api';
    import { getFormContext } from '$lib/FormContext';
    import FormContextProvider from '$lib/FormContextProvider.svelte';

    import Button from '$components/Button.svelte';
    import Link from '$components/Link.svelte';

    let setupTabId = 0;

    let readApiKey = '';
    let readApiKeyValid = false;

    let eventKey = '';
    let writeAuthId = '';
    let writeAuthSecret = '';

    const selectedEventStore = getContext(state.SELECTED_EVENT_KEY);
    $: eventKey = $selectedEventStore;

    $: eventDescription = eventKey ? `event: ${eventKey}` : '';
    $: eventKeyValid = eventKey && eventKey.match(/^\d{4}[a-z]{3,7}$/) != null;

    let formContext = getFormContext();

    function saveReadApiKey() {
        formContext.withSubmit(async () => {
            await tbaRead({key: readApiKey, route: 'status'});
            localStorage.readKeyTmp = readApiKey;
            readApiKeyValid = true;
            setupTabId = 1;
        });
    }

    function saveEventKeys() {
        formContext.withSubmit(async () => {
            await new Promise((r) => setTimeout(r, 500));
            $selectedEventStore = eventKey;
        });
    }

    onMount(() => {
        // todo: move to proper storage
        if (localStorage.readKeyTmp) {
            readApiKey = localStorage.readKeyTmp;
            readApiKeyValid = true;
            setupTabId = 1;
        }
    });
</script>

<style>
input {
    font-family: monospace;
}
</style>

<TabGroup>
    <Tab bind:group={setupTabId} name="tba" value={0}>TBA API Key</Tab>
    {#if readApiKeyValid}
    <Tab bind:group={setupTabId} name="event" value={1}>Event Keys</Tab>
    {/if}
    <!-- Tab Panels --->
    <svelte:fragment slot="panel">
        <div class="space-y-3">
        {#if setupTabId === 0}
        <FormContextProvider bind:context={formContext}>
            <label class="label">
                <p>Enter a TBA API key.</p>
                <p>You can create a key at <Link href="https://www.thebluealliance.com/account"/></p>
                <input class="input" title="Input (text)" type="text" placeholder="API Key" bind:value={readApiKey} />
            </label>
            <p>
                <Button color="success" disabled={readApiKey == ''} on:click={saveReadApiKey}>Save</Button>
                <Button class="float-right" on:click={() => readApiKey = ''}>Clear</Button>
            </p>
        </FormContextProvider>
        {:else if setupTabId === 1}
        <FormContextProvider bind:context={formContext}>
            <label class="label">
                <p>Enter the event key (e.g. <code>2023cmptx</code>)</p>
                <input class="input" title="Event key" type="text" placeholder="Event key" bind:value={eventKey} />
            </label>
            <p class={eventKeyValid ? '' : 'text-error-500'}>{eventDescription}&nbsp;</p>
            <p>Enter the event write keys. You can find these at <Link href="https://www.thebluealliance.com/account"/></p>
            <label class="label">
                <span>Auth ID:</span>
                <input class="input" title="Auth ID" type="text" placeholder="Auth ID" bind:value={writeAuthId} />
            </label>
            <label class="label">
                <span>Auth secret:</span>
                <input class="input" title="Auth secret" type="text" placeholder="Auth secret" bind:value={writeAuthSecret} />
            </label>
            <p>
                <Button color="success" disabled={!eventKeyValid} on:click={saveEventKeys}>Save</Button>
                <Button class="float-right">Reset</Button>
            </p>
        </FormContextProvider>
        {/if}
        </div>
    </svelte:fragment>
</TabGroup>

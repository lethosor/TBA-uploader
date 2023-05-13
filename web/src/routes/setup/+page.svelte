<script lang="ts">
    import { TabGroup, Tab } from '@skeletonlabs/skeleton';

    import Button from '$components/Button.svelte';
    import Link from '$components/Link.svelte';

    let setupTabId = 0;

    let readApiKey = '';
    let readApiKeyValid = false;

    let eventKey = '';
    let writeAuthId = '';
    let writeAuthSecret = '';

    $: eventDescription = eventKey ? `event: ${eventKey}` : '';
    $: eventKeyValid = eventKey.match(/^\d{4}[a-z]{3,7}$/) != null;

    function saveReadApiKey() {
        setTimeout(() => {
            readApiKeyValid = true;
            setupTabId = 1;
        }, 500);
    }
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
        <!-- <div class="card p-4 w-full text-token space-y-4"> -->
            <label class="label">
                <p>Enter a TBA API key.</p>
                <p>You can create a key at <Link href="https://www.thebluealliance.com/account"/></p>
                <input class="input" title="Input (text)" type="text" placeholder="API Key" bind:value={readApiKey} />
            </label>
            <p>
                <Button color="success" disabled={readApiKey == ''} on:click={saveReadApiKey}>Save</Button>
                <Button class="float-right" on:click={() => readApiKey = ''}>Clear</Button>
            </p>
        {:else if setupTabId === 1}
            <label class="label">
                <p>Enter the event key (e.g. <code>2023cmptx</code>)</p>
                <input class="input" title="Input (text)" type="text" placeholder="Event key" bind:value={eventKey} />
            </label>
            <p class={eventKeyValid ? '' : 'text-error-500'}>{eventDescription}&nbsp;</p>
            <p>Enter the event write keys. You can find these at <Link href="https://www.thebluealliance.com/account"/></p>
            <label class="label">
                <span>Auth ID:</span>
                <input class="input" title="Input (text)" type="text" placeholder="Auth ID" bind:value={writeAuthId} />
            </label>
            <label class="label">
                <span>Auth secret:</span>
                <input class="input" title="Input (text)" type="text" placeholder="Auth secret" bind:value={writeAuthSecret} />
            </label>
            <p>
                <Button color="success" disabled={!eventKeyValid}>Save</Button>
                <Button class="float-right">Reset</Button>
            </p>
        {/if}
        </div>
    </svelte:fragment>
</TabGroup>

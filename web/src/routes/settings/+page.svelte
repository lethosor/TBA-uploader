<script lang="ts">
    import { apiSettingsStore, API_SETTINGS_DEFAULT } from '$lib/api';

    import { SlideToggle } from '@skeletonlabs/skeleton';

    import Button from '$components/Button.svelte';
</script>

<div class="card p-4 space-y-3">
    <h3 class="pb-1">Network Settings

        <Button color="error" class="float-right" on:click={() => {apiSettingsStore.set({...API_SETTINGS_DEFAULT})}}>
            Restore defaults
        </Button>

    </h3>

    <label class="label">
        <p>TBA base URL</p>
        <input class="input" title="TBA base URL" type="text" placeholder="URL" bind:value={$apiSettingsStore.tbaUrl} />
    </label>

    <div class="align-top">
        <SlideToggle class="[&>label]:items-start" name="useProxy" active="bg-primary-500" bind:checked={$apiSettingsStore.useProxy}>
            <p>Proxy external requests through the server.</p>
            <p>By default, external requests are made from the client when possible.</p>
        </SlideToggle>
    </div>
</div>

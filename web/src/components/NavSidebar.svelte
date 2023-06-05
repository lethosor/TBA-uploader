<script lang="ts">
    import { drawerStore } from '@skeletonlabs/skeleton';

    import { page } from '$app/stores';
    import { HEADER_PAGES, SIDEBAR_PAGES } from '$lib/nav';

    export let isEventSelected = false;

    function drawerClose(): void {
        drawerStore.close();
    }

    $: activePath = $page.url.pathname;
</script>

<nav class="list-nav p-4">
    <ul>
        {#each SIDEBAR_PAGES as page (page.path)}
            <li>
                {#if !page.requiresEvent || isEventSelected}
                <a href={page.path} class={page.path == activePath ? 'bg-primary-active-token' : null} on:click={drawerClose}>
                    {page.name}
                </a>
                {/if}
            </li>
        {/each}
    </ul>
    <ul class="lg:hidden">
        <br/>
        {#each HEADER_PAGES as page (page.path)}
            <li>
                <a href={page.path} class={page.path == activePath ? 'bg-primary-active-token' : null} on:click={drawerClose}>
                    {page.name}
                </a>
            </li>
        {/each}
    </ul>
</nav>

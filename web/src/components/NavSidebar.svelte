<script lang="ts">
    import { drawerStore } from '@skeletonlabs/skeleton';

    import { page } from '$app/stores';
    import { SIDEBAR_PAGES } from '$lib/nav';

    function drawerClose(): void {
        drawerStore.close();
    }

    $: activePath = $page.url.pathname;
</script>

<nav class="list-nav p-4">
    <ul>
        {#each SIDEBAR_PAGES as page (page.path)}
            <li>
                {#if !page.requiresEvent}
                <a href={page.path} class={page.path == activePath ? 'bg-primary-active-token' : null} on:click={drawerClose}>
                    {page.name}
                </a>
                {/if}
            </li>
        {/each}
    </ul>
</nav>

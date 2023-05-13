<script lang="ts">
    import { drawerStore } from '@skeletonlabs/skeleton';

    import { page } from '$app/stores';
    import { ALL_PAGES } from '$lib/nav';

    function drawerClose(): void {
        drawerStore.close();
    }

    $: activePath = $page.url.pathname;
</script>

<nav class="list-nav p-4">
    <ul>
        {#each ALL_PAGES as page}
            <li key={page.path}>
                <a href={page.path} class={page.path == activePath ? 'bg-primary-active-token' : null} on:click={drawerClose}>
                    {page.name}
                </a>
            </li>
        {/each}
    </ul>
</nav>

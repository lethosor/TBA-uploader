<script lang="ts">
	import { getContext, setContext } from 'svelte';
	import { get } from 'svelte/store';
	import {
		AppBar,
		AppShell,
		Drawer,
		drawerStore,
		LightSwitch,
		localStorageStore,
	} from '@skeletonlabs/skeleton';

	import { page } from '$app/stores';
	import { apiSettingsStore, makeTbaApiClient } from '$lib/api';
	import { pageNameFromPath, HEADER_PAGES } from '$lib/nav';

	import Link from '$components/Link.svelte';
	import NavSidebar from '$components/NavSidebar.svelte';

	import * as state from '$routes/state';

	// from https://www.skeleton.dev/docs/get-started (manual)
	import '../theme.css';
	import '@skeletonlabs/skeleton/styles/all.css';
	import '../app.postcss';

	function drawerOpen(): void {
		drawerStore.open({});
	}

	const selectedEventStore = localStorageStore('selectedEvent', '');
	setContext(state.SELECTED_EVENT_KEY, selectedEventStore);
	$: selectedEventKey = $selectedEventStore;

	$: setContext(state.TBA_CLIENT_KEY, makeTbaApiClient({
		event: selectedEventKey,
		readKey: '',
		authId: '',
		authSecret: '',
	}));
</script>

<!-- Drawer -->
<Drawer width="auto">
	<h2 class="p-4">Navigation</h2>
	<hr />
	<NavSidebar />
</Drawer>

<!-- App Shell -->
<AppShell slotSidebarLeft="bg-surface-500/5 w-0 lg:w-64">
	<svelte:fragment slot="header">
		<!-- App Bar -->
		<AppBar>
			<svelte:fragment slot="lead">
				<div class="flex items-center">
					<button class="lg:hidden btn btn-sm mr-4" on:click={drawerOpen}>
						<span>
							<svg viewBox="0 0 100 80" class="fill-token w-4 h-4">
								<rect width="100" height="20" />
								<rect y="30" width="100" height="20" />
								<rect y="60" width="100" height="20" />
							</svg>
						</span>
					</button>
					<strong class="text-xl">
						<a href="/">TBA Uploader</a>
					</strong>
					{#if selectedEventKey}
					&nbsp;(<Link href={$apiSettingsStore.tbaUrl + '/event/' + selectedEventKey}>{selectedEventKey}</Link>)
					{/if}
				</div>
			</svelte:fragment>
			<svelte:fragment slot="trail">
				<LightSwitch />
				{#each HEADER_PAGES as page (page.path)}
				<a class="btn btn-sm" href={page.path}>{page.name}</a>
				{/each}
			</svelte:fragment>
		</AppBar>
	</svelte:fragment>
	<!-- Left Sidebar Slot -->
	<svelte:fragment slot="sidebarLeft">
		<NavSidebar />
	</svelte:fragment>
	<!-- Page Route Content -->
	<div class="container p-10 space-y-4">
		<h2>{pageNameFromPath($page.url.pathname)}</h2>
		<hr />
		<slot />
	</div>
</AppShell>

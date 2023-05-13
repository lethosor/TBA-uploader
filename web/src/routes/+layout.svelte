<script lang="ts">
	import { AppShell, AppBar, Drawer, drawerStore, LightSwitch } from '@skeletonlabs/skeleton';

	import { page } from '$app/stores';
	import { pageNameFromPath } from '$lib/nav';

	import NavSidebar from '$components/NavSidebar.svelte';

	// from https://www.skeleton.dev/docs/get-started (manual)
	import '../theme.css';
	import '@skeletonlabs/skeleton/styles/all.css';
	import '../app.postcss';

	function drawerOpen(): void {
		drawerStore.open({});
	}
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
					<strong class="text-xl"><a href="/">TBA Uploader</a></strong>
				</div>
			</svelte:fragment>
			<svelte:fragment slot="trail">
				<LightSwitch />
				<a class="btn btn-sm" href="/settings">Settings</a>
				<a class="btn btn-sm" href="/help">Help</a>
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

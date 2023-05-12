<script>
	import { getApiBaseUrl } from '$lib/api';
</script>

<svelte:head>
	<title>Home</title>
</svelte:head>

<section>
	<p>
		API URL: <code>{getApiBaseUrl()}</code>
	</p>
</section>

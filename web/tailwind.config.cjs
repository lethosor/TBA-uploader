/** @type {import('tailwindcss').Config} */
const config = {
	darkMode: 'class',
	content: [
		'./src/**/*.{html,js,svelte,ts}',
		// from https://www.skeleton.dev/docs/get-started (manual)
		require('path').join(require.resolve(
			'@skeletonlabs/skeleton'),
			'../**/*.{html,js,svelte,ts}'
		)
	],
	theme: {
		extend: {},
	},
	plugins: [
		...require('@skeletonlabs/skeleton/tailwind/skeleton.cjs')()
	]
};

module.exports = config;

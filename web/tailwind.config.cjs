/** @type {import('tailwindcss').Config} */
const config = {
	darkMode: 'class',
	content: [
		'./src/**/*.{html,js,svelte,ts}',
		// from https://www.skeleton.dev/docs/get-started (manual)
		require('path').join(require.resolve('@skeletonlabs/skeleton'), '../**/*.{html,js,svelte,ts}'),
	],
	safelist: [
		'variant-filled',
		'variant-filled-error',
		'variant-filled-primary',
		'variant-filled-secondary',
		'variant-filled-success',
		'variant-filled-warning',
	],
	theme: {
		extend: {},
	},
	plugins: [
		require('@tailwindcss/forms'),
		...require('@skeletonlabs/skeleton/tailwind/skeleton.cjs')()
	]
};

module.exports = config;

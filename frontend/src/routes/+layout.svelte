<script>
	import favicon from '$lib/assets/favicon.svg';
	import '$lib/global.css';

	let { children } = $props();
	import Nav from '$lib/Nav.svelte';
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

<Nav />

{@render children?.()}


<script>
	import favicon from '$lib/assets/favicon.svg';
	import '$lib/global.css';

	let { children } = $props();
	import Nav from '$lib/Nav.svelte';
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<link href="https://fonts.googleapis.com/css2?family=Cinzel+Decorative:wght@700&family=IM+Fell+English:ital@0;1&display=swap" rel="stylesheet">

</svelte:head>

<Nav />

{@render children?.()}


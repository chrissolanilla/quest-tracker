<script>
	//get the user somehow
	import { onMount } from 'svelte';
	import { getMe, logout } from '$lib/api';

	let me = null;
	console.log(me);
	let error = '';

	onMount(async () => {
		try {
			me = await getMe();
			console.log("me is now", me);
		} catch (e) {
			error = e.message || 'not logged in';
		}
	});

	async function doLogout() {
		await logout();
		me = null;
	}

	export function setMeNull(){
		me = null
	}

</script>

<nav>
	<a href="/">Home</a>
	<a href="/leaderboard">Leaderboard</a>
	{#if me}
		<a href="/profile">Profile</a>
		<a on:click={ doLogout() } href="/">Logout</a>
	{:else}
		<a href="/login">Login</a>
	{/if}
</nav>


<style style="scss">
	nav {
		font-size: 1.5rem;
		padding: 1rem;
		gap: 1rem;
		justify-content: space-between;
		display: flex;
		gap: 1rem;

		a {
			color: #ff7cc4;
		}
		a:visited {
			color: #ff7cc4;
		}
		background: black;
	}
</style>

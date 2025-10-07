<script>
	import { onMount } from 'svelte';
	import { getMe, listProjects, listProjectTasks } from '$lib/api';

	let me = null;
	let error = '';

	onMount(async () => {
		try {
			me = await getMe();
			const projects = await listProjects();
			console.log("projects are ", projects);
		} catch (e) {
			error = e.message || 'not logged in';
		}
	});

	// async function doLogout() {
	// 	await logout();
	// 	me = null;
	// 	Nav.setMeNull();
	// 	error = 'logged out';
	// }


</script>

<div class="container">
	<h1>Profile</h1>

	{#if me}
		<!--  i guess put their profile picture-->
		<img src={me.picture} />
		<p>Hi {me.name}! Asana GID: {me.user_id}</p>
		<!-- <button on:click={doLogout}>Log out</button> -->
	{:else if error}
		<p>{error}. <a href="/login">Log in</a></p>
	{:else}
		<p>Loading…</p>
	{/if}

</div>

<style lang="scss">
	.container {
		display: flex;
		flex-direction: column;
		justify-content: center;
		text-align: center;
	}
</style>


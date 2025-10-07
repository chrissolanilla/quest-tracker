<script>
	//make some api call that gets the first 10 users based on point system
	import { onMount } from 'svelte';
	import { getLeaderboard } from '$lib/api';

	let rows = [];
	onMount(async () => {
		rows = await getLeaderboard();
		console.log("we have ", rows);
	});
</script>

<h1>Leaderboard</h1>

<div>
	{#each rows as row, i}
		<div class="row">
			<strong>{i + 1}.</strong>
			<p> {row.name} : {row.points}</p>
		</div>
	{/each}
</div>


<style lang="scss">
	.row {
		display: flex;
		flex-direction: row;

		p {
			margin: 0;
		}
	}


	h1 {
		color: lightgreen;
	}
</style>

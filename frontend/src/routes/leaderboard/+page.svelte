<script>
	//make some api call that gets the first 10 users based on point system
	import { onMount } from 'svelte';
	import { getLeaderboard, syncMe } from '$lib/api';

	let rows = [];
	onMount(async () => {
		await syncMe();
		rows = await getLeaderboard();
		console.log("we have ", rows);
	});
</script>

<div class="container">
<h1>Leaderboard</h1>


		<table>
		{#each rows as row, i}
			<tr class="row">
			<th>
				<strong>{i + 1}.</strong>
			</th>
			<td>
				<p> {row.name} : {row.points}</p>
			</td>
			</tr>
		{/each}
		</table>

</div>


<style lang="scss">

	:root{
		--wood:#6e4a2f;
		--wood-dark:#573922;
		--gold:#f0d772;
		--parchment:#f6edd8;
		--parchment-edge:#e3d6b7;
	}

	.row {
		display: flex;
		flex-direction: row;

		p {
			margin: 0;
		}
	}

	.container {
		text-align: center;
		display: flex;
		flex-direction: column;
		justify-content: center;
		align-items: center;
	}

	table {
		background: var(--wood);
		}


	h1 {
		color: lightgreen;
	}
</style>

<script>
	import { onMount } from 'svelte';
	import { getMe, listProjects, listProjectTasks } from '$lib/api';
	import QuestCard from '$lib/QuestCard.svelte';

	let me = null;
	let error = '';
	let rank = null

	let projecttasks = [];
	let incompletetasks = [];
	let completedtasks = [];

	onMount(async () => {
		try {
			me = await getMe();
			const projects = await listProjects();
			console.log("projects are ", projects);
			console.log("projects[0].gid is ", projects[0].gid);
			projecttasks = await listProjectTasks(projects[0].gid);
			console.log("tasks are ", projecttasks);
			incompletetasks = getincompleteforme(projecttasks);
			console.log("incomplete tasks are ", incompletetasks);
			completedtasks = getcompletedforme(projecttasks);
			rank = getrank(completedtasks.length)
			console.log("rank is ", rank);



		} catch (e) {
			error = e.message || 'not logged in';
		}
	});

	function getincompleteforme(tasksarr) {
		return tasksarr.filter(t =>
			t &&
			t.completed === false &&
			(t.assignee?.gid === me.user_id)
		);
	}

	function getcompletedforme(tasksarr) {
		return tasksarr.filter(t =>
			t &&
			t.completed === true &&
			(t.assignee?.gid === me.user_id)
		);
	}

	function getrank(completedNum) {
		console.log("completedNum is ", completedNum);
		const ranks = {
			0: "Village Greenhorn 🪵",
			3: "Squire 🛡️",
			6: "Wanderer 🥾",
			10: "Dungeon Delver 🕯️",
			15: "Royal Knight 👑",
			20: "Hero ⚔️",
			30: "Legend 🌟",
			50: "Force of Nature 🌪️",
		};

		const thresholds = Object.keys(ranks)
			.map(Number)
			.sort((a, b) => a - b);

		let current = thresholds[0];
		for (const t of thresholds) {
			if (completedNum >= t) current = t;
		}

		const idx = thresholds.indexOf(current);
		const next = thresholds[idx + 1];

		return {
			name: ranks[current],
			currentLevel: current,
			nextLevel: next ?? null,
		};
	}


</script>

<div class="container">
	<h1>profile</h1>


	{#if me}
		<!-- <img src={me.photo.image_21x21} alt="profile picture" /> -->
		{#if rank}
		<h3>greetings, travelor {me.name}!</h3>
		<h4>rank: {rank.name}</h4>
		{/if}
		<!-- maybe show their rank or something -->


		<QuestCard title="incomplete quests">
			{#if incompletetasks.length === 0}
				<p>no incomplete quests. nice!</p>
			{:else}
				{#each incompletetasks as task}
					<div class="scroll">
						<img src="/scroll.png" alt="scroll" />
						<p>{task.name}</p>
					</div>
				{/each}
			{/if}
		</QuestCard>

		<QuestCard title="completed quests">
			{#if completedtasks.length === 0}
				<p>nothing here yet.</p>
			{:else}
				{#each completedtasks as task}
					<div class="scroll">
						<img src="/scroll.png" alt="scroll" />
						<p>{task.name}</p>
					</div>
				{/each}
			{/if}
		</QuestCard>
	{/if}
</div>

<style >
	.container {
		display: flex;
		flex-direction: column;
		justify-content: center;
		text-align: center;
	}

	.scroll {
		position: relative;
		/* width: min(680px, 90vw); */
		aspect-ratio: 3 / 4;
		display: grid;
		place-items: center;
		padding: clamp(16px, 3vw, 28px);

		max-height:8rem;

		p{
			z-index: 1;
		}

		img {
			position: absolute;
			inset: 0;
			width: 100%;
			height: 100%;
			object-fit: contain;
			pointer-events: none;
			z-index: 0;
		}

	}

	h3 , h4{
		color: #fff;
	}

	section {
		background-image: url('/questhang.svg');
		min-height: 400px;
		background-size: contain;
		background-repeat: no-repeat;
		background-position: center;
	}

	p {
		color: black;
	}
</style>


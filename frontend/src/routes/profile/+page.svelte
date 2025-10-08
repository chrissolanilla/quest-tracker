<script>
	import { onMount } from 'svelte';
	import { getMe, listProjects, listProjectTasks } from '$lib/api';

	let me = null;
	let error = '';

	let projectTasks = [];
	let incompleteTasks = [];
	let completedTasks = [];

	onMount(async () => {
		try {
			me = await getMe();
			const projects = await listProjects();
			console.log("projects are ", projects);
			projectTasks = await listProjectTasks(projects[0].gid);
			console.log("tasks are ", projectTasks);
			incompleteTasks = getIncompleteForMe(projectTasks);
			console.log("incomplete tasks are ", incompleteTasks);
			completedTasks = getCompletedForMe(projectTasks);



		} catch (e) {
			error = e.message || 'not logged in';
		}
	});

	function getIncompleteForMe(tasksArr) {
		return tasksArr.filter(t =>
			t &&
			t.completed === false &&
			(t.assignee?.gid === me.user_id)
		);
	}

	function getCompletedForMe(tasksArr) {
		return tasksArr.filter(t =>
			t &&
			t.completed === true &&
			(t.assignee?.gid === me.user_id)
		);
	}


</script>

<div class="container">
	<h1>Profile</h1>

	{#if me}
		<!--  i guess put their profile picture-->
		<img src={me.picture} />
		<h3>Greetings, travelor {me.name}!</h3>
		<!-- {me.user_id} -->
		<!-- <button on:click={doLogout}>Log out</button> -->
		<div class="incompleteQuests" >
			<section>
			<h2>Incomplete Quests</h2>
				{#each incompleteTasks as task}
					<p>task name: {task.name}</p>
				{/each}
			</section>
		</div>

		<div class="completeQuests">
			<section>
				<h2>Completed Quests</h2>
				{#each completedTasks as task}
					<p>task name: {task.name}</p>
				{/each}
			</section>
		</div>
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

	.incompleteQuests {
		display: flex;
		flex-direction: column;
		color: #fff;
		// text-align: left;
	}

	h3 {
		color: #fff;
	}
</style>


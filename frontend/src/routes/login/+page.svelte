<script>
	const API = import.meta.env.VITE_API_BASE ?? 'http://localhost:8080';
	function signIn() {
		window.location.href = `api/auth/asana/start`;
	}
</script>


<div class="container">
	<h1>Log in</h1>
	<button on:click={signIn}>Sign in with Asana</button>
</div>

<style lang="scss">
	.container {
		display: flex;
		flex-direction: column;
		justify-content: center;
		align-items: center;

		button {
			max-width: 5rem;
		}

		text-align: center;
	}
</style>


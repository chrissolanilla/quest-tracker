export const API_BASE = import.meta.env.VITE_API_BASE ?? 'http://localhost:8080';

export async function getLeaderboard(fetchFn = fetch) {
  const res = await fetchFn(`/api/leaderboard`, { credentials: 'include' });
  if (!res.ok) throw new Error('failed to load leaderboard');
  return res.json();
}

export async function getMe(fetchFn = fetch) {
	const res = await fetchFn(`/api/me`, { credentials: 'include' });
	if (!res.ok) throw new Error('not logged in');
	return res.json();
}

export async function logout() {
	await fetch(`/api/auth/logout`, { method: 'POST', credentials: 'include' });
}

export async function listProjects(fetchFn = fetch) {
	const res = await fetchFn(`/api/asana/projects`, { credentials:'include' });
	if (!res.ok) throw new Error('failed');
	return res.json();
}

export async function listProjectTasks(projectGid, fetchFn = fetch) {
	const res = await fetchFn(`/api/asana/projects/${projectGid}/tasks`, { credentials:'include' });
	if (!res.ok) throw new Error('failed');
	//array of tasks with custom_fields
	return res.json();
}

export async function syncMe(fetchFn = fetch) {
	const res = await fetchFn(`/api/asana/sync/me`, { method:'POST', credentials:'include' });
	if (!res.ok) throw new Error('sync failed');
}



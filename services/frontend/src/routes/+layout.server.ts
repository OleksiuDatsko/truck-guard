import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ locals, cookies }) => {
	let user = locals.user;
	const token = cookies.get('session');	

	if (user && token) {
		try {
			const coreUser = await locals.coreClient.getUser(user.id);
			if (coreUser) {
				user = {
					...user,
					...coreUser
				};
			} 
		} catch (e) {
			console.error('Failed to fetch core user', e);
		}
	}

	return {
		user
	};
};
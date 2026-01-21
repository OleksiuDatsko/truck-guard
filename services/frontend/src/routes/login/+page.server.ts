import { fail, redirect } from '@sveltejs/kit';
import type { Actions } from './$types';

export const actions: Actions = {
    default: async ({ request, cookies, locals }) => {
        const data = await request.formData();
        const username = data.get('username') as string;
        const password = data.get('password') as string;

        if (!username || !password) {
            return fail(400, { message: 'Username and password are required' });
        }

        const result = await locals.authClient.login(username, password);

        if (!result) {
            console.error('AuthClient.login failed', {result, username, password});
            return fail(401, { message: 'Invalid username or password', status: 401 });
        }

        cookies.set('session', result.token, {
            path: '/',
            httpOnly: true,
            sameSite: 'strict',
            maxAge: 60 * 60 * 24
        });
        console.log('Login successful', {result, username, password});

        throw redirect(303, '/');
    }
};
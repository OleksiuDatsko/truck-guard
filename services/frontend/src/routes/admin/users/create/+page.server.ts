import { fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { can } from '$lib/auth';

export const load: PageServerLoad = async ({ locals, cookies }) => {
    // Permission check
    if (!can(locals.user, 'create:users')) {
        throw redirect(303, '/admin/users?error=You do not have permission to create users');
    }

    const roles = await locals.authClient.getRoles();

    return {
        roles
    };
};

export const actions: Actions = {
    default: async ({ request, locals, cookies }) => {
        // Permission check
        if (!can(locals.user, 'create:users')) {
            return fail(403, { error: 'You do not have permission to create users' });
        }

        const data = await request.formData();
        const username = data.get('username') as string;
        const password = data.get('password') as string;
        const role = data.get('role') as string;
        const firstName = data.get('first_name') as string;
        const lastName = data.get('last_name') as string;
        const thirdName = data.get('third_name') as string;
        const phone = data.get('phone_number') as string;
        const email = data.get('email') as string;
        const notes = data.get('notes') as string;

        if (!username || !password || !role) {
             return fail(400, { error: 'Username, password and role are required' });
        }

        const result = await locals.coreClient.createUser({
            username,
            password,
            role,
            first_name: firstName,
            last_name: lastName,
            third_name: thirdName,
            phone_number: phone,
            email,
            notes
        });

        if (!result) {
            return fail(500, { error: 'Failed to create user. It might already exist.' });
        }

        throw redirect(303, '/admin/users');
    }
};

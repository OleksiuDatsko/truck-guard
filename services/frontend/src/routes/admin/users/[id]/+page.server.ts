import { fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { AuthClient } from '$lib/server/auth-client';
import { CoreClient } from '$lib/server/core-client';

export const load: PageServerLoad = async ({ params, locals, cookies }) => {
     if (!locals.user?.permissions?.includes('update:users')) {
        throw redirect(303, '/admin/users');
    }

    const { id } = params;

    const roles = await locals.authClient.getRoles();

    const allAuthUsers = await locals.authClient.listUsers();
    const authUser = allAuthUsers.find((u: { id: string; }) => u.id == id);

    if (!authUser) {
        throw new Error('User not found in Auth service');
    }

    const coreUser = await locals.coreClient.getUser(id);

    return {
        user: {
            ...authUser,
            profile: coreUser || {}
        },
        roles
    };
};

export const actions: Actions = {
    default: async ({ request, params, locals, cookies }) => {
        if (!locals.user?.permissions?.includes('update:users')) {
            return fail(403, { error: 'You do not have permission to update users' });
        }

        const { id } = params;
        const data = await request.formData();
        
        const roleId = data.get('role_id');
        const firstName = data.get('first_name') as string;
        const lastName = data.get('last_name') as string;
        const thirdName = data.get('third_name') as string;
        const phone = data.get('phone_number') as string;
        const email = data.get('email') as string;
        const notes = data.get('notes') as string;

        if (roleId) {
             const success = await locals.authClient.updateUserRole(id, Number(roleId));
             if (!success) {
                 return fail(500, { error: 'Failed to update user role' });
             }
        }

        const result = await locals.coreClient.updateUser(id, {
            first_name: firstName,
            last_name: lastName,
            third_name: thirdName,
            phone_number: phone,
            email,
            notes
        });

        if (!result) {
            return fail(500, { error: 'Failed to update user profile' });
        }

        throw redirect(303, '/admin/users');
    }
};

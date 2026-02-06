import type { PageServerLoad, Actions } from './$types';
import { error, fail } from '@sveltejs/kit';
import { can } from '$lib/auth';

export const load: PageServerLoad = async ({ locals }) => {
    const authUsers = await locals.authClient.listUsers();
    if (!authUsers) {
        throw error(500, 'Failed to fetch users from Auth Service');
    }

    const coreUsers = await locals.coreClient.listUsers();
    if (!coreUsers) {
         console.error("Failed to fetch Core users");
    }

    const users = authUsers.map((authUser: any) => {
        const profile = coreUsers?.find((cu: any) => cu.auth_id === authUser.id);
        return {
            ...authUser,
            profile: profile || null
        };
    });

    return {
        users
    };
};

export const actions: Actions = {
    delete: async ({ request, locals }) => {
        if (!can(locals.user, 'delete:users')) {
            return fail(403, { error: 'You do not have permission to delete users' });
        }

        const data = await request.formData();
        const id = data.get('id') as string;

        console.log("Deleting user with id: ", id, data)
        if (!id) {
            return fail(400, { error: 'User ID is required' });
        }

        const success = await locals.coreClient.deleteUser(id);

        if (!success) {
            return fail(500, { error: 'Failed to delete user' });
        }

        return { success: true };
    }
};

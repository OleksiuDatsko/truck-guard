import type { PageServerLoad, Actions } from './$types';
import { fail } from '@sveltejs/kit';

export const load: PageServerLoad = async ({ locals }) => {
    if (!locals.user) {
        return { keys: [], permissions: [] };
    }

    try {
        const [keys, permissions] = await Promise.all([
            locals.authClient.getKeys(),
            locals.authClient.getPermissions()
        ]);

        return {
            keys,
            permissions,
            user: locals.user
        };
    } catch (e) {
        console.error('Failed to load keys data:', e);
        return { keys: [], permissions: [], error: 'Failed to load data' };
    }
};

export const actions: Actions = {
    create: async ({ request, locals }) => {
        const formData = await request.formData();
        const name = formData.get('name') as string;
        const permissions = formData.getAll('permissions') as string[];

        if (!name) {
            return fail(400, { error: 'Name is required' });
        }

        try {
            const result = await locals.authClient.createKey(name, permissions);
            return { success: true, newKey: result };
        } catch (e) {
            console.error('Failed to create key:', e);
            return fail(500, { error: 'Failed to create key' });
        }
    },

    update: async ({ request, locals }) => {
        const formData = await request.formData();
        const idStr = formData.get('id') as string;
        const ownerName = formData.get('owner_name') as string;
        const isActiveStr = formData.get('is_active') as string;

        if (!idStr || !ownerName) {
            return fail(400, { error: 'ID and Owner Name are required' });
        }

        const id = parseInt(idStr);
        const isActive = isActiveStr === 'true';

        try {
            await locals.authClient.updateKey(id, ownerName, isActive);
            return { success: true };
        } catch (e) {
            console.error('Failed to update key:', e);
            return fail(500, { error: 'Failed to update key' });
        }
    },

    delete: async ({ request, locals }) => {
        const formData = await request.formData();
        const idStr = formData.get('id') as string;

        if (!idStr) {
            return fail(400, { error: 'ID is required' });
        }

        const id = parseInt(idStr);

        try {
            await locals.authClient.deleteKey(id);
            return { success: true };
        } catch (e) {
            console.error('Failed to delete key:', e);
            return fail(500, { error: 'Failed to delete key' });
        }
    },

    assignPermissions: async ({ request, locals }) => {
        const formData = await request.formData();
        const idStr = formData.get('id') as string;
        const permissions = formData.getAll('permissions') as string[];

        if (!idStr) {
            return fail(400, { error: 'ID is required' });
        }

        const id = parseInt(idStr);

        try {
            await locals.authClient.assignKeyPermissions(id, permissions);
            return { success: true };
        } catch (e) {
            console.error('Failed to assign permissions to key:', e);
            return fail(500, { error: 'Failed to assign permissions' });
        }
    }
};

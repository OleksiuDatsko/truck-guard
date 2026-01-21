import type { PageServerLoad, Actions } from './$types';
import { error, fail } from '@sveltejs/kit';

import { can } from '$lib/auth';

export const load: PageServerLoad = async ({ locals }) => {
    if (!locals.user) {
        throw error(401, 'Unauthorized');
    }

    if (!can(locals.user, 'read:roles')) {
        throw error(403, 'Forbidden');
    }

    try {
        const [roles, permissions] = await Promise.all([
            locals.authClient.getRoles(),
            locals.authClient.getPermissions()
        ]);

        return {
            roles,
            permissions
        };
    } catch (e) {
        console.error('Error fetching roles/permissions:', e);
        throw error(500, 'Failed to load roles data');
    }
};

export const actions: Actions = {
    create: async ({ request, locals }) => {
        if (!can(locals.user, 'create:roles')) return fail(403, { error: 'Forbidden' });

        const data = await request.formData();
        const name = data.get('name') as string;
        const description = data.get('description') as string;

        if (!name) return fail(400, { error: 'Name is required' });

        try {
            await locals.authClient.createRole(name, description);
            return { success: true };
        } catch (e) {
            console.error('Create role error:', e);
            return fail(500, { error: 'Failed to create role' });
        }
    },

    update: async ({ request, locals }) => {
        if (!can(locals.user, 'update:roles')) return fail(403, { error: 'Forbidden' });

        const data = await request.formData();
        const id = Number(data.get('id'));
        const name = data.get('name') as string;
        const description = data.get('description') as string;

        if (!id || !name) return fail(400, { error: 'ID and Name required' });

        try {
            await locals.authClient.updateRole(id, name, description);
            return { success: true };
        } catch (e) {
            return fail(500, { error: 'Failed to update role' });
        }
    },

    delete: async ({ request, locals }) => {
        if (!can(locals.user, 'delete:roles')) return fail(403, { error: 'Forbidden' });

        const data = await request.formData();
        const id = Number(data.get('id'));

        if (!id) return fail(400, { error: 'ID required' });

        try {
            await locals.authClient.deleteRole(id);
            return { success: true };
        } catch (e) {
            return fail(500, { error: 'Failed to delete role' });
        }
    },

    assignPermissions: async ({ request, locals }) => {
        if (!can(locals.user, 'update:roles')) return fail(403, { error: 'Forbidden' });

        const data = await request.formData();
        const id = Number(data.get('id'));
        const permissions = data.getAll('permissions') as string[];

        if (!id) return fail(400, { error: 'ID required' });

        try {
            await locals.authClient.assignPermissions(id, permissions);
            return { success: true };
        } catch (e) {
            return fail(500, { error: 'Failed to assign permissions' });
        }
    }
};

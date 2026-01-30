import type { Actions, PageServerLoad } from './$types';
import { error, fail } from '@sveltejs/kit';

export const load: PageServerLoad = async ({ params, locals }) => {
    if (!locals.coreClient) {
        throw error(401, 'Unauthorized');
    }

    const { id } = params;
    try {
        const event = await locals.coreClient.getPlateEvent(id);
        if (!event) {
             throw error(404, 'Event not found');
        }
        return { event };
    } catch (e) {
        console.error('Failed to load plate event:', e);
        throw error(500, 'Failed to load event details');
    }
};

export const actions: Actions = {
    correct: async ({ request, params, locals }) => {
        if (!locals.coreClient || !locals.user) {
            return fail(401, { error: 'Unauthorized' });
        }

        if (!locals.user.permissions.includes('update:events')) {
            return fail(403, { error: 'Forbidden: Missing update:events permission' });
        }

        const data = await request.formData();
        const plate = data.get('plate');
        const { id } = params;

        if (!plate || typeof plate !== 'string') {
            return fail(400, { error: 'Invalid plate number' });
        }

        try {
            await locals.coreClient.correctPlate(id, plate);
            return { success: true };
        } catch (e) {
            console.error('Failed to correct plate:', e);
            return fail(500, { error: 'Failed to correct plate' });
        }
    }
};

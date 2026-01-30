import type { PageServerLoad } from './$types';
import type { SystemEvent } from '$lib/types/events';

export const load: PageServerLoad = async ({ params, locals }) => {
    const { id } = params;
    let event: SystemEvent | null = null;

    if (locals.coreClient) {
        try {
            event = await locals.coreClient.getSystemEvent<SystemEvent>(id);
        } catch (e) {
            console.error('Failed to fetch system event:', e);
        }
    }

    return {
        event
    };
};

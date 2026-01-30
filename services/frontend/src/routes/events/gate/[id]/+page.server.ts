import type { PageServerLoad } from './$types';
import type { GateEvent } from '$lib/types/events';

export const load: PageServerLoad = async ({ params, locals }) => {
    const { id } = params;
    let event: GateEvent | null = null;

    if (locals.coreClient) {
        try {
            event = await locals.coreClient.getGateEvent<GateEvent>(id);
        } catch (e) {
            console.error('Failed to fetch gate event:', e);
        }
    }

    return {
        event
    };
};

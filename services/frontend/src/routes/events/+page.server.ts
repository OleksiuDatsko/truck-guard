import type { PageServerLoad } from './$types';
import type { ApiResponse } from '$lib/types/events';

export const load: PageServerLoad = async ({ url, locals }) => {
    const tab = url.searchParams.get('tab') || 'plate';
    const page = Number(url.searchParams.get('page')) || 1;
    const limit = Number(url.searchParams.get('limit')) || 9;

    let events: ApiResponse<any> = {
        data: [],
        metadata: {
            total_items: 0,
            total_pages: 0,
            current_page: page,
            limit: limit
        }
    };

    const filters: Record<string, string | undefined> = {
        plate: url.searchParams.get('plate') || undefined,
        from: url.searchParams.get('from') || undefined,
        to: url.searchParams.get('to') || undefined,
        type: url.searchParams.get('type') || undefined,
        gate: url.searchParams.get('gate') || undefined,
    };

    if (locals.coreClient) {
        try {
            const res = await locals.coreClient.getEvents<ApiResponse<any>>(tab, page, limit, filters);
            if (res) {
                events = res;
            }
        } catch (e) {
            console.error('Failed to fetch events:', e);
        }
    }

    return {
        events,
        tab,
        page,
        limit
    };
};

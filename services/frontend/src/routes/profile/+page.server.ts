import type { Actions, PageServerLoad } from './$types';
import { fail } from '@sveltejs/kit';

export const load: PageServerLoad = async ({ locals }) => {
    const profile = await locals.coreClient.getMyProfile();
    return {
        profile: profile || {}
    };
};

export const actions: Actions = {
    default: async ({ request, locals }) => {
        const data = await request.formData();
        
        const firstName = data.get('first_name') as string;
        const lastName = data.get('last_name') as string;
        const thirdName = data.get('third_name') as string;
        const phone = data.get('phone_number') as string;
        const email = data.get('email') as string;
        const notes = data.get('notes') as string;

        const updatedProfile = await locals.coreClient.updateMyProfile({
            first_name: firstName,
            last_name: lastName,
            third_name: thirdName,
            phone_number: phone,
            email,
            notes
        });

        if (!updatedProfile) {
            return fail(500, { error: 'Failed to update profile' });
        }

        return { success: true };
    }
};

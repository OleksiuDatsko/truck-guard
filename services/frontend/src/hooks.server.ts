import type { Handle } from '@sveltejs/kit';
import { authClient } from '$lib/server/auth-client';
import { redirect } from '@sveltejs/kit';
import { CoreClient } from '$lib/server/core-client';

export const handle: Handle = async ({ event, resolve }) => {
    const session = event.cookies.get('session');

    // Default to null user
    event.locals.user = null;

    if (session) {
        // Validate token with Auth Service
        const user = await authClient.validate(session);
        if (user) {
            event.locals.user = user;
            event.locals.coreClient = new CoreClient(session)
        } else {
            // Invalid session, clear cookie
            event.cookies.delete('session', { path: '/' });
        }
    }

    // Protected routes logic
    if (!event.locals.user) {
        // If not logged in and not on login page, redirect to login
        if (!event.url.pathname.startsWith('/login') && !event.url.pathname.startsWith('/api')) {
             throw redirect(303, '/login');
        }
    } else {
        // If logged in and on login page, redirect to dashboard or home
        if (event.url.pathname === '/login') {
            throw redirect(303, '/');
        }
    }

    return await resolve(event);
};
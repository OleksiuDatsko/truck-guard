import type { Handle } from '@sveltejs/kit';
import { AuthClient } from '$lib/server/auth-client';
import { redirect } from '@sveltejs/kit';
import { CoreClient } from '$lib/server/core-client';

export const handle: Handle = async ({ event, resolve }) => {
    const session = event.cookies.get('session');

    // Init auth client with session (or undefined)
    
    // Default to null user
    event.locals.user = null;
    
    if (session) {
        // Validate token with Auth Service
        event.locals.authClient = new AuthClient(session);
        const user = await event.locals.authClient.validate();
        console.log('User:', user)
        console.log('Session:', session)
        if (user) {
            event.locals.user = user;
            event.locals.coreClient = new CoreClient(session)
        } else {
            // Invalid session, clear cookie
            console.log('Invalid session, clearing cookie');
            event.cookies.delete('session', { path: '/' });
        }
    } else {
        event.locals.authClient = new AuthClient();
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
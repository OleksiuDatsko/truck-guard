// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
	namespace App {
		// interface Error {}
		interface Locals {
			user: {
				id: string; // Auth ID
				username: string;
				permissions: string[];
				// Core Profile Data
				first_name?: string;
				last_name?: string;
				third_name?: string
				email?: string;
				phone_number?: string;
				notes?: string;
            } | null;
			coreClient;
		}
		// interface PageData {}
		// interface PageState {}
		// interface Platform {}
	}
}

export {};

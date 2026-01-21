import type { CoreClient, User } from "./server/core-client";
import type { AuthClient } from "./server/auth-client";

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
            } & User | null;
			authClient: AuthClient;
			coreClient: CoreClient;
		}
		// interface PageData {}
		// interface PageState {}
		// interface Platform {}
	}
}

export {};

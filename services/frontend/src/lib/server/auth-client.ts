import { jwtDecode } from "jwt-decode";

export interface UserProfile {
    id: string;
    username: string;
    role: string;
    permissions: string[];
}

export class AuthClient {
    private baseUrl: string;
    private cache: Map<string, { user: UserProfile, expiresAt: number }>;
    private cacheDuration = 60 * 1000;

    constructor(baseUrl: string = 'http://gateway/auth') {
        this.baseUrl = baseUrl;
        this.cache = new Map();
    }

    async login(username: string, password: string): Promise<{ token: string } | null> {
        try {
            const response = await fetch(`${this.baseUrl}/login`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username, password })
            });

            if (!response.ok) {
                return null;
            }

            return await response.json();
        } catch (error) {
            console.error('AuthClient.login error:', error);
            return null;
        }
    }

    async validate(token: string): Promise<UserProfile | null> {
        try {
            const now = Date.now();
            const cached = this.cache.get(token);

            if (cached && cached.expiresAt > now) {
                return cached.user;
            }

            const response = await fetch(`${this.baseUrl}/validate`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });

            if (!response.ok) {
                this.cache.delete(token);
                return null;
            }

            const permissionsHeader = response.headers.get('X-Permissions');
            const permissions = permissionsHeader ? permissionsHeader.split(',') : [];

            // Decode token to get basic info
            const decoded: any = jwtDecode(token);
            
            const user = {
                id: decoded.sub || decoded.user_id || '0',
                username: decoded.username || decoded.sub || 'unknown',
                role: decoded.role || 'user',
                permissions: permissions
            };

            this.cache.set(token, {
                user,
                expiresAt: now + this.cacheDuration
            });

            return user;
        } catch (error) {
            console.error('AuthClient.validate error:', error);
            return null;
        }
    }
}

export const authClient = new AuthClient();

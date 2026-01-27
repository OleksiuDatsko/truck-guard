export interface CoreUser {
    ID: number;
    auth_id: string; // Змінено на string, оскільки в методах використовується authId: string
    first_name: string;
    last_name: string;
    third_name: string;
    phone: string;
    email: string;
    notes: string;
    role: string | { id: number; name: string; description: string };
}

export class CoreClient {
    private baseUrl: string;
    private token?: string;

    constructor(token?: string | null, baseUrl: string = 'http://gateway/api') {
        this.baseUrl = baseUrl;
        this.token = token || undefined;
    }

    // --- User Management ---

    async listUsers(): Promise<CoreUser[]> {
        return this.fetchWithAuth<CoreUser[]>('/users/');
    }

    async getUser(authId: string): Promise<CoreUser | null> {
        if (!authId) return null;
        return this.fetchWithAuth<CoreUser>(`/users/by-auth-id/${authId}`);
    }

    async createUser(data: Omit<CoreUser, 'ID'>): Promise<CoreUser | null> {
        return this.fetchWithAuth<CoreUser>('/users/', 'POST', data);
    }

    async updateUser(authId: string, data: Partial<CoreUser>): Promise<CoreUser | null> {
        return this.fetchWithAuth<CoreUser>(`/users/${authId}`, 'PUT', data);
    }

    async deleteUser(authId: string): Promise<boolean> {
        return this.fetchWithAuth<boolean>(`/users/${authId}`, 'DELETE');
    }

    async getMyProfile(): Promise<CoreUser | null> {
        return this.fetchWithAuth<CoreUser>('/users/me');
    }

    async updateMyProfile(data: Partial<CoreUser>): Promise<CoreUser | null> {
        return this.fetchWithAuth<CoreUser>('/users/me', 'PUT', data);
    }

    private async fetchWithAuth<T>(endpoint: string, method: string = 'GET', body?: any): Promise<T> {
        if (!this.token) {
            throw new Error('CoreClient: No token provided');
        }

        try {
            const options: RequestInit = {
                method,
                headers: {
                    'Authorization': `Bearer ${this.token}`,
                    'Content-Type': 'application/json'
                }
            };

            if (body) {
                options.body = JSON.stringify(body);
            }

            const response = await fetch(`${this.baseUrl}${endpoint}`, options);
            console.log({response, options})
            if (!response.ok) {
                // throw new Error(`Request failed with status ${response.status}`);
                return undefined as T 
            }

            if (response.status === 204) {
                return true as unknown as T;
            }

            const text = await response.text();
            if (!text) return true as unknown as T;

            return JSON.parse(text);
        } catch (error) {
            console.error(`CoreClient request to ${endpoint} failed:`, error);
            throw error;
        }
    }
}

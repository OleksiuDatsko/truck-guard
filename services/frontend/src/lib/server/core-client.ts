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

    // --- Events ---
    async getEvents<T>(type: string, page: number = 1, limit: number = 10, filters?: Record<string, string | undefined>): Promise<T> {
        const query = new URLSearchParams({
            page: page.toString(),
            limit: limit.toString(),
        });

        if (filters) {
            Object.entries(filters).forEach(([key, value]) => {
                if (value) query.append(key, value);
            });
        }

        return this.fetchWithAuth<T>(`/events/${type}?${query.toString()}`);
    }

    async getGateEvent<T>(id: string | number): Promise<T> {
        return this.fetchWithAuth<T>(`/events/gate/${id}`);
    }

    async getSystemEvent<T>(id: string | number): Promise<T> {
        return this.fetchWithAuth<T>(`/events/system/${id}`);
    }

    async getPlateEvent<T>(id: string | number): Promise<T> {
        return this.fetchWithAuth<T>(`/events/plate/${id}`);
    }

    async correctPlate(id: string | number, newPlate: string): Promise<boolean> {
        return this.fetchWithAuth<boolean>(`/events/plate/${id}`, 'PATCH', { plate_corrected: newPlate });
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

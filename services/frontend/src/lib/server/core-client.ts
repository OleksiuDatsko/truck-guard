export interface CoreUser {
    ID: number;
    auth_id: number;
    first_name: string;
    last_name: string;
    third_name: string;
    phone: string;
    email: string;
    notes: string;
    role: string;
}

export class CoreClient {
    private baseUrl: string;
    private token: string;

    constructor(token: string, baseUrl: string = 'http://gateway/api') {
        this.baseUrl = baseUrl;
        this.token = token
    }

    async getUser(authId: string): Promise<CoreUser | null> {
        if (!authId) return null;
        try {
            const response = await fetch(`${this.baseUrl}/users/by-auth-id/${authId}`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${this.token}`
                }
            });

            if (!response.ok) {
                return null;
            }
            return await response.json();
        } catch (error) {
            console.error('CoreClient.getUser error:', error);
            return null;
        }
    }
}


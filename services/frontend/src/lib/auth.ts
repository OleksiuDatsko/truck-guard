export type User = {
    permissions: string[];
};

export function can(user: User | null | undefined, permission: string): boolean {
    if (!user || !user.permissions) return false;
    
    // Admin has everything
    if (user.permissions.includes('admin')) return true;

    for (const p of user.permissions) {
        if (p === permission) return true;

        // Hierarchical logic for action:resource
        const [actionUser, resourceUser] = p.split(':');
        const [actionReq, resourceReq] = permission.split(':');

        if (actionUser && resourceUser && actionReq && resourceReq) {
            if (resourceUser === '*' || resourceUser === resourceReq) {
                const levels: Record<string, number> = {
                    'read': 1,
                    'create': 2,
                    'update': 3,
                    'delete': 4,
                    'manage': 5,
                    'admin': 10
                };

                const userLevel = levels[actionUser] || 0;
                const reqLevel = levels[actionReq] || 0;

                if (userLevel >= reqLevel) return true;
            }
        }
    }

    return false;
}

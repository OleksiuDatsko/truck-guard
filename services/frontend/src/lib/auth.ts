export type User = {
    permissions: string[];
};

export function can(user: User | null | undefined, permission: string): boolean {
    if (!user) return false;
    return user.permissions.includes(permission);
}

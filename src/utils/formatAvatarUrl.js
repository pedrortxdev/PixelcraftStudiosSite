export const getAvatarUrl = (avatarPath) => {
    if (!avatarPath) return null;
    if (avatarPath.startsWith('http')) return avatarPath;
    const baseUrl = import.meta.env.VITE_API_URL?.replace('/api/v1', '') || 'https://api.pixelcraft-studio.store';
    return `${baseUrl}${avatarPath}`;
};

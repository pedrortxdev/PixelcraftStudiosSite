import React from 'react';

/**
 * Role configuration with display labels and colors
 */
const roleConfig = {
    PARTNER: {
        label: 'Parceiro',
        color: '#00bd65',
        shadow: 'rgba(0, 189, 101, 0.4)'
    },
    CLIENT: {
        label: 'Cliente',
        color: '#00d415',
        shadow: 'rgba(0, 212, 21, 0.4)'
    },
    CLIENT_VIP: {
        label: 'Cliente VIP',
        color: '#00bd13',
        shadow: 'rgba(0, 189, 19, 0.4)'
    },
    SUPPORT: {
        label: 'Suporte',
        color: '#fbff00',
        shadow: 'rgba(251, 255, 0, 0.4)'
    },
    ADMIN: {
        label: 'Administração',
        color: '#bd005b',
        shadow: 'rgba(189, 0, 91, 0.4)'
    },
    DEVELOPMENT: {
        label: 'Desenvolvimento',
        color: '#0047bd',
        shadow: 'rgba(0, 71, 189, 0.4)'
    },
    ENGINEERING: {
        label: 'Engenharia',
        color: '#6a00ff',
        shadow: 'rgba(106, 0, 255, 0.4)'
    },
    DIRECTION: {
        label: 'Direção',
        color: '#ff3f00',
        shadow: 'rgba(255, 63, 0, 0.4)'
    },
};

/**
 * Role hierarchy for determining highest role
 */
const roleHierarchy = {
    PARTNER: 1,
    CLIENT: 2,
    CLIENT_VIP: 3,
    SUPPORT: 4,
    ADMIN: 5,
    DEVELOPMENT: 6,
    ENGINEERING: 7,
    DIRECTION: 8,
};

/**
 * Get the highest role from a list of roles
 */
export const getHighestRole = (roles) => {
    if (!roles || roles.length === 0) return null;

    return roles.reduce((highest, role) => {
        if (!highest) return role;
        return (roleHierarchy[role] || 0) > (roleHierarchy[highest] || 0) ? role : highest;
    }, null);
};

/**
 * Check if user has any admin-level role
 */
export const hasAdminAccess = (roles) => {
    if (!roles || roles.length === 0) return false;
    const adminRoles = ['SUPPORT', 'ADMIN', 'DEVELOPMENT', 'ENGINEERING', 'DIRECTION'];
    return roles.some(role => adminRoles.includes(role));
};

/**
 * RoleBadge component displays a user's role with Minecraft-style font
 */
const RoleBadge = ({ role, roles, size = 'normal', showBorder = true, customColor = null, customLabel = null }) => {
    // If roles array is provided, use highest role
    const displayRole = role || getHighestRole(roles);

    if (!displayRole) return null;

    // Check if it's a custom role (not in roleConfig)
    const config = roleConfig[displayRole] || {
        label: customLabel || displayRole,
        color: customColor || '#999999',
        shadow: customColor ? `${customColor}66` : 'rgba(153, 153, 153, 0.4)'
    };

    const sizeStyles = {
        small: {
            fontSize: '0.65rem',
            padding: '1px 6px',
        },
        normal: {
            fontSize: '0.75rem',
            padding: '2px 8px',
        },
        large: {
            fontSize: '0.875rem',
            padding: '3px 10px',
        },
    };

    const badgeStyle = {
        fontFamily: "'MinecraftFont', 'Courier New', monospace",
        fontSize: sizeStyles[size].fontSize,
        color: config.color,
        textShadow: `0 0 8px ${config.shadow}, 0 0 15px ${config.shadow}`,
        border: showBorder ? `1px solid ${config.color}50` : 'none',
        padding: sizeStyles[size].padding,
        borderRadius: '4px',
        backgroundColor: `${config.color}15`,
        display: 'inline-flex',
        alignItems: 'center',
        gap: '4px',
        whiteSpace: 'nowrap',
        letterSpacing: '0.02em',
        fontWeight: 'bold',
        lineHeight: '1.4',
        overflow: 'visible',
    };

    return (
        <span style={badgeStyle} title={`Cargo: ${config.label}`} className="minecraft-text-fix">
            {config.label}
        </span>
    );
};

/**
 * RoleBadgeList component shows all user roles
 */
export const RoleBadgeList = ({ roles, maxDisplay = 3 }) => {
    if (!roles || roles.length === 0) return null;

    // Sort roles by hierarchy (highest first)
    const sortedRoles = [...roles].sort((a, b) =>
        (roleHierarchy[b] || 0) - (roleHierarchy[a] || 0)
    );

    const displayRoles = sortedRoles.slice(0, maxDisplay);
    const remaining = sortedRoles.length - maxDisplay;

    return (
        <div style={{ display: 'flex', gap: '6px', flexWrap: 'wrap', alignItems: 'center' }}>
            {displayRoles.map(role => (
                <RoleBadge key={role} role={role} size="small" />
            ))}
            {remaining > 0 && (
                <span style={{
                    fontSize: '0.7rem',
                    color: '#888',
                    fontStyle: 'italic'
                }}>
                    +{remaining}
                </span>
            )}
        </div>
    );
};

/**
 * Get role display info
 */
export const getRoleInfo = (role) => roleConfig[role] || null;

/**
 * Get all roles config (for admin UI)
 */
export const getAllRolesConfig = () => roleConfig;

export default RoleBadge;

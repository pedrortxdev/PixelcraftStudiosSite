/**
 * Role Utility Functions
 * Helper functions for role hierarchy and permission checks
 */

import { ROLE_HIERARCHY, ROLE_DISPLAY_CONFIG } from '../constants/roles';

/**
 * Get the highest role from a list of roles
 * @param {string[]} roles - Array of role strings
 * @returns {string|null} - Highest role or null if empty
 */
export function getHighestRole(roles) {
  if (!roles || roles.length === 0) {
    return null;
  }

  let highest = roles[0];
  let highestLevel = ROLE_HIERARCHY[highest] || 0;

  for (const role of roles) {
    const level = ROLE_HIERARCHY[role] || 0;
    if (level > highestLevel) {
      highest = role;
      highestLevel = level;
    }
  }

  return highest;
}

/**
 * Get hierarchy level for a role
 * @param {string} role - Role string
 * @returns {number} - Hierarchy level (0 if not found)
 */
export function getRoleLevel(role) {
  return ROLE_HIERARCHY[role] || 0;
}

/**
 * Check if source roles can edit target roles
 * @param {string[]} sourceRoles - Roles of the user performing the action
 * @param {string[]} targetRoles - Roles of the user being edited
 * @returns {boolean} - True if can edit
 */
export function canEditRole(sourceRoles, targetRoles) {
  if (!sourceRoles || sourceRoles.length === 0) {
    return false;
  }

  // DIRECTION can edit anyone
  if (sourceRoles.includes('DIRECTION')) {
    return true;
  }

  const sourceHighest = getHighestRole(sourceRoles);
  const sourceLevel = getRoleLevel(sourceHighest);

  // Check if source level is higher than all target roles
  for (const targetRole of targetRoles) {
    const targetLevel = getRoleLevel(targetRole);
    if (targetLevel >= sourceLevel) {
      return false;
    }
  }

  return true;
}

/**
 * Check if a user can edit a specific role
 * @param {string[]} userRoles - User's roles
 * @param {string} targetRole - Role to check
 * @returns {boolean} - True if can edit
 */
export function canEditSpecificRole(userRoles, targetRole) {
  if (!userRoles || userRoles.length === 0) {
    return false;
  }

  // DIRECTION can edit any role
  if (userRoles.includes('DIRECTION')) {
    return true;
  }

  const userHighest = getHighestRole(userRoles);
  const userLevel = getRoleLevel(userHighest);
  const targetLevel = getRoleLevel(targetRole);

  return userLevel > targetLevel;
}

/**
 * Get display information for a role
 * @param {string} role - Role string
 * @returns {object} - Display info (label, color, description)
 */
export function getRoleDisplayInfo(role) {
  return ROLE_DISPLAY_CONFIG[role] || {
    label: role,
    color: '#999999',
    description: 'Cargo desconhecido',
  };
}

/**
 * Sort roles by hierarchy (highest first)
 * @param {string[]} roles - Array of role strings
 * @returns {string[]} - Sorted array
 */
export function sortRolesByHierarchy(roles) {
  return [...roles].sort((a, b) => {
    const levelA = getRoleLevel(a);
    const levelB = getRoleLevel(b);
    return levelB - levelA; // Descending order
  });
}

/**
 * Check if a role is an admin role
 * @param {string} role - Role string
 * @returns {boolean} - True if admin role
 */
export function isAdminRole(role) {
  const adminRoles = ['SUPPORT', 'ADMIN', 'DEVELOPMENT', 'ENGINEERING', 'DIRECTION'];
  return adminRoles.includes(role);
}

/**
 * Check if user has any admin role
 * @param {string[]} roles - User's roles
 * @returns {boolean} - True if has admin role
 */
export function hasAdminRole(roles) {
  if (!roles || roles.length === 0) {
    return false;
  }
  return roles.some(role => isAdminRole(role));
}

/**
 * Filter roles that user can edit
 * @param {string[]} userRoles - User's roles
 * @param {string[]} allRoles - All available roles
 * @returns {string[]} - Roles that can be edited
 */
export function getEditableRoles(userRoles, allRoles) {
  return allRoles.filter(role => canEditSpecificRole(userRoles, role));
}

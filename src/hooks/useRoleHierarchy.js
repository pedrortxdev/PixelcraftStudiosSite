/**
 * useRoleHierarchy Hook
 * Manages role hierarchy logic and edit permissions
 */

import { useMemo } from 'react';
import { useAuth } from '../context/AuthContext';
import { getHighestRole, getRoleLevel, canEditSpecificRole } from '../utils/roleUtils';

export function useRoleHierarchy() {
  const { user } = useAuth();

  // Get user's roles
  const userRoles = useMemo(() => {
    if (!user || !user.roles) {
      return [];
    }
    return user.roles;
  }, [user]);

  // Calculate highest hierarchy level
  const userLevel = useMemo(() => {
    const highest = getHighestRole(userRoles);
    return getRoleLevel(highest);
  }, [userRoles]);

  // Check if user is DIRECTION
  const isDirection = useMemo(() => {
    return userRoles.includes('DIRECTION');
  }, [userRoles]);

  // Function to check if user can edit a specific role
  const canEdit = useMemo(() => {
    return (targetRole) => {
      return canEditSpecificRole(userRoles, targetRole);
    };
  }, [userRoles]);

  return {
    userLevel,
    userRoles,
    isDirection,
    canEdit,
  };
}

/**
 * usePermissions Hook
 * Manages permissions for a specific role with optimistic updates
 */

import { useState, useEffect, useCallback, useRef } from 'react';
import { rolesAPI } from '../services/api';
import { RESOURCE_LABELS, ACTION_LABELS } from '../constants/permissions';
import { useAuth } from '../context/AuthContext';

export function usePermissions(role) {
  const { user } = useAuth();
  const [permissions, setPermissions] = useState([]);
  const [resources, setResources] = useState([]);
  const [actions, setActions] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  
  // Use provided role or current user's highest role
  const targetRole = role || user?.highest_role || (user?.roles && user.roles[0]);

  // Keep track of previous permissions for rollback
  const previousPermissionsRef = useRef([]);

  // Load resources and actions (only once)
  useEffect(() => {
    const loadMetadata = async () => {
      try {
        const [resourcesRes, actionsRes] = await Promise.all([
          rolesAPI.getAvailableResources(),
          rolesAPI.getAvailableActions(),
        ]);

        // Transform resources
        const resourcesData = (resourcesRes.resources || []).map(resource => ({
          type: resource,
          label: RESOURCE_LABELS[resource] || resource,
        }));

        // Transform actions
        const actionsData = (actionsRes.actions || []).map(action => ({
          type: action,
          label: ACTION_LABELS[action] || action,
        }));

        setResources(resourcesData);
        setActions(actionsData);
      } catch (err) {
        console.error('Error loading metadata:', err);
      }
    };

    loadMetadata();
  }, []);

  // Load permissions when role changes
  useEffect(() => {
    if (!targetRole) {
      setPermissions([]);
      return;
    }

    const loadPermissions = async () => {
      setIsLoading(true);
      setError(null);

      try {
        const perms = await rolesAPI.getRolePermissions(targetRole);
        setPermissions(perms || []);
        previousPermissionsRef.current = perms || [];
      } catch (err) {
        console.error('Error loading permissions:', err);
        setError(err.message || 'Erro ao carregar permissões');
      } finally {
        setIsLoading(false);
      }
    };

    loadPermissions();
  }, [targetRole]);

  // Check if a permission exists
  const hasPermission = useCallback((resource, action) => {
    if (!action) {
      // Fallback for calls like hasPermission('view_cpf')
      // Map legacy slugs to Resource/Action if needed
      if (resource === 'view_cpf') return permissions.some(p => p.resource === 'USERS' && p.action === 'VIEW_CPF');
      return false;
    }
    return permissions.some(
      perm => (perm.resource === resource || perm.resource === 'ALL') && (perm.action === action || perm.action === 'MANAGE')
    );
  }, [permissions]);

  // Add permission with optimistic update
  const addPermission = useCallback(async (resource, action) => {
    if (!targetRole) return;

    // Save current state for rollback
    previousPermissionsRef.current = permissions;

    // Optimistic update
    const newPermission = {
      id: `temp-${Date.now()}`,
      role: targetRole,
      resource,
      action,
      created_at: new Date().toISOString(),
    };
    setPermissions(prev => [...prev, newPermission]);

    try {
      await rolesAPI.addRolePermission(targetRole, resource, action);
      
      // Reload to get actual data from server
      const updatedPerms = await rolesAPI.getRolePermissions(targetRole);
      setPermissions(updatedPerms || []);
      previousPermissionsRef.current = updatedPerms || [];
    } catch (err) {
      console.error('Error adding permission:', err);
      
      // Rollback on error
      setPermissions(previousPermissionsRef.current);
      throw err;
    }
  }, [targetRole, permissions]);

  // Remove permission with optimistic update
  const removePermission = useCallback(async (resource, action) => {
    if (!targetRole) return;

    // Save current state for rollback
    previousPermissionsRef.current = permissions;

    // Optimistic update
    setPermissions(prev => 
      prev.filter(perm => !(perm.resource === resource && perm.action === action))
    );

    try {
      await rolesAPI.removeRolePermission(targetRole, resource, action);
      
      // Reload to get actual data from server
      const updatedPerms = await rolesAPI.getRolePermissions(targetRole);
      setPermissions(updatedPerms || []);
      previousPermissionsRef.current = updatedPerms || [];
    } catch (err) {
      console.error('Error removing permission:', err);
      
      // Rollback on error
      setPermissions(previousPermissionsRef.current);
      throw err;
    }
  }, [targetRole, permissions]);

  return {
    permissions,
    resources,
    actions,
    addPermission,
    removePermission,
    hasPermission,
    isLoading,
    error,
  };
}

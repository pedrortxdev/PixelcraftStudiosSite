/**
 * useRoles Hook
 * Manages roles state and operations
 */

import { useState, useEffect, useCallback } from 'react';
import { rolesAPI } from '../services/api';
import { ROLE_DISPLAY_CONFIG, ROLE_HIERARCHY } from '../constants/roles';

// Global cache outside the hook to prevent duplicate fetches across components
let cachedRoles = null;
let fetchPromise = null;

export function useRoles() {
  const [roles, setRoles] = useState(cachedRoles || []);
  const [selectedRole, setSelectedRole] = useState(null);
  const [isLoading, setIsLoading] = useState(!cachedRoles);
  const [error, setError] = useState(null);

  const loadRoles = useCallback(async (forceRefresh = false) => {
    if (cachedRoles && !forceRefresh) {
      setRoles(cachedRoles);
      setIsLoading(false);
      return;
    }

    if (fetchPromise && !forceRefresh) {
      setIsLoading(true);
      try {
        const data = await fetchPromise;
        setRoles(data);
      } catch (err) {
        setError(err.message || 'Erro ao carregar cargos');
      } finally {
        setIsLoading(false);
      }
      return;
    }

    setIsLoading(true);
    setError(null);

    fetchPromise = (async () => {
      // Load default roles
      const response = await rolesAPI.getAvailableRoles();

      // Load custom roles
      let customRolesData = [];
      try {
        const customResponse = await rolesAPI.getCustomRoles();
        customRolesData = customResponse.roles || [];
      } catch (err) {
        console.warn('Could not load custom roles:', err);
      }

      // Transform default roles
      const defaultRoles = response.roles.map(roleItem => ({
        role: roleItem.role,
        description: roleItem.description,
        label: ROLE_DISPLAY_CONFIG[roleItem.role]?.label || roleItem.role,
        color: ROLE_DISPLAY_CONFIG[roleItem.role]?.color || '#999999',
        hierarchyLevel: ROLE_HIERARCHY[roleItem.role] || 0,
        isCustom: false,
      }));

      // Transform custom roles
      const customRoles = customRolesData.map(customRole => ({
        role: customRole.role_name,
        description: customRole.description || 'Cargo customizado',
        label: customRole.display_name,
        color: customRole.color || '#999999',
        hierarchyLevel: customRole.hierarchy_level || 0,
        isCustom: true,
      }));

      // Combine and sort by hierarchy level
      const allRoles = [...defaultRoles, ...customRoles].sort((a, b) => b.hierarchyLevel - a.hierarchyLevel);

      cachedRoles = allRoles;
      return allRoles;
    })();

    try {
      const data = await fetchPromise;
      setRoles(data);
    } catch (err) {
      console.error('Error loading roles:', err);
      setError(err.message || 'Erro ao carregar cargos');
      cachedRoles = null;
      fetchPromise = null;
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    loadRoles();
  }, [loadRoles]);

  const selectRole = useCallback((role) => {
    setSelectedRole(role);
  }, []);

  const refresh = useCallback(async () => {
    await loadRoles(true);
  }, [loadRoles]);

  return {
    roles,
    selectedRole,
    selectRole,
    isLoading,
    error,
    refresh,
  };
}

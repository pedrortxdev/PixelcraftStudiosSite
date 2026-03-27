/**
 * RolesList Component
 * Displays a list of roles in the sidebar
 */

import React, { useMemo } from 'react';
import RoleCard from './RoleCard';
import { sortRolesByHierarchy } from '../../../utils/roleUtils';

const RolesList = ({
  roles,
  selectedRole,
  onRoleSelect,
  searchQuery,
  filterResource,
  userHierarchyLevel,
  canEdit,
}) => {
  // Filter and sort roles
  const filteredRoles = useMemo(() => {
    let filtered = [...roles];

    // Apply search filter
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(role =>
        role.label.toLowerCase().includes(query) ||
        role.description.toLowerCase().includes(query) ||
        role.role.toLowerCase().includes(query)
      );
    }

    // Apply resource filter (if needed - would require permissions data)
    // This would be implemented if we pass permissions data to this component

    // Sort by hierarchy (highest first)
    const roleNames = filtered.map(r => r.role);
    const sorted = sortRolesByHierarchy(roleNames);
    
    return sorted.map(roleName => 
      filtered.find(r => r.role === roleName)
    ).filter(Boolean);
  }, [roles, searchQuery]);

  if (filteredRoles.length === 0) {
    return (
      <div style={{
        padding: '20px',
        textAlign: 'center',
        color: '#888',
        fontSize: '0.9rem',
      }}>
        Nenhum cargo encontrado
      </div>
    );
  }

  return (
    <div style={{ padding: '8px' }}>
      {filteredRoles.map(role => (
        <RoleCard
          key={role.role}
          role={role}
          isSelected={selectedRole === role.role}
          isEditable={canEdit(role.role)}
          onClick={() => onRoleSelect(role.role)}
        />
      ))}
    </div>
  );
};

export default RolesList;

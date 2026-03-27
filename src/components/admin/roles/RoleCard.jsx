/**
 * RoleCard Component
 * Displays a role card in the roles list
 */

import React from 'react';
import RoleBadge from '../../RoleBadge';
import { Lock } from 'lucide-react';

const RoleCard = ({ role, isSelected, isEditable, onClick }) => {
  return (
    <div
      onClick={onClick}
      className={`
        role-card
        ${isSelected ? 'selected' : ''}
        ${!isEditable ? 'locked' : ''}
      `}
      style={{
        padding: '16px',
        marginBottom: '8px',
        borderRadius: '8px',
        border: isSelected ? '2px solid #00d4ff' : '1px solid #333',
        backgroundColor: isSelected ? '#1a1a2e' : '#0f0f1a',
        cursor: 'pointer',
        transition: 'all 0.2s ease',
        position: 'relative',
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '8px' }}>
        <RoleBadge 
          role={role.role} 
          size="normal"
          customColor={role.isCustom ? role.color : null}
          customLabel={role.isCustom ? role.label : null}
        />
        {!isEditable && (
          <Lock size={16} color="#888" title="Você não pode editar este cargo" />
        )}
      </div>
      
      <div style={{ fontSize: '0.85rem', color: '#aaa', marginBottom: '4px' }}>
        {role.description}
      </div>
      
      <div style={{ fontSize: '0.75rem', color: '#666' }}>
        Nível: {role.hierarchyLevel}
      </div>
    </div>
  );
};

export default RoleCard;

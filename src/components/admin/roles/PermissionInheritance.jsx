/**
 * PermissionInheritance Component
 * Manages permission inheritance between roles
 */

import React, { useState } from 'react';
import { GitBranch, Trash2, AlertCircle, ChevronDown, ChevronUp } from 'lucide-react';
import api from '../../../services/api';

const ROLE_HIERARCHY = {
  DIRECTION: ['ENGINEERING', 'DEVELOPMENT', 'ADMIN', 'SUPPORT'],
  ENGINEERING: ['DEVELOPMENT', 'ADMIN', 'SUPPORT'],
  DEVELOPMENT: ['ADMIN', 'SUPPORT'],
  ADMIN: ['SUPPORT'],
  SUPPORT: []
};

const PermissionInheritance = ({ currentRole, onSuccess, onError }) => {
  const [loading, setLoading] = useState(false);
  const [selectedSource, setSelectedSource] = useState('');
  const [isExpanded, setIsExpanded] = useState(false);

  const availableSources = ROLE_HIERARCHY[currentRole] || [];

  const handleInherit = async () => {
    if (!selectedSource) {
      onError('Selecione um cargo para herdar permissões');
      return;
    }

    try {
      setLoading(true);
      const response = await api.roles.inheritPermissions(currentRole, selectedSource);
      onSuccess(`${response.inherited_count} permissões herdadas de ${selectedSource}`);
      setSelectedSource('');
      setIsExpanded(false);
    } catch (error) {
      onError(error.message || 'Erro ao herdar permissões');
    } finally {
      setLoading(false);
    }
  };

  const handleRemoveInherited = async () => {
    if (!confirm('Tem certeza que deseja remover todas as permissões herdadas?')) {
      return;
    }

    try {
      setLoading(true);
      const response = await api.roles.removeInheritedPermissions(currentRole);
      onSuccess(`${response.removed_count} permissões herdadas removidas`);
    } catch (error) {
      onError(error.message || 'Erro ao remover permissões herdadas');
    } finally {
      setLoading(false);
    }
  };

  if (availableSources.length === 0) {
    return null;
  }

  return (
    <div style={{
      backgroundColor: '#1a1a2e',
      border: '1px solid #333',
      borderRadius: '6px',
      marginBottom: '16px',
    }}>
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        style={{
          width: '100%',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '12px 16px',
          backgroundColor: 'transparent',
          border: 'none',
          color: '#00d4ff',
          cursor: 'pointer',
          fontSize: '0.9rem',
          fontWeight: '500',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <GitBranch size={16} />
          <span>Herança de Permissões</span>
        </div>
        {isExpanded ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
      </button>

      {isExpanded && (
        <div style={{ padding: '0 16px 16px 16px' }}>
          <p style={{
            color: '#888',
            fontSize: '0.8rem',
            marginBottom: '12px',
          }}>
            Herde todas as permissões de um cargo inferior
          </p>

          <div style={{ display: 'flex', gap: '8px', marginBottom: '12px' }}>
            <select
              value={selectedSource}
              onChange={(e) => setSelectedSource(e.target.value)}
              disabled={loading}
              style={{
                flex: 1,
                padding: '8px',
                backgroundColor: '#0f0f1a',
                border: '1px solid #333',
                borderRadius: '4px',
                color: '#fff',
                fontSize: '0.85rem',
              }}
            >
              <option value="">Selecione um cargo...</option>
              {availableSources.map(role => (
                <option key={role} value={role}>{role}</option>
              ))}
            </select>

            <button
              onClick={handleInherit}
              disabled={loading || !selectedSource}
              style={{
                padding: '8px 16px',
                backgroundColor: selectedSource && !loading ? '#00d4ff' : '#333',
                color: selectedSource && !loading ? '#000' : '#666',
                border: 'none',
                borderRadius: '4px',
                cursor: selectedSource && !loading ? 'pointer' : 'not-allowed',
                fontWeight: '500',
                fontSize: '0.85rem',
              }}
            >
              {loading ? 'Herdando...' : 'Herdar'}
            </button>
          </div>

          <button
            onClick={handleRemoveInherited}
            disabled={loading}
            style={{
              width: '100%',
              padding: '8px',
              backgroundColor: 'transparent',
              color: '#ff4444',
              border: '1px solid #ff4444',
              borderRadius: '4px',
              cursor: loading ? 'not-allowed' : 'pointer',
              fontSize: '0.85rem',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              gap: '6px',
            }}
          >
            <Trash2 size={14} />
            {loading ? 'Removendo...' : 'Remover Herdadas'}
          </button>
        </div>
      )}
    </div>
  );
};

export default PermissionInheritance;

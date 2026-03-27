/**
 * RoleDetailPanel Component
 * Displays details and permissions for a selected role
 */

import React, { useState, useEffect } from 'react';
import RoleBadge from '../../RoleBadge';
import PermissionsMatrix from './PermissionsMatrix';
import PermissionInheritance from './PermissionInheritance';
import { Edit, Lock, Shield, Users, TrendingUp } from 'lucide-react';

const RoleDetailPanel = ({
  role,
  roleInfo,
  permissions,
  resources,
  actions,
  canEdit,
  onPermissionChange,
  isLoading,
}) => {
  const [editMode, setEditMode] = useState(false);
  const [notification, setNotification] = useState(null);
  const [stats, setStats] = useState({
    totalPermissions: 0,
    inheritedPermissions: 0,
    directPermissions: 0
  });

  useEffect(() => {
    if (permissions) {
      const inherited = permissions.filter(p => p.is_inherited).length;
      const total = permissions.length;
      setStats({
        totalPermissions: total,
        inheritedPermissions: inherited,
        directPermissions: total - inherited
      });
    }
  }, [permissions]);

  if (!role) {
    return (
      <div style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        height: '100%',
        color: '#888',
        fontSize: '1.1rem',
      }}>
        Selecione um cargo para visualizar suas permissões
      </div>
    );
  }

  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      {/* Header */}
      <div style={{
        padding: '24px',
        borderBottom: '1px solid #333',
        backgroundColor: '#0f0f1a',
      }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '12px' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
            <RoleBadge 
              role={role} 
              size="large"
              customColor={roleInfo?.isCustom ? roleInfo.color : null}
              customLabel={roleInfo?.isCustom ? roleInfo.label : null}
            />
            {!canEdit && (
              <div style={{
                display: 'flex',
                alignItems: 'center',
                gap: '6px',
                fontSize: '0.85rem',
                color: '#888',
              }}>
                <Lock size={16} />
                <span>Somente leitura</span>
              </div>
            )}
          </div>
          
          {canEdit && (
            <button
              onClick={() => setEditMode(!editMode)}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '8px',
                padding: '8px 16px',
                backgroundColor: editMode ? '#00d415' : '#1a1a2e',
                border: editMode ? '1px solid #00d415' : '1px solid #333',
                borderRadius: '6px',
                color: editMode ? '#000' : '#fff',
                cursor: 'pointer',
                fontSize: '0.9rem',
                fontWeight: 'bold',
                transition: 'all 0.2s ease',
              }}
            >
              <Edit size={16} />
              {editMode ? 'Modo de Edição' : 'Editar Permissões'}
            </button>
          )}
        </div>
        
        {roleInfo && (
          <div style={{ fontSize: '0.9rem', color: '#aaa', marginBottom: '16px' }}>
            {roleInfo.description}
          </div>
        )}

        {/* Stats Cards */}
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(150px, 1fr))',
          gap: '12px',
          marginTop: '16px',
        }}>
          <div style={{
            backgroundColor: '#1a1a2e',
            border: '1px solid #333',
            borderRadius: '8px',
            padding: '12px',
          }}>
            <div style={{
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
              marginBottom: '8px',
            }}>
              <Shield size={16} color="#00d4ff" />
              <span style={{ fontSize: '0.75rem', color: '#888', textTransform: 'uppercase' }}>
                Total
              </span>
            </div>
            <div style={{ fontSize: 'var(--title-h4)', fontWeight: 'bold', color: '#00d4ff' }}>
              {stats.totalPermissions}
            </div>
            <div style={{ fontSize: '0.75rem', color: '#666' }}>
              Permissões
            </div>
          </div>

          <div style={{
            backgroundColor: '#1a1a2e',
            border: '1px solid #333',
            borderRadius: '8px',
            padding: '12px',
          }}>
            <div style={{
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
              marginBottom: '8px',
            }}>
              <TrendingUp size={16} color="#00ff88" />
              <span style={{ fontSize: '0.75rem', color: '#888', textTransform: 'uppercase' }}>
                Diretas
              </span>
            </div>
            <div style={{ fontSize: 'var(--title-h4)', fontWeight: 'bold', color: '#00ff88' }}>
              {stats.directPermissions}
            </div>
            <div style={{ fontSize: '0.75rem', color: '#666' }}>
              Configuradas
            </div>
          </div>

          <div style={{
            backgroundColor: '#1a1a2e',
            border: '1px solid #333',
            borderRadius: '8px',
            padding: '12px',
          }}>
            <div style={{
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
              marginBottom: '8px',
            }}>
              <Users size={16} color="#ff9900" />
              <span style={{ fontSize: '0.75rem', color: '#888', textTransform: 'uppercase' }}>
                Herdadas
              </span>
            </div>
            <div style={{ fontSize: 'var(--title-h4)', fontWeight: 'bold', color: '#ff9900' }}>
              {stats.inheritedPermissions}
            </div>
            <div style={{ fontSize: '0.75rem', color: '#666' }}>
              De outros cargos
            </div>
          </div>
        </div>
      </div>

      {/* Permissions Matrix */}
      <div style={{
        flex: 1,
        overflowY: 'auto',
        padding: '24px',
      }}>
        {!canEdit && !editMode && (
          <div style={{
            padding: '12px',
            backgroundColor: '#1a1a2e',
            border: '1px solid #333',
            borderRadius: '6px',
            marginBottom: '16px',
            fontSize: '0.85rem',
            color: '#aaa',
          }}>
            ℹ️ Você não pode editar este cargo devido à hierarquia de permissões.
            Apenas cargos superiores podem modificar cargos inferiores.
          </div>
        )}

        {/* Permission Inheritance */}
        {canEdit && (
          <div style={{ marginBottom: '24px' }}>
            <PermissionInheritance
              currentRole={role}
              onSuccess={(message) => {
                setNotification({ message, type: 'success' });
                // Trigger refresh of permissions
                window.location.reload();
              }}
              onError={(message) => {
                setNotification({ message, type: 'error' });
              }}
            />
          </div>
        )}
        
        <PermissionsMatrix
          role={role}
          permissions={permissions}
          resources={resources}
          actions={actions}
          editMode={editMode}
          onPermissionToggle={onPermissionChange}
          isLoading={isLoading}
        />
      </div>
    </div>
  );
};

export default RoleDetailPanel;

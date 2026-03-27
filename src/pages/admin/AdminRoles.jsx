/**
 * AdminRoles Page
 * Main page for role and permission management
 */

import React, { useState, useCallback } from 'react';
import { Search, X, History, Download, Upload, Settings, Bell, BookOpen } from 'lucide-react';
import { useRoles } from '../../hooks/useRoles';
import { usePermissions } from '../../hooks/usePermissions';
import { useRoleHierarchy } from '../../hooks/useRoleHierarchy';
import RolesList from '../../components/admin/roles/RolesList';
import RoleDetailPanel from '../../components/admin/roles/RoleDetailPanel';
import NotificationToast from '../../components/admin/roles/NotificationToast';
import AuditLogViewer from '../../components/admin/roles/AuditLogViewer';
import ExportImportModal from '../../components/admin/roles/ExportImportModal';
import CustomRolesManager from '../../components/admin/roles/CustomRolesManager';
import NotificationsPanel from '../../components/admin/roles/NotificationsPanel';
import TemplatesLibrary from '../../components/admin/roles/TemplatesLibrary';

const AdminRoles = () => {
  const [searchQuery, setSearchQuery] = useState('');
  const [notification, setNotification] = useState(null);
  const [showAuditLog, setShowAuditLog] = useState(false);
  const [showExportImport, setShowExportImport] = useState(false);
  const [activeTab, setActiveTab] = useState('permissions'); // permissions, audit, custom, notifications, templates

  // Hooks
  const { roles, selectedRole, selectRole, isLoading: rolesLoading } = useRoles();
  const { canEdit } = useRoleHierarchy();
  const {
    permissions,
    resources,
    actions,
    addPermission,
    removePermission,
    isLoading: permissionsLoading,
  } = usePermissions(selectedRole);

  // Get selected role info
  const selectedRoleInfo = roles.find(r => r.role === selectedRole);

  // Handle role selection
  const handleRoleSelect = useCallback((role) => {
    selectRole(role);
  }, [selectRole]);

  // Handle permission toggle
  const handlePermissionToggle = useCallback(async (resource, action) => {
    try {
      // Check if permission exists
      const hasPermission = permissions.some(
        perm => perm.resource === resource && perm.action === action
      );

      if (hasPermission) {
        await removePermission(resource, action);
        setNotification({
          message: 'Permissão removida com sucesso',
          type: 'success',
        });
      } else {
        await addPermission(resource, action);
        setNotification({
          message: 'Permissão adicionada com sucesso',
          type: 'success',
        });
      }
    } catch (error) {
      setNotification({
        message: error.message || 'Erro ao modificar permissão',
        type: 'error',
      });
    }
  }, [permissions, addPermission, removePermission]);

  // Handle search change
  const handleSearchChange = (e) => {
    setSearchQuery(e.target.value);
  };

  // Handle clear filters
  const handleClearFilters = () => {
    setSearchQuery('');
  };

  return (
    <div style={{
      minHeight: '100vh',
      backgroundColor: '#0a0a0f',
      color: '#fff',
    }}>
      {/* Header */}
      <div style={{
        padding: '24px',
        borderBottom: '1px solid #333',
        backgroundColor: '#0f0f1a',
      }}>
        <h1 style={{
          fontSize: '1.8rem',
          fontFamily: "'MinecraftFont', monospace",
          color: '#00d4ff',
          marginBottom: '16px',
          lineHeight: '1.4',
          overflow: 'visible',
        }} className="minecraft-text-fix">
          Gerenciamento de Cargos
        </h1>
        
        {/* Search Bar and Actions */}
        <div style={{ display: 'flex', gap: '12px', alignItems: 'center', flexWrap: 'wrap' }}>
          <div style={{
            flex: 1,
            position: 'relative',
            minWidth: '250px',
          }}>
            <Search
              size={18}
              color="#888"
              style={{
                position: 'absolute',
                left: '12px',
                top: '50%',
                transform: 'translateY(-50%)',
              }}
            />
            <input
              type="text"
              placeholder="Buscar cargos..."
              value={searchQuery}
              onChange={handleSearchChange}
              style={{
                width: '100%',
                padding: '10px 12px 10px 40px',
                backgroundColor: '#1a1a2e',
                border: '1px solid #333',
                borderRadius: '6px',
                color: '#fff',
                fontSize: '0.9rem',
              }}
            />
          </div>
          
          {searchQuery && (
            <button
              onClick={handleClearFilters}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '6px',
                padding: '10px 16px',
                backgroundColor: '#1a1a2e',
                border: '1px solid #333',
                borderRadius: '6px',
                color: '#fff',
                cursor: 'pointer',
                fontSize: '0.9rem',
              }}
            >
              <X size={16} />
              Limpar
            </button>
          )}

          <button
            onClick={() => setShowExportImport(true)}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '6px',
              padding: '10px 16px',
              backgroundColor: '#00ff88',
              color: '#000',
              border: 'none',
              borderRadius: '6px',
              cursor: 'pointer',
              fontSize: '0.9rem',
              fontWeight: 'bold',
            }}
          >
            <Download size={16} />
            <Upload size={16} />
            Exportar/Importar
          </button>
        </div>

        {/* Tabs */}
        <div style={{
          display: 'flex',
          gap: '8px',
          marginTop: '20px',
          borderBottom: '1px solid #333',
          overflowX: 'auto',
        }}>
          <button
            onClick={() => setActiveTab('permissions')}
            style={{
              padding: '12px 20px',
              backgroundColor: 'transparent',
              border: 'none',
              borderBottom: activeTab === 'permissions' ? '2px solid #00d4ff' : '2px solid transparent',
              color: activeTab === 'permissions' ? '#00d4ff' : '#888',
              cursor: 'pointer',
              fontSize: '0.9rem',
              fontWeight: 'bold',
              whiteSpace: 'nowrap',
            }}
          >
            Permissões
          </button>
          <button
            onClick={() => setActiveTab('audit')}
            style={{
              padding: '12px 20px',
              backgroundColor: 'transparent',
              border: 'none',
              borderBottom: activeTab === 'audit' ? '2px solid #00d4ff' : '2px solid transparent',
              color: activeTab === 'audit' ? '#00d4ff' : '#888',
              cursor: 'pointer',
              fontSize: '0.9rem',
              fontWeight: 'bold',
              display: 'flex',
              alignItems: 'center',
              gap: '6px',
              whiteSpace: 'nowrap',
            }}
          >
            <History size={16} />
            Histórico
          </button>
          <button
            onClick={() => setActiveTab('custom')}
            style={{
              padding: '12px 20px',
              backgroundColor: 'transparent',
              border: 'none',
              borderBottom: activeTab === 'custom' ? '2px solid #00d4ff' : '2px solid transparent',
              color: activeTab === 'custom' ? '#00d4ff' : '#888',
              cursor: 'pointer',
              fontSize: '0.9rem',
              fontWeight: 'bold',
              display: 'flex',
              alignItems: 'center',
              gap: '6px',
              whiteSpace: 'nowrap',
            }}
          >
            <Settings size={16} />
            Cargos Customizados
          </button>
          <button
            onClick={() => setActiveTab('notifications')}
            style={{
              padding: '12px 20px',
              backgroundColor: 'transparent',
              border: 'none',
              borderBottom: activeTab === 'notifications' ? '2px solid #00d4ff' : '2px solid transparent',
              color: activeTab === 'notifications' ? '#00d4ff' : '#888',
              cursor: 'pointer',
              fontSize: '0.9rem',
              fontWeight: 'bold',
              display: 'flex',
              alignItems: 'center',
              gap: '6px',
              whiteSpace: 'nowrap',
            }}
          >
            <Bell size={16} />
            Notificações
          </button>
          <button
            onClick={() => setActiveTab('templates')}
            style={{
              padding: '12px 20px',
              backgroundColor: 'transparent',
              border: 'none',
              borderBottom: activeTab === 'templates' ? '2px solid #00d4ff' : '2px solid transparent',
              color: activeTab === 'templates' ? '#00d4ff' : '#888',
              cursor: 'pointer',
              fontSize: '0.9rem',
              fontWeight: 'bold',
              display: 'flex',
              alignItems: 'center',
              gap: '6px',
              whiteSpace: 'nowrap',
            }}
          >
            <BookOpen size={16} />
            Templates
          </button>
        </div>
      </div>

      {/* Main Content */}
      {activeTab === 'permissions' ? (
        <div style={{
          display: 'grid',
          gridTemplateColumns: '350px 1fr',
          height: 'calc(100vh - 200px)',
        }}
        className="roles-main-content"
        >
          {/* Sidebar - Roles List */}
          <div style={{
            borderRight: '1px solid #333',
            backgroundColor: '#0f0f1a',
            overflowY: 'auto',
          }}
          className="roles-sidebar"
          >
            {rolesLoading ? (
              <div style={{
                padding: '20px',
                textAlign: 'center',
                color: '#888',
              }}>
                Carregando cargos...
              </div>
            ) : (
              <RolesList
                roles={roles}
                selectedRole={selectedRole}
                onRoleSelect={handleRoleSelect}
                searchQuery={searchQuery}
                filterResource={null}
                userHierarchyLevel={0}
                canEdit={canEdit}
              />
            )}
          </div>

          {/* Main Panel - Role Details */}
          <div style={{ backgroundColor: '#0a0a0f' }} className="roles-detail-panel">
            <RoleDetailPanel
              role={selectedRole}
              roleInfo={selectedRoleInfo}
              permissions={permissions}
              resources={resources}
              actions={actions}
              canEdit={canEdit(selectedRole)}
              onPermissionChange={handlePermissionToggle}
              isLoading={permissionsLoading}
            />
          </div>
        </div>
      ) : activeTab === 'audit' ? (
        <div style={{ padding: '24px' }}>
          <AuditLogViewer />
        </div>
      ) : activeTab === 'custom' ? (
        <div style={{ padding: '24px' }}>
          <CustomRolesManager
            onSuccess={(message) => setNotification({ message, type: 'success' })}
            onError={(message) => setNotification({ message, type: 'error' })}
          />
        </div>
      ) : activeTab === 'notifications' ? (
        <div style={{ padding: '24px' }}>
          <NotificationsPanel />
        </div>
      ) : activeTab === 'templates' ? (
        <div style={{ padding: '24px' }}>
          <TemplatesLibrary
            onSuccess={(message) => setNotification({ message, type: 'success' })}
            onError={(message) => setNotification({ message, type: 'error' })}
          />
        </div>
      ) : null}

      {/* Responsive Styles */}
      <style>{`
        @media (max-width: 768px) {
          .roles-main-content {
            grid-template-columns: 1fr !important;
            height: auto !important;
          }
          
          .roles-sidebar {
            border-right: none !important;
            border-bottom: 1px solid #333 !important;
            max-height: 400px !important;
          }
          
          .roles-detail-panel {
            min-height: 500px !important;
          }
        }
        
        @media (max-width: 480px) {
          .roles-sidebar {
            max-height: 300px !important;
          }
        }
      `}</style>

      {/* Modals */}
      {showExportImport && (
        <ExportImportModal
          onClose={() => setShowExportImport(false)}
          onSuccess={(message) => {
            setNotification({ message, type: 'success' });
            setShowExportImport(false);
          }}
          onError={(message) => {
            setNotification({ message, type: 'error' });
          }}
        />
      )}

      {/* Notification Toast */}
      {notification && (
        <NotificationToast
          message={notification.message}
          type={notification.type}
          duration={3000}
          onClose={() => setNotification(null)}
        />
      )}
    </div>
  );
};

export default AdminRoles;

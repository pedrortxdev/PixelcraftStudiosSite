/**
 * PermissionsMatrix Component
 * Displays a matrix of permissions (resources × actions)
 */

import React, { useMemo } from 'react';
import { Check, Loader } from 'lucide-react';
import { RESOURCE_CATEGORIES, getResourceCategory } from '../../../constants/permissions';

const PermissionsMatrix = ({
  role,
  permissions,
  resources,
  actions,
  editMode,
  onPermissionToggle,
  isLoading,
}) => {
  // Group resources by category
  const resourcesByCategory = useMemo(() => {
    const grouped = {};
    
    resources.forEach(resource => {
      const category = getResourceCategory(resource.type);
      if (!grouped[category]) {
        grouped[category] = [];
      }
      grouped[category].push(resource);
    });
    
    return grouped;
  }, [resources]);

  // Check if permission exists
  const hasPermission = (resource, action) => {
    return permissions.some(
      perm => perm.resource === resource && perm.action === action
    );
  };

  // Handle checkbox change
  const handleToggle = (resource, action) => {
    if (!editMode) return;
    onPermissionToggle(resource, action);
  };

  if (!role) {
    return (
      <div style={{
        padding: '40px',
        textAlign: 'center',
        color: '#888',
      }}>
        Selecione um cargo para visualizar suas permissões
      </div>
    );
  }

  return (
    <div style={{ 
      overflowX: 'auto',
      maxHeight: '600px',
      overflowY: 'auto',
    }}>
      {Object.entries(resourcesByCategory).map(([category, categoryResources]) => (
        <div key={category} style={{ marginBottom: '32px' }}>
          <h3 style={{
            fontSize: '1rem',
            color: '#00d4ff',
            marginBottom: '12px',
            fontFamily: "'MinecraftFont', monospace",
            lineHeight: '1.4',
            overflow: 'visible',
          }} className="minecraft-text-fix">
            {category}
          </h3>
          
          <table style={{
            width: '100%',
            borderCollapse: 'collapse',
            backgroundColor: '#0f0f1a',
            border: '1px solid #333',
          }}>
            <thead>
              <tr style={{ backgroundColor: '#1a1a2e' }}>
                <th style={{
                  padding: '12px',
                  textAlign: 'left',
                  borderBottom: '1px solid #333',
                  color: '#aaa',
                  fontSize: '0.85rem',
                }}>
                  Recurso
                </th>
                {actions.map(action => (
                  <th
                    key={action.type}
                    style={{
                      padding: '12px',
                      textAlign: 'center',
                      borderBottom: '1px solid #333',
                      borderLeft: '1px solid #333',
                      color: '#aaa',
                      fontSize: '0.85rem',
                    }}
                    title={action.label}
                  >
                    {action.label}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {categoryResources.map(resource => (
                <tr key={resource.type} style={{ borderBottom: '1px solid #222' }}>
                  <td style={{
                    padding: '12px',
                    color: '#ddd',
                    fontSize: '0.9rem',
                  }}>
                    {resource.label}
                  </td>
                  {actions.map(action => {
                    const checked = hasPermission(resource.type, action.type);
                    return (
                      <td
                        key={action.type}
                        style={{
                          padding: '12px',
                          textAlign: 'center',
                          borderLeft: '1px solid #222',
                        }}
                      >
                        <label style={{
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          cursor: editMode ? 'pointer' : 'default',
                        }}>
                          <input
                            type="checkbox"
                            checked={checked}
                            onChange={() => handleToggle(resource.type, action.type)}
                            disabled={!editMode || isLoading}
                            style={{
                              width: '18px',
                              height: '18px',
                              cursor: editMode ? 'pointer' : 'not-allowed',
                              accentColor: '#00d4ff',
                            }}
                          />
                          {isLoading && (
                            <Loader
                              size={14}
                              color="#00d4ff"
                              style={{
                                marginLeft: '8px',
                                animation: 'spin 1s linear infinite',
                              }}
                            />
                          )}
                        </label>
                      </td>
                    );
                  })}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ))}
      
      <style>{`
        @keyframes spin {
          from { transform: rotate(0deg); }
          to { transform: rotate(360deg); }
        }
      `}</style>
    </div>
  );
};

export default PermissionsMatrix;

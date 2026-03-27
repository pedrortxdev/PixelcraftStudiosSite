/**
 * CustomRolesManager Component
 * Create and manage custom roles
 */

import React, { useState, useEffect } from 'react';
import { Plus, Trash2, Edit2, Save, X } from 'lucide-react';
import api from '../../../services/api';

const CustomRolesManager = ({ onSuccess, onError }) => {
  const [customRoles, setCustomRoles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [editingRole, setEditingRole] = useState(null);
  const [formData, setFormData] = useState({
    role_name: '',
    display_name: '',
    description: '',
    color: '#999999',
    hierarchy_level: 1
  });

  useEffect(() => {
    loadCustomRoles();
  }, []);

  const loadCustomRoles = async () => {
    try {
      setLoading(true);
      const response = await api.roles.getCustomRoles();
      setCustomRoles(response.roles || []);
    } catch (error) {
      console.error('Failed to load custom roles:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = async () => {
    if (!formData.role_name || !formData.display_name || !formData.hierarchy_level) {
      onError('Preencha todos os campos obrigatórios');
      return;
    }

    try {
      await api.roles.createCustomRole(formData);
      onSuccess('Cargo customizado criado com sucesso');
      setShowCreateForm(false);
      setFormData({
        role_name: '',
        display_name: '',
        description: '',
        color: '#999999',
        hierarchy_level: 1
      });
      loadCustomRoles();
    } catch (error) {
      onError(error.message || 'Erro ao criar cargo customizado');
    }
  };

  const handleDelete = async (roleId) => {
    if (!confirm('Tem certeza que deseja deletar este cargo customizado?')) {
      return;
    }

    try {
      await api.roles.deleteCustomRole(roleId);
      onSuccess('Cargo customizado deletado com sucesso');
      loadCustomRoles();
    } catch (error) {
      onError(error.message || 'Erro ao deletar cargo customizado');
    }
  };

  return (
    <div style={{
      backgroundColor: '#0f0f1a',
      borderRadius: '8px',
      padding: '20px',
    }}>
      <div style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        marginBottom: '20px',
      }}>
        <h3 style={{
          fontSize: '1.2rem',
          color: '#00d4ff',
          fontFamily: "'MinecraftFont', monospace",
          lineHeight: '1.4',
          overflow: 'visible',
        }} className="minecraft-text-fix">
          Cargos Customizados
        </h3>

        <button
          onClick={() => setShowCreateForm(!showCreateForm)}
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
            padding: '10px 16px',
            backgroundColor: showCreateForm ? '#ff4444' : '#00ff88',
            color: '#000',
            border: 'none',
            borderRadius: '6px',
            cursor: 'pointer',
            fontWeight: 'bold',
            fontSize: '0.9rem',
          }}
        >
          {showCreateForm ? (
            <>
              <X size={16} />
              Cancelar
            </>
          ) : (
            <>
              <Plus size={16} />
              Novo Cargo
            </>
          )}
        </button>
      </div>

      {/* Create Form */}
      {showCreateForm && (
        <div style={{
          backgroundColor: '#1a1a2e',
          border: '1px solid #333',
          borderRadius: '8px',
          padding: '20px',
          marginBottom: '20px',
        }}>
          <h4 style={{ color: '#fff', marginBottom: '16px', fontSize: '1rem' }}>
            Criar Novo Cargo
          </h4>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
            <div>
              <label style={{ display: 'block', color: '#888', fontSize: '0.85rem', marginBottom: '6px' }}>
                Nome do Cargo (ID) *
              </label>
              <input
                type="text"
                value={formData.role_name}
                onChange={(e) => setFormData({ ...formData, role_name: e.target.value.toUpperCase().replace(/\s/g, '_') })}
                placeholder="CUSTOM_ROLE"
                style={{
                  width: '100%',
                  padding: '10px',
                  backgroundColor: '#0f0f1a',
                  border: '1px solid #333',
                  borderRadius: '6px',
                  color: '#fff',
                  fontSize: '0.9rem',
                }}
              />
            </div>

            <div>
              <label style={{ display: 'block', color: '#888', fontSize: '0.85rem', marginBottom: '6px' }}>
                Nome de Exibição *
              </label>
              <input
                type="text"
                value={formData.display_name}
                onChange={(e) => setFormData({ ...formData, display_name: e.target.value })}
                placeholder="Cargo Customizado"
                style={{
                  width: '100%',
                  padding: '10px',
                  backgroundColor: '#0f0f1a',
                  border: '1px solid #333',
                  borderRadius: '6px',
                  color: '#fff',
                  fontSize: '0.9rem',
                }}
              />
            </div>

            <div>
              <label style={{ display: 'block', color: '#888', fontSize: '0.85rem', marginBottom: '6px' }}>
                Descrição
              </label>
              <textarea
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                placeholder="Descrição do cargo..."
                rows={3}
                style={{
                  width: '100%',
                  padding: '10px',
                  backgroundColor: '#0f0f1a',
                  border: '1px solid #333',
                  borderRadius: '6px',
                  color: '#fff',
                  fontSize: '0.9rem',
                  resize: 'vertical',
                }}
              />
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
              <div>
                <label style={{ display: 'block', color: '#888', fontSize: '0.85rem', marginBottom: '6px' }}>
                  Cor
                </label>
                <input
                  type="color"
                  value={formData.color}
                  onChange={(e) => setFormData({ ...formData, color: e.target.value })}
                  style={{
                    width: '100%',
                    height: '42px',
                    padding: '4px',
                    backgroundColor: '#0f0f1a',
                    border: '1px solid #333',
                    borderRadius: '6px',
                    cursor: 'pointer',
                  }}
                />
              </div>

              <div>
                <label style={{ display: 'block', color: '#888', fontSize: '0.85rem', marginBottom: '6px' }}>
                  Nível de Hierarquia * (1-10)
                </label>
                <input
                  type="number"
                  min="1"
                  max="10"
                  value={formData.hierarchy_level}
                  onChange={(e) => setFormData({ ...formData, hierarchy_level: parseInt(e.target.value) })}
                  style={{
                    width: '100%',
                    padding: '10px',
                    backgroundColor: '#0f0f1a',
                    border: '1px solid #333',
                    borderRadius: '6px',
                    color: '#fff',
                    fontSize: '0.9rem',
                  }}
                />
              </div>
            </div>

            <button
              onClick={handleCreate}
              style={{
                width: '100%',
                padding: '12px',
                backgroundColor: '#00ff88',
                color: '#000',
                border: 'none',
                borderRadius: '6px',
                cursor: 'pointer',
                fontWeight: 'bold',
                fontSize: '0.9rem',
                marginTop: '8px',
              }}
            >
              <Save size={16} style={{ marginRight: '8px', verticalAlign: 'middle' }} />
              Criar Cargo
            </button>
          </div>
        </div>
      )}

      {/* Custom Roles List */}
      {loading ? (
        <div style={{ textAlign: 'center', padding: '40px', color: '#888' }}>
          Carregando cargos customizados...
        </div>
      ) : customRoles.length === 0 ? (
        <div style={{
          textAlign: 'center',
          padding: '40px',
          color: '#888',
          backgroundColor: '#1a1a2e',
          border: '1px solid #333',
          borderRadius: '8px',
        }}>
          Nenhum cargo customizado criado ainda
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
          {customRoles.map((role) => (
            <div
              key={role.id}
              style={{
                backgroundColor: '#1a1a2e',
                border: '1px solid #333',
                borderRadius: '8px',
                padding: '16px',
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
              }}
            >
              <div style={{ flex: 1 }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginBottom: '8px' }}>
                  <div
                    style={{
                      width: '12px',
                      height: '12px',
                      borderRadius: '50%',
                      backgroundColor: role.color,
                    }}
                  />
                  <span style={{ color: '#fff', fontWeight: 'bold', fontSize: '1rem' }}>
                    {role.display_name}
                  </span>
                  <span style={{
                    padding: '2px 8px',
                    backgroundColor: '#0f0f1a',
                    border: '1px solid #333',
                    borderRadius: '4px',
                    fontSize: '0.75rem',
                    color: '#888',
                  }}>
                    {role.role_name}
                  </span>
                  <span style={{
                    padding: '2px 8px',
                    backgroundColor: '#00d4ff',
                    color: '#000',
                    borderRadius: '4px',
                    fontSize: '0.75rem',
                    fontWeight: 'bold',
                  }}>
                    Nível {role.hierarchy_level}
                  </span>
                </div>
                {role.description && (
                  <div style={{ color: '#888', fontSize: '0.85rem' }}>
                    {role.description}
                  </div>
                )}
                <div style={{ color: '#666', fontSize: '0.75rem', marginTop: '8px' }}>
                  Criado em: {new Date(role.created_at).toLocaleDateString('pt-BR')}
                </div>
              </div>

              <button
                onClick={() => handleDelete(role.id)}
                style={{
                  padding: '8px 12px',
                  backgroundColor: '#ff4444',
                  color: '#fff',
                  border: 'none',
                  borderRadius: '6px',
                  cursor: 'pointer',
                  display: 'flex',
                  alignItems: 'center',
                  gap: '6px',
                }}
              >
                <Trash2 size={14} />
                Deletar
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default CustomRolesManager;

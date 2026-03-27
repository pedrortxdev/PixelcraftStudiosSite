/**
 * TemplatesLibrary Component
 * Manage and apply permission templates
 */

import React, { useState, useEffect } from 'react';
import { BookOpen, Save, Download, Upload, Trash2, Check } from 'lucide-react';
import api from '../../../services/api';

const TemplatesLibrary = ({ onSuccess, onError }) => {
  const [templates, setTemplates] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showSaveForm, setShowSaveForm] = useState(false);
  const [selectedRoles, setSelectedRoles] = useState([]);
  const [formData, setFormData] = useState({
    template_name: '',
    description: '',
    is_public: true
  });

  const allRoles = ['DIRECTION', 'ENGINEERING', 'DEVELOPMENT', 'ADMIN', 'SUPPORT'];

  useEffect(() => {
    loadTemplates();
  }, []);

  const loadTemplates = async () => {
    try {
      setLoading(true);
      const response = await api.roles.getTemplates();
      setTemplates(response.templates || []);
    } catch (error) {
      console.error('Failed to load templates:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSaveTemplate = async () => {
    if (!formData.template_name || selectedRoles.length === 0) {
      onError('Preencha o nome e selecione pelo menos um cargo');
      return;
    }

    try {
      // Export permissions for selected roles
      const exportData = await api.roles.exportPermissions(selectedRoles);

      // Save as template
      await api.roles.saveTemplate({
        template_name: formData.template_name,
        description: formData.description,
        template_data: exportData,
        is_public: formData.is_public
      });

      onSuccess('Template salvo com sucesso');
      setShowSaveForm(false);
      setFormData({
        template_name: '',
        description: '',
        is_public: true
      });
      setSelectedRoles([]);
      loadTemplates();
    } catch (error) {
      onError(error.message || 'Erro ao salvar template');
    }
  };

  const handleApplyTemplate = async (template) => {
    if (!confirm(`Aplicar o template "${template.template_name}"? Isso irá importar as permissões configuradas.`)) {
      return;
    }

    try {
      await api.roles.importPermissions(template.template_data, false);
      onSuccess('Template aplicado com sucesso');
    } catch (error) {
      onError(error.message || 'Erro ao aplicar template');
    }
  };

  const handleDownloadTemplate = (template) => {
    const blob = new Blob([JSON.stringify(template.template_data, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `template-${template.template_name.toLowerCase().replace(/\s/g, '-')}.json`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const toggleRole = (role) => {
    setSelectedRoles(prev =>
      prev.includes(role)
        ? prev.filter(r => r !== role)
        : [...prev, role]
    );
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
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
          <BookOpen size={20} color="#00d4ff" />
          <h3 style={{
            fontSize: '1.2rem',
            color: '#00d4ff',
            fontFamily: "'MinecraftFont', monospace",
            lineHeight: '1.4',
            overflow: 'visible',
          }} className="minecraft-text-fix">
            Biblioteca de Templates
          </h3>
        </div>

        <button
          onClick={() => setShowSaveForm(!showSaveForm)}
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
            padding: '10px 16px',
            backgroundColor: showSaveForm ? '#ff4444' : '#00ff88',
            color: '#000',
            border: 'none',
            borderRadius: '6px',
            cursor: 'pointer',
            fontWeight: 'bold',
            fontSize: '0.9rem',
          }}
        >
          {showSaveForm ? 'Cancelar' : (
            <>
              <Save size={16} />
              Salvar Template
            </>
          )}
        </button>
      </div>

      {/* Save Template Form */}
      {showSaveForm && (
        <div style={{
          backgroundColor: '#1a1a2e',
          border: '1px solid #333',
          borderRadius: '8px',
          padding: '20px',
          marginBottom: '20px',
        }}>
          <h4 style={{ color: '#fff', marginBottom: '16px', fontSize: '1rem' }}>
            Criar Novo Template
          </h4>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
            <div>
              <label style={{ display: 'block', color: '#888', fontSize: '0.85rem', marginBottom: '6px' }}>
                Nome do Template *
              </label>
              <input
                type="text"
                value={formData.template_name}
                onChange={(e) => setFormData({ ...formData, template_name: e.target.value })}
                placeholder="Meu Template de Permissões"
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
                placeholder="Descrição do template..."
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

            <div>
              <label style={{ display: 'block', color: '#888', fontSize: '0.85rem', marginBottom: '6px' }}>
                Selecione os cargos para incluir *
              </label>
              <div style={{
                display: 'flex',
                flexWrap: 'wrap',
                gap: '10px',
              }}>
                {allRoles.map(role => (
                  <button
                    key={role}
                    onClick={() => toggleRole(role)}
                    style={{
                      padding: '8px 16px',
                      backgroundColor: selectedRoles.includes(role) ? '#00d4ff' : '#0f0f1a',
                      color: selectedRoles.includes(role) ? '#000' : '#fff',
                      border: '1px solid #333',
                      borderRadius: '6px',
                      cursor: 'pointer',
                      fontWeight: 'bold',
                      fontSize: '0.85rem',
                    }}
                  >
                    {role}
                  </button>
                ))}
              </div>
            </div>

            <label style={{
              display: 'flex',
              alignItems: 'center',
              gap: '10px',
              color: '#fff',
              fontSize: '0.9rem',
              cursor: 'pointer',
            }}>
              <input
                type="checkbox"
                checked={formData.is_public}
                onChange={(e) => setFormData({ ...formData, is_public: e.target.checked })}
                style={{ cursor: 'pointer' }}
              />
              Template público (visível para todos os administradores)
            </label>

            <button
              onClick={handleSaveTemplate}
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
              Salvar Template
            </button>
          </div>
        </div>
      )}

      {/* Templates List */}
      {loading ? (
        <div style={{ textAlign: 'center', padding: '40px', color: '#888' }}>
          Carregando templates...
        </div>
      ) : templates.length === 0 ? (
        <div style={{
          textAlign: 'center',
          padding: '40px',
          color: '#888',
          backgroundColor: '#1a1a2e',
          border: '1px solid #333',
          borderRadius: '8px',
        }}>
          <BookOpen size={32} color="#666" style={{ marginBottom: '12px' }} />
          <div>Nenhum template disponível</div>
        </div>
      ) : (
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))',
          gap: '16px',
        }}>
          {templates.map((template) => (
            <div
              key={template.id}
              style={{
                backgroundColor: '#1a1a2e',
                border: '1px solid #333',
                borderRadius: '8px',
                padding: '20px',
                display: 'flex',
                flexDirection: 'column',
              }}
            >
              <div style={{
                display: 'flex',
                alignItems: 'flex-start',
                justifyContent: 'space-between',
                marginBottom: '12px',
              }}>
                <div style={{ flex: 1 }}>
                  <h4 style={{
                    color: '#fff',
                    fontSize: '1rem',
                    marginBottom: '6px',
                    fontWeight: 'bold',
                  }}>
                    {template.template_name}
                  </h4>
                  {template.is_public && (
                    <span style={{
                      padding: '2px 8px',
                      backgroundColor: '#00d4ff',
                      color: '#000',
                      borderRadius: '4px',
                      fontSize: '0.7rem',
                      fontWeight: 'bold',
                    }}>
                      PÚBLICO
                    </span>
                  )}
                </div>
              </div>

              {template.description && (
                <p style={{
                  color: '#888',
                  fontSize: '0.85rem',
                  marginBottom: '12px',
                  lineHeight: '1.5',
                }}>
                  {template.description}
                </p>
              )}

              <div style={{
                color: '#666',
                fontSize: '0.75rem',
                marginBottom: '16px',
              }}>
                Criado em: {new Date(template.created_at).toLocaleDateString('pt-BR')}
              </div>

              <div style={{
                display: 'flex',
                gap: '8px',
                marginTop: 'auto',
              }}>
                <button
                  onClick={() => handleApplyTemplate(template)}
                  style={{
                    flex: 1,
                    padding: '10px',
                    backgroundColor: '#00ff88',
                    color: '#000',
                    border: 'none',
                    borderRadius: '6px',
                    cursor: 'pointer',
                    fontWeight: 'bold',
                    fontSize: '0.85rem',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    gap: '6px',
                  }}
                >
                  <Check size={14} />
                  Aplicar
                </button>

                <button
                  onClick={() => handleDownloadTemplate(template)}
                  style={{
                    padding: '10px',
                    backgroundColor: '#1a1a2e',
                    border: '1px solid #333',
                    borderRadius: '6px',
                    color: '#00d4ff',
                    cursor: 'pointer',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                  }}
                  title="Baixar template"
                >
                  <Download size={14} />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default TemplatesLibrary;

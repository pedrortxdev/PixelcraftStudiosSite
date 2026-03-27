/**
 * ExportImportModal Component
 * Export and import permission configurations
 */

import React, { useState } from 'react';
import { Download, Upload, X, Copy, Check } from 'lucide-react';
import api from '../../../services/api';
import { useFocusTrap } from '../../../hooks/useFocusTrap';
import { copyToClipboard } from '../../../utils/clipboard';

const ExportImportModal = ({ onClose, onSuccess, onError }) => {
  const modalRef = useFocusTrap(true);
  const [activeTab, setActiveTab] = useState('export');
  const [exportData, setExportData] = useState(null);
  const [importData, setImportData] = useState('');
  const [overwrite, setOverwrite] = useState(false);
  const [loading, setLoading] = useState(false);
  const [copied, setCopied] = useState(false);
  const [selectedRoles, setSelectedRoles] = useState([]);

  const allRoles = ['DIRECTION', 'ENGINEERING', 'DEVELOPMENT', 'ADMIN', 'SUPPORT'];

  const handleExport = async () => {
    try {
      setLoading(true);
      const data = await api.roles.exportPermissions(selectedRoles);
      setExportData(data);
    } catch (error) {
      onError(error.message || 'Erro ao exportar permissões');
    } finally {
      setLoading(false);
    }
  };

  const handleCopy = async () => {
    const success = await copyToClipboard(JSON.stringify(exportData, null, 2));
    if (success) {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } else {
      onError('Erro ao copiar para a área de transferência');
    }
  };

  const handleDownload = () => {
    const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `permissions-export-${new Date().toISOString().split('T')[0]}.json`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const handleImport = async () => {
    try {
      const data = JSON.parse(importData);
      setLoading(true);
      const result = await api.roles.importPermissions(data, overwrite);
      onSuccess(result.message);
      onClose();
    } catch (error) {
      if (error instanceof SyntaxError) {
        onError('JSON inválido. Verifique o formato dos dados.');
      } else {
        onError(error.message || 'Erro ao importar permissões');
      }
    } finally {
      setLoading(false);
    }
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
      position: 'fixed',
      top: 0,
      left: 0,
      right: 0,
      bottom: 0,
      backgroundColor: 'rgba(0, 0, 0, 0.8)',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      zIndex: 1000,
      padding: '20px',
    }}>
      <div
        ref={modalRef}
        style={{
          backgroundColor: '#0f0f1a',
          borderRadius: '12px',
          width: '100%',
          maxWidth: '600px',
          maxHeight: '90vh',
          overflow: 'auto',
        }}>
        {/* Header */}
        <div style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          padding: '20px',
          borderBottom: '1px solid #333',
        }}>
          <h3 style={{
            fontSize: '1.2rem',
            color: '#00d4ff',
            fontFamily: "'MinecraftFont', monospace",
            lineHeight: '1.4',
            overflow: 'visible',
          }} className="minecraft-text-fix">
            Exportar / Importar Permissões
          </h3>
          <button
            onClick={onClose}
            style={{
              background: 'none',
              border: 'none',
              color: '#888',
              cursor: 'pointer',
              padding: '5px',
            }}
            aria-label="Fechar">
            <X size={24} />
          </button>
        </div>

        {/* Tabs */}
        <div style={{
          display: 'flex',
          borderBottom: '1px solid #333',
        }}>
          <button
            onClick={() => setActiveTab('export')}
            style={{
              flex: 1,
              padding: '15px',
              backgroundColor: activeTab === 'export' ? '#1a1a2e' : 'transparent',
              color: activeTab === 'export' ? '#00d4ff' : '#888',
              border: 'none',
              borderBottom: activeTab === 'export' ? '2px solid #00d4ff' : 'none',
              cursor: 'pointer',
              fontWeight: 'bold',
            }}
          >
            <Download size={16} style={{ marginRight: '8px', verticalAlign: 'middle' }} />
            Exportar
          </button>
          <button
            onClick={() => setActiveTab('import')}
            style={{
              flex: 1,
              padding: '15px',
              backgroundColor: activeTab === 'import' ? '#1a1a2e' : 'transparent',
              color: activeTab === 'import' ? '#00d4ff' : '#888',
              border: 'none',
              borderBottom: activeTab === 'import' ? '2px solid #00d4ff' : 'none',
              cursor: 'pointer',
              fontWeight: 'bold',
            }}
          >
            <Upload size={16} style={{ marginRight: '8px', verticalAlign: 'middle' }} />
            Importar
          </button>
        </div>

        {/* Content */}
        <div style={{ padding: '20px' }}>
          {activeTab === 'export' ? (
            <>
              <p style={{ color: '#888', fontSize: '0.9rem', marginBottom: '15px' }}>
                Selecione os cargos para exportar suas permissões:
              </p>

              <div style={{
                display: 'flex',
                flexWrap: 'wrap',
                gap: '10px',
                marginBottom: '20px',
              }}>
                {allRoles.map(role => (
                  <button
                    key={role}
                    onClick={() => toggleRole(role)}
                    style={{
                      padding: '8px 16px',
                      backgroundColor: selectedRoles.includes(role) ? '#00d4ff' : '#1a1a2e',
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

              <button
                onClick={handleExport}
                disabled={loading || selectedRoles.length === 0}
                style={{
                  width: '100%',
                  padding: '12px',
                  backgroundColor: selectedRoles.length > 0 && !loading ? '#00ff88' : '#333',
                  color: selectedRoles.length > 0 && !loading ? '#000' : '#666',
                  border: 'none',
                  borderRadius: '6px',
                  cursor: selectedRoles.length > 0 && !loading ? 'pointer' : 'not-allowed',
                  fontWeight: 'bold',
                  marginBottom: '20px',
                }}
              >
                {loading ? 'Exportando...' : 'Exportar Permissões'}
              </button>

              {exportData && (
                <div>
                  <div style={{
                    display: 'flex',
                    gap: '10px',
                    marginBottom: '10px',
                  }}>
                    <button
                      onClick={handleCopy}
                      style={{
                        flex: 1,
                        padding: '10px',
                        backgroundColor: '#1a1a2e',
                        color: '#fff',
                        border: '1px solid #333',
                        borderRadius: '6px',
                        cursor: 'pointer',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        gap: '8px',
                      }}
                    >
                      {copied ? <Check size={16} /> : <Copy size={16} />}
                      {copied ? 'Copiado!' : 'Copiar JSON'}
                    </button>
                    <button
                      onClick={handleDownload}
                      style={{
                        flex: 1,
                        padding: '10px',
                        backgroundColor: '#00d4ff',
                        color: '#000',
                        border: 'none',
                        borderRadius: '6px',
                        cursor: 'pointer',
                        fontWeight: 'bold',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        gap: '8px',
                      }}
                    >
                      <Download size={16} />
                      Baixar Arquivo
                    </button>
                  </div>

                  <pre style={{
                    backgroundColor: '#1a1a2e',
                    border: '1px solid #333',
                    borderRadius: '6px',
                    padding: '15px',
                    color: '#00ff88',
                    fontSize: '0.8rem',
                    maxHeight: '300px',
                    overflow: 'auto',
                  }}>
                    {JSON.stringify(exportData, null, 2)}
                  </pre>
                </div>
              )}
            </>
          ) : (
            <>
              <p style={{ color: '#888', fontSize: '0.9rem', marginBottom: '15px' }}>
                Cole o JSON de configuração de permissões:
              </p>

              <textarea
                value={importData}
                onChange={(e) => setImportData(e.target.value)}
                placeholder='{"permissions": {...}}'
                style={{
                  width: '100%',
                  minHeight: '300px',
                  padding: '15px',
                  backgroundColor: '#1a1a2e',
                  border: '1px solid #333',
                  borderRadius: '6px',
                  color: '#00ff88',
                  fontSize: '0.85rem',
                  fontFamily: 'monospace',
                  resize: 'vertical',
                  marginBottom: '15px',
                }}
              />

              <label style={{
                display: 'flex',
                alignItems: 'center',
                gap: '10px',
                color: '#fff',
                fontSize: '0.9rem',
                marginBottom: '20px',
                cursor: 'pointer',
              }}>
                <input
                  type="checkbox"
                  checked={overwrite}
                  onChange={(e) => setOverwrite(e.target.checked)}
                  style={{ cursor: 'pointer' }}
                />
                Sobrescrever permissões existentes
              </label>

              <button
                onClick={handleImport}
                disabled={loading || !importData.trim()}
                style={{
                  width: '100%',
                  padding: '12px',
                  backgroundColor: importData.trim() && !loading ? '#00ff88' : '#333',
                  color: importData.trim() && !loading ? '#000' : '#666',
                  border: 'none',
                  borderRadius: '6px',
                  cursor: importData.trim() && !loading ? 'pointer' : 'not-allowed',
                  fontWeight: 'bold',
                }}
              >
                {loading ? 'Importando...' : 'Importar Permissões'}
              </button>
            </>
          )}
        </div>
      </div>
    </div>
  );
};

export default ExportImportModal;

/**
 * AuditLogViewer Component
 * Displays permission change history
 */

import React, { useState, useEffect } from 'react';
import { Clock, User, Filter } from 'lucide-react';
import api from '../../../services/api';

const AuditLogViewer = ({ roleFilter = '' }) => {
  const [logs, setLogs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [selectedRole, setSelectedRole] = useState(roleFilter);

  useEffect(() => {
    loadLogs();
  }, [page, selectedRole]);

  const loadLogs = async () => {
    try {
      setLoading(true);
      const response = await api.roles.getAuditLog(page, 50, selectedRole);
      setLogs(response.logs || []);
      setTotal(response.total || 0);
    } catch (error) {
      console.error('Failed to load audit log:', error);
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (dateString) => {
    const date = new Date(dateString);
    return date.toLocaleString('pt-BR');
  };

  const getOperationColor = (operation) => {
    switch (operation) {
      case 'ADD': return '#00ff88';
      case 'REMOVE': return '#ff4444';
      case 'INHERIT': return '#00d4ff';
      case 'REMOVE_INHERITED': return '#ff9900';
      default: return '#888';
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
          Histórico de Mudanças
        </h3>

        <div style={{ display: 'flex', gap: '10px', alignItems: 'center' }}>
          <Filter size={16} color="#888" />
          <select
            value={selectedRole}
            onChange={(e) => {
              setSelectedRole(e.target.value);
              setPage(1);
            }}
            style={{
              padding: '8px 12px',
              backgroundColor: '#1a1a2e',
              border: '1px solid #333',
              borderRadius: '6px',
              color: '#fff',
              fontSize: '0.9rem',
            }}
          >
            <option value="">Todos os cargos</option>
            <option value="DIRECTION">Direção</option>
            <option value="ENGINEERING">Engenharia</option>
            <option value="DEVELOPMENT">Desenvolvimento</option>
            <option value="ADMIN">Admin</option>
            <option value="SUPPORT">Suporte</option>
          </select>
        </div>
      </div>

      {loading ? (
        <div style={{ textAlign: 'center', padding: '40px', color: '#888' }}>
          Carregando histórico...
        </div>
      ) : logs.length === 0 ? (
        <div style={{ textAlign: 'center', padding: '40px', color: '#888' }}>
          Nenhuma mudança registrada
        </div>
      ) : (
        <>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
            {logs.map((log) => (
              <div
                key={log.id}
                style={{
                  backgroundColor: '#1a1a2e',
                  border: '1px solid #333',
                  borderRadius: '6px',
                  padding: '15px',
                }}
              >
                <div style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'flex-start',
                  marginBottom: '10px',
                }}>
                  <div>
                    <span style={{
                      color: getOperationColor(log.operation),
                      fontWeight: 'bold',
                      fontSize: '0.9rem',
                    }}>
                      {log.operation}
                    </span>
                    <span style={{ color: '#888', margin: '0 8px' }}>•</span>
                    <span style={{ color: '#fff', fontWeight: 'bold' }}>
                      {log.role}
                    </span>
                  </div>
                  <div style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '6px',
                    color: '#888',
                    fontSize: '0.85rem',
                  }}>
                    <Clock size={14} />
                    {formatDate(log.performed_at)}
                  </div>
                </div>

                <div style={{ color: '#ccc', fontSize: '0.9rem', marginBottom: '8px' }}>
                  {log.resource} • {log.action}
                </div>

                {log.reason && (
                  <div style={{
                    color: '#888',
                    fontSize: '0.85rem',
                    fontStyle: 'italic',
                  }}>
                    {log.reason}
                  </div>
                )}

                {log.performed_by && (
                  <div style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '6px',
                    color: '#888',
                    fontSize: '0.85rem',
                    marginTop: '8px',
                  }}>
                    <User size={14} />
                    Realizado por: {log.performed_by}
                  </div>
                )}
              </div>
            ))}
          </div>

          {total > 50 && (
            <div style={{
              display: 'flex',
              justifyContent: 'center',
              gap: '10px',
              marginTop: '20px',
            }}>
              <button
                onClick={() => setPage(p => Math.max(1, p - 1))}
                disabled={page === 1}
                style={{
                  padding: '8px 16px',
                  backgroundColor: page === 1 ? '#1a1a2e' : '#00d4ff',
                  color: page === 1 ? '#666' : '#000',
                  border: 'none',
                  borderRadius: '6px',
                  cursor: page === 1 ? 'not-allowed' : 'pointer',
                  fontWeight: 'bold',
                }}
              >
                Anterior
              </button>
              <span style={{ color: '#888', padding: '8px 16px' }}>
                Página {page} de {Math.ceil(total / 50)}
              </span>
              <button
                onClick={() => setPage(p => p + 1)}
                disabled={page >= Math.ceil(total / 50)}
                style={{
                  padding: '8px 16px',
                  backgroundColor: page >= Math.ceil(total / 50) ? '#1a1a2e' : '#00d4ff',
                  color: page >= Math.ceil(total / 50) ? '#666' : '#000',
                  border: 'none',
                  borderRadius: '6px',
                  cursor: page >= Math.ceil(total / 50) ? 'not-allowed' : 'pointer',
                  fontWeight: 'bold',
                }}
              >
                Próxima
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
};

export default AuditLogViewer;

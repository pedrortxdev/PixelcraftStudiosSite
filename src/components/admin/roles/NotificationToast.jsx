/**
 * NotificationToast Component
 * Displays temporary notifications
 */

import React, { useEffect } from 'react';
import { CheckCircle, XCircle, X } from 'lucide-react';

const NotificationToast = ({ message, type = 'success', duration = 3000, onClose }) => {
  useEffect(() => {
    if (duration > 0) {
      const timer = setTimeout(() => {
        onClose();
      }, duration);

      return () => clearTimeout(timer);
    }
  }, [duration, onClose]);

  const styles = {
    success: {
      backgroundColor: '#1a4d2e',
      borderColor: '#00d415',
      iconColor: '#00d415',
    },
    error: {
      backgroundColor: '#4d1a1a',
      borderColor: '#ff3f00',
      iconColor: '#ff3f00',
    },
  };

  const style = styles[type] || styles.success;

  return (
    <div
      style={{
        position: 'fixed',
        top: '20px',
        right: '20px',
        zIndex: 9999,
        backgroundColor: style.backgroundColor,
        border: `2px solid ${style.borderColor}`,
        borderRadius: '8px',
        padding: '16px 20px',
        minWidth: '300px',
        maxWidth: '500px',
        boxShadow: `0 4px 12px rgba(0, 0, 0, 0.3), 0 0 20px ${style.borderColor}40`,
        animation: 'slideIn 0.3s ease-out',
        display: 'flex',
        alignItems: 'center',
        gap: '12px',
      }}
    >
      {type === 'success' ? (
        <CheckCircle size={24} color={style.iconColor} />
      ) : (
        <XCircle size={24} color={style.iconColor} />
      )}
      
      <div style={{ flex: 1, color: '#fff', fontSize: '0.9rem' }}>
        {message}
      </div>
      
      <button
        onClick={onClose}
        style={{
          background: 'none',
          border: 'none',
          cursor: 'pointer',
          padding: '4px',
          display: 'flex',
          alignItems: 'center',
          color: '#888',
        }}
       aria-label="Fechar">
        <X size={18} />
      </button>

      <style>{`
        @keyframes slideIn {
          from {
            transform: translateX(100%);
            opacity: 0;
          }
          to {
            transform: translateX(0);
            opacity: 1;
          }
        }
      `}</style>
    </div>
  );
};

export default NotificationToast;

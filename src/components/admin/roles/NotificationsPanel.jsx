/**
 * NotificationsPanel Component
 * Display permission change notifications
 */

import React, { useState, useEffect } from 'react';
import { Bell, Check, Clock, AlertCircle } from 'lucide-react';
import api from '../../../services/api';

const NotificationsPanel = () => {
  const [notifications, setNotifications] = useState([]);
  const [loading, setLoading] = useState(true);
  const [unreadCount, setUnreadCount] = useState(0);

  useEffect(() => {
    loadNotifications();
  }, []);

  const loadNotifications = async () => {
    try {
      setLoading(true);
      const response = await api.roles.getNotifications();
      const notifs = response.notifications || [];
      setNotifications(notifs);
      setUnreadCount(notifs.filter(n => !n.is_read).length);
    } catch (error) {
      console.error('Failed to load notifications:', error);
    } finally {
      setLoading(false);
    }
  };

  const markAsRead = async (notificationId) => {
    try {
      await api.roles.markNotificationAsRead(notificationId);
      setNotifications(prev =>
        prev.map(n => n.id === notificationId ? { ...n, is_read: true } : n)
      );
      setUnreadCount(prev => Math.max(0, prev - 1));
    } catch (error) {
      console.error('Failed to mark notification as read:', error);
    }
  };

  const markAllAsRead = async () => {
    const unreadNotifs = notifications.filter(n => !n.is_read);
    for (const notif of unreadNotifs) {
      await markAsRead(notif.id);
    }
  };

  const getNotificationIcon = (type) => {
    switch (type) {
      case 'PERMISSION_ADDED':
        return <Check size={16} color="#00ff88" />;
      case 'PERMISSION_REMOVED':
        return <AlertCircle size={16} color="#ff4444" />;
      case 'ROLE_ASSIGNED':
        return <Bell size={16} color="#00d4ff" />;
      default:
        return <Bell size={16} color="#888" />;
    }
  };

  const getNotificationColor = (type) => {
    switch (type) {
      case 'PERMISSION_ADDED':
        return '#00ff88';
      case 'PERMISSION_REMOVED':
        return '#ff4444';
      case 'ROLE_ASSIGNED':
        return '#00d4ff';
      default:
        return '#888';
    }
  };

  const formatDate = (dateString) => {
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now - date;
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return 'Agora mesmo';
    if (diffMins < 60) return `${diffMins} min atrás`;
    if (diffHours < 24) return `${diffHours}h atrás`;
    if (diffDays < 7) return `${diffDays}d atrás`;
    return date.toLocaleDateString('pt-BR');
  };

  return (
    <div style={{
      backgroundColor: '#0f0f1a',
      borderRadius: '8px',
      padding: '20px',
      maxHeight: '600px',
      display: 'flex',
      flexDirection: 'column',
    }}>
      <div style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        marginBottom: '20px',
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
          <Bell size={20} color="#00d4ff" />
          <h3 style={{
            fontSize: '1.2rem',
            color: '#00d4ff',
            fontFamily: "'MinecraftFont', monospace",
            lineHeight: '1.4',
            overflow: 'visible',
          }} className="minecraft-text-fix">
            Notificações
          </h3>
          {unreadCount > 0 && (
            <span style={{
              padding: '2px 8px',
              backgroundColor: '#ff4444',
              color: '#fff',
              borderRadius: '12px',
              fontSize: '0.75rem',
              fontWeight: 'bold',
            }}>
              {unreadCount}
            </span>
          )}
        </div>

        {unreadCount > 0 && (
          <button
            onClick={markAllAsRead}
            style={{
              padding: '6px 12px',
              backgroundColor: '#1a1a2e',
              border: '1px solid #333',
              borderRadius: '6px',
              color: '#00d4ff',
              cursor: 'pointer',
              fontSize: '0.85rem',
            }}
          >
            Marcar todas como lidas
          </button>
        )}
      </div>

      {loading ? (
        <div style={{ textAlign: 'center', padding: '40px', color: '#888' }}>
          Carregando notificações...
        </div>
      ) : notifications.length === 0 ? (
        <div style={{
          textAlign: 'center',
          padding: '40px',
          color: '#888',
          backgroundColor: '#1a1a2e',
          border: '1px solid #333',
          borderRadius: '8px',
        }}>
          <Bell size={32} color="#666" style={{ marginBottom: '12px' }} />
          <div>Nenhuma notificação</div>
        </div>
      ) : (
        <div style={{
          flex: 1,
          overflowY: 'auto',
          display: 'flex',
          flexDirection: 'column',
          gap: '10px',
        }}>
          {notifications.map((notification) => (
            <div
              key={notification.id}
              onClick={() => !notification.is_read && markAsRead(notification.id)}
              style={{
                backgroundColor: notification.is_read ? '#1a1a2e' : '#252540',
                border: `1px solid ${notification.is_read ? '#333' : getNotificationColor(notification.notification_type)}`,
                borderRadius: '8px',
                padding: '16px',
                cursor: notification.is_read ? 'default' : 'pointer',
                transition: 'all 0.2s ease',
                position: 'relative',
              }}
            >
              {!notification.is_read && (
                <div
                  style={{
                    position: 'absolute',
                    top: '12px',
                    right: '12px',
                    width: '8px',
                    height: '8px',
                    borderRadius: '50%',
                    backgroundColor: getNotificationColor(notification.notification_type),
                  }}
                />
              )}

              <div style={{
                display: 'flex',
                alignItems: 'flex-start',
                gap: '12px',
              }}>
                <div style={{
                  padding: '8px',
                  backgroundColor: '#0f0f1a',
                  borderRadius: '6px',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                }}>
                  {getNotificationIcon(notification.notification_type)}
                </div>

                <div style={{ flex: 1 }}>
                  <div style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '8px',
                    marginBottom: '6px',
                  }}>
                    <span style={{
                      padding: '2px 8px',
                      backgroundColor: '#0f0f1a',
                      border: '1px solid #333',
                      borderRadius: '4px',
                      fontSize: '0.75rem',
                      color: getNotificationColor(notification.notification_type),
                      fontWeight: 'bold',
                    }}>
                      {notification.role}
                    </span>
                    <span style={{
                      fontSize: '0.75rem',
                      color: '#666',
                      display: 'flex',
                      alignItems: 'center',
                      gap: '4px',
                    }}>
                      <Clock size={12} />
                      {formatDate(notification.created_at)}
                    </span>
                  </div>

                  <div style={{
                    color: notification.is_read ? '#aaa' : '#fff',
                    fontSize: '0.9rem',
                    lineHeight: '1.5',
                  }}>
                    {notification.message}
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default NotificationsPanel;

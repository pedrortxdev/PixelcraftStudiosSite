import React, { useEffect, useState } from 'react';
import { motion } from 'framer-motion';
import { useAuth } from '../../context/AuthContext';
import RoleBadge from '../../components/RoleBadge';
import {
    Headphones,
    MessageCircle,
    Clock,
    CheckCircle,
    AlertCircle,
    Send,
    Star,
    User,
    Filter,
    Search,
    RefreshCw,
    ChevronLeft
} from 'lucide-react';

import { adminAPI } from '../../services/api';
import { getAvatarUrl } from '../../utils/formatAvatarUrl';

const AdminSupport = () => {
    const { token, user } = useAuth();
    const [tickets, setTickets] = useState([]);
    const [loading, setLoading] = useState(true);
    const [selectedTicket, setSelectedTicket] = useState(null);
    const [newMessage, setNewMessage] = useState('');
    const [sending, setSending] = useState(false);
    const [filters, setFilters] = useState({ status: '', priority: '', search: '' });
    const [stats, setStats] = useState(null);
    const [page, setPage] = useState(1);
    const [totalPages, setTotalPages] = useState(1);

    const statusConfig = {
        OPEN: { label: 'Aberto', color: '#22c55e', icon: AlertCircle },
        IN_PROGRESS: { label: 'Em Atendimento', color: '#3b82f6', icon: MessageCircle },
        WAITING_RESPONSE: { label: 'Aguardando', color: '#eab308', icon: Clock },
        RESOLVED: { label: 'Resolvido', color: '#8b5cf6', icon: CheckCircle },
        CLOSED: { label: 'Fechado', color: '#6b7280', icon: CheckCircle },
    };

    const categoryLabels = {
        GENERAL: 'Dúvida Geral',
        SUBSCRIPTION: 'Assinatura',
        PAYMENT: 'Pagamento',
        TECHNICAL: 'Suporte Técnico',
        BILLING: 'Faturamento',
        OTHER: 'Outros',
    };

    const fetchTickets = async () => {
        setLoading(true);
        try {
            const params = new URLSearchParams({ page, limit: 20 });
            if (filters.status) params.append('status', filters.status);
            if (filters.priority) params.append('priority', filters.priority);

            const data = await adminAPI.getSupportTickets(params.toString());
            setTickets(data.tickets || []);
            setTotalPages(data.total_pages || 1);
        } catch (err) {
            console.error('Error fetching tickets:', err);
        } finally {
            setLoading(false);
        }
    };

    const fetchStats = async () => {
        try {
            const data = await adminAPI.getSupportStats();
            setStats(data);
        } catch (err) {
            console.error('Error fetching stats:', err);
        }
    };

    const fetchTicketDetails = async (ticketId) => {
        try {
            const data = await adminAPI.getSupportTicket(ticketId);
            setSelectedTicket(data);
        } catch (err) {
            console.error('Error fetching ticket:', err);
        }
    };

    const sendMessage = async () => {
        if (!newMessage.trim() || !selectedTicket) return;
        setSending(true);
        try {
            const msg = await adminAPI.sendSupportMessage(selectedTicket.id, newMessage);
            setSelectedTicket({
                ...selectedTicket,
                messages: [...(selectedTicket.messages || []), msg],
            });
            setNewMessage('');
        } catch (err) {
            console.error('Error sending message:', err);
        } finally {
            setSending(false);
        }
    };

    const updateStatus = async (status) => {
        if (!selectedTicket) return;
        try {
            await adminAPI.updateSupportStatus(selectedTicket.id, status);
            setSelectedTicket({ ...selectedTicket, status });
            fetchTickets();
            fetchStats();
        } catch (err) {
            console.error('Error updating status:', err);
            alert(`Erro ao atualizar status: ${err.message || 'Erro de conexão'}`);
        }
    };

    const assignToMe = async () => {
        if (!selectedTicket) return;
        try {
            await adminAPI.assignSupportTicket(selectedTicket.id, user.id);
            fetchTicketDetails(selectedTicket.id);
            fetchTickets();
        } catch (err) {
            console.error('Error assigning ticket:', err);
        }
    };

    const releaseTicket = async () => {
        if (!selectedTicket) return;
        try {
            await adminAPI.releaseSupportTicket(selectedTicket.id);
            fetchTicketDetails(selectedTicket.id);
            fetchTickets();
        } catch (err) {
            console.error('Error releasing ticket:', err);
        }
    };

    useEffect(() => {
        fetchTickets();
        fetchStats();
    }, [page, filters.status, filters.priority]);

    const renderStars = (priority) => {
        const fullStars = Math.floor(priority);
        const hasHalf = priority % 1 >= 0.5;
        return (
            <div style={{ display: 'flex', gap: '2px' }}>
                {[...Array(fullStars)].map((_, i) => (
                    <Star key={i} size={14} fill="#eab308" color="#eab308" />
                ))}
                {hasHalf && <Star size={14} fill="#eab308" color="#eab308" style={{ clipPath: 'inset(0 50% 0 0)' }} />}
            </div>
        );
    };

    const formatDate = (dateString) => {
        try {
            if (!dateString) return '—';
            const date = new Date(dateString);
            if (isNaN(date.getTime())) return '—';
            return new Intl.DateTimeFormat('pt-BR', { 
                day: '2-digit', month: '2-digit', year: 'numeric',
                hour: '2-digit', minute: '2-digit' 
            }).format(date);
        } catch (e) {
            return '—';
        }
    };

    const getFullAvatarUrl = (path) => {
        if (!path) return null;
        if (path.startsWith('http')) return path;
        return getAvatarUrl(path);
    };

    const ws = React.useRef(null);
    const [wsStatus, setWsStatus] = useState('DISCONNECTED'); // CONNECTING, CONNECTED, DISCONNECTED

    useEffect(() => {
        if (!selectedTicket?.id) return;

        let reconnectTimer = null;
        let shouldReconnect = true;
        let reconnectCount = 0;

        const connectWS = () => {
            if (!shouldReconnect) return;

            // Use the API_BASE_URL from api.js but replace http with ws
            const API_BASE_URL = 'https://api.pixelcraft-studio.store/api/v1';
            const wsBase = API_BASE_URL.replace(/^http/, 'ws');
            const wsUrl = `${wsBase}/ws?ticket_id=${selectedTicket.id}`;

            if (ws.current) {
                ws.current.onclose = null;
                ws.current.close();
            }

            console.log('Connecting to WS:', wsUrl);
            setWsStatus('CONNECTING');
            ws.current = new WebSocket(wsUrl);

            ws.current.onopen = () => {
                console.log('WS Connected');
                setWsStatus('CONNECTED');
                reconnectCount = 0;
            };

            const triggerReconnect = () => {
                setWsStatus('DISCONNECTED');
                if (shouldReconnect && reconnectCount < 5) {
                    const delay = Math.min(1000 * Math.pow(2, reconnectCount), 10000);
                    reconnectCount++;
                    console.log(`WS Disconnected. Reconnecting in ${delay}ms...`);
                    reconnectTimer = setTimeout(connectWS, delay);
                }
            };

            ws.current.onclose = triggerReconnect;
            ws.current.onerror = triggerReconnect;

            ws.current.onmessage = (event) => {
                try {
                    const data = JSON.parse(event.data);
                    console.log('WS Message:', data);
                    if (data.type === 'new_message') {
                        setSelectedTicket(prev => {
                            if (!prev || prev.id !== selectedTicket.id) return prev;
                            if (prev.messages.some(m => m.id === data.message.id)) return prev;
                            return {
                                ...prev,
                                messages: [...prev.messages, data.message]
                            };
                        });
                    } else if (data.type === 'ticket_updated') {
                        setSelectedTicket(prev => ({ ...prev, ...data.data }));
                        fetchTickets();
                    }
                } catch (err) {
                    console.error('WS Error:', err);
                }
            };
        };

        connectWS();

        return () => {
            shouldReconnect = false;
            if (reconnectTimer) clearTimeout(reconnectTimer);
            if (ws.current) {
                ws.current.onclose = null;
                ws.current.close();
                ws.current = null;
            }
        };
    }, [selectedTicket?.id]);

    const messagesEndRef = React.useRef(null);

    const scrollToBottom = () => {
        messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    };

    useEffect(() => {
        scrollToBottom();
    }, [selectedTicket?.messages]);

    const styles = {
        container: { display: 'flex', gap: '1.5rem', height: 'calc(100vh - 120px)' },
        sidebar: {
            width: '380px',
            background: 'rgba(15, 20, 35, 0.9)',
            borderRadius: '16px',
            border: '1px solid rgba(88, 58, 255, 0.2)',
            display: 'flex',
            flexDirection: 'column',
            overflow: 'hidden',
        },
        chatArea: {
            flex: 1,
            background: 'rgba(15, 20, 35, 0.9)',
            borderRadius: '16px',
            border: '1px solid rgba(88, 58, 255, 0.2)',
            display: 'flex',
            flexDirection: 'column',
            overflow: 'hidden',
        },
        statsBar: {
            display: 'flex',
            gap: '1rem',
            marginBottom: '1.5rem',
        },
        statCard: {
            flex: 1,
            padding: 'var(--btn-padding-md)',
            background: 'rgba(15, 20, 35, 0.8)',
            borderRadius: '12px',
            border: '1px solid rgba(88, 58, 255, 0.15)',
            textAlign: 'center',
        },
        filterBar: {
            padding: 'var(--btn-padding-md)',
            borderBottom: '1px solid rgba(255, 255, 255, 0.1)',
            display: 'flex',
            gap: '0.5rem',
            flexWrap: 'wrap',
        },
        select: {
            padding: '0.5rem',
            background: 'rgba(0, 0, 0, 0.3)',
            border: '1px solid rgba(88, 58, 255, 0.3)',
            borderRadius: '6px',
            color: '#F8F9FA',
            fontSize: '0.85rem',
        },
        ticketItem: {
            padding: 'var(--btn-padding-md)',
            borderBottom: '1px solid rgba(255, 255, 255, 0.05)',
            cursor: 'pointer',
            transition: 'background 0.2s',
        },
        messagesArea: {
            flex: 1,
            padding: '1.5rem',
            overflowY: 'auto',
            display: 'flex',
            flexDirection: 'column',
            gap: '1rem',
        },
        messageInput: {
            padding: 'var(--btn-padding-md)',
            borderTop: '1px solid rgba(255, 255, 255, 0.1)',
            display: 'flex',
            gap: '0.75rem',
        },
        input: {
            flex: 1,
            padding: '0.75rem 1rem',
            background: 'rgba(0, 0, 0, 0.3)',
            border: '1px solid rgba(88, 58, 255, 0.3)',
            borderRadius: '8px',
            color: '#F8F9FA',
            fontSize: '0.95rem',
        },
        sendBtn: {
            padding: '0.75rem 1.25rem',
            background: 'var(--gradient-primary)',
            border: 'none',
            borderRadius: '8px',
            color: 'white',
            fontWeight: 600,
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            gap: '0.5rem',
        },
        message: {
            maxWidth: '70%',
            padding: 'var(--btn-padding-md)',
            borderRadius: '12px',
            fontSize: '0.95rem',
            lineHeight: 1.5,
        },
        actionBtn: {
            padding: '0.5rem 0.75rem',
            background: 'rgba(88, 58, 255, 0.2)',
            border: '1px solid rgba(88, 58, 255, 0.4)',
            borderRadius: '6px',
            color: '#F8F9FA',
            cursor: 'pointer',
            fontSize: '0.8rem',
        },
    };

    return (
        <div style={{ padding: '1.5rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
                <h1 style={{ margin: 0, color: '#F8F9FA', fontSize: 'var(--title-h4)' }}>
                    <Headphones size={24} style={{ marginRight: '0.75rem', verticalAlign: 'middle' }} />
                    Canal de Atendimento
                </h1>
                <motion.button
                    style={styles.actionBtn}
                    whileHover={{ scale: 1.05 }}
                    onClick={() => { fetchTickets(); fetchStats(); }}
                >
                    <RefreshCw size={14} style={{ marginRight: '4px' }} /> Atualizar
                </motion.button>
            </div>

            {/* Stats Bar */}
            {stats && (
                <div style={styles.statsBar}>
                    <div style={styles.statCard}>
                        <div style={{ fontSize: '1.75rem', fontWeight: 700, color: '#22c55e' }}>{stats.open || 0}</div>
                        <div style={{ fontSize: '0.8rem', color: '#888' }}>Abertos</div>
                    </div>
                    <div style={styles.statCard}>
                        <div style={{ fontSize: '1.75rem', fontWeight: 700, color: '#3b82f6' }}>{stats.in_progress || 0}</div>
                        <div style={{ fontSize: '0.8rem', color: '#888' }}>Em Atendimento</div>
                    </div>
                    <div style={styles.statCard}>
                        <div style={{ fontSize: '1.75rem', fontWeight: 700, color: '#eab308' }}>{stats.waiting || 0}</div>
                        <div style={{ fontSize: '0.8rem', color: '#888' }}>Aguardando</div>
                    </div>
                    <div style={styles.statCard}>
                        <div style={{ fontSize: '1.75rem', fontWeight: 700, color: '#8b5cf6' }}>{stats.resolved || 0}</div>
                        <div style={{ fontSize: '0.8rem', color: '#888' }}>Resolvidos</div>
                    </div>
                </div>
            )}

            <div className={`admin-support-container ${selectedTicket ? 'ticket-selected' : ''}`} style={styles.container}>
                {/* Ticket List */}
                <div className="admin-support-sidebar" style={styles.sidebar}>
                    <div style={styles.filterBar}>
                        <select
                            style={styles.select}
                            value={filters.status}
                            onChange={(e) => setFilters({ ...filters, status: e.target.value })}
                        >
                            <option value="">Todos Status</option>
                            {Object.entries(statusConfig).map(([key, val]) => (
                                <option key={key} value={key}>{val.label}</option>
                            ))}
                        </select>
                        <select
                            style={styles.select}
                            value={filters.priority}
                            onChange={(e) => setFilters({ ...filters, priority: e.target.value })}
                        >
                            <option value="">Todas Prioridades</option>
                            <option value="5">⭐⭐⭐⭐⭐ (5)</option>
                            <option value="4">⭐⭐⭐⭐ (4)</option>
                            <option value="3">⭐⭐⭐ (3)</option>
                            <option value="2">⭐⭐ (2)</option>
                            <option value="1">⭐ (1)</option>
                        </select>
                    </div>
                    <div style={{ flex: 1, overflowY: 'auto' }}>
                        {loading ? (
                            <div style={{ padding: '2rem', textAlign: 'center', color: '#888' }}>Carregando...</div>
                        ) : tickets.length === 0 ? (
                            <div style={{ padding: '2rem', textAlign: 'center', color: '#888' }}>Nenhum ticket</div>
                        ) : (
                            tickets.map((ticket) => {
                                const status = statusConfig[ticket.status] || statusConfig.OPEN;
                                const StatusIcon = status.icon;
                                return (
                                    <motion.div
                                        key={ticket.id}
                                        style={{
                                            ...styles.ticketItem,
                                            background: selectedTicket?.id === ticket.id ? 'rgba(88, 58, 255, 0.15)' : 'transparent',
                                        }}
                                        whileHover={{ background: 'rgba(88, 58, 255, 0.1)' }}
                                        onClick={() => fetchTicketDetails(ticket.id)}
                                    >
                                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '0.5rem' }}>
                                            <span style={{ fontWeight: 600, color: '#F8F9FA', fontSize: '0.9rem', flex: 1 }}>
                                                {ticket.subject}
                                            </span>
                                            {renderStars(ticket.priority)}
                                        </div>
                                        <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '0.25rem' }}>
                                            <User size={12} color="#888" />
                                            <span style={{ fontSize: '0.75rem', color: '#888' }}>
                                                {ticket.user?.full_name || ticket.user?.username || 'Usuário'}
                                            </span>
                                            {ticket.user?.highest_role && <RoleBadge role={ticket.user.highest_role} size="small" />}
                                        </div>
                                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                            <span style={{ fontSize: '0.7rem', color: '#666' }}>
                                                {categoryLabels[ticket.category] || ticket.category}
                                                {ticket.assigned_staff && (
                                                    <span style={{ marginLeft: '6px', color: '#583AFF' }}>• {ticket.assigned_staff.username}</span>
                                                )}
                                            </span>
                                            <div style={{ display: 'flex', alignItems: 'center', gap: '4px', color: status.color, fontSize: '0.7rem' }}>
                                                <StatusIcon size={12} />
                                                {status.label}
                                            </div>
                                        </div>
                                    </motion.div>
                                );
                            })
                        )}
                    </div>
                </div>

                {/* Chat Area */}
                <div className="admin-support-chat" style={styles.chatArea}>
                    {selectedTicket ? (
                        <>
                            {/* Header */}
                            <div style={{ padding: '1rem 1.5rem', borderBottom: '1px solid rgba(255, 255, 255, 0.1)' }}>
                                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                                    <div>
                                        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                                            <button className="mobile-back-btn" onClick={() => setSelectedTicket(null)} style={{ border: 'none', color: '#F8F9FA', cursor: 'pointer' }}>
                                                <ChevronLeft size={20} />
                                            </button>
                                            <h3 style={{ margin: 0, color: '#F8F9FA', fontSize: '1.1rem' }}>{selectedTicket.subject}</h3>
                                        </div>
                                        <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginTop: '0.5rem' }}>
                                            <span style={{ fontSize: '0.8rem', color: '#888' }}>
                                                {selectedTicket.user?.full_name || 'Usuário'}
                                            </span>
                                            {selectedTicket.user?.highest_role && <RoleBadge role={selectedTicket.user.highest_role} size="small" />}
                                            {renderStars(selectedTicket.priority)}
                                            <div style={{
                                                width: '8px', height: '8px', borderRadius: '50%',
                                                background: wsStatus === 'CONNECTED' ? '#22c55e' : wsStatus === 'CONNECTING' ? '#eab308' : '#ef4444',
                                                marginLeft: '8px',
                                                boxShadow: wsStatus === 'CONNECTED' ? '0 0 8px #22c55e' : 'none'
                                            }} title={`WebSocket: ${wsStatus}`} />
                                        </div>
                                    </div>
                                    <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap', alignItems: 'center' }}>
                                        {/* Avatar of Attendant (Assignment) */}
                                        {selectedTicket.assigned_to && (
                                            <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginRight: '8px' }}>
                                                <div style={{
                                                    width: '28px', height: '28px', borderRadius: '50%',
                                                    background: 'rgba(88, 58, 255, 0.2)', display: 'flex', alignItems: 'center', justifyContent: 'center',
                                                    border: '1px solid rgba(88, 58, 255, 0.5)', overflow: 'hidden'
                                                }}>
                                                    {selectedTicket.assigned_staff?.avatar_url ? (
                                                        <img src={getFullAvatarUrl(selectedTicket.assigned_staff.avatar_url)} alt="Staff" style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
                                                    ) : (
                                                        <User size={14} color="#583AFF" />
                                                    )}
                                                </div>
                                                <span style={{ fontSize: '0.8rem', color: '#B8BDC7' }}>
                                                    {selectedTicket.assigned_staff?.full_name || (selectedTicket.assigned_to === user.id ? 'Você' : 'Staff')}
                                                </span>
                                            </div>
                                        )}

                                        {!selectedTicket.assigned_to ? (
                                            <motion.button style={styles.actionBtn} whileHover={{ scale: 1.05 }} onClick={assignToMe}>
                                                Resgatar Atendimento
                                            </motion.button>
                                        ) : selectedTicket.assigned_to === user.id ? (
                                            <motion.button
                                                style={{ ...styles.actionBtn, background: 'rgba(239, 68, 68, 0.15)', color: '#EF4444', borderColor: 'rgba(239, 68, 68, 0.3)' }}
                                                whileHover={{ scale: 1.05, background: 'rgba(239, 68, 68, 0.25)' }}
                                                onClick={releaseTicket}
                                            >
                                                Liberar Atendimento
                                            </motion.button>
                                        ) : null}

                                        <select
                                            style={styles.select}
                                            value={selectedTicket.status}
                                            onChange={(e) => updateStatus(e.target.value)}
                                        >
                                            {Object.entries(statusConfig).map(([key, val]) => (
                                                <option key={key} value={key}>{val.label}</option>
                                            ))}
                                        </select>

                                        {/* Close Button Added */}
                                        {selectedTicket.status !== 'CLOSED' && (
                                            <motion.button
                                                style={{
                                                    padding: '0.5rem 0.75rem',
                                                    background: '#EF4444',
                                                    border: 'none',
                                                    borderRadius: '6px',
                                                    color: 'white',
                                                    fontSize: '0.8rem',
                                                    fontWeight: 600,
                                                    cursor: 'pointer',
                                                    marginLeft: '8px'
                                                }}
                                                whileHover={{ scale: 1.05, background: '#DC2626' }}
                                                onClick={() => updateStatus('CLOSED')}
                                            >
                                                Fechar Atendimento
                                            </motion.button>
                                        )}
                                    </div>
                                </div>
                            </div>

                            {/* Messages */}
                            <div style={styles.messagesArea}>
                                {(selectedTicket.messages || []).map((msg) => {
                                    // Check for System Message
                                    if (msg.content.startsWith("🤖 **Sistema**:")) {
                                        return (
                                            <div key={msg.id} style={{ textAlign: 'center', margin: '1rem 0', opacity: 0.7 }}>
                                                <span style={{
                                                    fontSize: '0.75rem',
                                                    color: '#9CA3AF',
                                                    background: 'rgba(255, 255, 255, 0.05)',
                                                    padding: '4px 12px',
                                                    borderRadius: '12px',
                                                    border: '1px solid rgba(255, 255, 255, 0.05)'
                                                }}>
                                                    {msg.content.replace("🤖 **Sistema**:", "").trim()}
                                                </span>
                                            </div>
                                        );
                                    }

                                    return (
                                        <div
                                            key={msg.id}
                                            style={{
                                                ...styles.message,
                                                alignSelf: msg.is_staff ? 'flex-end' : 'flex-start',
                                                background: msg.is_staff ? 'rgba(26, 210, 255, 0.15)' : 'rgba(88, 58, 255, 0.2)',
                                                borderRight: msg.is_staff ? '3px solid #1AD2FF' : 'none',
                                                borderLeft: !msg.is_staff ? '3px solid #583AFF' : 'none',
                                            }}
                                        >
                                            <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '0.5rem' }}>
                                                <span style={{ fontWeight: 600, color: msg.is_staff ? '#1AD2FF' : '#583AFF', fontSize: '0.85rem' }}>
                                                    {msg.sender?.full_name || msg.sender?.username || (msg.is_staff ? 'Equipe' : 'Cliente')}
                                                </span>
                                                {msg.sender?.highest_role && <RoleBadge role={msg.sender.highest_role} size="small" />}
                                            </div>
                                            <p style={{ margin: 0, color: '#E0E0E0' }}>{msg.content || ''}</p>
                                            <span style={{ fontSize: '0.7rem', color: '#666', marginTop: '0.5rem', display: 'block' }}>
                                                {formatDate(msg.created_at || msg.createdAt)}
                                            </span>
                                        </div>
                                    )
                                })}
                                <div ref={messagesEndRef} />
                            </div>

                            {/* Input */}
                            {selectedTicket.status !== 'CLOSED' && (
                                <div style={styles.messageInput} className="admin-support-input-area">
                                    <input
                                        type="text"
                                        value={newMessage}
                                        onChange={(e) => setNewMessage(e.target.value)}
                                        placeholder={selectedTicket.assigned_to === user.id ? "Digite sua resposta..." : "Resgate o ticket para responder"}
                                        style={{
                                            ...styles.input,
                                            opacity: selectedTicket.assigned_to === user.id ? 1 : 0.5,
                                            cursor: selectedTicket.assigned_to === user.id ? 'text' : 'not-allowed'
                                        }}
                                        disabled={selectedTicket.assigned_to !== user.id}
                                        onKeyPress={(e) => e.key === 'Enter' && sendMessage()}
                                    />
                                    <motion.button
                                        style={styles.sendBtn}
                                        whileHover={{ scale: 1.05 }}
                                        whileTap={{ scale: 0.95 }}
                                        onClick={sendMessage}
                                        disabled={sending || selectedTicket.assigned_to !== user.id}
                                    >
                                        <Send size={18} /> {sending ? '...' : 'Enviar'}
                                    </motion.button>
                                </div>
                            )}
                        </>
                    ) : (
                        <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', flexDirection: 'column', color: '#888' }}>
                            <Headphones size={48} style={{ marginBottom: '1rem', opacity: 0.5 }} />
                            <p>Selecione um ticket para visualizar</p>
                        </div>
                    )}
                </div>
            </div>
        </div >
    );
};

export default AdminSupport;

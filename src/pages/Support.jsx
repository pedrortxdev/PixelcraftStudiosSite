import React, { useEffect, useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { getAvatarUrl } from '../utils/formatAvatarUrl';
import { useAuth } from '../context/AuthContext';
import { supportAPI } from '../services/api';
import DashboardLayout from '../components/DashboardLayout';
import { useToast } from '../context/ToastContext';
import RoleBadge from '../components/RoleBadge';
import {
    Headphones,
    Plus,
    MessageCircle,
    Clock,
    CheckCircle,
    AlertCircle,
    Send,
    ChevronRight,
    Star,
    X,
} from 'lucide-react';



const Support = () => {
    const { user, token } = useAuth();
    const toast = useToast();
    const [tickets, setTickets] = useState([]);
    const [loading, setLoading] = useState(true);
    const [selectedTicket, setSelectedTicket] = useState(null);
    const [showNewTicket, setShowNewTicket] = useState(false);
    const [newTicket, setNewTicket] = useState({ subject: '', category: 'GENERAL', content: '' });
    const [newMessage, setNewMessage] = useState('');
    const [sending, setSending] = useState(false);

    const categoryLabels = {
        GENERAL: 'Dúvida Geral',
        SUBSCRIPTION: 'Assinatura',
        PAYMENT: 'Pagamento',
        TECHNICAL: 'Suporte Técnico',
        BILLING: 'Faturamento',
        OTHER: 'Outros',
    };

    const statusConfig = {
        OPEN: { label: 'Aberto', color: '#22c55e', icon: AlertCircle },
        IN_PROGRESS: { label: 'Em Atendimento', color: '#3b82f6', icon: MessageCircle },
        WAITING_RESPONSE: { label: 'Aguardando Resposta', color: '#eab308', icon: Clock },
        RESOLVED: { label: 'Resolvido', color: '#8b5cf6', icon: CheckCircle },
        CLOSED: { label: 'Fechado', color: '#6b7280', icon: CheckCircle },
    };

    const fetchTickets = async () => {
        try {
            const data = await supportAPI.getMyTickets();
            setTickets(data.tickets || []);
        } catch (err) {
            console.error('Error fetching tickets:', err);
        } finally {
            setLoading(false);
        }
    };

    const fetchTicketDetails = async (ticketId) => {
        try {
            const data = await supportAPI.getTicket(ticketId);
            setSelectedTicket(data);
        } catch (err) {
            console.error('Error fetching ticket details:', err);
        }
    };



    const createTicket = async () => {
        if (newTicket.subject.trim().length < 5) {
            toast.error('O assunto deve ter pelo menos 5 caracteres.');
            return;
        }
        if (newTicket.content.trim().length < 10) {
            toast.error('A mensagem deve ter pelo menos 10 caracteres.');
            return;
        }

        setSending(true);
        try {
            const ticket = await supportAPI.createTicket(newTicket);
            setTickets([ticket, ...tickets]);
            setSelectedTicket(ticket);
            setShowNewTicket(false);
            setNewTicket({ subject: '', category: 'GENERAL', content: '' });
            fetchTickets();
        } catch (err) {
            console.error('Error creating ticket:', err);
            toast.error(err.message || 'Erro ao criar ticket.');
        } finally {
            setSending(false);
        }
    };

    const sendMessage = async () => {
        if (!newMessage.trim() || !selectedTicket) return;
        setSending(true);
        try {
            const msg = await supportAPI.sendMessage(selectedTicket.id, newMessage);
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

    const messagesEndRef = React.useRef(null);

    const scrollToBottom = () => {
        messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    };

    useEffect(() => {
        scrollToBottom();
    }, [selectedTicket?.messages]);

    const ws = React.useRef(null);
    const [wsStatus, setWsStatus] = useState('DISCONNECTED');

    useEffect(() => {
        if (!selectedTicket?.id) return;

        let reconnectTimer = null;
        let shouldReconnect = true;
        let reconnectCount = 0;

        const connectWS = () => {
            if (!shouldReconnect) return;
            const wsUrl = supportAPI.getWSUrl(selectedTicket.id);

            if (ws.current) {
                ws.current.onclose = null;
                ws.current.close();
            }

            console.log('Connecting to WS:', wsUrl);
            setWsStatus('CONNECTING');
            ws.current = new WebSocket(wsUrl);

            ws.current.onopen = () => {
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

    const closeTicket = async () => {
        if (!selectedTicket) return;
        if (!window.confirm('Tem certeza que deseja fechar este atendimento?')) return;

        try {
            await supportAPI.closeTicket(selectedTicket.id);
            fetchTickets();
        } catch (err) {
            console.error('Error closing ticket:', err);
        }
    };

    useEffect(() => {
        fetchTickets();
    }, []);

    const renderStars = (priority) => {
        const fullStars = Math.floor(priority);
        const hasHalf = priority % 1 >= 0.5;
        return (
            <div style={{ display: 'flex', gap: '2px' }}>
                {[...Array(fullStars)].map((_, i) => (
                    <Star key={i} size={12} fill="#eab308" color="#eab308" />
                ))}
                {hasHalf && <Star size={12} fill="#eab308" color="#eab308" style={{ clipPath: 'inset(0 50% 0 0)' }} />}
            </div>
        );
    };

    const styles = {
        container: { display: 'flex', gap: '2rem', height: 'calc(100vh - 200px)' },
        ticketList: {
            width: '350px',
            background: 'rgba(15, 20, 35, 0.8)',
            borderRadius: '16px',
            border: '1px solid rgba(88, 58, 255, 0.2)',
            overflow: 'hidden',
            display: 'flex',
            flexDirection: 'column',
        },
        ticketItem: {
            padding: 'var(--btn-padding-md)',
            borderBottom: '1px solid rgba(255, 255, 255, 0.05)',
            cursor: 'pointer',
            transition: 'background 0.2s',
        },
        chatArea: {
            flex: 1,
            background: 'rgba(15, 20, 35, 0.8)',
            borderRadius: '16px',
            border: '1px solid rgba(88, 58, 255, 0.2)',
            display: 'flex',
            flexDirection: 'column',
            overflow: 'hidden',
        },
        chatHeader: {
            padding: '1rem 1.5rem',
            borderBottom: '1px solid rgba(255, 255, 255, 0.1)',
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
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
            padding: '1rem 1.5rem',
            borderTop: '1px solid rgba(255, 255, 255, 0.1)',
            display: 'flex',
            gap: '1rem',
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
            padding: 'var(--btn-padding-md)',
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
        modal: {
            position: 'fixed',
            inset: 0,
            background: 'rgba(0, 0, 0, 0.8)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            zIndex: 1000,
        },
        modalContent: {
            background: 'linear-gradient(180deg, #0F1423 0%, #1A1F35 100%)',
            borderRadius: '16px',
            padding: '2rem',
            width: '500px',
            border: '1px solid rgba(88, 58, 255, 0.3)',
        },
    };

    return (
        <DashboardLayout title="Suporte">
            <div className="support-flex-layout" style={styles.container}>
                {/* Ticket List */}
                <div style={styles.ticketList}>
                    <div style={{ padding: 'var(--btn-padding-md)', borderBottom: '1px solid rgba(255, 255, 255, 0.1)' }}>
                        <motion.button
                            style={{ ...styles.sendBtn, width: '100%', justifyContent: 'center' }}
                            whileHover={{ scale: 1.02 }}
                            whileTap={{ scale: 0.98 }}
                            onClick={() => setShowNewTicket(true)}
                        >
                            <Plus size={18} /> Novo Ticket
                        </motion.button>
                    </div>
                    <div style={{ flex: 1, overflowY: 'auto' }}>
                        {loading ? (
                            <div style={{ padding: '2rem', textAlign: 'center', color: '#888' }}>Carregando...</div>
                        ) : tickets.length === 0 ? (
                            <div style={{ padding: '2rem', textAlign: 'center', color: '#888' }}>
                                Nenhum ticket encontrado
                            </div>
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
                                            <span style={{ fontWeight: 600, color: '#F8F9FA', fontSize: '0.9rem' }}>
                                                {ticket.subject}
                                            </span>
                                            {renderStars(ticket.priority)}
                                        </div>
                                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                            <span style={{ fontSize: '0.75rem', color: '#888' }}>
                                                {categoryLabels[ticket.category] || ticket.category}
                                            </span>
                                            <div style={{ display: 'flex', alignItems: 'center', gap: '4px', color: status.color, fontSize: '0.75rem' }}>
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
                <div style={styles.chatArea}>
                    {selectedTicket ? (
                        <>
                            <div style={styles.chatHeader}>
                                <div>
                                    <h3 style={{ margin: 0, color: '#F8F9FA', fontSize: '1.1rem' }}>{selectedTicket.subject}</h3>
                                    <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginTop: '4px' }}>
                                        <div style={{ fontSize: '0.8rem', color: '#888', display: 'flex', alignItems: 'center' }}>
                                            {categoryLabels[selectedTicket.category]} • {renderStars(selectedTicket.priority)}
                                            <div style={{
                                                width: '6px', height: '6px', borderRadius: '50%',
                                                background: wsStatus === 'CONNECTED' ? '#22c55e' : wsStatus === 'CONNECTING' ? '#eab308' : '#ef4444',
                                                marginLeft: '8px',
                                                boxShadow: wsStatus === 'CONNECTED' ? '0 0 6px #22c55e' : 'none'
                                            }} title={`WebSocket: ${wsStatus}`} />
                                        </div>
                                        {selectedTicket.assigned_staff && (
                                            <div style={{ display: 'flex', alignItems: 'center', gap: '6px', padding: '2px 8px', background: 'rgba(88, 58, 255, 0.1)', borderRadius: '12px', border: '1px solid rgba(88, 58, 255, 0.2)' }}>
                                                <div style={{ width: '20px', height: '20px', borderRadius: '50%', overflow: 'hidden', background: '#333' }}>
                                                    {selectedTicket.assigned_staff.avatar_url ? (
                                                        <img src={getAvatarUrl(selectedTicket.assigned_staff.avatar_url)} alt="Staff" style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
                                                    ) : (
                                                        <div style={{ width: '100%', height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#fff', fontSize: '10px' }}>
                                                            {(selectedTicket.assigned_staff.full_name || 'S').charAt(0)}
                                                        </div>
                                                    )}
                                                </div>
                                                <span style={{ fontSize: '0.8rem', color: '#B8BDC7' }}>
                                                    {selectedTicket.assigned_staff.full_name || 'Staff'}
                                                </span>
                                                {/* Role Badge Implemented */}
                                                {selectedTicket.assigned_staff.highest_role && (
                                                    <RoleBadge role={selectedTicket.assigned_staff.highest_role} size="small" />
                                                )}
                                            </div>
                                        )}
                                    </div>
                                </div>
                            </div>
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
                                                alignSelf: msg.is_staff ? 'flex-start' : 'flex-end',
                                                background: msg.is_staff ? 'rgba(88, 58, 255, 0.2)' : 'rgba(26, 210, 255, 0.15)',
                                                borderLeft: msg.is_staff ? '3px solid #583AFF' : 'none',
                                                borderRight: !msg.is_staff ? '3px solid #1AD2FF' : 'none',
                                            }}
                                        >
                                            <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '0.5rem' }}>
                                                <span style={{ fontWeight: 600, color: msg.is_staff ? '#583AFF' : '#1AD2FF', fontSize: '0.85rem' }}>
                                                    {msg.sender?.full_name || msg.sender?.username || (msg.is_staff ? 'Equipe' : 'Você')}
                                                </span>
                                                {msg.sender?.highest_role && <RoleBadge role={msg.sender.highest_role} size="small" />}
                                            </div>
                                            <p style={{ margin: 0, color: '#E0E0E0' }}>{msg.content}</p>
                                            <span style={{ fontSize: '0.7rem', color: '#666', marginTop: '0.5rem', display: 'block' }}>
                                                {new Date(msg.created_at).toLocaleString('pt-BR')}
                                            </span>
                                        </div>
                                    )
                                })}
                                <div ref={messagesEndRef} />
                            </div>
                            {selectedTicket.status !== 'CLOSED' && (
                                <div style={styles.messageInput}>
                                    <input
                                        type="text"
                                        value={newMessage}
                                        onChange={(e) => setNewMessage(e.target.value)}
                                        placeholder="Digite sua mensagem..."
                                        style={styles.input}
                                        onKeyDown={(e) => e.key === 'Enter' && sendMessage()}
                                    />
                                    <motion.button
                                        style={styles.sendBtn}
                                        whileHover={{ scale: 1.05 }}
                                        whileTap={{ scale: 0.95 }}
                                        onClick={sendMessage}
                                        disabled={sending}
                                    >
                                        <Send size={18} /> {sending ? 'Enviando...' : 'Enviar'}
                                    </motion.button>
                                </div>
                            )}
                        </>
                    ) : (
                        <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', flexDirection: 'column', color: '#888' }}>
                            <Headphones size={48} style={{ marginBottom: '1rem', opacity: 0.5 }} />
                            <p>Selecione um ticket ou crie um novo</p>
                        </div>
                    )}
                </div>
            </div>

            {/* New Ticket Modal */}
            <AnimatePresence>
                {showNewTicket && (
                    <motion.div
                        style={styles.modal}
                        className="mobile-bottom-sheet"
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        exit={{ opacity: 0 }}
                        onClick={() => setShowNewTicket(false)}
                    >
                        <motion.div
                            style={styles.modalContent}
                            className="mobile-bottom-sheet-content"
                            initial={{ scale: 0.9, opacity: 0, y: 50 }}
                            animate={{ scale: 1, opacity: 1, y: 0 }}
                            exit={{ scale: 0.9, opacity: 0, y: 50 }}
                            onClick={(e) => e.stopPropagation()}
                            drag="y"
                            dragConstraints={{ top: 0, bottom: 0 }}
                            dragElastic={{ top: 0, bottom: 0.5 }}
                            onDragEnd={(e, info) => {
                                if (info.offset.y > 100) setShowNewTicket(false);
                            }}
                        >
                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
                                <h2 style={{ margin: 0, color: '#F8F9FA' }}>Novo Ticket de Suporte</h2>
                                <X size={24} style={{ cursor: 'pointer', color: '#888' }} onClick={() => setShowNewTicket(false)} aria-label="Fechar" role="button" tabIndex={0} />
                            </div>
                            <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                                <div>
                                    <label style={{ display: 'block', marginBottom: '0.5rem', color: '#B8BDC7', fontSize: '0.9rem' }}>Assunto</label>
                                    <input
                                        type="text"
                                        value={newTicket.subject}
                                        onChange={(e) => setNewTicket({ ...newTicket, subject: e.target.value })}
                                        placeholder="Descreva brevemente o problema"
                                        style={{ ...styles.input, width: '100%', boxSizing: 'border-box' }}
                                    />
                                </div>
                                <div>
                                    <label style={{ display: 'block', marginBottom: '0.5rem', color: '#B8BDC7', fontSize: '0.9rem' }}>Categoria</label>
                                    <select
                                        value={newTicket.category}
                                        onChange={(e) => setNewTicket({ ...newTicket, category: e.target.value })}
                                        style={{ ...styles.input, width: '100%', boxSizing: 'border-box' }}
                                    >
                                        {Object.entries(categoryLabels).map(([key, label]) => (
                                            <option key={key} value={key}>{label}</option>
                                        ))}
                                    </select>
                                </div>
                                <div>
                                    <label style={{ display: 'block', marginBottom: '0.5rem', color: '#B8BDC7', fontSize: '0.9rem' }}>Mensagem</label>
                                    <textarea
                                        value={newTicket.content}
                                        onChange={(e) => setNewTicket({ ...newTicket, content: e.target.value })}
                                        placeholder="Descreva seu problema ou dúvida em detalhes..."
                                        rows={5}
                                        style={{ ...styles.input, width: '100%', boxSizing: 'border-box', resize: 'vertical' }}
                                    />
                                </div>
                                <motion.button
                                    style={{ ...styles.sendBtn, width: '100%', justifyContent: 'center', marginTop: '0.5rem' }}
                                    whileHover={{ scale: 1.02 }}
                                    whileTap={{ scale: 0.98 }}
                                    onClick={createTicket}
                                    disabled={sending}
                                >
                                    {sending ? 'Criando...' : 'Criar Ticket'}
                                </motion.button>
                            </div>
                        </motion.div>
                    </motion.div>
                )}
            </AnimatePresence>
        </DashboardLayout>
    );
};

export default Support;

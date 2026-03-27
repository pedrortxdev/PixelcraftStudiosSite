import React, { useState, useEffect, useRef, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Send, User, Maximize2, Minimize2, CheckCircle } from 'lucide-react';
import { adminAPI } from '../services/api'; // Ajuste o import conforme sua estrutura

// Helper para data (Mesma correção do ClientChat)
const formatChatDate = (dateString) => {
    if (!dateString) return '';
    try {
        const date = new Date(dateString);
        if (isNaN(date.getTime())) {
            const cleanDate = dateString.replace(/(\.\d{3})\d+/, '$1');
            const dateFallback = new Date(cleanDate);
            if (isNaN(dateFallback.getTime())) return '';
            return new Intl.DateTimeFormat('pt-BR', { 
                day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' 
            }).format(dateFallback);
        }
        
        return new Intl.DateTimeFormat('pt-BR', { 
            day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' 
        }).format(date);
    } catch (error) {
        return '';
    }
};

const SubscriptionChat = ({ subscriptionId }) => {
    const [messages, setMessages] = useState([]);
    const [newMessage, setNewMessage] = useState('');
    const [loading, setLoading] = useState(true);
    const [sending, setSending] = useState(false);
    const [isFullScreen, setIsFullScreen] = useState(false);
    
    const messagesEndRef = useRef(null);

    // Função de buscar mensagens (Polling)
    const fetchMessages = useCallback(async () => {
        try {
            // Assumindo que você criou esse método no adminAPI conforme o plano
            const data = await adminAPI.getSubscriptionChat(subscriptionId);
            setMessages(data || []);
        } catch (error) {
            console.error("Erro ao buscar chat admin", error);
        } finally {
            setLoading(false);
        }
    }, [subscriptionId]);

    useEffect(() => {
        fetchMessages();
        const interval = setInterval(fetchMessages, 5000);
        return () => clearInterval(interval);
    }, [subscriptionId, fetchMessages]);

    const scrollToBottom = () => {
        messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
    };

    useEffect(() => {
        scrollToBottom();
    }, [messages, isFullScreen]);

    const handleSend = async (e) => {
        e.preventDefault();
        if (!newMessage.trim() || sending) return;

        setSending(true);
        try {
            await adminAPI.sendSubscriptionMessage(subscriptionId, newMessage);
            setNewMessage('');
            await fetchMessages();
        } catch (error) {
            console.error("Erro ao enviar mensagem admin", error);
        } finally {
            setSending(false);
        }
    };

    // Estilos FullScreen vs Card
    // ... resto do código ...

    const containerStyle = isFullScreen ? {
        position: 'fixed',
        top: 0,
        bottom: 0,
        right: 0,
        left: '260px', // <--- Mesma largura da sidebar do admin
        zIndex: 9999,
        background: '#0F1219',
        display: 'flex',
        flexDirection: 'column',
        boxShadow: '-10px 0 30px rgba(0,0,0,0.5)'
    } : {
        height: '500px',
        display: 'flex',
        flexDirection: 'column',
        background: 'rgba(21, 26, 38, 0.6)',
        backdropFilter: 'blur(10px)',
        border: '1px solid rgba(255, 255, 255, 0.1)',
        borderRadius: '1rem',
        overflow: 'hidden'
    };

    return (
        <motion.div 
            layout
            style={containerStyle}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
        >
            {/* Header */}
            <div style={{ 
                padding: 'var(--btn-padding-md)', 
                borderBottom: '1px solid rgba(255, 255, 255, 0.1)',
                display: 'flex', 
                justifyContent: 'space-between', 
                alignItems: 'center',
                background: isFullScreen ? 'rgba(21, 26, 38, 0.95)' : 'transparent'
            }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                    <div style={{ color: '#E01A4F' }}>
                        <User size={20} />
                    </div>
                    <div>
                        <h3 style={{ margin: 0, fontSize: '1rem', fontWeight: 600, color: '#F8F9FA' }}>Chat com Cliente</h3>
                        <span style={{ fontSize: '0.75rem', color: '#B8BDC7' }}>Suporte Dedicado</span>
                    </div>
                </div>
                <button 
                    onClick={() => setIsFullScreen(!isFullScreen)}
                    style={{ 
                        background: 'rgba(255,255,255,0.05)', 
                        border: 'none', 
                        padding: '0.5rem', 
                        borderRadius: '0.5rem', 
                        color: '#B8BDC7', 
                        cursor: 'pointer',
                        display: 'flex', alignItems: 'center', justifyContent: 'center'
                    }}
                    title={isFullScreen ? "Minimizar" : "Tela Cheia"}
                >
                    {isFullScreen ? <Minimize2 size={18} /> : <Maximize2 size={18} />}
                </button>
            </div>

            {/* Lista de Mensagens */}
            <div style={{ 
                flex: 1, 
                overflowY: 'auto', 
                padding: '1.5rem', 
                display: 'flex', 
                flexDirection: 'column', 
                gap: '1rem' 
            }}>
                {loading ? (
                    <div style={{ textAlign: 'center', color: '#6C7384', marginTop: '2rem' }}>Carregando...</div>
                ) : (!messages || messages.length === 0) ? (
                    <div style={{ textAlign: 'center', color: '#6C7384', marginTop: '2rem', fontStyle: 'italic' }}>
                        Nenhuma mensagem neste chamado.
                    </div>
                ) : (
                    (Array.isArray(messages) ? messages : []).map((msg, idx) => {
                        if (!msg) return null;
                        // LÓGICA DE ADMIN (Inversa do Cliente)
                        const isAdmin = Boolean(msg.is_admin || msg.isAdmin);
                        // Se isAdmin é TRUE -> Sou EU (Direita). Se FALSE -> É o Cliente (Esquerda)
                        const isMe = isAdmin;

                        return (
                            <div key={idx} style={{ 
                                display: 'flex', 
                                flexDirection: 'column', 
                                alignItems: isMe ? 'flex-end' : 'flex-start' 
                            }}>
                                <div style={{ 
                                    maxWidth: '85%',
                                    padding: '0.75rem 1rem',
                                    borderRadius: '0.75rem',
                                    borderTopRightRadius: isMe ? '2px' : '0.75rem',
                                    borderTopLeftRadius: !isMe ? '2px' : '0.75rem',
                                    // Tema do Admin: Rosa/Laranja para "Eu", Cinza para Cliente
                                    background: isMe 
                                        ? 'var(--gradient-cta)' 
                                        : 'rgba(255, 255, 255, 0.05)',
                                    color: isMe ? 'white' : '#F8F9FA',
                                    boxShadow: '0 2px 8px rgba(0,0,0,0.2)',
                                    border: isMe ? 'none' : '1px solid rgba(255,255,255,0.05)'
                                }}>
                                    <div style={{ fontSize: '0.95rem', lineHeight: '1.5' }}>{msg.content || ''}</div>
                                </div>
                                <div style={{ 
                                    fontSize: '0.7rem', 
                                    color: '#6C7384', 
                                    marginTop: '0.25rem',
                                    marginRight: isMe ? '0.25rem' : 0,
                                    marginLeft: !isMe ? '0.25rem' : 0
                                }}>
                                    {isMe ? 'Você (Admin)' : 'Cliente'} • {formatChatDate(msg.created_at || msg.createdAt)}
                                </div>
                            </div>
                        );
                    })
                )}
                <div ref={messagesEndRef} />
            </div>

            {/* Input */}
            <form onSubmit={handleSend} style={{ 
                padding: 'var(--btn-padding-md)', 
                borderTop: '1px solid rgba(255, 255, 255, 0.1)',
                display: 'flex',
                gap: '0.75rem',
                background: isFullScreen ? 'rgba(21, 26, 38, 0.95)' : 'transparent'
            }}>
                <input 
                    type="text" 
                    value={newMessage}
                    onChange={(e) => setNewMessage(e.target.value)}
                    placeholder="Responder ao cliente..."
                    style={{
                        flex: 1,
                        background: 'rgba(0, 0, 0, 0.2)',
                        border: '1px solid rgba(255, 255, 255, 0.1)',
                        borderRadius: '0.5rem',
                        padding: '0.75rem',
                        color: '#F8F9FA',
                        outline: 'none',
                        fontFamily: 'inherit'
                    }}
                />
                <motion.button 
                    type="submit" 
                    disabled={sending || !newMessage.trim()}
                    whileHover={{ scale: 1.05 }}
                    whileTap={{ scale: 0.95 }}
                    style={{
                        background: 'var(--gradient-cta)',
                        border: 'none',
                        borderRadius: '0.5rem',
                        width: '3rem',
                        display: 'flex', alignItems: 'center', justifyContent: 'center',
                        color: 'white',
                        cursor: sending ? 'wait' : 'pointer',
                        opacity: sending ? 0.7 : 1
                    }}
                >
                    <Send size={18} />
                </motion.button>
            </form>
        </motion.div>
    );
};

export default SubscriptionChat;
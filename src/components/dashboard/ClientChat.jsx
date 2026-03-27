import React, { useState, useEffect, useRef, useCallback } from 'react';
import { createPortal } from 'react-dom'; // <--- IMPORTANTE: Importar o Portal
import { motion } from 'framer-motion';
import { Send, ArrowLeft, Maximize2, Minimize2 } from 'lucide-react';
import { subscriptionsAPI } from '../../services/api';

const formatChatDate = (dateString) => {
    if (!dateString) return '';
    try {
        const cleanDate = dateString.replace(/(\.\d{3})\d+/, '$1');
        const date = new Date(cleanDate);
        if (isNaN(date.getTime())) return '';
        return new Intl.DateTimeFormat('pt-BR', { 
            day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' 
        }).format(date);
    } catch (error) {
        console.error("Erro ao formatar data:", error);
        return '';
    }
};

const ClientChat = ({ subscriptionId, onBack }) => {
    const [messages, setMessages] = useState([]);
    const [newMessage, setNewMessage] = useState('');
    const [loading, setLoading] = useState(true);
    const [sending, setSending] = useState(false);
    const [isFullScreen, setIsFullScreen] = useState(false);
    
    const messagesEndRef = useRef(null);

    const fetchMessages = useCallback(async () => {
        try {
            const data = await subscriptionsAPI.getChatHistory(subscriptionId);
            setMessages(data || []);
        } catch (error) {
            console.error("Erro chat", error);
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
            await subscriptionsAPI.sendMessage(subscriptionId, newMessage);
            setNewMessage('');
            await fetchMessages();
        } catch (error) {
            console.error("Erro envio", error);
        } finally {
            setSending(false);
        }
    };

    // Estilos condicionados
    const containerStyle = isFullScreen ? {
        position: 'fixed',
        top: 0,
        bottom: 0,
        right: 0,
        left: '260px', // Respeita a Sidebar
        zIndex: 9999,
        background: '#0F1219',
        display: 'flex',
        flexDirection: 'column',
        boxShadow: '-10px 0 30px rgba(0,0,0,0.5)'
    } : {
        height: '500px',
        display: 'flex',
        flexDirection: 'column',
        background: 'rgba(0,0,0,0.2)',
        borderRadius: '16px',
        border: '1px solid rgba(255,255,255,0.05)',
        overflow: 'hidden'
    };

    // Conteúdo do Chat
    const chatContent = (
        <motion.div 
            // Removemos animação de scale no fullscreen pra evitar bugs visuais
            initial={isFullScreen ? { opacity: 0 } : { opacity: 0, scale: 0.95 }}
            animate={isFullScreen ? { opacity: 1 } : { opacity: 1, scale: 1 }}
            style={containerStyle}
        >
            {/* Header */}
            <div style={{ 
                display: 'flex', alignItems: 'center', justifyContent: 'space-between', 
                padding: '16px', borderBottom: '1px solid rgba(255,255,255,0.1)',
                background: 'rgba(21, 26, 38, 0.95)'
            }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                    {!isFullScreen && (
                        <button 
                            onClick={onBack}
                            style={{ background: 'none', border: 'none', color: '#B8BDC7', cursor: 'pointer' }}
                        >
                            <ArrowLeft size={20} />
                        </button>
                    )}
                    <div>
                        <h3 style={{ margin: 0, color: '#F8F9FA', fontSize: '1rem', fontWeight: 600 }}>Suporte Pixelcraft</h3>
                        {/* AQUI ESTAVA A MENTIRA - Removida ou Ajustada */}
                        <span style={{ fontSize: '0.75rem', color: '#B8BDC7' }}>
                            Canal de Atendimento
                        </span>
                    </div>
                </div>
                <button 
                    onClick={() => setIsFullScreen(!isFullScreen)}
                    style={{ background: 'rgba(255,255,255,0.1)', border: 'none', padding: '8px', borderRadius: '8px', color: '#F8F9FA', cursor: 'pointer' }}
                >
                    {isFullScreen ? <Minimize2 size={18} /> : <Maximize2 size={18} />}
                </button>
            </div>

            {/* Lista de Mensagens */}
            <div style={{ 
                flex: 1, overflowY: 'auto', padding: '20px', 
                display: 'flex', flexDirection: 'column', gap: '16px' 
            }}>
                {loading ? (
                    <div style={{ textAlign: 'center', color: '#6C7384', marginTop: '2rem' }}>Carregando...</div>
                ) : messages.length === 0 ? (
                    <div style={{ textAlign: 'center', color: '#6C7384', marginTop: '2rem', fontStyle: 'italic' }}>
                        Como podemos ajudar com seu projeto hoje?
                    </div>
                ) : (
                    messages.map((msg, idx) => {
                        const isAdmin = Boolean(msg.is_admin || msg.isAdmin);
                        const isMe = !isAdmin;
                        return (
                            <div key={idx} style={{ 
                                display: 'flex', flexDirection: 'column', 
                                alignItems: isMe ? 'flex-end' : 'flex-start' 
                            }}>
                                <div style={{ 
                                    maxWidth: '80%', padding: '12px 16px', borderRadius: '12px',
                                    borderTopRightRadius: isMe ? '2px' : '12px',
                                    borderTopLeftRadius: !isMe ? '2px' : '12px',
                                    background: isMe ? 'var(--gradient-primary)' : 'rgba(255, 255, 255, 0.08)',
                                    color: '#F8F9FA', boxShadow: '0 2px 10px rgba(0,0,0,0.1)'
                                }}>
                                    <div style={{ fontSize: '0.95rem', lineHeight: '1.5' }}>{msg.content}</div>
                                </div>
                                <div style={{ fontSize: '0.7rem', color: '#6C7384', marginTop: '4px' }}>
                                    {isAdmin ? 'Equipe' : 'Você'} • {formatChatDate(msg.created_at || msg.createdAt)}
                                </div>
                            </div>
                        );
                    })
                )}
                <div ref={messagesEndRef} />
            </div>

            {/* Input Area */}
            <form onSubmit={handleSend} style={{ 
                padding: '16px', background: 'rgba(21, 26, 38, 0.95)', 
                borderTop: '1px solid rgba(255,255,255,0.1)', display: 'flex', gap: '12px'
            }}>
                <input 
                    type="text" value={newMessage} onChange={(e) => setNewMessage(e.target.value)}
                    placeholder="Digite sua mensagem..."
                    style={{
                        flex: 1, background: 'rgba(0,0,0,0.3)', border: '1px solid rgba(255,255,255,0.1)',
                        borderRadius: '8px', padding: '12px', color: '#F8F9FA', outline: 'none'
                    }}
                />
                <button type="submit" disabled={sending || !newMessage.trim()} style={{
                    background: 'var(--gradient-cta)',
                    border: 'none', borderRadius: '8px', width: '48px',
                    display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'white', opacity: sending ? 0.7 : 1
                }}>
                    <Send size={20} />
                </button>
            </form>
        </motion.div>
    );

    // MÁGICA DO PORTAL: Se for FullScreen, joga direto no Body do navegador
    // Se não, renderiza normal dentro do card
    if (isFullScreen) {
        return createPortal(chatContent, document.body);
    }

    return chatContent;
};

export default ClientChat;
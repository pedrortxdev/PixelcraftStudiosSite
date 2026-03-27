import React, { createContext, useContext, useState, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { CheckCircle, AlertCircle, Info, X } from 'lucide-react';

const ToastContext = createContext(null);

export const useToast = () => {
    const ctx = useContext(ToastContext);
    if (!ctx) throw new Error('useToast must be used within ToastProvider');
    return ctx;
};

const ICONS = {
    success: CheckCircle,
    error: AlertCircle,
    info: Info,
};

const COLORS = {
    success: { bg: 'rgba(34, 197, 94, 0.15)', border: 'rgba(34, 197, 94, 0.4)', text: '#22C55E' },
    error: { bg: 'rgba(239, 68, 68, 0.15)', border: 'rgba(239, 68, 68, 0.4)', text: '#EF4444' },
    info: { bg: 'rgba(88, 58, 255, 0.15)', border: 'rgba(88, 58, 255, 0.4)', text: '#583AFF' },
};

export const ToastProvider = ({ children }) => {
    const [toasts, setToasts] = useState([]);

    const addToast = useCallback((message, type = 'info', duration = 4000) => {
        const id = Date.now() + Math.random();
        setToasts(prev => [...prev, { id, message, type }]);
        if (duration > 0) {
            setTimeout(() => removeToast(id), duration);
        }
    }, []);

    const removeToast = useCallback((id) => {
        setToasts(prev => prev.filter(t => t.id !== id));
    }, []);

    const toast = {
        success: (msg, dur) => addToast(msg, 'success', dur),
        error: (msg, dur) => addToast(msg, 'error', dur),
        info: (msg, dur) => addToast(msg, 'info', dur),
    };

    return (
        <ToastContext.Provider value={toast}>
            {children}
            <div style={{
                position: 'fixed', top: '2rem', left: '50%', transform: 'translateX(-50%)',
                zIndex: 10000, display: 'flex', flexDirection: 'column', gap: '0.75rem',
                pointerEvents: 'none', width: 'max-content', maxWidth: '90vw', alignItems: 'center'
            }}>
                <AnimatePresence>
                    {toasts.map(t => {
                        const c = COLORS[t.type] || COLORS.info;
                        const Icon = ICONS[t.type] || Info;
                        return (
                            <motion.div
                                key={t.id}
                                initial={{ opacity: 0, y: 30, scale: 0.95 }}
                                animate={{ opacity: 1, y: 0, scale: 1 }}
                                exit={{ opacity: 0, x: 80, scale: 0.9 }}
                                transition={{ duration: 0.25 }}
                                style={{
                                    display: 'flex', alignItems: 'center', gap: '0.75rem',
                                    padding: '0.75rem 1.25rem',
                                    background: 'var(--bg-card)',
                                    backdropFilter: 'blur(20px)',
                                    border: `1px solid ${c.border}`,
                                    borderRadius: '0px', // Sharp edges
                                    boxShadow: `4px 4px 0px 0px ${c.border.replace('0.4)', '0.15)')}`, // Brutalist drop shadow
                                    color: 'var(--text-primary)',
                                    fontFamily: 'var(--font-mono)',
                                    textTransform: 'uppercase',
                                    fontSize: '0.8rem', fontWeight: 700, letterSpacing: '0.5px',
                                    pointerEvents: 'auto',
                                }}
                            >
                                <Icon size={18} style={{ color: c.text, flexShrink: 0 }} />
                                <span style={{ flex: 1 }}>{t.message}</span>
                                <button
                                    onClick={() => removeToast(t.id)}
                                    style={{
                                        background: 'none', border: 'none', color: 'var(--text-secondary)',
                                        cursor: 'pointer', padding: '2px', flexShrink: 0,
                                    }}
                                >
                                    <X size={14} />
                                </button>
                            </motion.div>
                        );
                    })}
                </AnimatePresence>
            </div>
        </ToastContext.Provider>
    );
};

export default ToastProvider;

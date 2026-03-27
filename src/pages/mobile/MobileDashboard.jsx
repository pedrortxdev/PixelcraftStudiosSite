import React from 'react';
import { motion } from 'framer-motion';
import { Wallet, Activity, ShoppingBag, CreditCard } from 'lucide-react';
import DashboardLayout from '../../components/DashboardLayout';

const MobileDashboard = ({ stats, loading, error }) => {
    const styles = {
        container: {
            padding: 'env(safe-area-inset-top) 1rem calc(env(safe-area-inset-bottom) + 80px) 1rem', // Space for bottom nav
            display: 'flex',
            flexDirection: 'column',
            gap: '1rem',
        },
        header: {
            marginBottom: '1rem',
        },
        title: {
            fontFamily: 'var(--font-display)',
            fontSize: '2rem',
            color: '#F8F9FA',
            margin: 0,
            lineHeight: 1,
        },
        card: {
            background: 'rgba(255, 255, 255, 0.03)',
            border: '1px solid rgba(255, 255, 255, 0.05)',
            borderRadius: '16px',
            padding: '1.5rem',
            display: 'flex',
            flexDirection: 'column',
            position: 'relative',
            overflow: 'hidden',
        },
        primaryCard: {
            background: 'linear-gradient(135deg, rgba(88, 58, 255, 0.15) 0%, rgba(26, 210, 255, 0.05) 100%)',
            border: '1px solid rgba(88, 58, 255, 0.3)',
        },
        cardLabel: {
            fontFamily: 'var(--font-mono)',
            fontSize: '0.75rem',
            color: 'var(--text-secondary)',
            textTransform: 'uppercase',
            letterSpacing: '0.5px',
            marginBottom: '0.5rem',
            zIndex: 1,
        },
        cardValue: {
            fontFamily: 'var(--font-display)',
            fontSize: '2.5rem',
            fontWeight: 700,
            color: '#FFFFFF',
            lineHeight: 1,
            zIndex: 1,
        },
        grid2Col: {
            display: 'grid',
            gridTemplateColumns: '1fr 1fr',
            gap: '1rem',
        },
        listSection: {
            marginTop: '1rem',
            background: 'rgba(255, 255, 255, 0.02)',
            borderRadius: '16px',
            overflow: 'hidden',
            border: '1px solid rgba(255, 255, 255, 0.05)',
        },
        listHeader: {
            padding: '1rem',
            fontFamily: 'var(--font-display)',
            fontSize: '1.25rem',
            color: '#F8F9FA',
            borderBottom: '1px solid rgba(255, 255, 255, 0.05)',
        },
        listItem: {
            padding: '1rem',
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            borderBottom: '1px solid rgba(255, 255, 255, 0.03)',
        },
        muted: {
            fontSize: '0.85rem',
            color: 'var(--text-muted)',
        },
    };

    if (loading) {
        return (
            <DashboardLayout title="Dashboard">
                <div style={{ padding: '2rem', textAlign: 'center', color: '#888' }}>Carregando dados...</div>
            </DashboardLayout>
        );
    }

    return (
        <DashboardLayout title="Dashboard">
            <div style={styles.container}>

                <div style={styles.header}>
                    <h1 style={styles.title}>Visão Geral</h1>
                </div>

                {error && (
                    <div style={{ padding: '1rem', background: 'rgba(239, 68, 68, 0.1)', color: '#EF4444', borderRadius: '12px', fontSize: '0.9rem' }}>
                        {error}
                    </div>
                )}

                {/* Saldo Principal */}
                <motion.div style={{ ...styles.card, ...styles.primaryCard }} initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }}>
                    <div style={{ position: 'absolute', right: '-10px', top: '-10px', opacity: 0.1 }}>
                        <Wallet size={120} />
                    </div>
                    <span style={styles.cardLabel}>Capital Disponível</span>
                    <span style={{ ...styles.cardValue, color: '#1AD2FF' }}>
                        {stats ? `R$ ${stats.balance.toFixed(2)}` : 'R$ 0.00'}
                    </span>
                </motion.div>

                {/* Volume de Gastos */}
                <motion.div style={styles.card} initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.1 }}>
                    <div style={{ position: 'absolute', right: '-10px', top: '-10px', opacity: 0.05 }}>
                        <CreditCard size={100} />
                    </div>
                    <span style={styles.cardLabel}>Volume Negociado</span>
                    <span style={styles.cardValue}>
                        {stats ? `R$ ${stats.total_spent.toFixed(2)}` : 'R$ 0.00'}
                    </span>
                </motion.div>

                {/* Duas colunas menores */}
                <div style={styles.grid2Col}>
                    <motion.div style={styles.card} initial={{ opacity: 0, scale: 0.95 }} animate={{ opacity: 1, scale: 1 }} transition={{ delay: 0.2 }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '8px', color: '#80FFEA' }}>
                            <Activity size={20} />
                            <span style={{ fontSize: '1.25rem', fontWeight: 'bold' }}>{stats?.active_subscriptions || 0}</span>
                        </div>
                        <span style={{ fontSize: '0.75rem', color: '#888', textTransform: 'uppercase' }}>Ativos</span>
                    </motion.div>

                    <motion.div style={styles.card} initial={{ opacity: 0, scale: 0.95 }} animate={{ opacity: 1, scale: 1 }} transition={{ delay: 0.3 }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '8px', color: '#FF6B35' }}>
                            <ShoppingBag size={20} />
                            <span style={{ fontSize: '1.25rem', fontWeight: 'bold' }}>{stats?.products_purchased || 0}</span>
                        </div>
                        <span style={{ fontSize: '0.75rem', color: '#888', textTransform: 'uppercase' }}>Compras</span>
                    </motion.div>
                </div>

                {/* Lista de Pagamentos Recentes */}
                <div style={styles.listSection}>
                    <div style={styles.listHeader}>Últimas Transações</div>
                    {stats?.recent_payments?.length > 0 ? (
                        stats.recent_payments.map((p, idx) => (
                            <div key={idx} style={styles.listItem}>
                                <div style={{ display: 'flex', flexDirection: 'column', gap: '2px' }}>
                                    <span style={{ fontSize: '0.9rem', color: '#F8F9FA' }}>{p.description}</span>
                                    <span style={{ fontSize: '0.75rem', color: '#888' }}>{p.status}</span>
                                </div>
                                <span style={{ fontSize: '0.9rem', fontWeight: 600, color: '#22c55e' }}>
                                    R$ {p.amount.toFixed(2)}
                                </span>
                            </div>
                        ))
                    ) : (
                        <div style={{ padding: '1.5rem', textAlign: 'center', color: '#888', fontSize: '0.9rem' }}>
                            Nenhuma transação recente.
                        </div>
                    )}
                </div>

            </div>
        </DashboardLayout>
    );
};

export default MobileDashboard;

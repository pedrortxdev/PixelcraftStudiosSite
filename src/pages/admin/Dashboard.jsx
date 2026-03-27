import { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
    Users,
    ShoppingBag,
    DollarSign,
    TrendingUp,
    Package,
    ArrowUpRight,
    ArrowDownRight,
    RefreshCw,
    X,
} from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import { adminAPI } from '../../services/api';

function AdminDashboard() {
    const { user, loading: authLoading } = useAuth();
    const [stats, setStats] = useState(null);
    const [dataLoading, setDataLoading] = useState(true);
    const [refreshing, setRefreshing] = useState(false);
    const [selectedOrderItems, setSelectedOrderItems] = useState(null); // For Multiple Items modal

    const fetchStats = async () => {
        try {
            // Fetch data with individual error handling
            const statsPromises = [
                adminAPI.getStats(),
                adminAPI.getRecentOrders(),
                adminAPI.listTransactions({ limit: 5 })
            ];

            const [statsData, recentOrdersData, transactionsData] = await Promise.all(
                statsPromises.map(promise =>
                    promise.catch(error => {
                        console.error('API call failed:', error);
                        return null;
                    })
                )
            );

            // Set stats with safe defaults
            setStats({
                totalRevenue: statsData?.totalRevenue || 0,
                totalUsers: statsData?.totalUsers || 0,
                usersGrowthCount: statsData?.usersGrowthCount || 0, // NEW field
                totalProducts: statsData?.activeProducts || 0,
                totalSales: statsData?.totalSales || 0,
                recentOrders: (Array.isArray(recentOrdersData) ? recentOrdersData : []).map(order => ({
                    id: Math.random().toString(36).substr(2, 9),
                    user: order?.userName || 'Usuário Desconhecido',
                    product: order?.productName || 'Produto Desconhecido',
                    amount: order?.value || 0,
                    status: order?.status ? order.status.toLowerCase() : 'pending',
                    date: new Date().toISOString().split('T')[0]
                })),
                recentTransactions: (Array.isArray(transactionsData?.data) ? transactionsData.data : []).map(tx => ({
                    id: tx.id,
                    user: tx.user_name || tx.user_email || 'Usuário',
                    amount: tx.amount,
                    status: tx.status,
                    type: tx.type,
                    adjustment_type: tx.adjustment_type,
                    date: tx.created_at
                })),
                revenueGrowth: statsData?.revenueGrowthPct || 0,
                userGrowth: statsData?.usersGrowthPct || 0,
                salesGrowth: statsData?.salesGrowthPct || 0,
            });
            setDataLoading(false);
        } catch (error) {
            console.error('Error fetching stats:', error);
            setStats({
                totalRevenue: 0,
                totalUsers: 0,
                totalProducts: 0,
                totalSales: 0,
                recentOrders: [],
                topProducts: [],
                revenueGrowth: 0,
                userGrowth: 0,
                salesGrowth: 0,
            });
            setDataLoading(false);
        }
    };

    // Fetch admin stats
    useEffect(() => {
        if (authLoading || !user) return;
        fetchStats();
    }, [authLoading, user]);

    const handleRefresh = async () => {
        setRefreshing(true);
        try {
            await adminAPI.refreshStats();
            await fetchStats();
        } catch (error) {
            console.error('Error refreshing stats:', error);
        } finally {
            setRefreshing(false);
        }
    };

    const formatCurrency = (value) => {
        return new Intl.NumberFormat('pt-BR', {
            style: 'currency',
            currency: 'BRL'
        }).format(value);
    };

    const styles = {
        container: {
            color: '#F8F9FA',
            // Layout handles padding
        },
        header: {
            marginBottom: '2rem',
        },
        title: {
            fontSize: 'var(--title-h3)',
            fontWeight: 800,
            background: 'linear-gradient(135deg, #F8F9FA 0%, #583AFF 100%)',
            WebkitBackgroundClip: 'text',
            WebkitTextFillColor: 'transparent',
        },
        statsGrid: {
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))',
            gap: '1.5rem',
            marginBottom: '3rem',
        },
        statCard: {
            background: 'rgba(15, 18, 25, 0.6)',
            backdropFilter: 'blur(20px)',
            border: '1px solid rgba(255, 255, 255, 0.1)',
            borderRadius: '16px',
            padding: '1.5rem',
            position: 'relative',
            overflow: 'hidden',
        },
        statIcon: {
            width: '48px',
            height: '48px',
            borderRadius: '12px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            marginBottom: '1rem',
        },
        statValue: {
            fontSize: 'var(--title-h3)',
            fontWeight: 700,
            color: '#F8F9FA',
            marginBottom: '0.5rem',

        },
        statLabel: {
            fontSize: '0.875rem',
            color: '#B8BDC7',
            marginBottom: '0.75rem',
        },
        statGrowth: {
            display: 'flex',
            alignItems: 'center',
            gap: '0.5rem',
            fontSize: '0.875rem',
            fontWeight: 600,
        },
        contentGrid: {
            display: 'grid',
            gridTemplateColumns: '1fr 1fr',
            gap: '2rem',
        },
        card: {
            background: 'rgba(15, 18, 25, 0.6)',
            backdropFilter: 'blur(20px)',
            border: '1px solid rgba(255, 255, 255, 0.1)',
            borderRadius: '16px',
            padding: '2rem',
        },
        cardTitle: {
            fontSize: '1.25rem',
            fontWeight: 700,
            color: '#F8F9FA',
            marginBottom: '1.5rem',

        },
        table: {
            width: '100%',
            borderCollapse: 'collapse',
        },
        th: {
            textAlign: 'left',
            padding: '1rem 0.5rem',
            borderBottom: '1px solid rgba(255, 255, 255, 0.1)',
            color: '#B8BDC7',
            fontSize: '0.875rem',
            fontWeight: 600,
        },
        td: {
            padding: '1rem 0.5rem',
            borderBottom: '1px solid rgba(255, 255, 255, 0.05)',
            color: '#F8F9FA',
            fontSize: '0.875rem',
        },
        statusBadge: {
            padding: '0.25rem 0.75rem',
            borderRadius: '1rem',
            fontSize: '0.75rem',
            fontWeight: 600,
            display: 'inline-block',
        },
        statusCompleted: {
            background: 'rgba(34, 197, 94, 0.1)',
            color: '#22C55E',
            border: '1px solid rgba(34, 197, 94, 0.3)',
        },
        statusPending: {
            background: 'rgba(255, 159, 64, 0.1)',
            color: '#FF9F40',
            border: '1px solid rgba(255, 159, 64, 0.3)',
        },
        loadingText: { textAlign: 'center', padding: '4rem', color: '#B8BDC7' }
    };

    if (authLoading || dataLoading) {
        return (
            <div style={styles.container}>
                <div style={styles.header}>
                    <h1 style={styles.title}>Admin Dashboard</h1>
                </div>
                <div style={styles.loadingText}>
                    Carregando dados do dashboard...
                </div>
            </div>
        );
    }

    return (
        <div style={styles.container}>
            <div style={{ ...styles.header, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <h1 style={styles.title}>Admin Dashboard</h1>
                <motion.button
                    style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: '0.5rem',
                        padding: '0.6rem 1rem',
                        background: 'rgba(88, 58, 255, 0.1)',
                        border: '1px solid rgba(88, 58, 255, 0.3)',
                        borderRadius: '0.5rem',
                        color: '#583AFF',
                        fontSize: '0.875rem',
                        fontWeight: 600,
                        cursor: 'pointer'
                    }}
                    whileHover={{ background: 'rgba(88, 58, 255, 0.2)' }}
                    whileTap={{ scale: 0.95 }}
                    onClick={handleRefresh}
                    disabled={refreshing}
                >
                    <RefreshCw size={16} style={{ animation: refreshing ? 'spin 1s linear infinite' : 'none' }} />
                    {refreshing ? 'Atualizando...' : 'Atualizar Dados'}
                </motion.button>
            </div>

            {stats && (
                <>
                    {/* Stats Grid */}
                    <div style={styles.statsGrid}>
                        <motion.div
                            initial={{ opacity: 0, y: 20 }}
                            animate={{ opacity: 1, y: 0 }}
                            style={styles.statCard}
                        >
                            <div style={{ ...styles.statIcon, background: 'rgba(224, 26, 79, 0.1)' }}>
                                <DollarSign size={24} color="#E01A4F" />
                            </div>
                            <div style={styles.statValue}>{formatCurrency(stats.totalRevenue)}</div>
                            <div style={styles.statLabel}>Receita Total</div>
                        </motion.div>

                        <motion.div
                            initial={{ opacity: 0, y: 20 }}
                            animate={{ opacity: 1, y: 0 }}
                            transition={{ delay: 0.1 }}
                            style={styles.statCard}
                        >
                            <div style={{ ...styles.statIcon, background: 'rgba(88, 58, 255, 0.1)' }}>
                                <Users size={24} color="#583AFF" />
                            </div>
                            <div style={styles.statValue}>
                                {stats.totalUsers || 0}
                            </div>
                            <div style={styles.statLabel}>Usuários Totais</div>
                        </motion.div>

                        <motion.div
                            initial={{ opacity: 0, y: 20 }}
                            animate={{ opacity: 1, y: 0 }}
                            transition={{ delay: 0.2 }}
                            style={styles.statCard}
                        >
                            <div style={{ ...styles.statIcon, background: 'rgba(255, 107, 53, 0.1)' }}>
                                <Package size={24} color="#FF6B35" />
                            </div>
                            <div style={styles.statValue}>{stats.totalProducts}</div>
                            <div style={styles.statLabel}>Produtos Ativos</div>
                        </motion.div>

                        <motion.div
                            initial={{ opacity: 0, y: 20 }}
                            animate={{ opacity: 1, y: 0 }}
                            transition={{ delay: 0.3 }}
                            style={styles.statCard}
                        >
                            <div style={{ ...styles.statIcon, background: 'rgba(34, 197, 94, 0.1)' }}>
                                <ShoppingBag size={24} color="#22C55E" />
                            </div>
                            <div style={styles.statValue}>{stats.totalSales}</div>
                            <div style={styles.statLabel}>Vendas Totais</div>
                        </motion.div>
                    </div>

                    {/* Content Grid */}
                    <div style={styles.contentGrid}>
                        {/* Recent Orders */}
                        <motion.div
                            initial={{ opacity: 0, x: -20 }}
                            animate={{ opacity: 1, x: 0 }}
                            style={styles.card}
                        >
                            <h2 style={styles.cardTitle}>Pedidos Recentes</h2>
                            <table style={styles.table}>
                                <thead>
                                    <tr>
                                        <th style={styles.th}>Usuário</th>
                                        <th style={styles.th}>Produto</th>
                                        <th style={styles.th}>Valor</th>
                                        <th style={styles.th}>Status</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {(stats?.recentOrders || []).map((order) => (
                                        <tr key={order?.id || Math.random().toString(36).substr(2, 9)}>
                                            <td style={styles.td}>{order?.user || 'N/A'}</td>
                                            <td style={styles.td}>
                                                {order?.product === 'Multiple Items' ? (
                                                    <button 
                                                        onClick={() => setSelectedOrderItems(order.items || [])}
                                                        style={{ 
                                                            background: 'none', 
                                                            border: 'none', 
                                                            color: '#1AD2FF', 
                                                            cursor: 'pointer',
                                                            textDecoration: 'underline',
                                                            padding: 0,
                                                            fontSize: 'inherit'
                                                        }}
                                                    >
                                                        Múltiplos Itens
                                                    </button>
                                                ) : (
                                                    order?.product || 'N/A'
                                                )}
                                            </td>
                                            <td style={styles.td}>{formatCurrency(order?.amount || 0)}</td>
                                            <td style={styles.td}>
                                                <span
                                                    style={{
                                                        ...styles.statusBadge,
                                                        ...((order?.status === 'completed' || order?.status === 'paid')
                                                            ? styles.statusCompleted
                                                            : styles.statusPending),
                                                    }}
                                                >
                                                    {(order?.status === 'completed' || order?.status === 'paid') ? 'Concluído' : 'Pendente'}
                                                </span>
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </motion.div>

                        {/* Recent Transactions */}
                        <motion.div
                            initial={{ opacity: 0, x: 20 }}
                            animate={{ opacity: 1, x: 0 }}
                            style={styles.card}
                        >
                            <h2 style={styles.cardTitle}>Transações Recentes</h2>
                            <table style={styles.table}>
                                <thead>
                                    <tr>
                                        <th style={styles.th}>Usuário</th>
                                        <th style={styles.th}>Tipo</th>
                                        <th style={styles.th}>Valor</th>
                                        <th style={styles.th}>Status</th>
                                        <th style={styles.th}>Data</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {(stats?.recentTransactions || []).map((tx) => (
                                        <tr key={tx.id}>
                                            <td style={styles.td}>{tx.user}</td>
                                            <td style={styles.td}>
                                                <span style={{ 
                                                    fontSize: '0.7rem', 
                                                    padding: '2px 6px', 
                                                    borderRadius: '4px',
                                                    background: tx.adjustment_type === 'Teste' ? 'rgba(239, 68, 68, 0.1)' : 'rgba(88, 58, 255, 0.1)',
                                                    color: tx.adjustment_type === 'Teste' ? '#EF4444' : '#583AFF',
                                                    border: `1px solid ${tx.adjustment_type === 'Teste' ? 'rgba(239, 68, 68, 0.2)' : 'rgba(88, 58, 255, 0.2)'}`
                                                }}>
                                                    {tx.adjustment_type || (tx.type === 'deposit' ? 'Depósito' : tx.type)}
                                                </span>
                                            </td>
                                            <td style={{ ...styles.td, color: tx.status === 'completed' ? '#22C55E' : '#F8F9FA' }}>
                                                {formatCurrency(tx.amount)}
                                            </td>
                                            <td style={styles.td}>
                                                <span
                                                    style={{
                                                        ...styles.statusBadge,
                                                        ...(tx.status === 'completed' ? styles.statusCompleted : styles.statusPending),
                                                    }}
                                                >
                                                    {tx.status === 'completed' ? 'Pago' : 'Pendente'}
                                                </span>
                                            </td>
                                            <td style={styles.td}>{new Date(tx.date).toLocaleDateString('pt-BR')}</td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </motion.div>
                    </div>
                </>
            )}

            {/* Multiple Items Floating Modal */}
            <AnimatePresence>
                {selectedOrderItems && (
                    <motion.div
                        initial={{ opacity: 0, scale: 0.9, y: 20 }}
                        animate={{ opacity: 1, scale: 1, y: 0 }}
                        exit={{ opacity: 0, scale: 0.9, y: 20 }}
                        drag
                        dragMomentum={false}
                        style={{
                            position: 'fixed',
                            top: '20%',
                            left: '35%',
                            width: '400px',
                            background: 'rgba(21, 26, 38, 0.95)',
                            backdropFilter: 'blur(20px)',
                            border: '1px solid rgba(88, 58, 255, 0.3)',
                            borderRadius: '16px',
                            padding: '1.5rem',
                            zIndex: 9999,
                            boxShadow: '0 20px 50px rgba(0,0,0,0.5)',
                            cursor: 'grab'
                        }}
                    >
                        <div style={{ 
                            display: 'flex', 
                            justifyContent: 'space-between', 
                            alignItems: 'center', 
                            marginBottom: '1.5rem',
                            borderBottom: '1px solid rgba(255,255,255,0.1)',
                            paddingBottom: '0.75rem'
                        }}>
                            <h3 style={{ margin: 0, fontSize: '1.1rem', fontWeight: 700 }}>Itens do Pedido</h3>
                            <button 
                                onClick={() => setSelectedOrderItems(null)}
                                style={{ background: 'none', border: 'none', color: '#B8BDC7', cursor: 'pointer' }}
                            >
                                <X size={20} />
                            </button>
                        </div>
                        
                        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                            {selectedOrderItems.map((item, idx) => (
                                <div key={idx} style={{ 
                                    display: 'flex', 
                                    alignItems: 'center', 
                                    gap: '0.75rem',
                                    background: 'rgba(255,255,255,0.03)',
                                    padding: '0.75rem',
                                    borderRadius: '8px'
                                }}>
                                    <Package size={16} color="#583AFF" />
                                    <span style={{ fontSize: '0.95rem' }}>{item}</span>
                                </div>
                            ))}
                        </div>
                        
                        <div style={{ marginTop: '1.5rem', fontSize: '0.8rem', color: '#6C7384', textAlign: 'center' }}>
                            Segure e arraste para mover esta janela
                        </div>
                    </motion.div>
                )}
            </AnimatePresence>
        </div>
    );
}

export default AdminDashboard;

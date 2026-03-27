import React, { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import { useNavigate } from 'react-router-dom';
import { adminAPI } from '../../services/api';
import {
    Search,
    Filter,
    MoreVertical,
    Calendar,
    User,
    CreditCard,
    Package,
    ArrowRight
} from 'lucide-react';

const Orders = () => {
    const navigate = useNavigate();
    const [orders, setOrders] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [searchTerm, setSearchTerm] = useState('');

    useEffect(() => {
        const fetchOrders = async () => {
            try {
                setLoading(true);
                const data = await adminAPI.getActiveSubscriptions();
                setOrders(Array.isArray(data) ? data : []);
            } catch (err) {
                console.error('Error fetching orders:', err);
                setError('Falha ao carregar pedidos. Tente novamente.');
                setOrders([]);
            } finally {
                setLoading(false);
            }
        };

        fetchOrders();
    }, []);

    const formatCurrency = (value) => {
        return new Intl.NumberFormat('pt-BR', {
            style: 'currency',
            currency: 'BRL'
        }).format(value);
    };

    const formatDate = (dateString) => {
        if (!dateString) return '—';
        return new Intl.DateTimeFormat('pt-BR').format(new Date(dateString));
    };

    const getStageColor = (stage) => {
        const colors = {
            'Planejamento': '#3B82F6', // Blue
            'Desenvolvimento': '#8B5CF6', // Purple
            'Otimização': '#F59E0B', // Amber
            'Testes': '#EC4899', // Pink
            'Entrega': '#10B981', // Emerald
        };
        return colors[stage] || '#6B7280'; // Gray default
    };

    const filteredOrders = orders.filter(order =>
        (order?.userName || '').toString().toLowerCase().includes(searchTerm.toLowerCase()) ||
        (order?.userEmail || '').toString().toLowerCase().includes(searchTerm.toLowerCase()) ||
        (order?.planName || '').toString().toLowerCase().includes(searchTerm.toLowerCase())
    );

    const styles = {
        container: {
            color: '#F8F9FA',
        },
        header: {
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: '2rem',
        },
        title: {
            fontSize: 'var(--title-h3)',
            fontWeight: 800,
            background: 'linear-gradient(135deg, #F8F9FA 0%, #E01A4F 100%)',
            WebkitBackgroundClip: 'text',
            WebkitTextFillColor: 'transparent',
        },
        card: {
            background: 'rgba(21, 26, 38, 0.6)',
            backdropFilter: 'blur(10px)',
            border: '1px solid rgba(255, 255, 255, 0.1)',
            borderRadius: '1rem',
            overflow: 'hidden',
        },
        toolbar: {
            padding: '1.5rem',
            display: 'flex',
            gap: '1rem',
            borderBottom: '1px solid rgba(255, 255, 255, 0.05)',
        },
        searchContainer: {
            position: 'relative',
            flex: 1,
            maxWidth: '400px',
        },
        searchInput: {
            width: '100%',
            background: 'rgba(0, 0, 0, 0.2)',
            border: '1px solid rgba(255, 255, 255, 0.1)',
            borderRadius: '0.5rem',
            padding: '0.75rem 1rem 0.75rem 2.5rem',
            color: '#F8F9FA',
            outline: 'none',
        },
        searchIcon: {
            position: 'absolute',
            left: '0.75rem',
            top: '50%',
            transform: 'translateY(-50%)',
            color: '#B8BDC7',
        },
        table: {
            width: '100%',
            borderCollapse: 'collapse',
        },
        th: {
            textAlign: 'left',
            padding: '1rem 1.5rem',
            color: '#B8BDC7',
            fontSize: '0.875rem',
            fontWeight: 600,
            borderBottom: '1px solid rgba(255, 255, 255, 0.1)',
        },
        td: {
            padding: '1rem 1.5rem',
            borderBottom: '1px solid rgba(255, 255, 255, 0.05)',
            color: '#F8F9FA',
            fontSize: '0.9rem',
        },
        userCell: {
            display: 'flex',
            flexDirection: 'column',
        },
        userEmail: {
            fontSize: '0.8rem',
            color: '#B8BDC7',
        },
        stageBadge: (stage) => ({
            display: 'inline-flex',
            alignItems: 'center',
            padding: '0.25rem 0.75rem',
            borderRadius: '1rem',
            fontSize: '0.75rem',
            fontWeight: 600,
            background: `${getStageColor(stage)}20`,
            color: getStageColor(stage),
            border: `1px solid ${getStageColor(stage)}40`,
        }),
        actionButton: {
            display: 'inline-flex',
            alignItems: 'center',
            gap: '0.5rem',
            padding: 'var(--btn-padding-sm)',
            background: 'rgba(224, 26, 79, 0.1)',
            color: '#E01A4F',
            border: '1px solid rgba(224, 26, 79, 0.3)',
            borderRadius: '0.5rem',
            cursor: 'pointer',
            fontSize: '0.875rem',
            fontWeight: 600,
            transition: 'all 0.2s',
        },
    };

    if (loading) {
        return (
            <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '400px', color: '#B8BDC7' }}>
                Carregando pedidos...
            </div>
        );
    }

    if (error) {
        return (
            <div style={{ padding: '2rem', background: 'rgba(239, 68, 68, 0.1)', color: '#EF4444', borderRadius: '0.5rem' }}>
                {error}
            </div>
        );
    }

    return (
        <div style={styles.container}>
            <div style={styles.header}>
                <h1 style={styles.title}>Pedidos & Assinaturas</h1>
            </div>

            <motion.div
                style={styles.card}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
            >
                <div style={styles.toolbar}>
                    <div style={styles.searchContainer}>
                        <Search size={18} style={styles.searchIcon} />
                        <input
                            type="text"
                            placeholder="Buscar por usuário ou plano..."
                            style={styles.searchInput}
                            value={searchTerm}
                            onChange={(e) => setSearchTerm(e.target.value)}
                        />
                    </div>
                </div>

                <div style={{ overflowX: 'auto' }}>
                    <table style={styles.table} className="mobile-stacked-table">
                        <thead>
                            <tr>
                                <th style={styles.th}>Usuário</th>
                                <th style={styles.th}>Plano</th>
                                <th style={styles.th}>Valor</th>
                                <th style={styles.th}>Etapa do Projeto</th>
                                <th style={styles.th}>Próx. Fatura</th>
                                <th style={styles.th}>Ações</th>
                            </tr>
                        </thead>
                        <tbody>
                            {filteredOrders.length > 0 ? (
                                filteredOrders.map((order) => (
                                    <motion.tr
                                        key={order.id}
                                        initial={{ opacity: 0 }}
                                        animate={{ opacity: 1 }}
                                        whileHover={{ background: 'rgba(255, 255, 255, 0.02)' }}
                                    >
                                        <td style={styles.td} data-label="Usuário">
                                            <div style={styles.userCell}>
                                                <span style={{ fontWeight: 600 }}>{order?.userName || 'Usuário Desconhecido'}</span>
                                                <span style={styles.userEmail}>{order?.userEmail || 'Sem email'}</span>
                                            </div>
                                        </td>
                                        <td style={styles.td} data-label="Plano">
                                            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                                                <Package size={16} color="#B8BDC7" />
                                                {order?.planName || 'Plano Desconhecido'}
                                            </div>
                                        </td>
                                        <td style={styles.td} data-label="Valor">{formatCurrency(order?.price || 0)}</td>
                                        <td style={styles.td} data-label="Etapa">
                                            <span style={styles.stageBadge(order?.projectStage)}>
                                                {order?.projectStage || 'Não iniciado'}
                                            </span>
                                        </td>
                                        <td style={styles.td} data-label="Próx. Fatura">
                                            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: '#B8BDC7' }}>
                                                <Calendar size={16} />
                                                {formatDate(order?.nextBillingDate)}
                                            </div>
                                        </td>
                                        <td style={styles.td} data-label="Ações">
                                            <motion.button
                                                style={styles.actionButton}
                                                whileHover={{ background: 'rgba(224, 26, 79, 0.2)', scale: 1.05 }}
                                                whileTap={{ scale: 0.95 }}
                                                onClick={() => order?.id && navigate(`/admin/subscriptions/${order.id}`)}
                                            >
                                                Gerenciar
                                                <ArrowRight size={16} />
                                            </motion.button>
                                        </td>
                                    </motion.tr>
                                ))
                            ) : (
                                <tr>
                                    <td colSpan="6" style={{ ...styles.td, textAlign: 'center', padding: '3rem', color: '#B8BDC7' }}>
                                        Nenhum pedido encontrado.
                                    </td>
                                </tr>
                            )}
                        </tbody>
                    </table>
                </div>
            </motion.div>
        </div>
    );
};

export default Orders;

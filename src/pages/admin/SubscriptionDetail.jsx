import React, { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { adminAPI } from '../../services/api';
import SubscriptionChat from '../../components/SubscriptionChat';
import {
    ArrowLeft,
    Save,
    Plus,
    CheckCircle,
    AlertCircle,
    Clock,
    FileText,
    Settings,
    Info
} from 'lucide-react';

const SubscriptionDetail = () => {
    const { id } = useParams();
    const navigate = useNavigate();

    const [loading, setLoading] = useState(true);
    const [sub, setSub] = useState(null);
    const [logs, setLogs] = useState([]);
    const [error, setError] = useState(null);
    const [successMsg, setSuccessMsg] = useState('');

    // Form states
    const [status, setStatus] = useState('');
    const [projectStage, setProjectStage] = useState('');
    const [nextBillingDate, setNextBillingDate] = useState('');
    const [newLogMessage, setNewLogMessage] = useState('');
    const [updating, setUpdating] = useState(false);
    const [addingLog, setAddingLog] = useState(false);

    const stages = ["Planejamento", "Desenvolvimento", "Otimização", "Testes", "Entrega"];
    const statuses = ["ACTIVE", "CANCELED", "PAST_DUE", "COMPLETED"];

    const formatDate = (dateString) => {
        try {
            if (!dateString) return '—';
            const date = new Date(dateString);
            if (isNaN(date.getTime())) return '—';
            return new Intl.DateTimeFormat('pt-BR').format(date);
        } catch (e) {
            return '—';
        }
    };

    const fetchDetails = useCallback(async () => {
        try {
            setLoading(true);
            const data = await adminAPI.getSubscriptionDetails(id);
            // data structure: { subscription: {...}, logs: [...] }
            const subscriptionData = data?.subscription || null;
            const logsData = Array.isArray(data?.logs) ? data.logs : [];

            setSub(subscriptionData);
            setLogs(logsData);

            // Initialize form with safe defaults
            setStatus(subscriptionData?.status || '');
            setProjectStage(subscriptionData?.project_stage || 'Planejamento');
            // Format date for input type="datetime-local" (YYYY-MM-DDTHH:mm)
            if (subscriptionData?.next_billing_date) {
                setNextBillingDate(new Date(subscriptionData.next_billing_date).toISOString().slice(0, 10));
            }
        } catch (err) {
            console.error('Error fetching details:', err);
            setError('Erro ao carregar detalhes da assinatura.');
        } finally {
            setLoading(false);
        }
    }, [id]);

    useEffect(() => {
        fetchDetails();
    }, [id, fetchDetails]);

    const handleUpdate = async (e) => {
        e.preventDefault();
        setUpdating(true);
        setError(null);
        setSuccessMsg('');

        try {
            const payload = {
                status,
                projectStage,
                nextBillingDate: nextBillingDate ? new Date(nextBillingDate).toISOString() : null
            };

            await adminAPI.updateSubscription(id, payload);
            setSuccessMsg('Assinatura atualizada com sucesso!');

            // Refresh data
            const data = await adminAPI.getSubscriptionDetails(id);
            setSub(data.subscription);
        } catch (err) {
            console.error('Update error:', err);
            setError('Falha ao atualizar assinatura.');
        } finally {
            setUpdating(false);
        }
    };

    const handleAddLog = async (e) => {
        e.preventDefault();
        if (!newLogMessage.trim()) return;

        setAddingLog(true);
        try {
            await adminAPI.addSubscriptionLog(id, newLogMessage);
            setNewLogMessage('');
            // Refresh logs
            const data = await adminAPI.getSubscriptionDetails(id);
            setLogs(Array.isArray(data?.logs) ? data.logs : []);
        } catch (err) {
            console.error('Log error:', err);
            setError('Falha ao adicionar log.');
        } finally {
            setAddingLog(false);
        }
    };

    const styles = {
        container: {
            maxWidth: '1400px',
            margin: '0 auto',
            color: '#F8F9FA',
            padding: '2rem',
        },
        header: {
            display: 'flex',
            alignItems: 'center',
            gap: '1rem',
            marginBottom: '2rem',
        },
        backButton: {
            background: 'rgba(255, 255, 255, 0.05)',
            border: '1px solid rgba(255, 255, 255, 0.1)',
            borderRadius: '0.5rem',
            padding: '0.5rem',
            color: '#B8BDC7',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
        },
        title: {
            fontSize: 'var(--title-h4)',
            fontWeight: 700,
        },
        grid: {
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(400px, 1fr))',
            gap: '2rem',
            alignItems: 'start',
        },
        column: {
            display: 'flex',
            flexDirection: 'column',
            gap: '2rem',
        },
        card: {
            background: 'rgba(21, 26, 38, 0.6)',
            backdropFilter: 'blur(10px)',
            border: '1px solid rgba(255, 255, 255, 0.1)',
            borderRadius: '1rem',
            padding: '1.5rem',
            display: 'flex',
            flexDirection: 'column',
            gap: '1.5rem',
        },
        cardHeader: {
            display: 'flex',
            alignItems: 'center',
            gap: '0.75rem',
            marginBottom: '0.5rem',
            borderBottom: '1px solid rgba(255, 255, 255, 0.1)',
            paddingBottom: '1rem',
        },
        cardTitle: {
            fontSize: '1.1rem',
            fontWeight: 600,
            color: '#E01A4F',
            margin: 0,
        },
        infoRow: {
            display: 'flex',
            justifyContent: 'space-between',
            padding: '0.75rem 0',
            borderBottom: '1px solid rgba(255, 255, 255, 0.05)',
        },
        label: {
            color: '#B8BDC7',
            fontSize: '0.9rem',
        },
        value: {
            fontWeight: 500,
            color: '#F8F9FA',
        },
        formGroup: {
            display: 'flex',
            flexDirection: 'column',
            gap: '0.5rem',
        },
        input: {
            background: 'rgba(0, 0, 0, 0.2)',
            border: '1px solid rgba(255, 255, 255, 0.1)',
            borderRadius: '0.5rem',
            padding: '0.75rem',
            color: '#F8F9FA',
            outline: 'none',
            fontFamily: 'inherit',
        },
        select: {
            background: 'rgba(0, 0, 0, 0.2)',
            border: '1px solid rgba(255, 255, 255, 0.1)',
            borderRadius: '0.5rem',
            padding: '0.75rem',
            color: '#F8F9FA',
            outline: 'none',
            cursor: 'pointer',
        },
        button: {
            background: 'var(--gradient-cta)',
            border: 'none',
            borderRadius: '0.5rem',
            padding: '0.75rem',
            color: 'white',
            fontWeight: 600,
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: '0.5rem',
            marginTop: '1rem',
        },
        alert: (type) => ({
            padding: 'var(--btn-padding-md)',
            borderRadius: '0.5rem',
            background: type === 'error' ? 'rgba(239, 68, 68, 0.1)' : 'rgba(34, 197, 94, 0.1)',
            color: type === 'error' ? '#EF4444' : '#22C55E',
            marginBottom: '1rem',
            display: 'flex',
            alignItems: 'center',
            gap: '0.5rem',
        }),
    };

    if (loading) return <div style={{ padding: '4rem', textAlign: 'center', color: '#B8BDC7' }}>Carregando detalhes...</div>;
    if (!sub) return <div style={{ padding: '4rem', textAlign: 'center', color: '#EF4444' }}>Assinatura não encontrada.</div>;

    return (
        <div style={styles.container}>
            <div style={styles.header}>
                <motion.button
                    style={styles.backButton}
                    whileHover={{ scale: 1.05, background: 'rgba(255, 255, 255, 0.1)' }}
                    whileTap={{ scale: 0.95 }}
                    onClick={() => navigate('/admin/orders')}
                >
                    <ArrowLeft size={20} />
                </motion.button>
                <h1 style={styles.title}>Gerenciar Assinatura</h1>
            </div>

            {error && (
                <div style={styles.alert('error')}>
                    <AlertCircle size={20} />
                    {error}
                </div>
            )}

            {successMsg && (
                <div style={styles.alert('success')}>
                    <CheckCircle size={20} />
                    {successMsg}
                </div>
            )}

            <div style={styles.grid}>
                {/* LEFT COLUMN */}
                <div style={styles.column}>
                    {/* INFO CARD */}
                    <motion.div
                        style={styles.card}
                        initial={{ opacity: 0, y: 20 }}
                        animate={{ opacity: 1, y: 0 }}
                    >
                        <div style={styles.cardHeader}>
                            <Info size={20} color="#E01A4F" />
                            <h2 style={styles.cardTitle}>Informações</h2>
                        </div>

                        <div>
                            <div style={styles.infoRow}>
                                <span style={styles.label}>ID</span>
                                <span style={{ ...styles.value, fontSize: '0.8rem', fontFamily: 'monospace' }}>{sub.id}</span>
                            </div>
                            <div style={styles.infoRow}>
                                <span style={styles.label}>Plano</span>
                                <span style={styles.value}>{sub.plan?.name || sub.planName}</span>
                            </div>
                            <div style={styles.infoRow}>
                                <span style={styles.label}>Preço</span>
                                <span style={styles.value}>
                                    {new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(sub.pricePerMonth)}
                                </span>
                            </div>
                            <div style={styles.infoRow}>
                                <span style={styles.label}>Usuário</span>
                                <span style={{ ...styles.value, fontSize: '0.8rem', fontFamily: 'monospace' }}>
                                    {sub.user?.full_name || sub.user?.username || sub.user_name || 'Usuário Desconhecido'}
                                </span>
                            </div>
                            <div style={styles.infoRow}>
                                <span style={styles.label}>Email</span>
                                <span style={styles.value}>{sub.user?.email || sub.user_email || 'Sem email'}</span>
                            </div>
                            <div style={styles.infoRow}>
                                <span style={styles.label}>Criado em</span>
                                <span style={styles.value}>
                                    {formatDate(sub.createdAt || sub.created_at)}
                                </span>
                            </div>
                        </div>
                    </motion.div>

                    {/* MANAGEMENT FORM */}
                    <motion.div
                        style={styles.card}
                        initial={{ opacity: 0, y: 20 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ delay: 0.1 }}
                    >
                        <div style={styles.cardHeader}>
                            <Settings size={20} color="#E01A4F" />
                            <h2 style={styles.cardTitle}>Gerenciamento</h2>
                        </div>

                        <form onSubmit={handleUpdate} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                            <div style={styles.formGroup}>
                                <label style={styles.label}>Status</label>
                                <select
                                    style={styles.select}
                                    value={status}
                                    onChange={(e) => setStatus(e.target.value)}
                                >
                                    {statuses.map(s => <option key={s} value={s}>{s}</option>)}
                                </select>
                            </div>

                            <div style={styles.formGroup}>
                                <label style={styles.label}>Etapa do Projeto</label>
                                <select
                                    style={styles.select}
                                    value={projectStage}
                                    onChange={(e) => setProjectStage(e.target.value)}
                                >
                                    {stages.map(s => <option key={s} value={s}>{s}</option>)}
                                </select>
                            </div>

                            <div style={styles.formGroup}>
                                <label style={styles.label}>Próxima Fatura</label>
                                <input
                                    type="date"
                                    style={styles.input}
                                    value={nextBillingDate}
                                    onChange={(e) => setNextBillingDate(e.target.value)}
                                />
                            </div>

                            <motion.button
                                type="submit"
                                style={styles.button}
                                whileHover={{ scale: 1.02 }}
                                whileTap={{ scale: 0.98 }}
                                disabled={updating}
                            >
                                {updating ? 'Salvando...' : (
                                    <>
                                        <Save size={18} />
                                        Salvar Alterações
                                    </>
                                )}
                            </motion.button>
                        </form>
                    </motion.div>
                </div>

                {/* RIGHT COLUMN */}
                <div style={styles.column}>
                    {/* CHAT COMPONENT */}
                    <SubscriptionChat subscriptionId={id} />

                    {/* PROJECT LOGS */}
                    <motion.div
                        style={styles.card}
                        initial={{ opacity: 0, y: 20 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ delay: 0.2 }}
                    >
                        <div style={styles.cardHeader}>
                            <Clock size={20} color="#E01A4F" />
                            <h2 style={styles.cardTitle}>Timeline do Projeto</h2>
                        </div>

                        <form onSubmit={handleAddLog} style={{ display: 'flex', gap: '1rem', marginBottom: '1.5rem' }}>
                            <input
                                type="text"
                                style={{ ...styles.input, flex: 1 }}
                                placeholder="Adicionar novo registro..."
                                value={newLogMessage}
                                onChange={(e) => setNewLogMessage(e.target.value)}
                            />
                            <motion.button
                                type="submit"
                                style={{ ...styles.button, marginTop: 0, width: 'auto', padding: '0 1rem' }}
                                whileHover={{ scale: 1.05 }}
                                whileTap={{ scale: 0.95 }}
                                disabled={addingLog || !newLogMessage.trim()}
                            >
                                <Plus size={18} />
                            </motion.button>
                        </form>

                        <div style={{ maxHeight: '300px', overflowY: 'auto', paddingRight: '0.5rem' }}>
                            {logs.length > 0 ? (
                                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                                    {logs.map((log) => (
                                        <div key={log.id} style={{
                                            background: 'rgba(255, 255, 255, 0.03)',
                                            padding: '0.75rem',
                                            borderRadius: '8px',
                                            borderLeft: '3px solid #E01A4F'
                                        }}>
                                            <p style={{ color: '#fff', fontSize: '0.9rem', marginBottom: '0.25rem' }}>{log.message}</p>
                                            <span style={{ color: 'rgba(255,255,255,0.5)', fontSize: '0.75rem' }}>
                                                {new Date(log.createdAt || log.created_at).toLocaleString()}
                                            </span>
                                        </div>
                                    ))}
                                </div>
                            ) : (
                                <div style={{ textAlign: 'center', color: '#B8BDC7', padding: '2rem' }}>
                                    Nenhum registro encontrado.
                                </div>
                            )}
                        </div>
                    </motion.div>
                </div>
            </div>
        </div>
    );
};

export default SubscriptionDetail;

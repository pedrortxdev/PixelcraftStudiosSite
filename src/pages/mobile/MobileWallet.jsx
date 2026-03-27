import React, { useEffect, useState, useRef } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
    Wallet as WalletIcon, ArrowDownCircle, ArrowUpCircle, CreditCard,
    AlertCircle, Loader2, X, Copy, CheckCircle, QrCode
} from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import { usersAPI, dashboardAPI, depositAPI, walletAPI } from '../../services/api';
import DashboardLayout from '../../components/DashboardLayout';
import { useToast } from '../../context/ToastContext';
import { copyToClipboard } from '../../utils/clipboard';
import confetti from 'canvas-confetti';

const MobileWallet = () => {
    const navigate = useNavigate();
    const { user } = useAuth();
    const toast = useToast();
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [stats, setStats] = useState(null);
    const [profile, setProfile] = useState(null);
    const [transactions, setTransactions] = useState([]);

    // Deposit Modal State
    const [isDepositModalOpen, setIsDepositModalOpen] = useState(false);
    const [depositAmount, setDepositAmount] = useState('');
    const [depositStep, setDepositStep] = useState('input');
    const [qrCodeBase64, setQrCodeBase64] = useState('');
    const [qrCodeCopyPaste, setQrCodeCopyPaste] = useState('');
    const [activeTxId, setActiveTxId] = useState(null);
    const pollInterval = useRef(null);

    const styles = {
        container: {
            padding: 'env(safe-area-inset-top) 1rem calc(env(safe-area-inset-bottom) + 80px) 1rem',
            display: 'flex',
            flexDirection: 'column',
            gap: '1rem',
        },
        header: {
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '1rem 0',
        },
        title: {
            fontSize: '1.5rem',
            fontFamily: 'var(--font-display)',
            color: '#F8F9FA',
            margin: 0,
        },
        balanceCard: {
            background: 'linear-gradient(135deg, rgba(88, 58, 255, 0.1) 0%, rgba(26, 210, 255, 0.05) 100%)',
            borderRadius: '20px',
            padding: '2rem 1.5rem',
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            border: '1px solid rgba(88, 58, 255, 0.2)',
            position: 'relative',
            overflow: 'hidden',
        },
        balanceLabel: {
            color: 'var(--text-secondary)',
            fontSize: '0.9rem',
            textTransform: 'uppercase',
            letterSpacing: '1px',
            marginBottom: '0.5rem',
        },
        balanceValue: {
            fontFamily: 'var(--font-display)',
            fontSize: '3.5rem',
            lineHeight: 1,
            fontWeight: 'bold',
            color: '#FFFFFF',
        },
        actionButtonsRow: {
            display: 'flex',
            gap: '1rem',
            justifyContent: 'center',
            marginTop: '1.5rem',
        },
        circleButtonContainer: {
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            gap: '0.5rem',
        },
        circleButton: {
            width: '60px',
            height: '60px',
            borderRadius: '30px',
            border: 'none',
            background: 'rgba(255,255,255,0.05)',
            color: '#F8F9FA',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            boxShadow: '0 4px 15px rgba(0,0,0,0.2)',
        },
        actionLabel: {
            fontSize: '0.8rem',
            color: 'var(--text-secondary)',
            fontWeight: 500,
        },
        listSection: {
            marginTop: '1rem',
        },
        listHeader: {
            fontSize: '1.1rem',
            fontFamily: 'var(--font-display)',
            color: '#F8F9FA',
            marginBottom: '1rem',
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
        },
        listItem: {
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '1rem',
            background: 'rgba(255,255,255,0.02)',
            borderRadius: '16px',
            marginBottom: '0.5rem',
            border: '1px solid rgba(255,255,255,0.03)',
        },
        itemIcon: {
            width: '40px',
            height: '40px',
            borderRadius: '12px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            background: 'rgba(88, 58, 255, 0.1)',
            color: '#583AFF',
        },
        itemDetails: {
            flex: 1,
            marginLeft: '1rem',
            display: 'flex',
            flexDirection: 'column',
        },
        itemName: {
            fontSize: '0.95rem',
            fontWeight: 500,
            color: '#F8F9FA',
        },
        itemDate: {
            fontSize: '0.75rem',
            color: '#888',
        },
        itemAmount: {
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'flex-end',
        },
        amountValue: {
            fontSize: '1rem',
            fontWeight: 600,
        },
        itemStatus: {
            fontSize: '0.7rem',
            textTransform: 'uppercase',
            padding: '2px 6px',
            borderRadius: '4px',
            marginTop: '4px',
        },

        // Modal
        modalOverlay: {
            position: 'fixed', top: 0, left: 0, right: 0, bottom: 0,
            background: 'rgba(0, 0, 0, 0.7)', backdropFilter: 'blur(5px)',
            display: 'flex', alignItems: 'flex-end', zIndex: 1000,
        },
        modalSheet: {
            background: 'var(--bg-secondary)', width: '100%',
            borderTopLeftRadius: '24px', borderTopRightRadius: '24px',
            padding: '2rem 1.5rem calc(env(safe-area-inset-bottom) + 1.5rem) 1.5rem',
            position: 'relative',
        },
        pullIndicator: {
            width: '40px', height: '4px', background: 'rgba(255,255,255,0.2)',
            borderRadius: '2px', position: 'absolute', top: '12px', left: '50%',
            transform: 'translateX(-50%)'
        },
        inputSection: {
            marginTop: '1.5rem',
        },
        label: { display: 'block', color: 'var(--text-secondary)', marginBottom: '0.5rem', fontSize: '0.9rem' },
        input: {
            width: '100%', padding: '1rem', background: 'rgba(0,0,0,0.3)',
            border: '1px solid rgba(88, 58, 255, 0.3)', borderRadius: '12px',
            color: '#F8F9FA', fontSize: '1.5rem', outline: 'none', textAlign: 'center',
        },
        primaryBtn: {
            width: '100%', padding: '1rem', background: 'var(--gradient-primary)',
            border: 'none', borderRadius: '12px', color: 'white', fontWeight: 700, fontSize: '1.1rem',
            cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '0.5rem',
            marginTop: '1.5rem'
        },
        copyInput: { display: 'flex', background: 'rgba(0,0,0,0.3)', borderRadius: '12px', border: '1px solid rgba(88, 58, 255, 0.3)', overflow: 'hidden' },
        copyInputArea: { flex: 1, padding: '1rem', background: 'transparent', border: 'none', color: '#B8BDC7', outline: 'none', fontSize: '0.85rem' },
        copyBtn: { background: 'rgba(88, 58, 255, 0.2)', border: 'none', borderLeft: '1px solid rgba(88, 58, 255, 0.3)', padding: '0 1rem', color: '#1AD2FF' },
        qrContainer: {
            background: 'white', padding: '1rem', borderRadius: '16px', margin: '1.5rem auto',
            width: '200px', height: '200px', display: 'flex', alignItems: 'center', justifyContent: 'center',
        }
    };

    const loadData = async () => {
        try {
            setLoading(true);
            setError(null);
            const [statsResp, userResp, txList] = await Promise.all([
                dashboardAPI.getStats(),
                usersAPI.getMe(),
                walletAPI.getTransactions(),
            ]);
            setStats(statsResp || null);
            setProfile(userResp || null);
            setTransactions(txList || []);
        } catch (err) {
            console.error('Erro ao carregar carteira:', err);
            setError('Erro ao carregar dados.');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadData();
        return () => clearInterval(pollInterval.current);
    }, []);

    useEffect(() => {
        if (depositStep === 'qr' && activeTxId) {
            pollInterval.current = setInterval(async () => {
                try {
                    const status = await walletAPI.checkTransactionStatus(activeTxId);
                    if (status === 'COMPLETED') {
                        setDepositStep('success');
                        clearInterval(pollInterval.current);

                        confetti({
                            particleCount: 150, spread: 80, origin: { y: 0.6 },
                            colors: ['#583AFF', '#1AD2FF', '#E01A4F', '#80FFEA', '#25c972']
                        });

                        const [userResp, txList] = await Promise.all([
                            usersAPI.getMe(),
                            walletAPI.getTransactions()
                        ]);
                        setProfile(userResp);
                        setTransactions(txList || []);

                        setTimeout(() => {
                            setIsDepositModalOpen(false);
                            setDepositStep('input');
                            setDepositAmount('');
                            setActiveTxId(null);
                        }, 3000);
                    }
                } catch (err) {
                    console.error('Polling error:', err);
                }
            }, 5000);
        } else {
            if (pollInterval.current) clearInterval(pollInterval.current);
        }
        return () => clearInterval(pollInterval.current);
    }, [depositStep, activeTxId]);

    const handleOpenDeposit = () => {
        setDepositStep('input');
        setDepositAmount('');
        setIsDepositModalOpen(true);
    };

    const handleGeneratePix = async () => {
        if (!depositAmount || parseFloat(depositAmount) < 5) {
            toast.error('Valor mínimo de R$ 5,00');
            return;
        }
        setDepositStep('loading');
        try {
            const res = await depositAPI.create({ amount: parseFloat(depositAmount), method: 'pix' });
            setQrCodeBase64(res.qr_code_base64 || res.qr_code);
            setQrCodeCopyPaste(res.qr_code || res.qr_code_text);
            setActiveTxId(res.transaction_id || res.id);
            setDepositStep('qr');
        } catch (err) {
            console.error('Erro ao gerar Pix:', err);
            toast.error(err.message || 'Erro ao gerar Pix.');
            setDepositStep('input');
        }
    };

    const balance = typeof stats?.balance === 'number' ? stats.balance : (profile?.balance ?? 0);

    const getTransactionLabel = (p) => p.description || (p.type === 'deposit' ? 'Depósito via Pix' : 'Transação');

    const getStatusColor = (status) => {
        const s = String(status || '').toUpperCase();
        if (s === 'COMPLETED') return { color: '#22c55e', bg: 'rgba(34, 197, 94, 0.1)' };
        if (s === 'FAILED') return { color: '#ef4444', bg: 'rgba(239, 68, 68, 0.1)' };
        return { color: '#eab308', bg: 'rgba(234, 179, 8, 0.1)' };
    };

    const getStatusLabel = (status) => {
        const s = String(status || '').toUpperCase();
        if (s === 'COMPLETED') return 'Sucesso';
        if (s === 'FAILED') return 'Falhou';
        if (s === 'REFUNDED') return 'Reembolso';
        return 'Pendente';
    };

    if (loading) {
        return (
            <DashboardLayout title="Carteira">
                <div style={{ ...styles.container, justifyContent: 'center', alignItems: 'center', minHeight: '60vh' }}>
                    <Loader2 size={40} color="#583AFF" style={{ animation: 'spin 1s linear infinite' }} />
                    <p style={{ color: '#888', marginTop: '1rem' }}>Carregando carteira...</p>
                </div>
            </DashboardLayout>
        )
    }

    return (
        <DashboardLayout title="Carteira">
            <div style={styles.container}>
                <div style={styles.header}>
                    <h1 style={styles.title}>Carteira</h1>
                </div>

                {/* Card Principal - App Style */}
                <div style={styles.balanceCard}>
                    <div style={{ position: 'absolute', left: '-20px', bottom: '-20px', opacity: 0.05 }}>
                        <WalletIcon size={180} />
                    </div>

                    <span style={styles.balanceLabel}>Saldo Atual</span>
                    <span style={styles.balanceValue}>R$ {Number(balance).toFixed(2)}</span>

                    <div style={styles.actionButtonsRow}>
                        <div style={styles.circleButtonContainer}>
                            <button style={{ ...styles.circleButton, background: 'var(--gradient-primary)' }} onClick={handleOpenDeposit}>
                                <ArrowDownCircle size={24} />
                            </button>
                            <span style={styles.actionLabel}>Depositar</span>
                        </div>
                        <div style={styles.circleButtonContainer}>
                            <button style={styles.circleButton} onClick={() => toast.info('Saques estarão disponíveis em breve!')}>
                                <ArrowUpCircle size={24} />
                            </button>
                            <span style={styles.actionLabel}>Sacar</span>
                        </div>
                    </div>
                </div>

                {/* Histórico estilo List View */}
                <div style={styles.listSection}>
                    <div style={styles.listHeader}>
                        Histórico
                    </div>

                    {transactions.length === 0 ? (
                        <div style={{ textAlign: 'center', padding: '2rem', color: '#888', background: 'rgba(255,255,255,0.02)', borderRadius: '16px' }}>
                            Nenhuma movimentação.
                        </div>
                    ) : (
                        transactions.map((p, idx) => {
                            const isPositive = p.amount >= 0;
                            const st = getStatusColor(p.status);

                            return (
                                <div key={idx} style={styles.listItem}>
                                    <div style={{ ...styles.itemIcon, background: isPositive ? 'rgba(34, 197, 94, 0.1)' : 'rgba(239, 68, 68, 0.1)', color: isPositive ? '#22c55e' : '#ef4444' }}>
                                        {isPositive ? <ArrowDownCircle size={20} /> : <ArrowUpCircle size={20} />}
                                    </div>
                                    <div style={styles.itemDetails}>
                                        <span style={styles.itemName}>{getTransactionLabel(p)}</span>
                                        <span style={styles.itemDate}>{new Date(p.created_at).toLocaleString('pt-BR')}</span>
                                    </div>
                                    <div style={styles.itemAmount}>
                                        <span style={{ ...styles.amountValue, color: isPositive ? '#22c55e' : '#F8F9FA' }}>
                                            {isPositive ? '+' : ''} R$ {Number(p.amount || 0).toFixed(2)}
                                        </span>
                                        <span style={{ ...styles.itemStatus, background: st.bg, color: st.color }}>
                                            {getStatusLabel(p.status)}
                                        </span>
                                    </div>
                                </div>
                            )
                        })
                    )}
                </div>

                {/* Modal Bottom Sheet Nativo */}
                <AnimatePresence>
                    {isDepositModalOpen && (
                        <motion.div
                            style={styles.modalOverlay}
                            initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}
                        >
                            <motion.div
                                style={styles.modalSheet}
                                initial={{ y: '100%' }} animate={{ y: 0 }} exit={{ y: '100%' }}
                                transition={{ type: 'spring', damping: 25, stiffness: 200 }}
                                drag="y"
                                dragConstraints={{ top: 0, bottom: 0 }}
                                dragElastic={{ top: 0, bottom: 0.8 }}
                                onDragEnd={(e, info) => {
                                    if (info.offset.y > 100) setIsDepositModalOpen(false);
                                }}
                            >
                                <div style={styles.pullIndicator} />
                                <div style={{ position: 'absolute', right: '1.5rem', top: '1.5rem', color: '#888' }} onClick={() => setIsDepositModalOpen(false)}>
                                    <X size={24} />
                                </div>

                                <h2 style={{ fontSize: '1.5rem', fontFamily: 'var(--font-display)', marginBottom: '1rem', marginTop: '1rem' }}>
                                    Adicionar Saldo
                                </h2>

                                {depositStep === 'input' && (
                                    <div style={styles.inputSection}>
                                        <label style={styles.label}>Qual valor deseja depositar?</label>
                                        <input
                                            type="number"
                                            style={styles.input}
                                            placeholder="R$ 0,00"
                                            min="5"
                                            value={depositAmount}
                                            onChange={(e) => setDepositAmount(e.target.value)}
                                        />
                                        <button style={styles.primaryBtn} onClick={handleGeneratePix}>
                                            <QrCode size={20} /> Gerar Pix
                                        </button>
                                    </div>
                                )}

                                {depositStep === 'loading' && (
                                    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', padding: '3rem 0' }}>
                                        <Loader2 size={40} style={{ color: '#583AFF', animation: 'spin 1s linear infinite', marginBottom: '1rem' }} />
                                        <p style={{ color: '#888' }}>Gerando cobrança instantânea...</p>
                                    </div>
                                )}

                                {depositStep === 'qr' && (
                                    <div style={{ textAlign: 'center', padding: '1rem 0' }}>
                                        <div style={styles.qrContainer}>
                                            {qrCodeBase64 && <img src={`data:image/png;base64,${qrCodeBase64}`} alt="Pix QR" style={{ width: '100%' }} />}
                                        </div>
                                        <label style={styles.label}>Pix Copia e Cola</label>
                                        <div style={styles.copyInput}>
                                            <input type="text" readOnly value={qrCodeCopyPaste} style={styles.copyInputArea} />
                                            <button style={styles.copyBtn} onClick={async () => { await copyToClipboard(qrCodeCopyPaste); toast.success('Copiado!'); }}>
                                                <Copy size={20} />
                                            </button>
                                        </div>
                                        <div style={{ marginTop: '1.5rem', display: 'flex', justifyContent: 'center', alignItems: 'center', gap: '8px', color: '#583AFF' }}>
                                            <Loader2 size={16} style={{ animation: 'spin 1s linear infinite' }} />
                                            <span style={{ fontSize: '0.9rem' }}>Aguardando pagamento...</span>
                                        </div>
                                    </div>
                                )}

                                {depositStep === 'success' && (
                                    <div style={{ textAlign: 'center', padding: '3rem 0' }}>
                                        <motion.div initial={{ scale: 0 }} animate={{ scale: 1 }} style={{ display: 'inline-flex', padding: '1rem', borderRadius: '50%', background: 'rgba(34, 197, 94, 0.2)', color: '#22c55e', marginBottom: '1.5rem' }}>
                                            <CheckCircle size={64} />
                                        </motion.div>
                                        <h3 style={{ fontSize: '1.5rem', margin: '0 0 0.5rem 0', color: '#F8F9FA' }}>Pagamento Aprovado!</h3>
                                        <p style={{ color: '#888' }}>Seu saldo já foi atualizado na carteira.</p>
                                    </div>
                                )}

                            </motion.div>
                        </motion.div>
                    )}
                </AnimatePresence>
            </div>
        </DashboardLayout>
    );
};

export default MobileWallet;

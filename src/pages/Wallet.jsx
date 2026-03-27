import React, { useEffect, useState, useRef } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
  Wallet as WalletIcon, ArrowDownCircle, ArrowUpCircle, CreditCard,
  AlertCircle, Loader2, X, Copy, CheckCircle, QrCode, ArrowLeft
} from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { usersAPI, dashboardAPI, depositAPI, walletAPI } from '../services/api';
import DashboardLayout from '../components/DashboardLayout';
import { useToast } from '../context/ToastContext';
import { copyToClipboard } from '../utils/clipboard';
import confetti from 'canvas-confetti';
import { useMobile } from '../hooks/useMobile';
import MobileWallet from './mobile/MobileWallet';

const Wallet = () => {
  const isMobile = useMobile();
  const navigate = useNavigate();
  const { user } = useAuth();
  const toast = useToast(); // used for initialBalance fallback
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
  const [initialBalance, setInitialBalance] = useState(0);
  const [activeTxId, setActiveTxId] = useState(null);
  const pollInterval = useRef(null);

  const styles = {
    backButton: {
      width: '48px', height: '48px', borderRadius: '50%',
      background: 'var(--bg-card)', backdropFilter: 'blur(10px)',
      border: '1px solid var(--border-input)', display: 'flex', alignItems: 'center', justifyContent: 'center',
      color: 'var(--text-primary)', cursor: 'pointer', transition: 'all var(--transition-normal)', boxShadow: 'var(--shadow-card)',
      marginRight: '1rem',
    },
    summaryGrid: { display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '1rem', marginBottom: '1rem' },
    card: {
      background: 'var(--bg-card)', backdropFilter: 'blur(10px)',
      border: '1px solid var(--border-card)', borderRadius: 'var(--radius-lg)', padding: '1.25rem',
      boxShadow: 'var(--shadow-card)', transition: 'all var(--transition-normal)',
    },
    cardLabel: { color: 'var(--text-secondary)', fontSize: '0.85rem', marginBottom: '0.5rem' },
    cardValueRow: { display: 'flex', alignItems: 'center', gap: '0.75rem' },
    cardValue: { fontSize: '1.35rem', fontWeight: 800, color: 'var(--text-primary)' },
    actionsRow: { display: 'flex', alignItems: 'center', gap: '0.75rem', flexWrap: 'wrap', marginTop: '0.5rem' },
    actionBtn: {
      display: 'inline-flex', alignItems: 'center', gap: '0.5rem',
      background: 'var(--gradient-primary)', color: '#fff', border: 'none',
      borderRadius: '0.5rem', padding: '0.6rem 0.9rem', cursor: 'pointer', opacity: 1,
      transition: 'transform 0.2s', fontWeight: 600
    },
    actionBtnSecondary: {
      display: 'inline-flex', alignItems: 'center', gap: '0.5rem', background: 'transparent',
      color: '#F8F9FA', border: '1px solid rgba(88, 58, 255, 0.35)', borderRadius: '0.5rem',
      padding: '0.6rem 0.9rem', cursor: 'not-allowed'
    },
    actionMuted: { color: '#AEB7CD', fontSize: '0.8rem', marginTop: '0.5rem' },
    panel: {
      background: 'var(--bg-card)', backdropFilter: 'blur(10px)',
      border: '1px solid var(--border-card)', borderRadius: 'var(--radius-lg)', padding: '1.25rem',
      boxShadow: 'var(--shadow-card)', marginTop: '1rem'
    },
    panelTitle: { fontSize: '1.1rem', fontWeight: 700, color: 'var(--text-primary)', marginBottom: '1rem' },
    listItem: {
      display: 'grid', gridTemplateColumns: '1fr 160px 140px 160px', alignItems: 'center',
      gap: '0.75rem', padding: '0.75rem', borderRadius: '0.75rem',
      background: 'rgba(255,255,255,0.02)', border: '1px solid rgba(255,255,255,0.05)',
      marginBottom: '0.5rem', color: '#E2E7F1'
    },
    badge: { padding: '0.25rem 0.5rem', borderRadius: '999px', fontSize: '0.8rem' },
    badgeOk: { background: 'rgba(37, 201, 114, 0.15)', color: '#25c972', border: '1px solid rgba(37, 201, 114, 0.4)' },
    badgeFail: { background: 'rgba(255, 77, 79, 0.15)', color: '#ff4d4f', border: '1px solid rgba(255, 77, 79, 0.4)' },
    badgePend: { background: 'rgba(255, 193, 7, 0.15)', color: '#ffc107', border: '1px solid rgba(255, 193, 7, 0.4)' },
    loadingContainer: { display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '200px', flexDirection: 'column', gap: '1rem' },
    errorContainer: { display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '200px', flexDirection: 'column', gap: '1rem', textAlign: 'center' },
    retryButton: { padding: 'var(--btn-padding-md)', background: 'var(--gradient-primary)', border: 'none', borderRadius: '0.5rem', color: 'white', fontWeight: 600, cursor: 'pointer' },

    // Modal Styles
    modalOverlay: {
      position: 'fixed', top: 0, left: 0, right: 0, bottom: 0,
      background: 'rgba(0, 0, 0, 0.7)', backdropFilter: 'blur(5px)',
      display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 1000,
    },
    modalContent: {
      background: 'var(--bg-secondary)', border: '1px solid var(--border-input)',
      borderRadius: 'var(--radius-lg)', padding: '2rem', width: '90%', maxWidth: '450px',
      position: 'relative', boxShadow: 'var(--shadow-modal)',
    },
    modalClose: { position: 'absolute', top: '1rem', right: '1rem', cursor: 'pointer', color: 'var(--text-secondary)' },
    modalTitle: { fontSize: 'var(--title-h4)', fontWeight: 700, color: 'var(--text-primary)', marginBottom: '1.5rem', textAlign: 'center' },
    inputGroup: { marginBottom: '1.5rem' },
    label: { display: 'block', color: 'var(--text-secondary)', marginBottom: '0.5rem', fontSize: '0.9rem' },
    input: {
      width: '100%', padding: '0.75rem', background: 'var(--bg-input)',
      border: '1px solid var(--border-input)', borderRadius: 'var(--radius-sm)',
      color: 'var(--text-primary)', fontSize: '1.1rem', outline: 'none',
    },
    primaryBtn: {
      width: '100%', padding: '0.875rem', background: 'var(--gradient-primary)',
      border: 'none', borderRadius: '0.5rem', color: 'white', fontWeight: 700, fontSize: '1rem',
      cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '0.5rem',
    },
    copyInput: { display: 'flex', gap: '0.5rem', marginTop: '0.5rem' },
    copyBtn: { background: 'rgba(88, 58, 255, 0.2)', border: '1px solid rgba(88, 58, 255, 0.3)', color: '#1AD2FF', borderRadius: '0px', padding: '0 1rem', cursor: 'pointer' },
    qrContainer: {
      background: 'white',
      padding: '1rem',
      borderRadius: '0px',
      margin: '1.5rem auto',
      width: '180px',
      height: '180px',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      border: '2px solid var(--border-card)',
      boxShadow: '4px 4px 0px 0px rgba(255, 255, 255, 0.1)'
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
      setInitialBalance(userResp?.balance || 0);
    } catch (err) {
      console.error('Erro ao carregar carteira:', err);
      setError('Erro ao carregar dados da carteira.');
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

            // Tátil + Visual success
            confetti({
              particleCount: 150,
              spread: 80,
              origin: { y: 0.6 },
              colors: ['#583AFF', '#1AD2FF', '#E01A4F', '#80FFEA', '#25c972']
            });

            // Sync UI profile and transactions
            const userResp = await usersAPI.getMe();
            setProfile(userResp);
            const txList = await walletAPI.getTransactions();
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
    setInitialBalance(profile?.balance || 0);
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
      toast.error(err.message || 'Erro ao gerar Pix. Tente novamente.');
      setDepositStep('input');
    }
  };

  const handleCopyCode = async () => {
    await copyToClipboard(qrCodeCopyPaste);
    toast.success('Código Pix copiado!');
  };

  const payments = transactions;
  const balance = typeof stats?.balance === 'number' ? stats.balance : (profile?.balance ?? null);
  const username = profile?.full_name || profile?.name || 'Usuário';

  const renderStatus = (status) => {
    const s = String(status || '').toUpperCase();
    if (s === 'COMPLETED') return <span style={{ ...styles.badge, ...styles.badgeOk }}>Concluído</span>;
    if (s === 'FAILED') return <span style={{ ...styles.badge, ...styles.badgeFail }}>Falhou</span>;
    if (s === 'REFUNDED') return <span style={{ ...styles.badge, ...styles.badgePend }}>Reembolsado</span>;
    return <span style={{ ...styles.badge, ...styles.badgePend }}>Pendente</span>;
  };

  const getTransactionLabel = (p) => {
    if (p.description) return p.description;
    if (p.type === 'deposit') return 'Depósito via Pix';
    return 'Transação';
  };

  const headerStart = (
    <div style={styles.backButton} onClick={() => navigate(-1)} title="Voltar">
      <ArrowLeft size={22} />
    </div>
  );

  if (isMobile) {
    return <MobileWallet />;
  }

  return (
    <DashboardLayout title="Carteira" headerStart={headerStart}>
      <div style={styles.summaryGrid} className="wallet-summary-grid mobile-swipe-carousel">
        <motion.div style={styles.card} whileHover={{ scale: 1.01 }}>
          <div style={styles.cardLabel}>Saldo Disponível</div>
          <div style={styles.cardValueRow}>
            <WalletIcon size={20} color="#1AD2FF" />
            <div style={styles.cardValue}>{balance != null ? `R$ ${Number(balance).toFixed(2)}` : '—'}</div>
          </div>
        </motion.div>
        <motion.div style={styles.card} whileHover={{ scale: 1.01 }}>
          <div style={styles.cardLabel}>Conta</div>
          <div style={styles.cardValueRow}>
            <CreditCard size={20} color="#583AFF" />
            <div style={styles.cardValue}>{username}</div>
          </div>
        </motion.div>
        <motion.div style={styles.card} whileHover={{ scale: 1.01 }}>
          <div style={styles.cardLabel}>Ações</div>
          <div style={styles.actionsRow}>
            <button style={styles.actionBtn} onClick={handleOpenDeposit} title="Depositar via Pix">
              <ArrowDownCircle size={18} /> Depositar
            </button>
            <button style={styles.actionBtnSecondary} aria-disabled title="Sacar (em breve)">
              <ArrowUpCircle size={18} /> Sacar
            </button>
          </div>
          <div style={styles.actionMuted}>Saques em breve.</div>
        </motion.div>
      </div>

      <section style={styles.panel}>
        <div style={styles.panelTitle}>Histórico de Transações</div>
        {loading ? (
          <div style={styles.loadingContainer}>
            <Loader2 size={32} style={{ color: '#583AFF', animation: 'spin 1s linear infinite' }} />
            <p style={{ color: '#B8BDC7' }}>Carregando transações...</p>
          </div>
        ) : error ? (
          <div style={styles.errorContainer}>
            <AlertCircle size={32} style={{ color: '#EF4444' }} />
            <p style={{ color: '#EF4444' }}>{error}</p>
            <button onClick={() => window.location.reload()} style={styles.retryButton}>Tentar Novamente</button>
          </div>
        ) : payments.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '1.5rem', color: '#B8BDC7', fontStyle: 'italic' }}>
            Nenhuma transação encontrada.
          </div>
        ) : (
          <div>
            {(payments || []).map((p, idx) => (
              <div key={idx} style={styles.listItem} className="mobile-list-item-fintech">
                <div style={{ fontWeight: 600 }}>{getTransactionLabel(p)}</div>
                <div style={{ color: '#AEB7CD' }}>{new Date(p.created_at).toLocaleString('pt-BR')}</div>
                <div style={{ fontWeight: 700, color: p.amount >= 0 ? '#25c972' : '#ff4d4f' }}>
                  {`${p.amount >= 0 ? '+' : ''} R$ ${Number(p.amount || 0).toFixed(2)}`}
                </div>
                <div>{renderStatus(p.status)}</div>
              </div>
            ))}
          </div>
        )}
      </section>

      {/* MODAL */}
      <AnimatePresence>
        {isDepositModalOpen && (
          <motion.div
            style={styles.modalOverlay}
            className="mobile-bottom-sheet"
            initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}
          >
            <motion.div
              style={styles.modalContent}
              className="mobile-bottom-sheet-content"
              initial={{ scale: 0.9, opacity: 0, y: 50 }}
              animate={{ scale: 1, opacity: 1, y: 0 }}
              exit={{ scale: 0.9, opacity: 0, y: 50 }}
              drag="y"
              dragConstraints={{ top: 0, bottom: 0 }}
              dragElastic={{ top: 0, bottom: 0.5 }}
              onDragEnd={(e, info) => {
                if (info.offset.y > 100) setIsDepositModalOpen(false);
              }}
            >
              <div style={styles.modalClose} onClick={() => setIsDepositModalOpen(false)}>
                <X size={24} />
              </div>
              <h2 style={styles.modalTitle}>Adicionar Saldo (Pix)</h2>

              {depositStep === 'input' && (
                <>
                  <div style={styles.inputGroup}>
                    <label style={styles.label}>Valor do Depósito (R$)</label>
                    <input
                      type="number"
                      style={styles.input}
                      placeholder="5.00"
                      min="5"
                      value={depositAmount}
                      onChange={(e) => setDepositAmount(e.target.value)}
                    />
                  </div>
                  <button style={styles.primaryBtn} onClick={handleGeneratePix}>
                    <QrCode size={20} /> Gerar QR Code Pix
                  </button>
                </>
              )}

              {depositStep === 'loading' && (
                <div style={styles.loadingContainer}>
                  <Loader2 size={40} style={{ color: '#583AFF', animation: 'spin 1s linear infinite' }} />
                  <p style={{ color: '#F8F9FA' }}>Gerando cobrança...</p>
                </div>
              )}

              {depositStep === 'qr' && (
                <div style={{ textAlign: 'center' }}>
                  <p style={{ color: '#B8BDC7', marginBottom: '1rem' }}>
                    Escaneie o QR Code no app do seu banco.
                  </p>
                  <div style={styles.qrContainer}>
                    {qrCodeBase64 && (
                      <img
                        src={`data:image/png;base64,${qrCodeBase64}`}
                        alt="Pix QR Code"
                        style={{ width: '100%', height: '100%' }}
                      />
                    )}
                  </div>
                  <div style={styles.inputGroup}>
                    <label style={styles.label}>Pix Copia e Cola</label>
                    <div style={styles.copyInput}>
                      <input
                        type="text"
                        readOnly
                        value={qrCodeCopyPaste}
                        style={{ ...styles.input, fontSize: '0.9rem' }}
                      />
                      <button style={styles.copyBtn} onClick={handleCopyCode} title="Copiar">
                        <Copy size={20} />
                      </button>
                    </div>
                  </div>
                  <div style={{ marginTop: '1.5rem', fontSize: '0.9rem', color: '#583AFF', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '0.5rem' }}>
                    <Loader2 size={16} style={{ animation: 'spin 1s linear infinite' }} />
                    Aguardando pagamento...
                  </div>
                </div>
              )}

              {depositStep === 'success' && (
                <div style={{ textAlign: 'center', padding: '2rem 0' }}>
                  <motion.div
                    initial={{ scale: 0 }} animate={{ scale: 1 }}
                    style={{ display: 'inline-flex', padding: 'var(--btn-padding-md)', borderRadius: '50%', background: 'rgba(37, 201, 114, 0.2)', color: '#25c972', marginBottom: '1rem' }}
                  >
                    <CheckCircle size={48} />
                  </motion.div>
                  <h3 style={{ color: '#F8F9FA', fontSize: 'var(--title-h4)', marginBottom: '0.5rem' }}>Pagamento Confirmado!</h3>
                  <p style={{ color: '#B8BDC7' }}>Seu saldo já foi liberado.</p>
                </div>
              )}
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>
    </DashboardLayout>
  );
};

export default Wallet;
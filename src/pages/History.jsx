import React, { useEffect, useState } from 'react';
import { motion } from 'framer-motion';
import {
  CreditCard,
  ShoppingBag,
  AlertCircle,
  Loader2,
} from 'lucide-react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { historyAPI, walletAPI } from '../services/api';
import DashboardLayout from '../components/DashboardLayout';

const HistoryPage = () => {
  const navigate = useNavigate();
  const { user } = useAuth(); // Keep if needed
  const [subs, setSubs] = useState([]);
  const [libraryItems, setLibraryItems] = useState([]);
  const [transactions, setTransactions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  // === Estilos Específicos ===
  const styles = {
    column: { display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(350px, 1fr))', gap: '1.5rem' },
    panel: {
      background: 'var(--bg-card)',
      backdropFilter: 'blur(10px)',
      border: '1px solid var(--border-card)',
      borderRadius: 'var(--radius-lg)',
      padding: '1.25rem',
      boxShadow: 'var(--shadow-card)',
    },
    panelTitle: { fontSize: '1.1rem', fontWeight: 700, color: 'var(--text-primary)', marginBottom: '1rem' },

    timeline: { position: 'relative', paddingLeft: '1rem' },
    line: { position: 'absolute', left: '10px', top: '0', bottom: '0', width: '2px', background: 'rgba(255,255,255,0.08)' },
    itemRow: {
      position: 'relative',
      display: 'grid',
      gridTemplateColumns: '1fr auto',
      alignItems: 'center',
      gap: '0.75rem',
      padding: '0.75rem',
      borderRadius: '0.75rem',
      background: 'rgba(255,255,255,0.02)',
      border: '1px solid rgba(255,255,255,0.05)',
      marginBottom: '0.5rem',
      color: '#E2E7F1'
    },
    dot: {
      position: 'absolute',
      left: '-5px',
      top: '50%',
      transform: 'translateY(-50%)',
      width: '10px',
      height: '10px',
      borderRadius: '50%',
      background: 'var(--gradient-primary)',
      boxShadow: '0 0 12px rgba(88,58,255,0.6)'
    },
    actionLink: { color: '#A3E8FF', textDecoration: 'none', fontWeight: 600 },
    muted: { color: '#AEB7CD', fontSize: '0.9rem' },
    badge: {
      padding: '0.25rem 0.5rem',
      borderRadius: '999px',
      fontSize: '0.85rem',
      background: 'rgba(88, 58, 255, 0.15)',
      color: '#583AFF',
      border: '1px solid rgba(88, 58, 255, 0.3)'
    },

    loadingContainer: {
      display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '300px', flexDirection: 'column', gap: '1rem',
    },
    errorContainer: {
      display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '300px', flexDirection: 'column', gap: '1rem', textAlign: 'center',
    },
    retryButton: {
      padding: 'var(--btn-padding-md)', background: 'var(--gradient-primary)', border: 'none', borderRadius: '0.5rem',
      color: 'white', fontWeight: 600, cursor: 'pointer',
    },
  };

  useEffect(() => {
    const load = async () => {
      try {
        setLoading(true);
        setError(null);
        const [res, txs] = await Promise.all([
          historyAPI.getMyHistory(),
          walletAPI.getTransactions()
        ]);
        setSubs(Array.isArray(res?.subscriptions) ? res.subscriptions : []);
        setLibraryItems(Array.isArray(res?.products) ? res.products : []);
        setTransactions(Array.isArray(txs) ? txs : []);
      } catch (err) {
        console.error('Erro ao carregar histórico:', err);
        setError('Não foi possível carregar seu histórico.');
      } finally {
        setLoading(false);
      }
    };
    load();
  }, []);

  return (
    <DashboardLayout title="Histórico">
      {loading ? (
        <div style={styles.loadingContainer}>
          <Loader2 size={48} style={{ color: '#583AFF', animation: 'spin 1s linear infinite' }} />
          <p style={{ color: '#B8BDC7', fontSize: '1.1rem' }}>Carregando histórico...</p>
        </div>
      ) : error ? (
        <div style={styles.errorContainer}>
          <AlertCircle size={48} style={{ color: '#EF4444' }} />
          <p style={{ color: '#EF4444', fontSize: '1.1rem' }}>{error}</p>
          <button onClick={() => window.location.reload()} style={styles.retryButton}>Tentar Novamente</button>
        </div>
      ) : (
        <div style={styles.column} className="history-column-grid">
          <section style={styles.panel}>
            <div style={styles.panelTitle}>Assinaturas</div>
            <div style={styles.timeline}>
              <div style={styles.line} />
              {subs.length > 0 ? (
                subs.map((s, idx) => (
                  <motion.div key={idx} style={styles.itemRow} whileHover={{ scale: 1.01 }}>
                    <div style={styles.dot} />
                    <div>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                        <CreditCard size={18} color="#583AFF" />
                        <span style={{ fontWeight: 700 }}>{s?.plan_name || 'Assinatura'}</span>
                      </div>
                      <div style={styles.muted}>
                        Criado em: {s?.created_at ? new Date(s.created_at).toLocaleDateString('pt-BR') : '—'}
                      </div>
                    </div>
                    <span style={styles.badge}>
                      {s?.price_per_month ? `R$ ${Number(s.price_per_month).toFixed(2)}` : '—'}
                    </span>
                  </motion.div>
                ))
              ) : (
                <div style={{ ...styles.muted, fontStyle: 'italic', textAlign: 'center', padding: 'var(--btn-padding-md)' }}>
                  Nenhuma assinatura encontrada.
                </div>
              )}
            </div>
          </section>

          <section style={styles.panel}>
            <div style={styles.panelTitle}>Produtos Comprados</div>
            <div style={styles.timeline}>
              <div style={styles.line} />
              {libraryItems.length > 0 ? (
                libraryItems.map((item, idx) => (
                  <motion.div key={idx} style={styles.itemRow} whileHover={{ scale: 1.01 }}>
                    <div style={styles.dot} />
                    <div>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                        <ShoppingBag size={18} color="#1AD2FF" />
                        <span style={{ fontWeight: 700 }}>{item?.name || 'Produto'}</span>
                      </div>
                      <div style={styles.muted}>
                        Tipo: {item?.type || '—'}
                        {item?.price && ` • Preço: R$ ${Number(item.price).toFixed(2)}`}
                      </div>
                    </div>
                    <Link to="/downloads" style={styles.actionLink}>
                      Ver downloads
                    </Link>
                  </motion.div>
                ))
              ) : (
                <div style={{ ...styles.muted, fontStyle: 'italic', textAlign: 'center', padding: 'var(--btn-padding-md)' }}>
                  Nenhuma compra encontrada.
                </div>
              )}
            </div>
          </section>

          <section style={styles.panel}>
            <div style={styles.panelTitle}>Adição de Fundos (Carteira)</div>
            <div style={styles.timeline}>
              <div style={styles.line} />
              {transactions.length > 0 ? (
                transactions.map((tx, idx) => (
                  <motion.div key={idx} style={styles.itemRow} whileHover={{ scale: 1.01 }}>
                    <div style={styles.dot} />
                    <div>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                        <CreditCard size={18} color="#25C972" />
                        <span style={{ fontWeight: 700 }}>{tx.type === 'deposit' ? 'Depósito via Pix' : 'Transação'}</span>
                      </div>
                      <div style={styles.muted}>
                        Data: {tx.created_at ? new Date(tx.created_at).toLocaleDateString('pt-BR') : '—'}
                      </div>
                    </div>
                    <span style={{ ...styles.badge, background: tx.status === 'COMPLETED' ? 'rgba(37, 201, 114, 0.15)' : 'rgba(255, 193, 7, 0.15)', color: tx.status === 'COMPLETED' ? '#25c972' : '#ffc107', border: `1px solid ${tx.status === 'COMPLETED' ? 'rgba(37, 201, 114, 0.4)' : 'rgba(255, 193, 7, 0.4)'}` }}>
                      {tx.status === 'COMPLETED' ? 'Concluído' : 'Pendente'} {` R$ ${Number(tx.amount || 0).toFixed(2)}`}
                    </span>
                  </motion.div>
                ))
              ) : (
                <div style={{ ...styles.muted, fontStyle: 'italic', textAlign: 'center', padding: 'var(--btn-padding-md)' }}>
                  Nenhuma transação encontrada.
                </div>
              )}
            </div>
          </section>
        </div>
      )}
    </DashboardLayout>
  );
};

export default HistoryPage;
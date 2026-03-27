import React, { useEffect, useState } from 'react';
import { motion } from 'framer-motion';
import { useAuth } from '../context/AuthContext';
import { dashboardAPI } from '../services/api';
import {
  Wallet,
  CreditCard,
  ShoppingBag,
  Activity,
} from 'lucide-react';
import DashboardLayout from '../components/DashboardLayout';
import { useMobile } from '../hooks/useMobile';
import MobileDashboard from './mobile/MobileDashboard';

const Dashboard = () => {
  const isMobile = useMobile();
  const { user } = useAuth(); // used for checking ? maybe not needed here if only layout uses it
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const styles = {
    // Bento Grid Layout replacing standard 4-columns
    bentoGrid: {
      display: 'grid',
      gridTemplateColumns: 'repeat(12, 1fr)',
      gap: 'var(--bento-gap, 1.5rem)',
      marginBottom: '3rem'
    },
    bentoBlock: {
      background: 'rgba(255, 255, 255, 0.02)',
      border: '1px solid rgba(255, 255, 255, 0.05)',
      padding: 'var(--panel-padding, 2rem)',
      position: 'relative',
      overflow: 'hidden',
      display: 'flex',
      flexDirection: 'column',
      justifyContent: 'space-between',
      minHeight: 'var(--bento-min-height, 200px)'
    },
    bentoLarge: { gridColumn: 'span 12', '@media (min-width: 1024px)': { gridColumn: 'span 8' } },
    bentoSmall: { gridColumn: 'span 12', '@media (min-width: 1024px)': { gridColumn: 'span 4' } },
    bentoMedium: { gridColumn: 'span 12', '@media (min-width: 1024px)': { gridColumn: 'span 6' } },

    cardLabel: {
      fontFamily: 'var(--font-mono)',
      color: 'var(--text-secondary)',
      fontSize: '0.875rem',
      textTransform: 'uppercase',
      letterSpacing: '1px',
      marginBottom: '1rem'
    },
    cardValueRow: { display: 'flex', alignItems: 'center', gap: '1rem', marginTop: 'auto' },
    cardValue: {
      fontFamily: 'var(--font-display)',
      fontSize: 'var(--card-value-size, 4rem)',
      lineHeight: '1',
      fontWeight: 700,
      color: 'var(--text-primary)'
    },
    panelsGrid: { display: 'grid', gridTemplateColumns: '1fr', gap: '1.5rem', '@media (min-width: 1024px)': { gridTemplateColumns: '7fr 5fr' } },
    panel: {
      background: 'rgba(255, 255, 255, 0.02)',
      border: '1px solid rgba(255, 255, 255, 0.05)',
      padding: 'var(--panel-padding, 2rem)',
      minHeight: '400px',
      display: 'flex',
      flexDirection: 'column'
    },
    panelTitle: {
      fontFamily: 'var(--font-display)',
      fontSize: '2rem',
      color: 'var(--text-primary)',
      textTransform: 'uppercase',
      marginBottom: '2rem',
      lineHeight: 1
    },
    listItem: {
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'space-between',
      padding: '1.25rem',
      background: 'rgba(255,255,255,0.02)',
      border: '1px solid rgba(255,255,255,0.05)',
      marginBottom: '0.75rem',
      transition: 'all 0.3s ease'
    },
    muted: { fontFamily: 'var(--font-mono)', color: 'var(--text-muted)', fontSize: '0.9rem' },
  };

  useEffect(() => {
    const load = async () => {
      try {
        setLoading(true);
        const data = await dashboardAPI.getStats();
        setStats(data);
      } catch (err) {
        console.error('Erro ao carregar dashboard:', err);
        setError('Erro ao carregar seus dados.');
      } finally {
        setLoading(false);
      }
    };
    load();
  }, []);

  if (isMobile) {
    return <MobileDashboard stats={stats} loading={loading} error={error} />;
  }

  return (
    <DashboardLayout title="Dashboard">
      {/* SUMMARY CARDS */}
      {/* BENTO GRID (RESPONSIVE) */}
      <style dangerouslySetInnerHTML={{
        __html: `
        .bento-dash-large { grid-column: span 12; }
        .bento-dash-small { grid-column: span 12; }
        .bento-dash-half { grid-column: span 12; }
        .dash-panel-grid { grid-template-columns: 1fr; }
        
        @media (min-width: 1024px) {
          .bento-dash-large { grid-column: span 8; }
          .bento-dash-small { grid-column: span 4; }
          .bento-dash-half { grid-column: span 6; }
          .dash-panel-grid { grid-template-columns: 7fr 5fr; }
        }
      `}} />

      <div style={styles.bentoGrid} className="mobile-swipe-carousel">

        {/* BIG BLOCK - MAIN BALANCE */}
        <motion.div
          className="bento-dash-large stat-card-v4"
          style={{ ...styles.bentoBlock }}
          whileHover={{ borderColor: 'rgba(26, 210, 255, 0.4)' }}
        >
          <div style={{ position: 'absolute', top: '-100px', right: '-50px', opacity: 0.1 }} className="stat-icon-v4-bg">
            <Wallet size={250} color="#1AD2FF" />
          </div>
          <div style={styles.cardLabel}>CAPITAL DISPONÍVEL (SALDO)</div>
          <div style={styles.cardValueRow}>
            <div style={{ ...styles.cardValue, color: '#1AD2FF' }}>
              {stats ? `R$ ${stats.balance.toFixed(2)}` : 'R$ 0.00'}
            </div>
          </div>
        </motion.div>

        {/* SMALL BLOCK - ACTIVE SUBS */}
        <motion.div
          className="bento-dash-small stat-card-v4"
          style={{ ...styles.bentoBlock }}
          whileHover={{ borderColor: 'rgba(128, 255, 234, 0.4)' }}
        >
          <div style={{ position: 'absolute', top: '-50px', right: '-20px', opacity: 0.15 }} className="stat-icon-v4-bg">
            <Activity size={150} color="#80FFEA" />
          </div>
          <div style={styles.cardLabel}>INFRAESTRUTURAS ATIVAS</div>
          <div style={styles.cardValueRow}>
            <div style={styles.cardValue}>{stats ? stats.active_subscriptions : '0'}</div>
          </div>
        </motion.div>

        {/* HALF BLOCK - PURCHASES */}
        <motion.div
          className="bento-dash-half stat-card-v4"
          style={{ ...styles.bentoBlock }}
          whileHover={{ borderColor: 'rgba(255, 107, 53, 0.4)' }}
        >
          <div style={{ position: 'absolute', bottom: '-50px', right: '-20px', opacity: 0.1 }} className="stat-icon-v4-bg">
            <ShoppingBag size={180} color="#FF6B35" />
          </div>
          <div style={styles.cardLabel}>ASSETS ADQUIRIDOS</div>
          <div style={styles.cardValueRow}>
            <div style={styles.cardValue}>{stats ? stats.products_purchased : '0'}</div>
          </div>
        </motion.div>

        {/* HALF BLOCK - TOTAL SPENT */}
        <motion.div
          className="bento-dash-half stat-card-v4"
          style={{ ...styles.bentoBlock }}
          whileHover={{ borderColor: 'rgba(224, 26, 79, 0.4)' }}
        >
          <div style={{ position: 'absolute', bottom: '-20px', left: '-20px', opacity: 0.05 }} className="stat-icon-v4-bg">
            <CreditCard size={180} color="#E01A4F" />
          </div>
          <div style={{ ...styles.cardLabel, textAlign: 'right' }}>VOLUME NEGOCIADO</div>
          <div style={{ ...styles.cardValueRow, justifyContent: 'flex-end' }}>
            <div style={{ ...styles.cardValue, fontSize: '3rem', color: 'var(--text-secondary)' }}>
              {stats ? `R$ ${stats.total_spent.toFixed(2)}` : 'R$ 0.00'}
            </div>
          </div>
        </motion.div>

      </div>

      {/* PANELS */}
      <div style={styles.panelsGrid} className="dash-panel-grid">
        <motion.section
          style={styles.panel}
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1 }}
        >
          <div style={styles.panelTitle}>Pagamentos Recentes</div>
          {loading && <div style={styles.muted}>Carregando...</div>}
          {error && <div style={{ color: '#EF4444', padding: 'var(--btn-padding-md)', background: 'rgba(239, 68, 68, 0.1)', borderRadius: '0.5rem' }}>{error}</div>}
          {stats && Array.isArray(stats.recent_payments) && stats.recent_payments.length > 0 ? (
            (stats.recent_payments || []).map((p, idx) => (
              <motion.div
                key={idx}
                style={styles.listItem}
                whileHover={{
                  background: 'rgba(88, 58, 255, 0.1)',
                  borderColor: 'rgba(88, 58, 255, 0.3)'
                }}
                initial={{ opacity: 0, x: -20 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: idx * 0.05 }}
              >
                <span>{p.description}</span>
                <span style={styles.muted}>{`R$ ${p.amount.toFixed(2)} • ${p.status}`}</span>
              </motion.div>
            ))
          ) : (
            !loading && <div style={{ ...styles.muted, padding: '2rem', textAlign: 'center', background: 'rgba(255,255,255,0.03)', borderRadius: '0.75rem' }}>Nenhum pagamento recente.</div>
          )}
        </motion.section>

        <motion.section
          style={styles.panel}
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2 }}
        >
          <div style={styles.panelTitle}>Gastos Mensais</div>
          {stats && Array.isArray(stats.monthly_spending) && stats.monthly_spending.length > 0 ? (
            <div>
              {(() => {
                const maxAmount = Math.max(...stats.monthly_spending.map(m => m.amount), 1);
                return stats.monthly_spending.map((m, idx) => (
                  <div key={idx} style={{ display: 'grid', gridTemplateColumns: '120px 1fr auto', alignItems: 'center', gap: '1rem', marginBottom: '1rem' }}>
                    <span style={styles.muted}>{m.month}</span>
                    <div style={{ height: '12px', background: 'rgba(255,255,255,0.08)', borderRadius: '999px', overflow: 'hidden' }}>
                      <motion.div
                        style={{
                          height: '100%',
                          background: 'linear-gradient(90deg, #583AFF 0%, #1AD2FF 100%)',
                          borderRadius: '999px'
                        }}
                        initial={{ width: 0 }}
                        animate={{ width: `${(m.amount / maxAmount) * 100}%` }}
                        transition={{ duration: 1, delay: idx * 0.1 }}
                      />
                    </div>
                    <span style={{ ...styles.muted, minWidth: '80px', textAlign: 'right', fontSize: '0.85rem' }}>
                      {new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(m.amount)}
                    </span>
                  </div>
                ));
              })()}
            </div>
          ) : (
            !loading && <div style={{ ...styles.muted, padding: '2rem', textAlign: 'center', background: 'rgba(255,255,255,0.03)', borderRadius: '0.75rem' }}>Sem dados de gastos.</div>
          )}
        </motion.section>
      </div>
    </DashboardLayout>
  );
};

export default Dashboard;
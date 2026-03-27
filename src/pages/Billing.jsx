// src/pages/Billing.jsx
import React, { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import { useAuth } from '../context/AuthContext';
import { billingAPI } from '../services/api';
import DashboardLayout from '../components/DashboardLayout';
import { useToast } from '../context/ToastContext';
import {
  FileText,
  CreditCard,
  Eye,
  Download as DownloadIcon,
  Calendar,
  AlertCircle,
  CheckCircle,
} from 'lucide-react';

const Billing = () => {
  const [nextInvoice, setNextInvoice] = useState(null);
  const [overdueInvoices, setOverdueInvoices] = useState([]);
  const [paidInvoices, setPaidInvoices] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const toast = useToast();

  useEffect(() => {
    const loadInvoices = async () => {
      try {
        setLoading(true);
        setError(null);

        const response = await billingAPI.getInvoices();

        // A API retorna null para arrays vazios, então convertemos para array vazio
        setNextInvoice(response.next_invoice || null);
        setOverdueInvoices(Array.isArray(response.overdue_invoices) ? response.overdue_invoices : []);
        setPaidInvoices(Array.isArray(response.paid_invoices) ? response.paid_invoices : []);
      } catch (err) {
        console.error('Erro ao carregar faturas:', err);
        setError('Não foi possível carregar suas faturas. Verifique sua conexão.');
      } finally {
        setLoading(false);
      }
    };

    loadInvoices();
  }, []);

  // --- Estilos Específicos da Página ---
  const styles = {
    panel: {
      background: 'rgba(21, 26, 38, 0.6)',
      backdropFilter: 'blur(10px)',
      border: '1px solid rgba(88, 58, 255, 0.2)',
      borderRadius: '1rem',
      padding: '1.5rem',
      boxShadow: '0 8px 32px rgba(0, 0, 0, 0.3)',
      marginBottom: '2rem',
    },
    panelTitle: {
      fontSize: '1.25rem',
      fontWeight: 700,
      background: 'linear-gradient(135deg, #F8F9FA 0%, #583AFF 50%, #1AD2FF 100%)',
      WebkitBackgroundClip: 'text',
      WebkitTextFillColor: 'transparent',
      marginBottom: '1.5rem'
    },
    invoiceItem: {
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'space-between',
      padding: '1.25rem',
      borderRadius: '0.875rem',
      background: 'rgba(255,255,255,0.04)',
      border: '1px solid rgba(255,255,255,0.06)',
      marginBottom: '1rem',
      transition: 'all 0.3s ease'
    },
    muted: { color: '#B8BDC7', fontSize: '0.9rem' },
    statusBadge: {
      padding: '4px 12px',
      borderRadius: '20px',
      fontSize: '12px',
      fontWeight: 700,
      textTransform: 'uppercase',
      letterSpacing: '0.5px',
    },
    emptyState: {
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      padding: '3rem',
      textAlign: 'center',
      color: '#B8BDC7',
    },
    section: {
      marginBottom: '2.5rem',
    },
    sectionHeader: {
      display: 'flex',
      alignItems: 'center',
      gap: '0.75rem',
      marginBottom: '1.25rem',
      color: '#F8F9FA',
      fontSize: '1.125rem',
      fontWeight: 600,
    },
  };

  const getStatusStyle = (status) => {
    switch (status) {
      case 'due':
        return { ...styles.statusBadge, background: 'rgba(234, 179, 8, 0.15)', color: '#EAB308' };
      case 'overdue':
        return { ...styles.statusBadge, background: 'rgba(239, 68, 68, 0.15)', color: '#EF4444' };
      case 'paid':
        return { ...styles.statusBadge, background: 'rgba(34, 197, 94, 0.15)', color: '#22C55E' };
      default:
        return { ...styles.statusBadge, background: 'rgba(107, 114, 128, 0.15)', color: '#6B7280' };
    }
  };

  const formatDate = (dateString) => {
    if (!dateString) return '—';
    const date = new Date(dateString);
    return date.toLocaleDateString('pt-BR');
  };

  const formatPrice = (price) => {
    if (price == null) return 'R$ 0,00';
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(price);
  };

  const handleView = (invoice) => {
    toast.info(`Fatura: ${invoice.plan_name} — ${formatPrice(invoice.amount)} — Vencimento: ${formatDate(invoice.due_date)}`);
  };

  const handleDownloadPDF = (invoice) => {
    toast.info('Download de PDF será disponibilizado em breve.');
  };

  return (
    <DashboardLayout title="Faturas">
      {loading ? (
        <div style={{ display: 'flex', justifyContent: 'center', padding: '3rem', color: '#B8BDC7' }}>
          Carregando suas faturas...
        </div>
      ) : error ? (
        <div style={{
          background: 'rgba(239, 68, 68, 0.1)',
          border: '1px solid rgba(239, 68, 68, 0.3)',
          borderRadius: '1rem',
          padding: '1.5rem',
          color: '#EF4444'
        }}>
          {error}
        </div>
      ) : (
        <>
          {/* Próxima Fatura */}
          {nextInvoice && (
            <motion.section style={{ ...styles.panel, borderLeft: '4px solid #EAB308' }}>
              <div style={{ ...styles.sectionHeader, color: '#EAB308' }}>
                <AlertCircle size={20} />
                Próxima Fatura
              </div>
              <motion.div
                style={styles.invoiceItem}
                whileHover={{
                  background: 'rgba(234, 179, 8, 0.08)',
                  borderColor: 'rgba(234, 179, 8, 0.3)',
                }}
              >
                <div>
                  <div style={{ fontSize: '1.1rem', fontWeight: 600, color: '#F8F9FA', marginBottom: '0.25rem' }}>
                    {nextInvoice.plan_name}
                  </div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '1.5rem', marginTop: '0.5rem' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: '#B8BDC7' }}>
                      <Calendar size={14} />
                      <span style={styles.muted}>{formatDate(nextInvoice.due_date)}</span>
                    </div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: '#B8BDC7' }}>
                      <CreditCard size={14} />
                      <span style={styles.muted}>Assinatura</span>
                    </div>
                  </div>
                </div>

                <div style={{ display: 'flex', alignItems: 'center', gap: '1.5rem' }}>
                  <div style={{ textAlign: 'right' }}>
                    <div style={{ fontSize: '1.25rem', fontWeight: 800, color: '#F8F9FA' }}>
                      {formatPrice(nextInvoice.amount)}
                    </div>
                    <div style={getStatusStyle(nextInvoice.status)}>
                      A vencer
                    </div>
                  </div>
                  <div style={{ display: 'flex', gap: '0.75rem' }}>
                    <motion.button
                      style={{
                        width: '36px',
                        height: '36px',
                        borderRadius: '10px',
                        background: 'rgba(255,255,255,0.08)',
                        border: 'none',
                        color: '#EAB308',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        cursor: 'pointer',
                      }}
                      whileHover={{ background: 'rgba(234, 179, 8, 0.2)' }}
                      title="Visualizar"
                      onClick={() => handleView(nextInvoice)}
                    >
                      <Eye size={16} />
                    </motion.button>
                  </div>
                </div>
              </motion.div>
            </motion.section>
          )}

          {/* Faturas Vencidas */}
          {overdueInvoices.length > 0 && (
            <motion.section style={{ ...styles.panel, borderLeft: '4px solid #EF4444' }}>
              <div style={{ ...styles.sectionHeader, color: '#EF4444' }}>
                <AlertCircle size={20} />
                Faturas Vencidas ({overdueInvoices.length})
              </div>
              {overdueInvoices.map((invoice, idx) => (
                <motion.div
                  key={invoice.subscription_id}
                  style={styles.invoiceItem}
                  whileHover={{
                    background: 'rgba(239, 68, 68, 0.08)',
                    borderColor: 'rgba(239, 68, 68, 0.3)',
                  }}
                  initial={{ opacity: 0, x: -20 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ delay: idx * 0.05 }}
                >
                  <div>
                    <div style={{ fontSize: '1.1rem', fontWeight: 600, color: '#F8F9FA', marginBottom: '0.25rem' }}>
                      {invoice.plan_name}
                    </div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '1.5rem', marginTop: '0.5rem' }}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: '#B8BDC7' }}>
                        <Calendar size={14} />
                        <span style={styles.muted}>{formatDate(invoice.due_date)}</span>
                      </div>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: '#B8BDC7' }}>
                        <CreditCard size={14} />
                        <span style={styles.muted}>Assinatura</span>
                      </div>
                    </div>
                  </div>

                  <div style={{ display: 'flex', alignItems: 'center', gap: '1.5rem' }}>
                    <div style={{ textAlign: 'right' }}>
                      <div style={{ fontSize: '1.25rem', fontWeight: 800, color: '#F8F9FA' }}>
                        {formatPrice(invoice.amount)}
                      </div>
                      <div style={getStatusStyle(invoice.status)}>
                        Vencida
                      </div>
                    </div>
                    <div style={{ display: 'flex', gap: '0.75rem' }}>
                      <motion.button
                        style={{
                          width: '36px',
                          height: '36px',
                          borderRadius: '10px',
                          background: 'rgba(255,255,255,0.08)',
                          border: 'none',
                          color: '#EF4444',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          cursor: 'pointer',
                        }}
                        whileHover={{ background: 'rgba(239, 68, 68, 0.2)' }}
                        title="Visualizar"
                        onClick={() => handleView(invoice)}
                      >
                        <Eye size={16} />
                      </motion.button>
                      <motion.button
                        style={{
                          width: '36px',
                          height: '36px',
                          borderRadius: '10px',
                          background: 'rgba(255,255,255,0.08)',
                          border: 'none',
                          color: '#583AFF',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          cursor: 'pointer',
                        }}
                        whileHover={{ background: 'rgba(88, 58, 255, 0.2)' }}
                        title="Baixar PDF"
                        onClick={() => handleDownloadPDF(invoice)}
                      >
                        <DownloadIcon size={16} />
                      </motion.button>
                    </div>
                  </div>
                </motion.div>
              ))}
            </motion.section>
          )}

          {/* Faturas Pagas */}
          {paidInvoices.length > 0 && (
            <motion.section style={{ ...styles.panel, borderLeft: '4px solid #22C55E' }}>
              <div style={{ ...styles.sectionHeader, color: '#22C55E' }}>
                <CheckCircle size={20} />
                Faturas Pagas ({paidInvoices.length})
              </div>
              {paidInvoices.map((invoice, idx) => (
                <motion.div
                  key={invoice.subscription_id}
                  style={styles.invoiceItem}
                  whileHover={{
                    background: 'rgba(34, 197, 94, 0.08)',
                    borderColor: 'rgba(34, 197, 94, 0.3)',
                  }}
                  initial={{ opacity: 0, x: -20 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ delay: idx * 0.05 }}
                >
                  <div>
                    <div style={{ fontSize: '1.1rem', fontWeight: 600, color: '#F8F9FA', marginBottom: '0.25rem' }}>
                      {invoice.plan_name}
                    </div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '1.5rem', marginTop: '0.5rem' }}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: '#B8BDC7' }}>
                        <Calendar size={14} />
                        <span style={styles.muted}>{formatDate(invoice.due_date)}</span>
                      </div>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: '#B8BDC7' }}>
                        <CreditCard size={14} />
                        <span style={styles.muted}>Assinatura</span>
                      </div>
                    </div>
                  </div>

                  <div style={{ display: 'flex', alignItems: 'center', gap: '1.5rem' }}>
                    <div style={{ textAlign: 'right' }}>
                      <div style={{ fontSize: '1.25rem', fontWeight: 800, color: '#F8F9FA' }}>
                        {formatPrice(invoice.amount)}
                      </div>
                      <div style={getStatusStyle(invoice.status)}>
                        Pago
                      </div>
                    </div>
                    <div style={{ display: 'flex', gap: '0.75rem' }}>
                      <motion.button
                        style={{
                          width: '36px',
                          height: '36px',
                          borderRadius: '10px',
                          background: 'rgba(255,255,255,0.08)',
                          border: 'none',
                          color: '#22C55E',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          cursor: 'pointer',
                        }}
                        whileHover={{ background: 'rgba(34, 197, 94, 0.2)' }}
                        title="Visualizar"
                        onClick={() => handleView(invoice)}
                      >
                        <Eye size={16} />
                      </motion.button>
                      <motion.button
                        style={{
                          width: '36px',
                          height: '36px',
                          borderRadius: '10px',
                          background: 'rgba(255,255,255,0.08)',
                          border: 'none',
                          color: '#583AFF',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          cursor: 'pointer',
                        }}
                        whileHover={{ background: 'rgba(88, 58, 255, 0.2)' }}
                        title="Baixar PDF"
                        onClick={() => handleDownloadPDF(invoice)}
                      >
                        <DownloadIcon size={16} />
                      </motion.button>
                    </div>
                  </div>
                </motion.div>
              ))}
            </motion.section>
          )}

          {/* Estado vazio */}
          {!nextInvoice && overdueInvoices.length === 0 && paidInvoices.length === 0 && !loading && (
            <div style={styles.emptyState}>
              <FileText size={64} opacity={0.3} />
              <h3 style={{ marginTop: '1rem', color: '#F8F9FA', fontWeight: 600 }}>Nenhuma fatura encontrada</h3>
              <p style={{ marginTop: '0.5rem' }}>Suas faturas de assinatura e compras aparecerão aqui.</p>
            </div>
          )}
        </>
      )}
    </DashboardLayout>
  );
};

export default Billing;
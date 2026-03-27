import { motion, AnimatePresence } from 'framer-motion';
import { Circle, Package, Calendar, CreditCard, Settings, ChevronDown, Clock, MessageCircle } from 'lucide-react';
import { useState } from 'react';
import ClientChat from './ClientChat';

function ProjectCard({ project, index }) {
  const [isExpanded, setIsExpanded] = useState(false);
  const [showChat, setShowChat] = useState(false);

  const getStatusStyle = () => {
    if (project.status === 'desenvolvimento') return {
      background: 'transparent',
      color: '#FFD700',
      border: '1px solid rgba(255, 215, 0, 0.3)',
    };
    if (project.status === 'concluido') return {
      background: 'transparent',
      color: '#3B82F6',
      border: '1px solid rgba(59, 130, 246, 0.3)',
    };
    return {
      background: 'transparent',
      color: '#EF4444',
      border: '1px solid rgba(239, 68, 68, 0.3)',
    };
  };

  const getStatusLabel = () => {
    if (project.status === 'desenvolvimento') return 'EM DESENVOLVIMENTO';
    if (project.status === 'concluido') return 'CONCLUÍDO';
    return 'CANCELADO';
  };

  const getBorderColor = () => {
    if (project.status === 'desenvolvimento') return 'rgba(255, 215, 0, 0.15)';
    if (project.status === 'concluido') return 'rgba(59, 130, 246, 0.15)';
    return 'rgba(239, 68, 68, 0.15)';
  };

  const getHoverBorder = () => {
    if (project.status === 'desenvolvimento') return 'rgba(255, 215, 0, 0.6)';
    if (project.status === 'concluido') return 'rgba(59, 130, 246, 0.6)';
    return 'rgba(239, 68, 68, 0.6)';
  };

  const getHoverGlow = () => {
    if (project.status === 'desenvolvimento') return '0 0 0 1px rgba(255, 215, 0, 0.3), 0 20px 60px rgba(255, 215, 0, 0.15)';
    if (project.status === 'concluido') return '0 0 0 1px rgba(59, 130, 246, 0.3), 0 20px 60px rgba(59, 130, 246, 0.15)';
    return '0 0 0 1px rgba(239, 68, 68, 0.3), 0 20px 60px rgba(239, 68, 68, 0.15)';
  };

  const styles = {
    card: {
      background: 'rgba(15, 18, 25, 0.6)',
      backdropFilter: 'blur(20px)',
      border: `1px solid ${getBorderColor()}`,
      borderRadius: '20px',
      padding: '28px',
      transition: 'all 0.5s cubic-bezier(0.4, 0, 0.2, 1)',
      position: 'relative',
      overflow: 'hidden',
    },
    name: {
      fontSize: '20px',
      fontWeight: 700,
      color: '#F8F9FA',
      marginBottom: '8px',

      letterSpacing: '-0.01em',
    },
    statusBadge: {
      display: 'flex',
      alignItems: 'center',
      gap: '8px',
      padding: '6px 14px',
      borderRadius: '6px',
      fontSize: '11px',
      fontWeight: 700,
      textTransform: 'uppercase',
      letterSpacing: '1px',

    },
    header: {
      display: 'flex',
      justifyContent: 'space-between',
      alignItems: 'flex-start',
      marginBottom: '24px',
    },
    progressSection: {
      marginBottom: '20px',
    },
    progressLabel: {
      fontSize: '11px',
      color: '#B8BDC7',
      marginBottom: '12px',
      textTransform: 'uppercase',
      letterSpacing: '1.5px',
      fontWeight: 600,

    },
    stagesBar: {
      display: 'flex',
      gap: '5px',
      marginBottom: '8px',
    },
    stageStep: {
      flex: 1,
      height: '7px',
      borderRadius: '12px',
      background: 'rgba(255, 255, 255, 0.05)',
      transition: 'all 0.5s ease',
    },
    stageStepActive: {
      background: 'rgba(224, 26, 79, 0.3)',
    },
    stageStepCurrent: {
      background: 'linear-gradient(90deg, #E01A4F 0%, #FF6B35 50%, #FFD700 100%)',
      animation: 'pulse 2s ease-in-out infinite',
    },
    stageLabels: {
      display: 'flex',
      justifyContent: 'space-between',
      fontSize: '10px',
      color: '#6C7384',

      fontWeight: 500,
    },
    actionsRow: {
      display: 'flex',
      gap: '8px',
      paddingTop: '5px',
      justifyContent: 'center',
    },
    expandButton: {
      width: '100%',
      padding: '10px',
      background: 'transparent',
      border: '1px solid rgba(255, 255, 255, 0.12)',
      borderRadius: '8px',
      color: '#E8E9EB',
      fontSize: '13px',
      fontWeight: 600,
      cursor: 'pointer',
      transition: 'all 0.3s ease',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      gap: '8px',

    },
    chevron: {
      transition: 'transform 0.3s ease',
    },
    chevronRotated: {
      transform: 'rotate(180deg)',
    },
    detailsSection: {
      paddingTop: '24px',
      borderTop: '1px solid rgba(255, 255, 255, 0.06)',
      marginTop: '20px',
    },
    detailsGrid: {
      display: 'grid',
      gridTemplateColumns: '1fr 1fr',
      gap: '16px',
      marginBottom: '24px',
    },
    detailItem: {
      display: 'flex',
      flexDirection: 'column',
      gap: '6px',
    },
    detailLabel: {
      fontSize: '10px',
      color: '#B8BDC7',
      textTransform: 'uppercase',
      letterSpacing: '1.5px',
      display: 'flex',
      alignItems: 'center',
      gap: '6px',
      fontWeight: 600,

    },
    detailValue: {
      fontSize: '15px',
      fontWeight: 600,
      color: '#F8F9FA',

    },
    updatesFeed: {
      marginBottom: '20px',
    },
    feedTitle: {
      fontSize: '11px',
      color: '#B8BDC7',
      textTransform: 'uppercase',
      letterSpacing: '1.5px',
      fontWeight: 600,
      marginBottom: '16px',

    },
    feedItem: {
      padding: '12px 0',
      borderBottom: '1px solid rgba(255, 255, 255, 0.04)',
      display: 'flex',
      gap: '12px',
      alignItems: 'flex-start',
    },
    feedTimestamp: {
      fontSize: '11px',
      color: '#6C7384',
      fontFamily: 'monospace',
      minWidth: '120px',
      display: 'flex',
      alignItems: 'center',
      gap: '6px',
    },
    feedMessage: {
      fontSize: '13px',
      color: '#E8E9EB',
      lineHeight: '1.6',

    },
    expandedActions: {
      display: 'flex',
      gap: '12px',
      marginTop: '20px',
    },
    actionButton: {
      flex: 1,
      padding: '12px',
      background: 'transparent',
      border: '1px solid rgba(255, 255, 255, 0.12)',
      borderRadius: '8px',
      color: '#E8E9EB',
      fontSize: '13px',
      fontWeight: 600,
      cursor: 'pointer',
      transition: 'all 0.3s ease',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      gap: '6px',

    },
  };

  return (
    <motion.div
      style={styles.card}
      whileHover={{
        borderColor: getHoverBorder(),
        boxShadow: getHoverGlow(),
        y: -4,
      }}
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.1, duration: 0.5, ease: [0.4, 0, 0.2, 1] }}
    >
      {/* HEADER */}
      <div style={styles.header}>
        <div>
          <h3 style={styles.name}>{project.name}</h3>
        </div>
        <div style={{
          ...styles.statusBadge,
          ...getStatusStyle(),
        }}>
          <Circle size={6} fill="currentColor" />
          {getStatusLabel()}
        </div>
      </div>

      {/* PROGRESS BAR */}
      <div style={styles.progressSection}>
        <div style={styles.progressLabel}>Progresso do Projeto</div>
        <div style={styles.stagesBar}>
          {project.stages.map((_, idx) => {
            const isActive = idx < project.currentStage;
            const isCurrent = idx === project.currentStage;
            return (
              <div
                key={idx}
                style={{
                  ...styles.stageStep,
                  ...(isActive ? styles.stageStepActive : {}),
                  ...(isCurrent ? styles.stageStepCurrent : {}),
                }}
              />
            );
          })}
        </div>
        <div style={styles.stageLabels}>
          {project.stages.map((stage, idx) => (
            <span
              key={idx}
              style={{
                color: idx <= project.currentStage ? '#F8F9FA' : '#6C7384',
                fontWeight: idx === project.currentStage ? 600 : 400,
              }}
            >
              {stage}
            </span>
          ))}
        </div>
      </div>

      {/* ACTIONS */}
      <div style={styles.actionsRow}>
        <motion.div
          style={styles.expandButton}
          whileHover={{
            background: 'rgba(255, 255, 255, 0.05)',
            borderColor: 'rgba(255, 255, 255, 0.2)',
          }}
          onClick={() => setIsExpanded(!isExpanded)}
        >
          Ver Detalhes
          <motion.div
            style={{
              ...styles.chevron,
              ...(isExpanded ? styles.chevronRotated : {}),
            }}
          >
            <ChevronDown size={16} />
          </motion.div>
        </motion.div>
      </div>

      {/* EXPANDED DETAILS SECTION */}
      <AnimatePresence>
        {isExpanded && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: 'auto', opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.3, ease: [0.4, 0, 0.2, 1] }}
            style={{ overflow: 'hidden' }}
          >
            <div style={styles.detailsSection}>
              {showChat ? (
                <ClientChat subscriptionId={project.id} onBack={() => setShowChat(false)} />
              ) : (
                <>
                  {/* KEY INFORMATION GRID */}
                  <div style={styles.detailsGrid}>
                    <div style={styles.detailItem}>
                      <div style={styles.detailLabel}>
                        <Package size={12} />
                        Plano Contratado
                      </div>
                      <div style={styles.detailValue}>{project.plan}</div>
                    </div>

                    <div style={styles.detailItem}>
                      <div style={styles.detailLabel}>
                        <Calendar size={12} />
                        Data de Início
                      </div>
                      <div style={styles.detailValue}>{project.startDate}</div>
                    </div>

                    <div style={styles.detailItem}>
                      <div style={styles.detailLabel}>
                        <CreditCard size={12} />
                        Próxima Cobrança
                      </div>
                      <div style={styles.detailValue}>{project.nextBilling}</div>
                    </div>
                  </div>

                  {/* UPDATES FEED */}
                  <div style={styles.updatesFeed}>
                    <div style={styles.feedTitle}>Diário de Bordo</div>

                    {(!project.logs || project.logs.length === 0) ? (
                      <div style={{ ...styles.feedMessage, fontStyle: 'italic', opacity: 0.7 }}>
                        Nenhuma atividade recente registrada.
                      </div>
                    ) : (
                      project.logs.map((log, idx) => {
                        const date = new Date(log.createdAt);
                        const formattedDate = new Intl.DateTimeFormat('pt-BR', {
                          day: '2-digit',
                          month: '2-digit',
                          hour: '2-digit',
                          minute: '2-digit'
                        }).format(date);

                        return (
                          <motion.div
                            key={log.id || idx}
                            style={styles.feedItem}
                            initial={{ opacity: 0, x: -20 }}
                            animate={{ opacity: 1, x: 0 }}
                            transition={{ delay: idx * 0.1 }}
                          >
                            <div style={styles.feedTimestamp}>
                              <Clock size={12} />
                              {formattedDate}
                            </div>
                            <div style={styles.feedMessage}>{log.message}</div>
                          </motion.div>
                        );
                      })
                    )}
                  </div>

                  {/* ACTION BUTTONS */}
                  <div style={styles.expandedActions}>
                    <motion.div
                      style={styles.actionButton}
                      whileHover={{
                        background: 'var(--gradient-cta)',
                        borderColor: '#E01A4F',
                        color: '#FFFFFF',
                        scale: 1.02,
                      }}
                      onClick={() => setShowChat(true)}
                    >
                      <MessageCircle size={16} />
                      Chat com Suporte
                    </motion.div>
                    <motion.div
                      style={styles.actionButton}
                      whileHover={{
                        background: 'linear-gradient(135deg, #FFD700 0%, #FF6B35 100%)',
                        borderColor: '#FFD700',
                        color: '#0A0E1A',
                        scale: 1.02,
                      }}
                    >
                      <CreditCard size={16} />
                      Ver Fatura
                    </motion.div>
                  </div>
                </>
              )}
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </motion.div>
  );
}

export default ProjectCard;

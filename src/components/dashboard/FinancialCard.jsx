import { motion } from 'framer-motion';

function FinancialCard({ data, type, index }) {
  const getHoverStyle = () => {
    if (type === 'balance') return {
      borderColor: 'rgba(224, 26, 79, 0.3)',
      boxShadow: '0 10px 30px rgba(224, 26, 79, 0.1)',
    };
    if (type === 'paid') return {
      borderColor: 'rgba(34, 197, 94, 0.3)',
      boxShadow: '0 10px 30px rgba(34, 197, 94, 0.1)',
    };
    return {
      borderColor: 'rgba(255, 215, 0, 0.3)',
      boxShadow: '0 10px 30px rgba(255, 215, 0, 0.1)',
    };
  };

  const styles = {
    card: {
      background: 'linear-gradient(135deg, rgba(21, 26, 38, 0.9) 0%, rgba(19, 24, 36, 0.95) 100%)',
      border: '1px solid rgba(224, 26, 79, 0.1)',
      borderRadius: '14px',
      padding: '24px',
      backdropFilter: 'blur(10px)',
      transition: 'all 0.3s',
    },
    label: {
      fontSize: '14px',
      color: 'var(--text-secondary)',
      marginBottom: '10px',
      textTransform: 'uppercase',
      letterSpacing: '0.5px',
    },
    amount: {
      fontSize: '34px',
      fontWeight: 700,
      background: 'var(--gradient-primary)',
      WebkitBackgroundClip: 'text',
      WebkitTextFillColor: 'transparent',
      marginBottom: '14px',
    },
    meta: {
      fontSize: '15px',
      color: 'var(--text-muted)',
      lineHeight: '1.6',
    },
  };

  return (
    <motion.div
      style={styles.card}
      whileHover={getHoverStyle()}
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.1 }}
    >
      <div style={styles.label}>{data.label}</div>
      <div style={styles.amount}>{data.amount}</div>
      <div style={styles.meta}>
        {type === 'balance' && 'Disponível para uso'}
        {type === 'paid' && (
          <>
            {data.date}<br />
            {data.description}
          </>
        )}
        {type === 'pending' && (
          <>
            Vencimento: {data.dueDate}<br />
            {data.description}
          </>
        )}
      </div>
    </motion.div>
  );
}

export default FinancialCard;

import { motion } from 'framer-motion';

function DigitalAssetCard({ asset, index }) {
  const styles = {
    card: {
      background: 'linear-gradient(135deg, rgba(21, 26, 38, 0.9) 0%, rgba(19, 24, 36, 0.95) 100%)',
      border: '1px solid rgba(224, 26, 79, 0.1)',
      borderRadius: '14px',
      padding: '24px',
      backdropFilter: 'blur(10px)',
      transition: 'all 0.3s',
    },
    iconBox: {
      fontSize: '38px',
      marginBottom: '14px',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      width: '67px',
      height: '67px',
      borderRadius: '14px',
      background: 'var(--gradient-primary)',
    },
    name: {
      fontSize: '17px',
      fontWeight: 600,
      color: 'var(--text-primary)',
      marginBottom: '8px',
    },
    size: {
      fontSize: '14px',
      color: 'var(--text-muted)',
      marginBottom: '20px',
    },
    button: {
      width: '100%',
      padding: '12px',
      background: 'transparent',
      border: '2px solid var(--accent-red)',
      borderRadius: '10px',
      color: 'var(--accent-red)',
      fontSize: '14px',
      fontWeight: 600,
      cursor: 'pointer',
      transition: 'all 0.3s',
    },
  };

  // Function to handle download action
  const handleDownload = () => {
    // In a real implementation, this would trigger a download
    // Download initiated
    // For now, we'll just show an alert
    alert(`Download de "${asset.name}" iniciado!`);
  };

  return (
    <motion.div
      style={styles.card}
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.1 }}
      whileHover={{
        borderColor: 'rgba(224, 26, 79, 0.3)',
        boxShadow: '0 10px 30px rgba(224, 26, 79, 0.1)',
      }}
    >
      <div style={styles.iconBox}>
        {asset.icon}
      </div>
      <div style={styles.name}>{asset.name}</div>
      <div style={styles.size}>{asset.size}</div>
      {asset.showDownload ? (
        <motion.button
          whileHover={{ scale: 1.02 }}
          whileTap={{ scale: 0.98 }}
          style={styles.button}
          onClick={handleDownload}
        >
          Baixar
        </motion.button>
      ) : (
        <div style={{
          ...styles.button,
          borderColor: 'rgba(255, 255, 255, 0.1)',
          color: 'var(--text-muted)',
          cursor: 'not-allowed',
        }}>
          Não disponível
        </div>
      )}
    </motion.div>
  );
}

export default DigitalAssetCard;
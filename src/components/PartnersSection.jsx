import { motion } from 'framer-motion';

function PartnersSection({ inline = false, variant = 'tri' }) {
  const partners = [
    { name: 'MagnoHost', logo: '/magnohost-logo.png', url: 'https://magnohost.com.br' },
    { name: 'Kingo Network', logo: '/kingonetwork-logo.png', url: 'https://kingo.network' },
    { name: 'Fatal-Info', logo: '/fatal-info-logo.png', url: 'https://fatal-info.com.br' }
  ];

  const styles = {
    section: {
      padding: '6rem 0',
      background: 'var(--bg-secondary)',
      borderTop: '1px solid rgba(248, 249, 250, 0.05)',
    },
    container: {
      maxWidth: '1400px',
      margin: '0 auto',
      padding: '0 2rem',
    },
    sectionTitle: {
      fontSize: 'clamp(2.5rem, 5vw, 4rem)',
      fontWeight: 900,
      marginBottom: '1rem',
      letterSpacing: '-0.02em',
      color: 'var(--text-primary)',
      textAlign: 'center',
    },
    sectionDescription: {
      fontSize: '1.25rem',
      color: 'var(--text-secondary)',
      fontWeight: 400,
      textAlign: 'center',
      marginBottom: '4rem',
    },
    partnersGrid: {
      display: 'grid',
      gridTemplateColumns: 'repeat(2, 1fr)',
      gridAutoRows: 'auto',
      justifyContent: 'center',
      alignItems: 'center',
      gap: '4rem',
      maxWidth: '600px',
      margin: '0 auto',
    },
    partnerCard: {
      padding: '2rem',
      background: 'rgba(248, 249, 250, 0.02)',
      borderRadius: '1rem',
      border: '1px solid rgba(248, 249, 250, 0.08)',
      transition: 'all 0.3s',
      cursor: 'pointer',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      minWidth: '200px',
    },
    partnerLogo: {
      width: '160px',
      height: '60px',
      objectFit: 'contain',
      filter: 'grayscale(100%) brightness(0.8)',
      transition: 'filter 0.3s',
    },
    divider: {
      width: '100%',
      height: '1px',
      background: 'linear-gradient(90deg, transparent 0%, rgba(224, 26, 79, 0.3) 50%, transparent 100%)',
      marginTop: '4rem',
      transformOrigin: 'center',
    },
  };

  if (inline) {
    const gridStyle = {
      ...styles.partnersGrid,
      ...(variant === 'row3' ? { gridTemplateColumns: 'repeat(3, 1fr)', maxWidth: '100%', gap: '2rem' } : {}),
    };
    return (
      <div style={{ marginTop: '2rem' }}>
        <div style={gridStyle}>
          {partners.map((partner, index) => (
            <motion.a
              key={partner.name}
              href={partner.url}
              target="_blank"
              rel="noopener noreferrer"
              initial={{ opacity: 0, scale: 0.9 }}
              whileInView={{ opacity: 1, scale: 1 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: index * 0.15 }}
              whileHover={{
                scale: 1.05,
                borderColor: 'rgba(224, 26, 79, 0.3)',
              }}
              style={{
                ...styles.partnerCard,
                ...(variant === 'tri' && index === 0 ? { gridColumn: '1 / span 2', justifySelf: 'center' } : {}),
                textDecoration: 'none',
              }}
              onMouseEnter={(e) => {
                const img = e.currentTarget.querySelector('img');
                if (img) img.style.filter = 'grayscale(0%) brightness(1)';
              }}
              onMouseLeave={(e) => {
                const img = e.currentTarget.querySelector('img');
                if (img) img.style.filter = 'grayscale(100%) brightness(0.8)';
              }}
            >
              <img
                src={partner.logo}
                alt={partner.name}
                style={{ ...styles.partnerLogo, ...(variant === 'row3' ? { width: '140px', height: '50px' } : {}) }}
              />
            </motion.a>
          ))}
        </div>
      </div>
    );
  }

  return (
    <section id="parceiros" style={styles.section}>
      <div style={styles.container}>
        <div style={styles.partnersGrid}>
          {partners.map((partner, index) => (
            <motion.a
              key={partner.name}
              href={partner.url}
              target="_blank"
              rel="noopener noreferrer"
              initial={{ opacity: 0, scale: 0.9 }}
              whileInView={{ opacity: 1, scale: 1 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: index * 0.15 }}
              whileHover={{
                scale: 1.05,
                borderColor: 'rgba(224, 26, 79, 0.3)',
              }}
              style={{
                ...styles.partnerCard,
                ...(index === 0 && {
                  gridColumn: '1 / span 2',
                  justifySelf: 'center',
                }),
                textDecoration: 'none',
              }}
              onMouseEnter={(e) => {
                const img = e.currentTarget.querySelector('img');
                if (img) img.style.filter = 'grayscale(0%) brightness(1)';
              }}
              onMouseLeave={(e) => {
                const img = e.currentTarget.querySelector('img');
                if (img) img.style.filter = 'grayscale(100%) brightness(0.8)';
              }}
            >
              <img
                src={partner.logo}
                alt={partner.name}
                style={styles.partnerLogo}
              />
            </motion.a>
          ))}
        </div>

        <motion.div
          initial={{ scaleX: 0 }}
          whileInView={{ scaleX: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 1, delay: 0.5 }}
          style={styles.divider}
        />
      </div>
    </section>
  );
}

export default PartnersSection;

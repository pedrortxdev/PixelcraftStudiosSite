import { motion } from 'framer-motion';
import { Zap, Box } from 'lucide-react';

function ProductsSection() {
  const products = [
    {
      icon: Zap,
      iconBg: 'linear-gradient(135deg, var(--accent-red) 0%, var(--accent-orange) 100%)',
      title: 'Plugins Exclusivos',
      description: 'Plugins desenvolvidos sob medida para as necessidades únicas do seu servidor. Licenças únicas disponíveis — compre uma vez, é seu para sempre.',
      features: [
        'Desenvolvimento customizado',
        'Otimização de performance',
        'Suporte técnico incluso'
      ],
      accentColor: 'var(--accent-red)',
    },
    {
      icon: Box,
      iconBg: 'linear-gradient(135deg, var(--accent-orange) 0%, var(--accent-yellow) 100%)',
      title: 'Mapas Sob Demanda',
      description: 'Mundos personalizados, spawns épicos e arenas competitivas construídas por especialistas. Cada mapa é uma obra de arte funcional.',
      features: [
        'Design profissional',
        'Otimizados para gameplay',
        'Licença de uso exclusiva'
      ],
      accentColor: 'var(--accent-orange)',
    },
  ];

  const styles = {
    section: {
      padding: '8rem 0',
      background: 'var(--bg-primary)',
      position: 'relative',
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
    productsGrid: {
      display: 'grid',
      gridTemplateColumns: 'repeat(auto-fit, minmax(450px, 1fr))',
      gap: '3rem',
      maxWidth: '1200px',
      margin: '0 auto',
    },
    productCard: {
      background: 'var(--bg-card)',
      padding: '3rem',
      borderRadius: '1rem',
      border: '1px solid rgba(248, 249, 250, 0.1)',
      transition: 'all 0.3s',
    },
    iconContainer: {
      width: '5rem',
      height: '5rem',
      borderRadius: '1rem',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      marginBottom: '1.5rem',
      boxShadow: '0 10px 30px rgba(0, 0, 0, 0.3)',
    },
    productTitle: {
      fontSize: '1.75rem',
      fontWeight: 800,
      marginBottom: '1rem',
      color: 'var(--text-primary)',
    },
    productDescription: {
      color: 'var(--text-secondary)',
      fontSize: '1.05rem',
      lineHeight: '1.7',
      marginBottom: '1.5rem',
    },
    featuresList: {
      listStyle: 'none',
      padding: 0,
      margin: 0,
    },
    featureItem: {
      display: 'flex',
      alignItems: 'center',
      gap: '0.75rem',
      marginBottom: '0.75rem',
      color: 'var(--text-secondary)',
      fontSize: '0.95rem',
    },
    featureBullet: {
      fontSize: '1.25rem',
    },
  };

  return (
    <section id="produtos" style={styles.section}>
      <div style={styles.container}>
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
        >
          <h2 style={styles.sectionTitle}>
            <span className="text-gradient">Produtos Exclusivos</span>
          </h2>
          <p style={styles.sectionDescription}>
            Soluções customizadas para elevar seu servidor
          </p>
        </motion.div>

        <div style={styles.productsGrid}>
          {products.map((product, index) => {
            const Icon = product.icon;
            return (
              <motion.div
                key={product.title}
                initial={{ opacity: 0, x: index % 2 === 0 ? -30 : 30 }}
                whileInView={{ opacity: 1, x: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.6 }}
                whileHover={{
                  y: -8,
                  boxShadow: `0 25px 50px ${product.accentColor}40`,
                }}
                style={styles.productCard}
              >
                <div
                  style={{
                    ...styles.iconContainer,
                    background: product.iconBg,
                  }}
                >
                  <Icon size={32} color="white" />
                </div>

                <h3 style={styles.productTitle}>{product.title}</h3>

                <p style={styles.productDescription}>
                  {product.description.split('—').map((part, i) => (
                    <span key={i}>
                      {i > 0 && '—'}
                      {i === 1 ? (
                        <strong style={{ color: product.accentColor }}>{part}</strong>
                      ) : (
                        part
                      )}
                    </span>
                  ))}
                </p>

                <ul style={styles.featuresList}>
                  {product.features.map((feature, i) => (
                    <li key={i} style={styles.featureItem}>
                      <span style={{ ...styles.featureBullet, color: product.accentColor }}>
                        ✦
                      </span>
                      {feature}
                    </li>
                  ))}
                </ul>
              </motion.div>
            );
          })}
        </div>
      </div>
    </section>
  );
}

export default ProductsSection;

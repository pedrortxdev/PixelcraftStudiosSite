import { motion } from 'framer-motion';
import { Link } from 'react-router-dom';

function Footer() {
  const currentYear = new Date().getFullYear();

  const styles = {
    footer: {
      padding: '6rem 0 3rem',
      background: 'var(--bg-primary)',
      borderTop: '1px solid var(--border-subtle)',
    },
    container: {
      maxWidth: '1400px',
      margin: '0 auto',
      padding: '0 2rem',
    },
    footerGrid: {
      display: 'grid',
      gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))',
      gap: '3rem',
      marginBottom: '4rem',
    },
    footerColumn: {
      display: 'flex',
      flexDirection: 'column',
    },
    footerTitle: {
      fontSize: '1rem',
      fontWeight: 700,
      color: 'var(--text-primary)',
      marginBottom: '1.5rem',
      letterSpacing: '-0.01em',
    },
    footerLink: {
      color: 'var(--text-secondary)',
      textDecoration: 'none',
      fontSize: '0.95rem',
      marginBottom: '0.75rem',
      transition: 'color var(--transition-normal)',
      cursor: 'pointer',
      display: 'block',
    },
    footerBottom: {
      paddingTop: '2rem',
      borderTop: '1px solid var(--border-subtle)',
      display: 'flex',
      justifyContent: 'space-between',
      alignItems: 'center',
      flexWrap: 'wrap',
      gap: '1rem',
    },
    copyright: {
      color: 'var(--text-muted)',
      fontSize: '0.875rem',
    },
    socialLinks: {
      display: 'flex',
      gap: '1.5rem',
    },
    socialLink: {
      width: '2.5rem',
      height: '2.5rem',
      borderRadius: '50%',
      border: '1px solid rgba(248, 249, 250, 0.1)',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      color: 'var(--text-secondary)',
      transition: 'all var(--transition-normal)',
      cursor: 'pointer',
      textDecoration: 'none',
    },
    logo: {
      fontSize: 'var(--title-h4)',
      fontWeight: 900,
      color: 'var(--text-primary)',
      marginBottom: '1rem',
      letterSpacing: '-0.02em',
    },
    tagline: {
      color: 'var(--text-secondary)',
      fontSize: '0.95rem',
      lineHeight: '1.6',
      maxWidth: '280px',
    },
  };

  const footerLinks = {
    'Serviços': [
      { label: 'Planos', to: '/#planos' },
      { label: 'Loja', to: '/shop' },
      { label: 'Suporte', to: '/suporte' },
    ],
    'Recursos': [
      { label: 'Dashboard', to: '/dashboard' },
      { label: 'Downloads', to: '/downloads' },
      { label: 'Projetos', to: '/projetos' },
    ],
    'Conta': [
      { label: 'Configurações', to: '/configuracoes' },
      { label: 'Carteira', to: '/carteira' },
      { label: 'Histórico', to: '/history' },
    ],
  };

  const socialLinks = [
    { name: 'Discord', url: 'https://discord.gg/pixelcraft', icon: 'D' },
    { name: 'Twitter', url: 'https://twitter.com/pixelcraft', icon: 'X' },
    { name: 'GitHub', url: 'https://github.com/pixelcraft', icon: 'G' },
  ];

  return (
    <footer style={styles.footer}>
      <div style={styles.container}>
        <div style={styles.footerGrid}>
          <div style={styles.footerColumn}>
            <div style={styles.logo}>Pixelcraft</div>
            <p style={styles.tagline}>
              A arquitetura do seu mundo. Servidores de Minecraft construídos com precisão e paixão.
            </p>
          </div>

          {Object.entries(footerLinks).map(([title, links]) => (
            <div key={title} style={styles.footerColumn}>
              <h4 style={styles.footerTitle}>{title}</h4>
              {links.map((link) => (
                <motion.div key={link.label} whileHover={{ x: 4 }}>
                  <Link
                    to={link.to}
                    style={styles.footerLink}
                    onMouseEnter={(e) => e.target.style.color = 'var(--accent-red)'}
                    onMouseLeave={(e) => e.target.style.color = 'var(--text-secondary)'}
                  >
                    {link.label}
                  </Link>
                </motion.div>
              ))}
            </div>
          ))}
        </div>

        <div style={styles.footerBottom}>
          <div style={styles.copyright}>
            © {currentYear} Pixelcraft. Todos os direitos reservados.
          </div>

          <div style={styles.socialLinks}>
            {socialLinks.map((social) => (
              <motion.a
                key={social.name}
                href={social.url}
                target="_blank"
                rel="noopener noreferrer"
                title={social.name}
                aria-label={social.name}
                style={styles.socialLink}
                whileHover={{
                  borderColor: 'var(--accent-red)',
                  color: 'var(--accent-red)',
                  scale: 1.1,
                }}
                whileTap={{ scale: 0.95 }}
              >
                {social.icon}
              </motion.a>
            ))}
          </div>
        </div>
      </div>
    </footer>
  );
}

export default Footer;


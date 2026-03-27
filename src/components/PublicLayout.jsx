import React from 'react';
import { motion } from 'framer-motion';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import Footer from './Footer';

const PublicLayout = ({ children }) => {
    const navigate = useNavigate();
    const location = useLocation();

    const isHome = location.pathname === '/';

    const styles = {
        nav: {
            position: 'fixed',
            top: 0,
            left: 0,
            right: 0,
            zIndex: 1000,
            background: 'rgba(10, 14, 26, 0.9)',
            backdropFilter: 'blur(20px)',
            borderBottom: '1px solid rgba(248, 249, 250, 0.05)',
            padding: '1.5rem 0',
        },
        container: {
            maxWidth: '1400px',
            margin: '0 auto',
            padding: '0 2rem',
        },
        navContent: {
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
        },
        logo: {
            display: 'flex',
            alignItems: 'center',
            gap: '0.75rem',
            fontSize: 'var(--title-h4)',
            fontWeight: 900,
            color: 'var(--text-primary)',
            cursor: 'pointer',
            letterSpacing: '-0.02em',
            textDecoration: 'none'
        },
        logoImage: {
            width: '32px',
            height: '32px',
            filter: 'drop-shadow(0 2px 8px rgba(224, 26, 79, 0.3))',
        },
        navLinks: {
            display: 'flex',
            gap: '3rem',
            alignItems: 'center',
        },
        navLink: {
            color: 'var(--text-secondary)',
            fontWeight: 500,
            textDecoration: 'none',
            fontSize: '0.95rem',
            transition: 'color 0.3s',
            cursor: 'pointer'
        },
        ctaButton: {
            display: 'inline-flex',
            alignItems: 'center',
            gap: '0.75rem',
            padding: '0.75rem 1.5rem',
            background: 'var(--gradient-primary)',
            color: 'white',
            border: 'none',
            borderRadius: '0.5rem',
            fontWeight: 700,
            fontSize: '0.95rem',
            cursor: 'pointer',
            transition: 'all 0.3s',
            boxShadow: '0 10px 40px var(--accent-glow)',
            textDecoration: 'none'
        },
        pageWrapper: {
            minHeight: '100vh',
            background: 'var(--gradient-bg)'
        }
    };

    const handleNavClick = (e, hash) => {
        e.preventDefault();
        if (isHome) {
            const element = document.getElementById(hash);
            if (element) {
                element.scrollIntoView({ behavior: 'smooth' });
            }
        } else {
            navigate('/#' + hash);
        }
    };

    return (
        <div style={styles.pageWrapper}>
            <nav style={styles.nav}>
                <div style={styles.container}>
                    <div style={styles.navContent}>
                        <Link to="/" style={styles.logo}>
                            <img src="/pixelcraft-logo.png" alt="Pixelcraft" style={styles.logoImage} />
                            Pixelcraft
                        </Link>

                        <div style={styles.navLinks} className="desktop-only">
                            <a href="#planos" onClick={(e) => handleNavClick(e, 'planos')} style={styles.navLink}>Planos</a>
                            <a href="#produtos" onClick={(e) => handleNavClick(e, 'produtos')} style={styles.navLink}>Serviços</a>
                            <a href="#parceiros" onClick={(e) => handleNavClick(e, 'parceiros')} style={styles.navLink}>Parceiros</a>
                        </div>

                        <motion.div
                            whileHover={{ scale: 1.05, boxShadow: '0 20px 60px var(--accent-glow)' }}
                            whileTap={{ scale: 0.98 }}
                        >
                            <Link
                                to="/dashboard"
                                style={styles.ctaButton}
                            >
                                Área do Cliente
                            </Link>
                        </motion.div>
                    </div>
                </div>
            </nav>

            <main>
                {children}
            </main>

            <Footer />
        </div>
    );
};

export default PublicLayout;

import { ArrowRight, Shield, Sparkles, Star, Zap, Rocket } from 'lucide-react';
import { motion } from 'framer-motion';
import './App.css';
import HeroIllustration from './components/HeroIllustration';
import PricingSection from './components/PricingSection';
import HomeBentoCategories from './components/HomeBentoCategories';
import HomeExpressShowcase from './components/HomeExpressShowcase';
import PartnersSection from './components/PartnersSection';
import PublicLayout from './components/PublicLayout';
import { useNavigate } from 'react-router-dom';

function App() {
  const navigate = useNavigate();

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
    },
    heroSection: {
      position: 'relative',
      minHeight: '80vh', // Shrink height
      display: 'flex',
      alignItems: 'center',
      paddingTop: '10rem', // Push top down a bit
      paddingBottom: '2rem', // Destroy huge bottom gap
      overflow: 'hidden',
    },
    heroGrid: {
      display: 'grid',
      gap: '6rem',
      alignItems: 'center',
    },
    heroTitle: {
      fontSize: 'clamp(2.5rem, 8vw, 6rem)', // Reduzido minimo de 3.5rem para 2.5rem para nao explodir mobile
      fontWeight: 900,
      lineHeight: 1.1,
      marginBottom: '2rem',
      letterSpacing: '-0.03em',
      color: 'var(--text-primary)',
    },
    heroDescription: {
      fontSize: '1.25rem',
      color: 'var(--text-secondary)',
      fontWeight: 400,
      lineHeight: 1.7,
      marginBottom: '3rem',
      maxWidth: '600px',
    },
    ctaButton: {
      display: 'inline-flex',
      alignItems: 'center',
      gap: '0.75rem',
      padding: '1.25rem 2.5rem',
      background: 'var(--gradient-primary)',
      color: 'white',
      border: 'none',
      borderRadius: '0.5rem',
      fontWeight: 700,
      fontSize: '1rem',
      cursor: 'pointer',
      transition: 'all 0.3s',
      boxShadow: '0 10px 40px var(--accent-glow)',
    },
    ctaSecondary: {
      display: 'inline-flex',
      alignItems: 'center',
      gap: '0.75rem',
      padding: '1.25rem 2.5rem',
      background: 'transparent',
      color: 'var(--text-primary)',
      border: '2px solid rgba(248, 249, 250, 0.2)',
      borderRadius: '0.5rem',
      fontWeight: 600,
      fontSize: '1rem',
      cursor: 'pointer',
      transition: 'all 0.3s',
    },
    badge: {
      display: 'inline-flex',
      alignItems: 'center',
      gap: '0.5rem',
      background: 'rgba(224, 26, 79, 0.1)',
      border: '1px solid rgba(224, 26, 79, 0.3)',
      color: 'var(--accent-red)',
      padding: 'var(--btn-padding-sm)',
      borderRadius: '4px', // Cyber brutalist feeling (not rounded huge anymore)
      fontWeight: 600,
      fontSize: '0.875rem',
      marginBottom: '2rem',
    },
    statsCard: {
      textAlign: 'center',
      padding: '1rem 1.5rem',
    },
    statsNumber: {
      fontSize: 'var(--title-h2)',
      fontWeight: 900,
      marginBottom: '0.25rem',
      background: 'linear-gradient(135deg, #F8F9FA 0%, #E01A4F 100%)',
      WebkitBackgroundClip: 'text',
      WebkitTextFillColor: 'transparent',
    },
    statsLabel: {
      fontSize: '0.875rem',
      color: 'var(--text-muted)',
      fontWeight: 500,
    },
    sectionTitle: {
      fontSize: 'clamp(2.5rem, 5vw, 4rem)',
      fontWeight: 900,
      marginBottom: '1rem',
      letterSpacing: '-0.02em',
      color: 'var(--text-primary)',
    },
    sectionDescription: {
      fontSize: '1.25rem',
      color: 'var(--text-secondary)',
      fontWeight: 400,
    },
  };

  const fadeInUp = {
    hidden: { opacity: 0, y: 30 },
    visible: { opacity: 1, y: 0 }
  };

  return (
    <PublicLayout>
      {/* HERO SECTION - "A Arquitetura do Seu Mundo" */}
      <section style={styles.heroSection}>
        <div style={styles.container}>
          <div style={styles.heroGrid} className="home-hero-grid">
            {/* Left: Content */}
            <div style={{ position: 'relative', zIndex: 10 }}>
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.6 }}
              >
                <div style={styles.badge}>
                  <Shield size={16} />
                  <span>Desde 2021 · +500 Servidores Construídos</span>
                </div>

                <h1 style={styles.heroTitle}>
                  Tudo que você precisa<br />
                  <span className="text-gradient">para monetizar.</span>
                </h1>

                <p style={styles.heroDescription}>
                  Sistemas prontos, infraestrutura robusta e scripts exclusivos que vendem. Coloque seu projeto no ar hoje e comece a faturar.
                </p>

                <div style={{ display: 'flex', gap: '1.5rem', marginBottom: '4rem', flexWrap: 'wrap' }}>
                  <motion.button
                    style={styles.ctaButton}
                    whileHover={{ scale: 1.05, boxShadow: '0 20px 60px var(--accent-glow)' }}
                    whileTap={{ scale: 0.98 }}
                    onClick={() => navigate('/shop')}
                    className="btn-premium"
                  >
                    <Rocket size={20} />
                    Explorar Catálogo
                    <ArrowRight size={20} />
                  </motion.button>
                  <motion.button
                    style={styles.ctaSecondary}
                    whileHover={{ scale: 1.03, borderColor: 'rgba(224, 26, 79, 0.5)' }}
                    whileTap={{ scale: 0.98 }}
                    onClick={() => document.getElementById('planos')?.scrollIntoView({ behavior: 'smooth' })}
                  >
                    <Star size={20} />
                    Ver Planos
                  </motion.button>
                </div>

                {/* Trust Indicators with premium styling */}
                <motion.div
                  initial="hidden"
                  animate="visible"
                  variants={fadeInUp}
                  transition={{ delay: 0.4 }}
                  style={{
                    display: 'flex',
                    gap: '0.5rem',
                    alignItems: 'center',
                    justifyContent: 'center',
                    flexWrap: 'wrap', // CRITICAL pra nao vazar fora da tela
                    background: 'rgba(255, 255, 255, 0.03)',
                    borderRadius: '4px', // Brutalista
                    padding: 'var(--btn-padding-md)',
                    border: '1px solid rgba(255, 255, 255, 0.06)',
                  }}
                >
                  <motion.div
                    style={styles.statsCard}
                    whileHover={{ scale: 1.05 }}
                  >
                    <div style={styles.statsNumber} className="mono-data">24/7</div>
                    <div style={styles.statsLabel} className="mono-data">UPTIME SLA</div>
                  </motion.div>
                  <div style={{ width: '1px', height: '3rem', background: 'rgba(248, 249, 250, 0.1)' }} />
                  <motion.div
                    style={styles.statsCard}
                    whileHover={{ scale: 1.05 }}
                  >
                    <div style={styles.statsNumber} className="mono-data">500+</div>
                    <div style={styles.statsLabel} className="mono-data">NODES</div>
                  </motion.div>
                  <div style={{ width: '1px', height: '3rem', background: 'rgba(248, 249, 250, 0.1)' }} />
                  <motion.div
                    style={styles.statsCard}
                    whileHover={{ scale: 1.05 }}
                  >
                    <div style={styles.statsNumber} className="mono-data">4Y</div>
                    <div style={styles.statsLabel} className="mono-data">RUNTIME</div>
                  </motion.div>
                </motion.div>
              </motion.div>
            </div>

            {/* Right: Static High-End Image - Hidden on Mobile */}
            <motion.div
              className="desktop-only"
              initial={{ opacity: 0, scale: 0.9, x: 100 }}
              animate={{ opacity: 1, scale: 1, x: -120 }} /* Negative X purposefully overlaps the left column! */
              transition={{ duration: 0.8 }}
              style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', zIndex: 1, mixBlendMode: 'screen' }}
            >
              <img
                src="/principal.png"
                alt="Pixelcraft 3D Hero"
                style={{
                  width: '100%',
                  maxWidth: '550px',
                  filter: 'drop-shadow(0 20px 40px rgba(0,0,0,0.5))'
                }}
              />
            </motion.div>
          </div>
        </div>

        <div style={{ position: 'absolute', inset: 0, opacity: 0.3, overflow: 'hidden', pointerEvents: 'none' }}>
          <motion.div
            animate={{
              scale: [1, 1.2, 1],
              opacity: [0.3, 0.5, 0.3],
            }}
            transition={{ duration: 8, repeat: Infinity }}
            style={{
              position: 'absolute',
              top: '-20%',
              right: '-10%',
              width: '600px',
              height: '600px',
              background: 'radial-gradient(circle, rgba(224, 26, 79, 0.3) 0%, transparent 70%)',
              filter: 'blur(80px)',
            }}
          />
          <motion.div
            animate={{
              scale: [1, 1.3, 1],
              opacity: [0.2, 0.4, 0.2],
            }}
            transition={{ duration: 10, repeat: Infinity }}
            style={{
              position: 'absolute',
              bottom: '-20%',
              left: '-10%',
              width: '500px',
              height: '500px',
              background: 'radial-gradient(circle, rgba(255, 107, 53, 0.3) 0%, transparent 70%)',
              filter: 'blur(80px)',
            }}
          />
        </div>
      </section>

      {/* BENTO GRID - FAST NAVIGATION */}
      <HomeBentoCategories />

      {/* EXPRESS SHOWCASE - FAST SALES */}
      <HomeExpressShowcase />

      {/* PRICING SECTION - ENGINEERING SERVICES */}
      <PricingSection />

      <PartnersSection />
    </PublicLayout>
  );
}

export default App;

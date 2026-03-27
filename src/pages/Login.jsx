import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { ArrowRight, CheckCircle, Box, KeyRound, X, Loader2, Mail, Lock } from 'lucide-react';
import PartnersSection from '../components/PartnersSection';
import PublicLayout from '../components/PublicLayout';
import Input from '../components/shared/Input';
import { useAuth } from '../context/AuthContext.jsx';
import { useNavigate, Link, useLocation } from 'react-router-dom';
import { authAPI } from '../services/api';

const Login = () => {
  const [formData, setFormData] = useState({ email: '', password: '' });
  const [error, setError] = useState(null);
  const [loginLoading, setLoginLoading] = useState(false);
  const { user, login } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    if (user) {
      const returnUrl = location.state?.returnUrl || '/dashboard';
      navigate(returnUrl, { replace: true });
    }
  }, [user, navigate, location]);

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
      textDecoration: 'none',
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
      textDecoration: 'none',
    },

    page: { background: 'var(--gradient-bg)', minHeight: '100vh' },
    registerContainer: { padding: '4rem 2rem' },
    heroGrid: {
      display: 'grid',
      gridTemplateColumns: '1fr 1fr',
      gap: '4rem',
      alignItems: 'center',
      maxWidth: '1400px',
      margin: '0 auto',
      paddingTop: '2rem',
    },

    formCard: {
      background: 'var(--bg-card)',
      borderRadius: '1rem',
      padding: '2rem',
      boxShadow: '0 10px 40px rgba(0, 0, 0, 0.2)',
      maxWidth: '420px',
      width: '100%',
      textAlign: 'left',
      border: '1px solid rgba(248, 249, 250, 0.05)',
    },
    title: {
      fontSize: 'var(--title-h3)',
      fontWeight: 900,
      color: 'var(--text-primary)',
      marginBottom: '0.75rem',
    },
    subtitle: {
      fontSize: '0.95rem',
      color: 'var(--text-secondary)',
      marginBottom: '1.5rem',
    },
    formGroup: { marginBottom: '1rem' },
    label: {
      display: 'block',
      fontSize: '0.85rem',
      color: 'var(--text-secondary)',
      marginBottom: '0.5rem',
      fontWeight: 500,
    },
    input: {
      width: '100%',
      padding: '0.75rem 1rem',
      borderRadius: '0.5rem',
      border: '1px solid rgba(248, 249, 250, 0.1)',
      background: 'var(--bg-secondary)',
      color: 'var(--text-primary)',
      fontSize: '1rem',
    },
    submitButton: {
      display: 'inline-flex',
      alignItems: 'center',
      justifyContent: 'center',
      gap: '0.75rem',
      padding: '0.9rem 2rem',
      background: 'var(--gradient-primary)',
      color: 'white',
      border: 'none',
      borderRadius: '0.5rem',
      fontWeight: 700,
      fontSize: '1rem',
      cursor: 'pointer',
      transition: 'all 0.3s',
      boxShadow: '0 10px 40px var(--accent-glow)',
      width: '100%',
      marginTop: '1rem',
    },

    featureCard: {
      background: 'var(--bg-card)',
      borderRadius: '1rem',
      padding: '2rem',
      border: '1px solid rgba(248, 249, 250, 0.05)',
      maxWidth: '560px',
      width: '100%',
    },
    featureHeader: { display: 'flex', alignItems: 'center', gap: '1rem', marginBottom: '1rem' },
    featureIconWrap: {
      width: '44px',
      height: '44px',
      borderRadius: '0.75rem',
      background: 'linear-gradient(135deg, rgba(224, 26, 79, 0.2), rgba(255, 107, 53, 0.2))',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      border: '1px solid rgba(224, 26, 79, 0.35)'
    },
    featureTitle: { fontSize: 'var(--title-h4)', fontWeight: 800, color: 'var(--text-primary)' },
    featureText: { fontSize: '0.95rem', color: 'var(--text-secondary)', lineHeight: 1.7 },
    bullet: { display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--text-secondary)' },
    bullets: { display: 'grid', gap: '0.5rem', marginTop: '1rem' },
  };

  function handleChange(e) {
    setFormData((prev) => ({ ...prev, [e.target.name]: e.target.value }));
  }

  async function handleSubmit(e) {
    e.preventDefault();
    setError(null);
    if (!formData.email || !formData.password) {
      setError('Preencha email e senha.');
      return;
    }
    try {
      setLoginLoading(true);
      await login({ email: formData.email, password: formData.password });

      const returnUrl = location.state?.returnUrl || '/dashboard';
      navigate(returnUrl, { replace: true });
    } catch (err) {
      setError(err.message || 'Falha no login. Verifique suas credenciais.');
    } finally {
      setLoginLoading(false);
    }
  }


  return (
    <PublicLayout>
      <div style={styles.page}>
        <div style={{ paddingTop: '2rem' }}>
          <div style={styles.registerContainer}>
            <div style={styles.heroGrid} className="auth-hero-grid">
              {/* Card de Login */}
              <motion.div
                initial={{ opacity: 0, x: -50 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ duration: 0.8, delay: 0.2 }}
                style={styles.formCard}
              >
                <h2 style={styles.title}>Bem Vindo de Volta!</h2>
                <p style={styles.subtitle}>Estamos Felizes em Vê-lo Aqui Novamente!</p>
                {error && (
                  <p style={{ background: 'rgba(224, 26, 79, 0.15)', color: 'var(--accent-red)', padding: '0.5rem 0.75rem', borderRadius: '0.5rem', marginBottom: '0.75rem' }}>{error}</p>
                )}
                <form onSubmit={handleSubmit}>
                  <Input
                    id="email"
                    name="email"
                    type="email"
                    label="Email"
                    icon={Mail}
                    value={formData.email}
                    onChange={handleChange}
                    required
                  />
                  <Input
                    id="password"
                    name="password"
                    type="password"
                    label="Senha"
                    icon={Lock}
                    value={formData.password}
                    onChange={handleChange}
                    required
                  />
                  <motion.button
                    type="submit"
                    style={{ ...styles.submitButton, opacity: loginLoading ? 0.7 : 1 }}
                    disabled={loginLoading}
                    whileHover={{ scale: 1.05, boxShadow: '0 20px 60px var(--accent-glow)' }}
                    whileTap={{ scale: 0.98 }}
                  >
                    {loginLoading ? <Loader2 size={20} style={{ animation: 'spin 1s linear infinite' }} /> : null}
                    {loginLoading ? 'Entrando...' : 'Entrar'}
                    {!loginLoading && <ArrowRight size={20} />}
                  </motion.button>
                  <button
                    type="button"
                    onClick={() => navigate('/reset-password')}
                    style={{
                      background: 'transparent',
                      border: 'none',
                      color: 'var(--accent-red)',
                      cursor: 'pointer',
                      marginTop: '1rem',
                      fontSize: '0.9rem',
                      display: 'flex',
                      alignItems: 'center',
                      gap: '0.5rem',
                      width: '100%',
                      justifyContent: 'center'
                    }}
                  >
                    <KeyRound size={16} />
                    Esqueci minha senha
                  </button>

                  <div style={{
                    marginTop: '2rem',
                    textAlign: 'center',
                    borderTop: '1px solid rgba(248, 249, 250, 0.05)',
                    paddingTop: '1.5rem',
                  }}>
                    <p style={{
                      color: 'var(--text-secondary)',
                      fontSize: '0.9rem',
                      marginBottom: '1rem',
                    }}>
                      Não tem conta?
                    </p>
                    <Link
                      to="/register"
                      style={{
                        color: 'var(--text-primary)',
                        fontWeight: 700,
                        textDecoration: 'none',
                        fontSize: '0.95rem',
                        display: 'inline-flex',
                        alignItems: 'center',
                        gap: '0.5rem',
                        transition: 'color 0.3s',
                      }}
                      onMouseEnter={(e) => e.target.style.color = 'var(--accent-red)'}
                      onMouseLeave={(e) => e.target.style.color = 'var(--text-primary)'}
                    >
                      Criar uma conta
                      <ArrowRight size={16} />
                    </Link>
                  </div>
                </form>
              </motion.div>

              {/* Card de Feature */}
              <motion.div
                initial={{ opacity: 0, x: 50 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ duration: 0.8, delay: 0.3 }}
                style={styles.featureCard}
              >
                <div style={styles.featureHeader}>
                  <div style={styles.featureIconWrap}>
                    <Box size={24} color="#F8F9FA" />
                  </div>
                  <div style={styles.featureTitle}>Mapas Sob Demanda</div>
                </div>
                <p style={styles.featureText}>
                  Mundos personalizados, spawns épicos e arenas competitivas construídas por especialistas.
                  Cada mapa é uma obra de arte funcional.
                </p>
                <div style={styles.bullets}>
                  <div style={styles.bullet}><CheckCircle size={18} color="#ff6b35" /> Design profissional</div>
                  <div style={styles.bullet}><CheckCircle size={18} color="#ff6b35" /> Otimizados para gameplay</div>
                  <div style={styles.bullet}><CheckCircle size={18} color="#ff6b35" /> Licença de uso exclusiva</div>
                </div>
              </motion.div>
            </div>

            {/* Logos dos Parceiros - linha com 3 itens */}
            <div style={{ marginTop: '3rem' }}>
              <PartnersSection inline variant="row3" />
            </div>
          </div>
        </div>
      </div>

    </PublicLayout>
  );
};

export default Login;
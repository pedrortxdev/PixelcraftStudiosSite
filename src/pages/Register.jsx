import React, { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import { ArrowRight, Shield } from 'lucide-react';
import { useNavigate, Link, useLocation } from 'react-router-dom';
import { useAuth } from '../context/AuthContext.jsx';
import HeroIllustration from '../components/HeroIllustration';
import PartnersSection from '../components/PartnersSection';
import PublicLayout from '../components/PublicLayout';
import Input from '../components/shared/Input';
import { User, Mail, Lock, FileText, Smartphone, Hash } from 'lucide-react';

const Register = () => {
  const [formData, setFormData] = useState({
    full_name: '',
    username: '',
    email: '',
    cpf: '',
    discord_handle: '',
    whatsapp_phone: '',
    password: '',
  });

  const [error, setError] = useState(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const { user, register } = useAuth();

  useEffect(() => {
    if (user) {
      const returnUrl = location.state?.returnUrl || '/dashboard';
      navigate(returnUrl, { replace: true });
    }
  }, [user, navigate, location]);

  const formatCPF = (value) => {
    return value
      .replace(/\D/g, '')
      .replace(/(\d{3})(\d)/, '$1.$2')
      .replace(/(\d{3})(\d)/, '$1.$2')
      .replace(/(\d{3})(\d{1,2})/, '$1-$2')
      .replace(/(-\d{2})\d+?$/, '$1');
  };

  const handleChange = (e) => {
    let { name, value } = e.target;
    if (name === 'cpf') {
      value = formatCPF(value);
    }
    setFormData({ ...formData, [name]: value });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError(null);

    if (!formData.full_name || !formData.username || !formData.email || !formData.password || !formData.cpf) {
      setError("Por favor, preencha todos os campos obrigatórios.");
      return;
    }

    if (formData.cpf.length < 14) {
      setError("CPF incompleto.");
      return;
    }

    setIsSubmitting(true);

    try {
      // Prepare payload with cleaned CPF
      const payload = {
        ...formData,
        cpf: formData.cpf.replace(/\D/g, '')
      };

      // Use AuthContext.register
      await register(payload);
      const returnUrl = location.state?.returnUrl || '/dashboard';
      navigate(returnUrl, { replace: true });

    } catch (err) {
      console.error("Erro de registro:", err);
      const errorMsg = err.message || "Erro de registro.";

      if (JSON.stringify(errorMsg).includes("'len' tag")) {
        setError("CPF inválido (tamanho incorreto).");
      } else {
        setError(errorMsg);
      }
    } finally {
      setIsSubmitting(false);
    }
  };

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
    registerContainer: {
      padding: '4rem 2rem',
      background: 'var(--gradient-bg)',
    },
    heroGrid: {
      display: 'grid',
      gridTemplateColumns: '1fr 1fr',
      gap: '4rem',
      alignItems: 'center',
      maxWidth: '1400px',
      margin: '0 auto',
      paddingTop: '0',
    },
    formCard: {
      background: 'var(--bg-card)',
      borderRadius: '1rem',
      padding: '2.25rem',
      boxShadow: '0 10px 40px rgba(0, 0, 0, 0.2)',
      maxWidth: '460px',
      width: '100%',
      textAlign: 'center',
      border: '1px solid rgba(248, 249, 250, 0.05)',
    },
    title: {
      fontSize: '2.25rem',
      fontWeight: 900,
      color: 'var(--text-primary)',
      marginBottom: '1rem',
    },
    subtitle: {
      fontSize: '1rem',
      color: 'var(--text-secondary)',
      marginBottom: '2rem',
    },
    heroContent: {
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      textAlign: 'center',
    },
    formGroup: {
      marginBottom: '1.5rem',
      textAlign: 'left',
    },
    label: {
      display: 'block',
      fontSize: '0.9rem',
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
      transition: 'border-color 0.3s, box-shadow 0.3s',
    },
    submitButton: {
      display: 'inline-flex',
      alignItems: 'center',
      justifyContent: 'center',
      gap: '0.75rem',
      padding: '1rem 2.5rem',
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
      marginTop: '1.5rem',
    },
    loginLink: {
      display: 'block',
      marginTop: '1.5rem',
      color: 'var(--text-secondary)',
      textDecoration: 'none',
      fontSize: '0.9rem',
    },
    message: {
      marginTop: '1rem',
      padding: '0.75rem',
      borderRadius: '0.5rem',
      fontSize: '0.9rem',
      fontWeight: 500,
    },
    errorMessage: {
      backgroundColor: 'rgba(224, 26, 79, 0.2)',
      color: 'var(--accent-red)',
    },
  };

  return (
    <PublicLayout>
      <div style={{ background: 'var(--gradient-bg)', paddingTop: '2rem' }}>
        <div style={styles.registerContainer}>
          <div style={styles.heroGrid} className="auth-hero-grid">
            <motion.div
              initial={{ opacity: 0, x: -50 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.8, delay: 0.2 }}
              style={styles.formCard}
            >
              <h2 style={styles.title}>Crie sua conta</h2>
              <p style={styles.subtitle}>Junte-se à comunidade Pixelcraft e comece sua jornada!</p>

              {error && <p style={{ ...styles.message, ...styles.errorMessage }}>{typeof error === 'string' ? error : "Erro no formulário"}</p>}

              <form onSubmit={handleSubmit}>
                <Input
                  id="full_name"
                  name="full_name"
                  type="text"
                  label="Nome Completo"
                  icon={User}
                  value={formData.full_name}
                  onChange={handleChange}
                  required
                />
                <Input
                  id="cpf"
                  name="cpf"
                  type="text"
                  label="CPF"
                  icon={FileText}
                  value={formData.cpf}
                  onChange={handleChange}
                  placeholder="000.000.000-00"
                  maxLength="14"
                  required
                />
                <Input
                  id="username"
                  name="username"
                  type="text"
                  label="Nome de Usuário"
                  icon={Hash}
                  value={formData.username}
                  onChange={handleChange}
                  required
                />
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
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                  <Input
                    id="discord_handle"
                    name="discord_handle"
                    type="text"
                    label="Discord (Opcional)"
                    value={formData.discord_handle}
                    onChange={handleChange}
                  />
                  <Input
                    id="whatsapp_phone"
                    name="whatsapp_phone"
                    type="text"
                    label="WhatsApp (Opcional)"
                    icon={Smartphone}
                    value={formData.whatsapp_phone}
                    onChange={handleChange}
                    placeholder="(00) 00000-0000"
                    maxLength="15"
                  />
                </div>
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
                  style={styles.submitButton}
                  whileHover={{ scale: 1.05, boxShadow: '0 20px 60px var(--accent-glow)' }}
                  whileTap={{ scale: 0.98 }}
                >
                  Registrar
                  <ArrowRight size={20} />
                </motion.button>
              </form>
              <a href="/login" style={styles.loginLink}>Já tem uma conta? Faça login aqui.</a>
            </motion.div>

            <motion.div
              initial={{ opacity: 0, x: 50 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.8, delay: 0.4 }}
              style={styles.heroContent}
            >
              <HeroIllustration />
              <PartnersSection inline />
            </motion.div>
          </div>
        </div>
      </div>
    </PublicLayout >
  );
};

export default Register;
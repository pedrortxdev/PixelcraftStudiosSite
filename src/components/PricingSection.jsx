import { motion } from 'framer-motion';
import { CheckCircle, Loader2, Sparkles, ArrowRight } from 'lucide-react';
import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { adminAPI } from '../services/api';

function PricingSection() {
  const navigate = useNavigate();
  const [plans, setPlans] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchPlans = async () => {
      try {
        const data = await adminAPI.getPlans(); // Using the public endpoint we exposed
        setPlans(data || []);
      } catch (error) {
        console.error('Failed to fetch plans:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchPlans();
  }, []);

  const hardcodedPlans = [
    {
      name: 'Plano Básico Minecraft',
      price: 140,
      popular: false,
      features: [
        'Instalação de Servidores',
        'Configuração de plugins básicos',
        'Otimização Inicial',
        'Tradução de todos os plugins',
        'Suporte via ticket no Discord'
      ]
    },
    {
      name: 'Plano Avançado Minecraft',
      price: 240,
      popular: false,
      features: [
        'Configuração e tradução de plugins avançados',
        'Otimização avançada para reduzir lag',
        'Suporte técnico prioritário 24h',
        'Correção de ResourcePacks (ItemsAdder/Oraxen)',
        'Análise para identificar melhorias'
      ]
    },
    {
      name: 'Plano Network Minecraft',
      price: 300,
      popular: true,
      features: [
        'Tudo do plano Avançado, e mais:',
        'Instalação de Proxy + integração de servidores',
        'Suporte 24/7 via WhatsApp',
        'Otimização de ponta a ponta'
      ]
    },
    {
      name: 'Plano Sócio Minecraft',
      price: 480,
      popular: false,
      features: [
        'Serviço completo e sob medida',
        'Configuração de plugins e mods',
        'Testes de desempenho contínuos',
        'Análise de segurança',
        'Consultoria estratégica para seu servidor'
      ]
    },
    {
      name: 'Plano Básico Ragnarok',
      price: 160,
      popular: false,
      features: [
        'Instalação e configuração do emulador (rAthena / Hercules).',
        'Tradução completa do servidor (itens, NPCs, skills e mensagens).',
        'Configuração inicial do servidor (rates, drops, classes e mapas).',
        'Suporte técnico via ticket no Discord.'
      ]
    },
    {
      name: 'Plano Premium Ragnarok',
      price: 280,
      popular: true,
      features: [
        'Tudo do Plano Básico',
        'Criação e configuração de patcher automático.',
        'Desenvolvimento e edição de scripts personalizados.',
        'Configuração avançada de balanceamento (PvP, WoE e economia).',
        'Otimização de desempenho e redução de lag.'
      ]
    }
  ];

  const displayPlans = plans.length > 0 ? plans : hardcodedPlans;

  const styles = {
    section: {
      padding: '8rem 0',
      background: 'var(--bg-secondary)',
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
      fontFamily: 'var(--font-display)',
      textTransform: 'uppercase'
    },
    sectionDescription: {
      fontSize: '1.25rem',
      color: 'var(--text-secondary)',
      fontWeight: 400,
      textAlign: 'center',
      marginBottom: '4rem',
      maxWidth: '600px',
      margin: '0 auto 4rem auto'
    },
    plansGrid: {
      display: 'grid',
      gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))',
      gap: '2rem',
      maxWidth: '1400px',
      margin: '0 auto',
    },
    planCard: {
      background: 'var(--bg-card)',
      border: '1px solid rgba(248, 249, 250, 0.1)',
      borderRadius: '1rem',
      padding: '2.5rem',
      position: 'relative',
      display: 'flex',
      flexDirection: 'column',
      transition: 'all 0.3s',
    },
    planCardPopular: {
      background: 'linear-gradient(135deg, rgba(224, 26, 79, 0.05) 0%, rgba(255, 107, 53, 0.05) 100%)',
      border: '2px solid var(--accent-red)',
      boxShadow: '0 0 40px rgba(224, 26, 79, 0.2)',
    },
    popularBadge: {
      position: 'absolute',
      top: '-1rem',
      left: '50%',
      transform: 'translateX(-50%)',
      background: 'var(--gradient-primary)',
      color: 'white',
      padding: '0.5rem 1.5rem',
      borderRadius: '2rem',
      fontSize: '0.875rem',
      fontWeight: 700,
      boxShadow: '0 10px 20px var(--accent-glow)',
    },
    planName: {
      fontSize: '1.75rem',
      fontWeight: 700,
      marginBottom: '0.5rem',
      color: 'var(--text-primary)',
    },
    planImage: {
      width: '100%',
      height: '140px',
      objectFit: 'cover',
      borderRadius: '8px',
      marginBottom: '1.5rem',
      backgroundColor: 'rgba(0,0,0,0.2)'
    },
    priceContainer: {
      display: 'flex',
      alignItems: 'baseline',
      marginBottom: '2rem',
    },
    currency: {
      fontSize: '1.25rem',
      fontWeight: 600,
      color: 'var(--text-secondary)',
    },
    price: {
      fontSize: 'var(--title-h1)',
      fontWeight: 900,
      margin: '0 0.5rem',
      color: 'var(--text-primary)',
    },
    period: {
      fontSize: '1rem',
      color: 'var(--text-muted)',
    },
    featuresList: {
      flex: 1,
      marginBottom: '2rem',
    },
    featureItem: {
      display: 'flex',
      alignItems: 'flex-start',
      gap: '0.75rem',
      marginBottom: '1rem',
    },
    featureText: {
      fontSize: '0.95rem',
      lineHeight: '1.5',
      color: 'var(--text-secondary)',
    },
    featureTextBold: {
      fontWeight: 600,
      color: 'var(--text-primary)',
    },
    ctaButton: {
      width: '100%',
      padding: 'var(--btn-padding-lg)',
      borderRadius: '0.5rem',
      border: 'none',
      fontSize: '1rem',
      fontWeight: 700,
      cursor: 'pointer',
      transition: 'all 0.3s',
    },
    ctaPrimary: {
      background: 'var(--gradient-primary)',
      color: 'white',
      boxShadow: '0 10px 20px var(--accent-glow)',
    },
    ctaSecondary: {
      background: 'transparent',
      color: 'var(--text-primary)',
      border: '2px solid rgba(248, 249, 250, 0.2)',
    },
  };

  return (
    <section id="planos" style={styles.section}>
      <div style={styles.container}>
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
        >
          <h2 style={styles.sectionTitle}>
            Não quer ter trabalho? <br /> <span className="text-gradient">Nós construímos para você.</span>
          </h2>
          <p style={styles.sectionDescription}>
            Contrate nossa engenharia. Instalação, otimização e suporte 24/7. Escolha seu plano e deixe a infraestrutura com a Pixelcraft.
          </p>
        </motion.div>

        <div style={styles.plansGrid} className="mobile-swipe-carousel">
          {loading ? (
            <div style={{ display: 'flex', justifyContent: 'center', width: '100%', padding: '4rem', gridColumn: '1/-1' }}>
              <Loader2 className="animate-spin" size={32} color="var(--accent-blue)" />
            </div>
          ) : displayPlans.map((plan, index) => (
            <motion.div
              key={plan.name}
              initial={{ opacity: 0, y: 30 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: index * 0.1 }}
              whileHover={{
                y: -8,
                boxShadow: plan.popular
                  ? '0 30px 60px rgba(224, 26, 79, 0.4)'
                  : '0 20px 40px rgba(248, 249, 250, 0.1)',
              }}
              style={{
                ...styles.planCard,
                ...(plan.popular ? styles.planCardPopular : {}),
              }}
            >
              {plan.popular && (
                <div style={styles.popularBadge}>
                  ⚡ Mais Popular
                </div>
              )}

              <h3 style={{
                ...styles.planName,
                marginTop: plan.popular ? '1rem' : '0',
                color: plan.popular ? 'var(--accent-red)' : 'var(--text-primary)',
              }}>
                {plan.name}
              </h3>

              {plan.imageUrl && (
                <img
                  src={plan.imageUrl}
                  alt={plan.name}
                  style={styles.planImage}
                  onError={(e) => {
                    e.target.style.display = 'none'; // Hide if fails to load
                  }}
                />
              )}

              <div style={styles.priceContainer}>
                <span style={styles.currency}>R$</span>
                <span style={styles.price}>{plan.price}</span>
                <span style={styles.period}>/mês</span>
              </div>

              <div style={styles.featuresList}>
                {plan.features.map((feature, i) => (
                  <div key={i} style={styles.featureItem}>
                    <CheckCircle
                      size={20}
                      style={{
                        color: plan.popular ? 'var(--accent-red)' : 'var(--accent-orange)',
                        flexShrink: 0,
                        marginTop: '2px',
                      }}
                    />
                    <span
                      style={{
                        ...styles.featureText,
                        ...(feature.includes(':') ? styles.featureTextBold : {}),
                      }}
                    >
                      {feature}
                    </span>
                  </div>
                ))}
              </div>

              <motion.button
                whileHover={{ scale: 1.03, boxShadow: plan.popular ? '0 15px 35px var(--accent-glow)' : '0 10px 25px rgba(255,255,255,0.1)' }}
                whileTap={{ scale: 0.97 }}
                onClick={() => navigate('/loja?view=subscriptions')}
                style={{
                  ...styles.ctaButton,
                  ...(plan.popular ? styles.ctaPrimary : styles.ctaSecondary),
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: '0.5rem'
                }}
              >
                Contratar Agora
                <ArrowRight size={18} />
              </motion.button>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}

export default PricingSection;

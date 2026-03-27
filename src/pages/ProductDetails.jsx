import { motion, AnimatePresence } from 'framer-motion';
import LoadingSpinner from '../components/shared/LoadingSpinner';
import {
  ArrowLeft,
  ShoppingCart,
  Check,
  Star,
  Shield,
  Zap,
  Download,
  Crown,
  X,
  CreditCard,
  Tag,
  Users,
  Sparkles
} from 'lucide-react';
import { useState, useEffect } from 'react';
import { useNavigate, useLocation, useParams } from 'react-router-dom';
import ReactMarkdown from 'react-markdown';
import { discountsAPI, checkoutAPI, productsAPI, aiAPI } from '../services/api';
import { useAuth } from '../context/AuthContext';
import { useCart } from '../context/CartContext';
import DashboardLayout from '../components/DashboardLayout';

function ProductDetails() {
  const navigate = useNavigate();
  const location = useLocation();
  const { id } = useParams();
  useAuth(); // User handled by layout
  const { addToCart } = useCart();

  const [product, setProduct] = useState(location.state?.product || null);
  const [loadingProduct, setLoadingProduct] = useState(!location.state?.product);

  // AI Formatting State
  const [formattedDescription, setFormattedDescription] = useState(null);
  const [isFormatting, setIsFormatting] = useState(false);

  const [quantity] = useState(1);
  const [showCheckoutModal, setShowCheckoutModal] = useState(false);
  const [couponCode, setCouponCode] = useState('');
  const [referralCode, setReferralCode] = useState('');
  const [appliedCoupon, setAppliedCoupon] = useState(null);
  const [discount, setDiscount] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const formatPrice = (value) => {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL'
    }).format(value);
  };

  // Fetch product if not in state
  useEffect(() => {
    if (!product && id) {
      const fetchProduct = async () => {
        try {
          setLoadingProduct(true);
          const data = await productsAPI.getById(id);

          // Normalize data
          const normalizedProduct = {
            ...data,
            priceValue: data.price, // Ensure raw number is available
            price: formatPrice(data.price), // Formatted string
            category: data.category?.name || data.type || 'Produto', // Handle missing category name
            // Use image_url from backend, or fallback
            image: data.image_url,
          };

          setProduct(normalizedProduct);
        } catch (err) {
          console.error("Failed to fetch product:", err);
          setError("Produto não encontrado.");
        } finally {
          setLoadingProduct(false);
        }
      };

      fetchProduct();
    }
  }, [id, product]);

  // AI Formatting Effect
  useEffect(() => {
    if (!product?.description) return;

    const raw = product.description;

    // Check cache
    const cacheKey = `ai_desc_${product.id || 'temp'}`;
    const cached = sessionStorage.getItem(cacheKey);
    if (cached) {
      setFormattedDescription(cached);
      return;
    }

    // Heuristic: If description is long (> 50 char) and has few newlines (< 3), assume it's a blob and format it.
    // Or if the user explicitely requested this feature everywhere, we might want to be aggressive.
    // Given the screenshot, it's a wall of text.
    const isMessy = raw.length > 50 && (raw.split('\n').length < 3 || !raw.includes('•') && !raw.includes('- '));

    if (isMessy) {
      setIsFormatting(true);
      // Small delay to allow UI to render first
      const timer = setTimeout(async () => {
        try {
          console.log("Formatting description with AI...");
          const res = await aiAPI.formatText(raw);
          if (res && res.formatted_text) {
            setFormattedDescription(res.formatted_text);
            sessionStorage.setItem(cacheKey, res.formatted_text);
          } else {
            setFormattedDescription(raw);
          }
        } catch (err) {
          console.error("AI Formatting failed:", err);
          setFormattedDescription(raw);
        } finally {
          setIsFormatting(false);
        }
      }, 500);
      return () => clearTimeout(timer);
    } else {
      setFormattedDescription(raw);
    }
  }, [product]);


  const handlePurchase = () => {
    setShowCheckoutModal(true);
  };

  const applyCoupon = async () => {
    if (!couponCode.trim()) return;

    setLoading(true);
    setError(null);

    try {
      const response = await discountsAPI.validate(couponCode, product.priceValue, [{ product_id: product.id, quantity: 1 }]);

      if (response.is_valid) {
        setAppliedCoupon({
          code: couponCode,
          discount: response.discount_amount,
          type: 'fixed'
        });
        setDiscount(response.discount_amount);
      } else {
        setError(response.message || 'Cupom inválido');
        setAppliedCoupon(null);
        setDiscount(0);
      }
    } catch (error) {
      setError(error.message || 'Erro ao validar cupom. Tente novamente.');
      setAppliedCoupon(null);
      setDiscount(0);
    } finally {
      setLoading(false);
    }
  };

  const finalPrice = product ? Math.max(0, (product.priceValue || 0) - discount) : 0;

  const handleCheckout = async () => {
    setLoading(true);
    setError(null);

    try {
      addToCart(product, quantity);

      const checkoutData = {
        cart: [{ product_id: product.id, quantity: quantity }],
        use_balance: true
      };

      if (appliedCoupon) {
        checkoutData.coupon_code = appliedCoupon.code;
      }

      if (referralCode) {
        checkoutData.referral_code = referralCode;
      }

      const response = await checkoutAPI.process(checkoutData);

      if (response.success) {
        alert(`Compra realizada com sucesso!\n\nTotal: ${formatPrice(finalPrice)}\n${discount > 0 ? `Desconto aplicado: ${formatPrice(discount)}\n` : ''}${referralCode ? `Código de referência: ${referralCode}\n` : ''}\nPagamento ID: ${response.payment_id}`);
        setShowCheckoutModal(false);
        navigate('/dashboard');
      } else {
        setError(response.message || 'Erro ao processar compra');
      }
    } catch (error) {
      setError(error.message || 'Erro ao processar compra. Tente novamente.');
    } finally {
      setLoading(false);
    }
  };

  if (loadingProduct) {
    return (
      <DashboardLayout title="Carregando...">
        <LoadingSpinner message="Carregando detalhes do produto..." fullHeight={false} />
      </DashboardLayout>
    );
  }

  if (!product) {
    return (
      <DashboardLayout title="Detalhes do Produto">
        <div style={{ textAlign: 'center', padding: '4rem', color: '#F8F9FA' }}>
          <h2>Produto não encontrado</h2>
          <button onClick={() => navigate('/loja')} style={{
            marginTop: '2rem', padding: 'var(--btn-padding-lg)', background: 'var(--gradient-cta)',
            border: 'none', borderRadius: '12px', color: '#FFFFFF', fontSize: '14px', fontWeight: 600, cursor: 'pointer'
          }}>
            Voltar para a Loja
          </button>
        </div>
      </DashboardLayout>
    );
  }

  const isSubscription = product.category === 'Assinatura' || product.type === 'SUBSCRIPTION';

  const styles = {
    contentWrapper: {
      maxWidth: '1400px',
      margin: '0 auto',
      position: 'relative',
      zIndex: 1,
    },
    backButton: {
      width: '48px', height: '48px', borderRadius: '50%',
      background: 'rgba(21, 26, 38, 0.6)',
      backdropFilter: 'blur(10px)',
      border: '1px solid rgba(88, 58, 255, 0.35)',
      display: 'flex', alignItems: 'center', justifyContent: 'center',
      color: '#F8F9FA', cursor: 'pointer', transition: 'all 0.3s',
      boxShadow: '0 4px 20px rgba(0, 0, 0, 0.3)',
      marginRight: '1rem',
    },
    breadcrumb: {
      fontSize: '14px', color: '#B8BDC7',
      display: 'flex', alignItems: 'center', gap: '0.5rem'
    },
    mainGrid: {
      display: 'grid',
      gridTemplateColumns: 'minmax(400px, 1fr) 1fr',
      gap: '4rem',
      alignItems: 'start',
    },
    leftColumn: { position: 'sticky', top: '3rem' },
    imageContainer: {
      width: '100%',
      height: '500px',
      background: 'rgba(15, 18, 25, 0.6)',
      backdropFilter: 'blur(20px)',
      border: '1px solid rgba(255, 255, 255, 0.08)',
      borderRadius: '24px',
      display: 'flex', alignItems: 'center', justifyContent: 'center',
      position: 'relative', overflow: 'hidden',
      boxShadow: '0 20px 60px rgba(0, 0, 0, 0.4)',
    },
    productImage: {
      width: '100%',
      height: '100%',
      objectFit: 'cover',
    },
    placeholderIcon: { fontSize: '180px', opacity: 0.2 },
    badge: {
      position: 'absolute', top: '24px', right: '24px', padding: '10px 20px',
      background: 'linear-gradient(135deg, #FFD700 0%, #FF6B35 100%)', borderRadius: '24px',
      fontSize: '12px', fontWeight: 700, color: '#0A0E1A', textTransform: 'uppercase', letterSpacing: '1.5px',
      display: 'flex', alignItems: 'center', gap: '6px', boxShadow: '0 0 30px rgba(255, 215, 0, 0.5)',
      zIndex: 2,
    },
    popularBadge: { background: 'var(--gradient-cta)', boxShadow: '0 0 30px rgba(224, 26, 79, 0.5)' },
    rightColumn: {},
    categoryTag: {
      display: 'inline-block', padding: '6px 16px', background: 'transparent',
      border: '1px solid rgba(224, 26, 79, 0.3)', borderRadius: '8px',
      fontSize: '12px', fontWeight: 600, color: '#E01A4F', textTransform: 'uppercase', letterSpacing: '1.5px',
      marginBottom: '1rem',
    },
    productName: {
      fontSize: '3.5rem', fontWeight: 900,
      background: 'linear-gradient(135deg, #F8F9FA 0%, #E01A4F 50%, #FFD700 100%)',
      WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent',
      marginBottom: '1.5rem', letterSpacing: '-0.03em', lineHeight: '1.1',
    },
    description: {
      fontSize: '1.125rem', color: '#B8BDC7', lineHeight: '1.8', marginBottom: '2.5rem',
      // Make sure markdown styles are applied
      whiteSpace: 'pre-wrap'
    },
    aiBadge: {
      display: 'inline-flex', alignItems: 'center', gap: '0.5rem',
      background: 'rgba(26, 210, 255, 0.1)', border: '1px solid rgba(26, 210, 255, 0.3)',
      color: '#1AD2FF', fontSize: '12px', padding: '4px 10px', borderRadius: '8px', marginBottom: '1rem'
    },
    priceSection: {
      background: 'rgba(15, 18, 25, 0.6)', backdropFilter: 'blur(20px)',
      border: '1px solid rgba(224, 26, 79, 0.2)', borderRadius: '20px', padding: '2rem', marginBottom: '2.5rem',
      boxShadow: '0 0 40px rgba(224, 26, 79, 0.1)',
    },
    priceLabel: { fontSize: '14px', color: '#B8BDC7', marginBottom: '0.5rem', textTransform: 'uppercase', letterSpacing: '1.5px', fontWeight: 600 },
    price: {
      fontSize: 'var(--title-h1)', fontWeight: 900,
      background: 'linear-gradient(135deg, #E01A4F 0%, #FF6B35 50%, #FFD700 100%)',
      WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent',
    },
    recurrence: { fontSize: '14px', color: '#B8BDC7', marginTop: '0.5rem' },
    ctaButton: {
      width: '100%', padding: '1.25rem 2rem',
      background: 'var(--gradient-cta)',
      border: 'none', borderRadius: '16px', color: '#FFFFFF', fontSize: '16px', fontWeight: 700,
      textTransform: 'uppercase', letterSpacing: '1.5px', cursor: 'pointer',
      display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '12px',
      boxShadow: '0 8px 32px rgba(224, 26, 79, 0.4)', marginBottom: '2.5rem',
    },
    featuresSection: {
      background: 'rgba(15, 18, 25, 0.4)', backdropFilter: 'blur(20px)',
      border: '1px solid rgba(255, 255, 255, 0.06)', borderRadius: '20px', padding: '2rem', marginBottom: '2.5rem',
    },
    featuresTitle: {
      fontSize: 'var(--title-h4)', fontWeight: 700, color: '#F8F9FA', marginBottom: '1.5rem',
      display: 'flex', alignItems: 'center', gap: '0.75rem',
    },
    featuresList: { display: 'flex', flexDirection: 'column', gap: '1rem' },
    featureItem: { display: 'flex', alignItems: 'flex-start', gap: '1rem', fontSize: '15px', color: '#E8E9EB', lineHeight: '1.7' },
    checkIcon: {
      width: '24px', height: '24px', borderRadius: '50%',
      background: 'var(--gradient-cta)',
      display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0, marginTop: '2px',
      boxShadow: '0 0 20px rgba(224, 26, 79, 0.3)',
    },
    trustBadges: { display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '1rem' },
    trustBadge: {
      background: 'rgba(15, 18, 25, 0.4)', backdropFilter: 'blur(20px)',
      border: '1px solid rgba(255, 255, 255, 0.06)', borderRadius: '12px', padding: '1.25rem',
      display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '0.5rem', textAlign: 'center',
    },
    trustBadgeIcon: {
      width: '40px', height: '40px', borderRadius: '50%',
      background: 'linear-gradient(135deg, rgba(224, 26, 79, 0.2) 0%, rgba(255, 107, 53, 0.2) 100%)',
      display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#E01A4F',
    },
    trustBadgeLabel: { fontSize: '12px', color: '#B8BDC7', fontWeight: 600 },
    // Modal Styles
    modalOverlay: {
      position: 'fixed', top: 0, left: 0, right: 0, bottom: 0,
      background: 'rgba(10, 14, 26, 0.8)', backdropFilter: 'blur(20px)',
      display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 1000, padding: '2rem',
    },
    modalContent: {
      background: 'rgba(15, 18, 25, 0.95)', backdropFilter: 'blur(40px)',
      border: '1px solid rgba(224, 26, 79, 0.2)', borderRadius: '24px', padding: '3rem',
      maxWidth: '600px', width: '100%', maxHeight: '90vh', overflowY: 'auto', position: 'relative',
      boxShadow: '0 0 80px rgba(224, 26, 79, 0.3)',
    },
    modalHeader: { display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' },
    modalTitle: {
      fontSize: '1.75rem', fontWeight: 700,
      background: 'linear-gradient(135deg, #F8F9FA 0%, #E01A4F 100%)',
      WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent',
    },
    closeButton: {
      width: '40px', height: '40px', borderRadius: '50%', background: 'transparent',
      border: '1px solid rgba(255, 255, 255, 0.12)', color: '#B8BDC7',
      cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center',
    },
    productSummary: {
      background: 'rgba(224, 26, 79, 0.05)', border: '1px solid rgba(224, 26, 79, 0.2)',
      borderRadius: '16px', padding: '1.5rem', marginBottom: '2rem',
    },
    summaryTitle: { fontSize: '1.125rem', fontWeight: 600, color: '#F8F9FA', marginBottom: '0.5rem' },
    summaryCategory: { fontSize: '14px', color: '#E01A4F', marginBottom: '1rem' },
    formSection: { marginBottom: '2rem' },
    sectionTitle: { fontSize: '1rem', fontWeight: 600, color: '#F8F9FA', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.5rem' },
    inputGroup: { display: 'flex', gap: '0.75rem', marginBottom: '1rem' },
    input: {
      flex: 1, padding: '1rem 1.25rem', background: 'rgba(255, 255, 255, 0.02)',
      border: '1px solid rgba(255, 255, 255, 0.12)', borderRadius: '12px',
      color: '#F8F9FA', fontSize: '14px', outline: 'none',
    },
    applyButton: {
      padding: '1rem 1.5rem', background: 'transparent', border: '1px solid rgba(224, 26, 79, 0.5)',
      borderRadius: '12px', color: '#E01A4F', fontSize: '14px', fontWeight: 600, cursor: 'pointer',
    },
    discountApplied: {
      background: 'rgba(34, 197, 94, 0.1)', border: '1px solid rgba(34, 197, 94, 0.3)',
      borderRadius: '12px', padding: 'var(--btn-padding-md)', color: '#22C55E', fontSize: '14px', fontWeight: 600,
      display: 'flex', alignItems: 'center', gap: '0.5rem',
    },
    priceBreakdown: {
      background: 'rgba(15, 18, 25, 0.6)', border: '1px solid rgba(255, 255, 255, 0.08)',
      borderRadius: '16px', padding: '1.5rem', marginBottom: '2rem',
    },
    priceRow: { display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.75rem', fontSize: '14px' },
    priceLabelModal: { color: '#B8BDC7' },
    priceValue: { color: '#F8F9FA', fontWeight: 600 },
    discountValue: { color: '#22C55E', fontWeight: 600 },
    finalPriceRow: { borderTop: '1px solid rgba(255, 255, 255, 0.12)', paddingTop: '1rem', marginTop: '1rem', marginBottom: 0 },
    finalPriceStyle: {
      fontSize: '1.25rem', fontWeight: 700,
      background: 'var(--gradient-cta)',
      WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent',
    },
    checkoutButton: {
      width: '100%', padding: '1.25rem 2rem',
      background: 'var(--gradient-cta)',
      border: 'none', borderRadius: '16px', color: '#FFFFFF', fontSize: '16px', fontWeight: 700,
      textTransform: 'uppercase', letterSpacing: '1.5px', cursor: 'pointer',
      display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '12px',
      boxShadow: '0 8px 32px rgba(224, 26, 79, 0.4)',
    },
    errorText: { color: '#EF4444', fontSize: '14px', marginTop: '0.5rem', },
    loadingText: { color: '#B8BDC7', fontSize: '14px', marginTop: '0.5rem', },
  };

  const headerStart = product ? (
    <div style={{ display: 'flex', alignItems: 'center' }}>
      <motion.div
        style={styles.backButton}
        whileHover={{
          background: 'rgba(26, 210, 255, 0.2)',
          borderColor: '#1AD2FF',
          color: '#1AD2FF',
          boxShadow: '0 0 20px rgba(26, 210, 255, 0.4)',
        }}
        onClick={() => navigate('/loja')}
        title="Voltar para a Loja"
      >
        <ArrowLeft size={20} />
      </motion.div>
      <div style={styles.breadcrumb}>
        <span style={{ opacity: 0.6 }}>Loja</span>
        {' / '}
        <span style={{ opacity: 0.6 }}>{product.category || 'Detalhes'}</span>
        {' / '}
        <span style={{ color: '#E01A4F', fontWeight: 600 }}>{product.name}</span>
      </div>
    </div>
  ) : null;

  return (
    <DashboardLayout headerStart={headerStart} title="">
      <div style={styles.contentWrapper}>
        <div style={styles.mainGrid}>
          {/* LEFT COLUMN - IMAGE */}
          <div style={styles.leftColumn}>
            <motion.div
              style={styles.imageContainer}
              initial={{ opacity: 0, scale: 0.95 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ duration: 0.5 }}
            >
              {product.popular && (
                <div style={{ ...styles.badge, ...styles.popularBadge }}>
                  <Crown size={14} />
                  Mais Popular
                </div>
              )}
              {product.featured && !product.popular && (
                <div style={styles.badge}>
                  <Star size={14} />
                  Destaque
                </div>
              )}

              {/* IMAGE LOGIC */}
              {product.image_url || product.image ? (
                <img
                  src={product.image_url || product.image}
                  alt={product.name}
                  style={styles.productImage}
                  onError={(e) => {
                    e.target.style.display = 'none';
                  }}
                />
              ) : (
                <div style={styles.placeholderIcon}>
                  {product.category === 'Assinatura' || product.type === 'SUBSCRIPTION' ? '💎' :
                    product.category === 'Plugin' || product.type === 'PLUGIN' ? '⚙️' :
                      product.category === 'Mapa' || product.type === 'MAP' ? '🗺️' :
                        product.category === 'Mod' || product.type === 'MOD' ? '🔧' :
                          product.category === 'Texture Pack' || product.type === 'TEXTUREPACK' ? '🎨' :
                            '🖥️'}
                </div>
              )}
            </motion.div>
          </div>

          {/* RIGHT COLUMN - DETAILS */}
          <motion.div
            style={styles.rightColumn}
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.5, delay: 0.2 }}
          >
            <div style={styles.categoryTag}>{product.category || product.type}</div>
            <h1 style={styles.productName}>{product.name}</h1>

            {/* Description Area */}
            {isFormatting ? (
              <div style={{ ...styles.description, opacity: 0.7 }}>
                <Sparkles size={16} className="inline-block mr-2 animate-pulse text-[#1AD2FF]" />
                Otimizando descrição com IA...
              </div>
            ) : (
              <div className="prose prose-invert max-w-none mb-10 text-[#B8BDC7]">
                {formattedDescription && formattedDescription !== product.description && (
                  <div style={styles.aiBadge}>
                    <Sparkles size={12} />
                    Formatado por IA
                  </div>
                )}
                <ReactMarkdown
                  components={{
                    // Override default element styles for better look
                    p: ({ node, ...props }) => <p style={{ marginBottom: '1rem', lineHeight: '1.8' }} {...props} />,
                    ul: ({ node, ...props }) => <ul style={{ listStyleType: 'disc', paddingLeft: '1.5rem', marginBottom: '1rem' }} {...props} />,
                    li: ({ node, ...props }) => <li style={{ marginBottom: '0.5rem' }} {...props} />,
                    strong: ({ node, ...props }) => <strong style={{ color: '#E01A4F', fontWeight: 600 }} {...props} />,
                  }}
                >
                  {formattedDescription || product.description || "Sem descrição disponível."}
                </ReactMarkdown>
              </div>
            )}

            {/* PRICE SECTION */}
            <div style={styles.priceSection}>
              <div style={styles.priceLabel}>
                {isSubscription ? 'Investimento Mensal' : 'Preço'}
              </div>
              <div style={styles.price}>{product.price}</div>
              {isSubscription && (
                <div style={styles.recurrence}>Cobrado mensalmente</div>
              )}
            </div>

            {/* CTA BUTTON */}
            <motion.button
              style={styles.ctaButton}
              whileHover={{
                scale: 1.02,
                boxShadow: '0 12px 48px rgba(224, 26, 79, 0.6)',
              }}
              whileTap={{ scale: 0.98 }}
              onClick={handlePurchase}
            >
              {isSubscription ? (
                <>
                  <Crown size={20} />
                  Contratar Plano
                </>
              ) : (
                <>
                  <ShoppingCart size={20} />
                  Adicionar ao Carrinho
                </>
              )}
            </motion.button>

            {/* FEATURES */}
            {product.features && product.features.length > 0 && (
              <div style={styles.featuresSection}>
                <div style={styles.featuresTitle}>
                  <Zap size={24} color="#E01A4F" />
                  O que está incluído
                </div>
                <div style={styles.featuresList}>
                  {product.features.map((feature, index) => (
                    <motion.div
                      key={index}
                      style={styles.featureItem}
                      initial={{ opacity: 0, x: -20 }}
                      animate={{ opacity: 1, x: 0 }}
                      transition={{ delay: 0.4 + (index * 0.1) }}
                    >
                      <div style={styles.checkIcon}>
                        <Check size={14} color="#FFFFFF" />
                      </div>
                      <span>{feature}</span>
                    </motion.div>
                  ))}
                </div>
              </div>
            )}

            {/* TRUST BADGES */}
            <div style={styles.trustBadges}>
              <motion.div
                style={styles.trustBadge}
                whileHover={{ scale: 1.05, borderColor: 'rgba(224, 26, 79, 0.3)' }}
              >
                <div style={styles.trustBadgeIcon}>
                  <Shield size={20} />
                </div>
                <div style={styles.trustBadgeLabel}>Compra Segura</div>
              </motion.div>

              <motion.div
                style={styles.trustBadge}
                whileHover={{ scale: 1.05, borderColor: 'rgba(224, 26, 79, 0.3)' }}
              >
                <div style={styles.trustBadgeIcon}>
                  <Zap size={20} />
                </div>
                <div style={styles.trustBadgeLabel}>Entrega Imediata</div>
              </motion.div>

              <motion.div
                style={styles.trustBadge}
                whileHover={{ scale: 1.05, borderColor: 'rgba(224, 26, 79, 0.3)' }}
              >
                <div style={styles.trustBadgeIcon}>
                  <Download size={20} />
                </div>
                <div style={styles.trustBadgeLabel}>Atualizações Grátis</div>
              </motion.div>
            </div>
          </motion.div>
        </div>
      </div>

      {/* CHECKOUT MODAL */}
      <AnimatePresence>
        {showCheckoutModal && (
          <motion.div
            style={styles.modalOverlay}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={() => setShowCheckoutModal(false)}
          >
            <motion.div
              style={styles.modalContent}
              initial={{ scale: 0.9, opacity: 0, y: 20 }}
              animate={{ scale: 1, opacity: 1, y: 0 }}
              exit={{ scale: 0.9, opacity: 0, y: 20 }}
              onClick={(e) => e.stopPropagation()}
            >
              <div style={styles.modalHeader}>
                <h2 style={styles.modalTitle}>Checkout</h2>
                <button
                  style={styles.closeButton}
                  onClick={() => setShowCheckoutModal(false)}
                >
                  <X size={20} />
                </button>
              </div>

              <div style={styles.productSummary}>
                <div style={styles.summaryTitle}>{product.name}</div>
                <div style={styles.summaryCategory}>{product.category || product.type}</div>
                <div style={styles.price}>{product.price}</div>
              </div>

              {/* CUPOM */}
              <div style={styles.formSection}>
                <div style={styles.sectionTitle}>
                  <Tag size={16} color="#E01A4F" />
                  Cupom de Desconto
                </div>
                <div style={styles.inputGroup}>
                  <input
                    type="text"
                    placeholder="Digite seu cupom"
                    style={styles.input}
                    value={couponCode}
                    onChange={(e) => setCouponCode(e.target.value.toUpperCase())}
                    disabled={!!appliedCoupon || loading}
                  />
                  <button
                    style={styles.applyButton}
                    onClick={applyCoupon}
                    disabled={!!appliedCoupon || loading}
                  >
                    {loading ? '...' : 'Aplicar'}
                  </button>
                </div>
                {appliedCoupon && (
                  <div style={styles.discountApplied}>
                    <Check size={16} />
                    Cupom <strong>{appliedCoupon.code}</strong> aplicado com sucesso!
                  </div>
                )}
                {error && <div style={styles.errorText}>{error}</div>}
              </div>

              {/* INDICAÇÃO */}
              <div style={styles.formSection}>
                <div style={styles.sectionTitle}>
                  <Users size={16} color="#E01A4F" />
                  Código de Indicação (Opcional)
                </div>
                <div style={styles.inputGroup}>
                  <input
                    type="text"
                    placeholder="Código de quem te indicou"
                    style={styles.input}
                    value={referralCode}
                    onChange={(e) => setReferralCode(e.target.value)}
                  />
                </div>
              </div>

              {/* RESUMO DE PREÇO */}
              <div style={styles.priceBreakdown}>
                <div style={styles.priceRow}>
                  <span style={styles.priceLabelModal}>Subtotal</span>
                  <span style={styles.priceValue}>{product.price}</span>
                </div>
                {discount > 0 && (
                  <div style={styles.priceRow}>
                    <span style={styles.priceLabelModal}>Desconto</span>
                    <span style={styles.discountValue}>- {formatPrice(discount)}</span>
                  </div>
                )}
                <div style={{ ...styles.priceRow, ...styles.finalPriceRow }}>
                  <span style={styles.priceLabelModal}>Total a Pagar</span>
                  <span style={styles.finalPriceStyle}>{formatPrice(finalPrice)}</span>
                </div>
              </div>

              {/* BUY BUTTON */}
              <button
                style={styles.checkoutButton}
                onClick={handleCheckout}
                disabled={loading}
              >
                {loading ? 'Processando...' : (
                  <>
                    <CreditCard size={20} />
                    Confirmar Compra
                  </>
                )}
              </button>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>
    </DashboardLayout>
  );
}

export default ProductDetails;
import { motion } from 'framer-motion';
import {
  ArrowLeft,
  CreditCard,
  Gift,
  Users,
  Tag,
  Check,
  X,
  Loader2,
  Trash2
} from 'lucide-react';
import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useCart } from '../context/CartContext';
import { useAuth } from '../context/AuthContext';
import { discountsAPI, checkoutAPI } from '../services/api';
import DashboardLayout from '../components/DashboardLayout';

function Checkout() {
  const navigate = useNavigate();
  const { user } = useAuth(); // Changed to not force redirect if useAuth allows it, but assuming it protects by default, let's just grab the user. Wait, if useAuth has a redirect inside, we need to handle it. Assuming it returns { user }.
  const { cart, cartTotal, clearCart, removeFromCart } = useCart();
  const [couponCode, setCouponCode] = useState('');
  const [referralCode, setReferralCode] = useState('');
  const [appliedCoupon, setAppliedCoupon] = useState(null);
  const [discount, setDiscount] = useState(0);
  const [loading, setLoading] = useState(false);
  const [validating, setValidating] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(false);

  // Redirect to shop if cart is empty
  useEffect(() => {
    if (cart.length === 0 && !success) {
      navigate('/loja');
    }
  }, [cart, navigate, success]);

  const finalPrice = Math.max(0, cartTotal - discount);

  const formatPrice = (value) => {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL'
    }).format(value);
  };

  const applyCoupon = async () => {
    if (!couponCode.trim()) return;

    setValidating(true);
    setError(null);

    try {
      const cartItems = cart.map(item => ({
        product_id: item.id,
        quantity: item.quantity
      }));
      const response = await discountsAPI.validate(couponCode, cartTotal, cartItems);

      if (response.is_valid) {
        setAppliedCoupon({
          code: couponCode,
          discount: response.discount_amount,
          type: 'fixed'
        });
        setDiscount(response.discount_amount);
        setError(null);
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
      setValidating(false);
    }
  };

  const handleCheckout = async () => {
    setLoading(true);
    setError(null);

    try {
      const checkoutData = {
        cart: cart.map(item => ({
          product_id: item.id,
          quantity: item.quantity
        })),
        use_balance: true,
        ...(referralCode.trim() && { referral_code: referralCode.trim() })
      };

      if (appliedCoupon) {
        checkoutData.coupon_code = appliedCoupon.code;
      }

      const response = await checkoutAPI.process(checkoutData);

      if (response.success) {
        setSuccess(true);
        clearCart();
      } else {
        setError(response.message || 'Erro ao processar compra. Tente novamente.');
      }
    } catch (error) {
      if (error.response?.status === 401) {
        // If unauthorized from API, redirect to login with returnUrl
        navigate('/login', { state: { returnUrl: '/checkout' } });
      } else {
        setError(error.message || 'Erro inesperado. Tente novamente.');
      }
    } finally {
      setLoading(false);
    }
  };

  // If user is not logged in, we intercept the UI to show a login prompt 
  // without losing the cart context (as it's stored in local storage)
  if (!user) {
    return (
      <DashboardLayout hideSidebar={true}>
        <div style={styles.checkoutContainer}>
          <div style={styles.checkoutWrapper}>
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              style={{
                background: 'var(--bg-card)',
                padding: '3rem',
                borderRadius: '1rem',
                textAlign: 'center',
                border: '1px solid var(--border-card)',
                boxShadow: 'var(--shadow-card)'
              }}
            >
              <h2 style={{ fontSize: 'var(--title-h2)', marginBottom: '1rem', fontFamily: 'var(--font-display)', color: 'var(--text-primary)' }}>FAÇA LOGIN PARA CONTINUAR</h2>
              <p style={{ color: 'var(--text-secondary)', marginBottom: '2rem' }}>
                Você possui <strong>{cart.length} itens</strong> no carrinho aguardando finalização.
              </p>
              <div style={{ display: 'flex', gap: '1rem', justifyContent: 'center' }}>
                <button
                  onClick={() => navigate('/login', { state: { returnUrl: '/checkout' } })}
                  style={{ ...styles.actionButton, background: 'var(--gradient-primary)', border: 'none' }}
                >
                  Fazer Login
                </button>
                <button
                  onClick={() => navigate('/cadastrar', { state: { returnUrl: '/checkout' } })}
                  style={{ ...styles.actionButton, background: 'transparent', border: '1px solid rgba(255,255,255,0.2)' }}
                >
                  Criar Conta
                </button>
              </div>
            </motion.div>
          </div>
        </div>
      </DashboardLayout>
    );
  }

  const handleRemoveItem = (id, name) => {
    if (window.confirm(`Tem certeza que deseja remover "${name}" do carrinho?`)) {
      removeFromCart(id);
    }
  };

  const styles = {
    backButton: {
      width: '48px', height: '48px', borderRadius: '50%',
      background: 'var(--bg-card)',
      backdropFilter: 'blur(10px)',
      border: '1px solid var(--border-input)',
      display: 'flex', alignItems: 'center', justifyContent: 'center',
      color: 'var(--text-primary)', cursor: 'pointer', transition: 'all var(--transition-normal)',
      boxShadow: 'var(--shadow-card)',
      marginRight: '1rem',
    },
    contentWrapper: {
      maxWidth: '1200px',
      margin: '0 auto',
      position: 'relative',
      zIndex: 1,
    },
    mainGrid: {
      display: 'grid',
      gridTemplateColumns: 'minmax(0, 1fr) 1fr', // Ensure flex items don't overflow
      gap: '3rem',
    },
    leftColumn: {},
    rightColumn: {
      background: 'rgba(15, 18, 25, 0.6)',
      backdropFilter: 'blur(20px)',
      border: '1px solid rgba(224, 26, 79, 0.2)',
      borderRadius: '20px',
      padding: '2rem',
      boxShadow: '0 0 40px rgba(224, 26, 79, 0.1)',
      height: 'fit-content'
    },
    sectionTitle: {
      fontSize: 'var(--title-h4)',
      fontWeight: 700,
      color: '#F8F9FA',
      marginBottom: '1.5rem',

      display: 'flex',
      alignItems: 'center',
      gap: '0.75rem',
    },
    cartItem: {
      display: 'flex',
      justifyContent: 'space-between',
      alignItems: 'center',
      padding: '1rem 0',
      borderBottom: '1px solid rgba(255, 255, 255, 0.08)',
    },
    itemInfo: {
      display: 'flex',
      alignItems: 'center',
      gap: '1rem',
    },
    itemImage: {
      width: '80px',
      height: '80px',
      background: 'rgba(224, 26, 79, 0.1)',
      borderRadius: '12px',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      fontSize: '24px',
      overflow: 'hidden',
      flexShrink: 0,
    },
    itemDetails: {
      display: 'flex',
      flexDirection: 'column',
      gap: '0.25rem',
    },
    itemName: {
      fontSize: '1.1rem',
      fontWeight: 600,
      color: '#F8F9FA',

    },
    itemCategory: {
      fontSize: '0.875rem',
      color: '#E01A4F',

    },
    itemPrice: {
      fontSize: '1.1rem',
      fontWeight: 600,
      color: '#F8F9FA',

    },
    quantity: {
      fontSize: '0.875rem',
      color: '#B8BDC7',

    },
    formSection: {
      marginBottom: '2rem',
    },
    inputGroup: {
      display: 'flex',
      gap: '0.75rem',
      marginBottom: '1rem',
    },
    input: {
      flex: 1,
      padding: '1rem 1.25rem',
      background: 'rgba(255, 255, 255, 0.02)',
      border: '1px solid rgba(255, 255, 255, 0.12)',
      borderRadius: '12px',
      color: '#F8F9FA',
      fontSize: '14px',

      outline: 'none',
      transition: 'all 0.3s',
    },
    applyButton: {
      padding: '1rem 1.5rem',
      background: 'transparent',
      border: '1px solid rgba(224, 26, 79, 0.5)',
      borderRadius: '12px',
      color: '#E01A4F',
      fontSize: '14px',
      fontWeight: 600,
      cursor: 'pointer',
      transition: 'all 0.3s',

      disabled: {
        opacity: 0.5,
        cursor: 'not-allowed',
      },
    },
    discountApplied: {
      background: 'rgba(34, 197, 94, 0.1)',
      border: '1px solid rgba(34, 197, 94, 0.3)',
      borderRadius: '12px',
      padding: 'var(--btn-padding-md)',
      color: '#22C55E',
      fontSize: '14px',
      fontWeight: 600,
      display: 'flex',
      alignItems: 'center',
      gap: '0.5rem',

      marginBottom: '1rem',
    },
    priceBreakdown: {
      background: 'rgba(15, 18, 25, 0.4)',
      border: '1px solid rgba(255, 255, 255, 0.08)',
      borderRadius: '16px',
      padding: '1.5rem',
      marginBottom: '2rem',
    },
    priceRow: {
      display: 'flex',
      justifyContent: 'space-between',
      alignItems: 'center',
      marginBottom: '0.75rem',
      fontSize: '14px',

    },
    priceLabel: {
      color: '#B8BDC7',
    },
    priceValue: {
      color: '#F8F9FA',
      fontWeight: 600,
    },
    discountValue: {
      color: '#22C55E',
      fontWeight: 600,
    },
    finalPriceRow: {
      borderTop: '1px solid rgba(255, 255, 255, 0.12)',
      paddingTop: '1rem',
      marginTop: '1rem',
      marginBottom: 0,
    },
    finalPrice: {
      fontSize: '1.25rem',
      fontWeight: 700,
      background: 'var(--gradient-cta)',
      WebkitBackgroundClip: 'text',
      WebkitTextFillColor: 'transparent',
    },
    checkoutButton: {
      width: '100%',
      padding: '1.25rem 2rem',
      background: 'var(--gradient-cta)',
      border: 'none',
      borderRadius: '16px',
      color: '#FFFFFF',
      fontSize: '16px',
      fontWeight: 700,
      textTransform: 'uppercase',
      letterSpacing: '1.5px',
      cursor: 'pointer',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      gap: '12px',

      boxShadow: '0 8px 32px rgba(224, 26, 79, 0.4)',
      disabled: {
        opacity: 0.5,
        cursor: 'not-allowed',
      },
    },
    errorText: {
      color: '#EF4444',
      fontSize: '14px',
      marginTop: '0.5rem',

    },
    successMessage: {
      background: 'rgba(34, 197, 94, 0.1)',
      border: '1px solid rgba(34, 197, 94, 0.3)',
      borderRadius: '12px',
      padding: '2rem',
      textAlign: 'center',
      color: '#22C55E',
      fontSize: '16px',
      fontWeight: 600,

      marginBottom: '2rem',
    },
    successIcon: {
      width: '60px',
      height: '60px',
      borderRadius: '50%',
      background: 'rgba(34, 197, 94, 0.2)',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      margin: '0 auto 1rem',
    },
  };

  const headerStart = (
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
  );

  if (success) {
    return (
      <DashboardLayout title="Compra Concluída" headerStart={headerStart}>
        <div style={styles.contentWrapper}>
          <div style={styles.successMessage}>
            <div style={styles.successIcon}>
              <Check size={32} color="#22C55E" />
            </div>
            <p>Sua compra foi realizada com sucesso!</p>
            <p style={{ fontSize: '14px', fontWeight: 400, marginTop: '0.5rem' }}>
              Você será redirecionado para o dashboard em breve...
            </p>
          </div>
        </div>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout title="Finalizar Compra" headerStart={headerStart}>
      <div style={styles.contentWrapper}>
        <div style={styles.mainGrid} className="checkout-main-grid">
          {/* LEFT COLUMN - CART ITEMS */}
          <div style={styles.leftColumn}>
            <h2 style={styles.sectionTitle}>Itens no Carrinho</h2>

            {cart.map((item, index) => (
              <motion.div
                key={item.id}
                style={styles.cartItem}
                initial={{ opacity: 0, x: -20 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: index * 0.1 }}
              >
                <div style={styles.itemInfo}>
                  <div style={styles.itemImage}>
                    {item.image_url ? (
                      <>
                        <img
                          src={item.image_url}
                          alt={item.name}
                          style={{ width: '100%', height: '100%', objectFit: 'cover' }}
                          onError={(e) => {
                            e.target.style.display = 'none';
                            e.target.nextSibling && (e.target.nextSibling.style.display = 'flex');
                          }}
                        />
                        <span style={{ display: 'none', width: '100%', height: '100%', alignItems: 'center', justifyContent: 'center', fontSize: 'var(--title-h4)' }}>🖼️</span>
                      </>
                    ) : (
                      <>
                        {(!item.category || item.category === 'Assinatura') && '💎'}
                        {item.category === 'Plugin' && '⚙️'}
                        {item.category === 'Mapa' && '🗺️'}
                        {item.category === 'Mod' && '🔧'}
                        {item.category === 'Texture Pack' && '🎨'}
                        {item.category === 'Servidor Pronto' && '🖥️'}
                      </>
                    )}
                  </div>
                  <div style={styles.itemDetails}>
                    <div style={styles.itemName}>{item.name}</div>
                    <div style={styles.itemCategory}>{item.category || item.type}</div>
                    <div style={styles.quantity}>Quantidade: {item.quantity}</div>
                  </div>
                </div>

                <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: '0.5rem' }}>
                  <div style={styles.itemPrice}>
                    {formatPrice(item.price * item.quantity)}
                  </div>
                  <button
                    onClick={() => handleRemoveItem(item.id, item.name)}
                    style={{
                      background: 'transparent',
                      border: 'none',
                      color: '#EF4444',
                      cursor: 'pointer',
                      display: 'flex',
                      alignItems: 'center',
                      gap: '4px',
                      fontSize: '12px',
                      fontWeight: 600,
                      padding: '4px 8px',
                      borderRadius: '6px',
                    }}
                    onMouseEnter={(e) => e.target.style.background = 'rgba(239, 68, 68, 0.1)'}
                    onMouseLeave={(e) => e.target.style.background = 'transparent'}
                    title="Remover item"
                  >
                    <Trash2 size={14} />
                    Remover
                  </button>
                </div>
              </motion.div>
            ))}

            {cart.length === 0 && (
              <div style={{ textAlign: 'center', padding: '2rem', color: '#B8BDC7' }}>
                Seu carrinho está vazio.
              </div>
            )}
          </div>

          {/* RIGHT COLUMN - CHECKOUT FORM */}
          <div style={styles.rightColumn}>
            <h2 style={styles.sectionTitle}>Resumo do Pedido</h2>

            {/* CUPOM DE DESCONTO */}
            <div style={styles.formSection}>
              <h3 style={{ ...styles.sectionTitle, fontSize: '1.125rem', marginBottom: '1rem' }}>
                <Gift size={18} color="#E01A4F" />
                Cupom de Desconto
              </h3>
              {!appliedCoupon ? (
                <div style={styles.inputGroup}>
                  <input
                    type="text"
                    placeholder="Digite seu cupom (ex: PIXELCRAFT10)"
                    value={couponCode}
                    onChange={(e) => setCouponCode(e.target.value)}
                    style={styles.input}
                    disabled={validating || loading}
                  />
                  <motion.button
                    style={{
                      ...styles.applyButton,
                      ...(validating ? styles.applyButton.disabled : {})
                    }}
                    whileHover={!validating && !loading ? {
                      background: 'rgba(224, 26, 79, 0.1)',
                      borderColor: '#E01A4F'
                    } : {}}
                    onClick={applyCoupon}
                    disabled={!couponCode.trim() || validating || loading}
                  >
                    {validating ? 'Aplicando...' : 'Aplicar'}
                  </motion.button>
                </div>
              ) : (
                <div style={styles.discountApplied}>
                  <Check size={16} />
                  Cupom "{appliedCoupon.code}" aplicado!
                  Desconto de {formatPrice(appliedCoupon.discount)}
                  <motion.button
                    style={{
                      background: 'none',
                      border: 'none',
                      color: '#22C55E',
                      cursor: 'pointer',
                      marginLeft: 'auto'
                    }}
                    whileHover={{ scale: 1.1 }}
                    onClick={() => {
                      setAppliedCoupon(null);
                      setDiscount(0);
                      setCouponCode('');
                    }}
                  >
                    <X size={16} />
                  </motion.button>
                </div>
              )}

              {error && !appliedCoupon && (
                <div style={styles.errorText}>{error}</div>
              )}
            </div>

            {/* CÓDIGO DE REFERÊNCIA */}
            <div style={styles.formSection}>
              <h3 style={{ ...styles.sectionTitle, fontSize: '1.125rem', marginBottom: '1rem' }}>
                <Users size={18} color="#E01A4F" />
                Código de Referência
              </h3>
              <input
                type="text"
                placeholder="Código do amigo que te indicou (opcional)"
                value={referralCode}
                onChange={(e) => setReferralCode(e.target.value)}
                style={styles.input}
                disabled={loading}
              />
              <p style={{
                fontSize: '12px',
                color: '#B8BDC7',
                marginTop: '0.5rem'
              }}>
                💡 Seu amigo ganhará créditos quando você usar o código dele
              </p>
            </div>

            {/* RESUMO DE PREÇOS */}
            <div style={styles.priceBreakdown}>
              <h3 style={{ ...styles.sectionTitle, fontSize: '1.125rem', marginBottom: '1.5rem' }}>
                <Tag size={18} color="#E01A4F" />
                Resumo do Pedido
              </h3>

              <div style={styles.priceRow}>
                <span style={styles.priceLabel}>Subtotal:</span>
                <span style={styles.priceValue}>{formatPrice(cartTotal)}</span>
              </div>

              {discount > 0 && (
                <div style={styles.priceRow}>
                  <span style={styles.priceLabel}>Desconto:</span>
                  <span style={styles.discountValue}>-{formatPrice(discount)}</span>
                </div>
              )}

              <div style={{ ...styles.priceRow, ...styles.finalPriceRow }}>
                <span style={{ fontSize: '1.125rem', fontWeight: 600, color: '#F8F9FA' }}>
                  Total:
                </span>
                <span style={styles.finalPrice}>{formatPrice(finalPrice)}</span>
              </div>
            </div>

            {/* BOTÃO DE CHECKOUT */}
            <motion.button
              style={{
                ...styles.checkoutButton,
                ...(loading ? styles.checkoutButton.disabled : {})
              }}
              whileHover={!loading ? {
                scale: 1.02,
                boxShadow: '0 12px 48px rgba(224, 26, 79, 0.6)',
              } : {}}
              whileTap={!loading ? { scale: 0.98 } : {}}
              onClick={handleCheckout}
              disabled={loading}
            >
              {loading ? (
                <>
                  <Loader2 size={20} style={{ animation: 'spin 1s linear infinite' }} />
                  Processando...
                </>
              ) : (
                <>
                  <CreditCard size={20} />
                  Finalizar Compra
                </>
              )}
            </motion.button>

            {error && (
              <div style={{ ...styles.errorText, marginTop: '1rem', textAlign: 'center' }}>
                {error}
              </div>
            )}
          </div>
        </div>
      </div>
    </DashboardLayout>
  );
}

export default Checkout;
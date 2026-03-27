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
  Trash2,
  Wallet,
  QrCode
} from 'lucide-react';
import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useCart } from '../context/CartContext';
import { useAuth } from '../context/AuthContext';
import { discountsAPI, checkoutAPI } from '../services/api';
import DashboardLayout from '../components/DashboardLayout';

function Checkout() {
  const navigate = useNavigate();
  const { user } = useAuth();
  const { cart, cartTotal, clearCart, removeFromCart } = useCart();
  const [couponCode, setCouponCode] = useState('');
  const [referralCode, setReferralCode] = useState('');
  const [appliedCoupon, setAppliedCoupon] = useState(null);
  const [discount, setDiscount] = useState(0);
  const [loading, setLoading] = useState(false);
  const [validating, setValidating] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(false);
  const [paymentMethod, setPaymentMethod] = useState('balance'); // 'balance' or 'direct'

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
        use_balance: paymentMethod === 'balance',
        ...(referralCode.trim() && { referral_code: referralCode.trim() })
      };

      if (appliedCoupon) {
        checkoutData.coupon_code = appliedCoupon.code;
      }

      const response = await checkoutAPI.process(checkoutData);

      if (response.success) {
        if (response.payment_gateway_url) {
          // Redirect to Mercado Pago
          window.location.href = response.payment_gateway_url;
          return;
        }
        setSuccess(true);
        clearCart();
      } else {
        setError(response.message || 'Erro ao processar compra. Tente novamente.');
      }
    } catch (error) {
      if (error.response?.status === 401) {
        navigate('/login', { state: { returnUrl: '/checkout' } });
      } else {
        setError(error.response?.data?.error || error.message || 'Erro inesperado. Tente novamente.');
      }
    } finally {
      setLoading(false);
    }
  };

  if (!user) {
    return (
      <DashboardLayout hideSidebar={true}>
        <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '60vh' }}>
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            style={{
              background: 'var(--bg-card)',
              padding: '3rem',
              borderRadius: '1rem',
              textAlign: 'center',
              border: '1px solid var(--border-card)',
              boxShadow: 'var(--shadow-card)',
              maxWidth: '500px'
            }}
          >
            <h2 style={{ fontSize: 'var(--title-h2)', marginBottom: '1rem', color: 'var(--text-primary)' }}>FAÇA LOGIN PARA CONTINUAR</h2>
            <p style={{ color: 'var(--text-secondary)', marginBottom: '2rem' }}>
              Você possui <strong>{cart.length} itens</strong> no carrinho aguardando finalização.
            </p>
            <div style={{ display: 'flex', gap: '1rem', justifyContent: 'center' }}>
              <button
                onClick={() => navigate('/login', { state: { returnUrl: '/checkout' } })}
                style={{ padding: '12px 24px', borderRadius: '8px', background: 'var(--gradient-primary)', border: 'none', color: 'white', fontWeight: 700, cursor: 'pointer' }}
              >
                Fazer Login
              </button>
              <button
                onClick={() => navigate('/cadastrar', { state: { returnUrl: '/checkout' } })}
                style={{ padding: '12px 24px', borderRadius: '8px', background: 'transparent', border: '1px solid rgba(255,255,255,0.2)', color: 'white', fontWeight: 700, cursor: 'pointer' }}
              >
                Criar Conta
              </button>
            </div>
          </motion.div>
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
      color: 'var(--text-primary)', cursor: 'pointer', transition: 'all 0.3s ease',
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
      gridTemplateColumns: '1fr 400px',
      gap: '3rem',
    },
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
    methodCard: {
      flex: 1,
      padding: '1.5rem',
      borderRadius: '16px',
      border: '1px solid rgba(255, 255, 255, 0.1)',
      background: 'rgba(255, 255, 255, 0.03)',
      cursor: 'pointer',
      transition: 'all 0.3s ease',
      display: 'flex',
      flexDirection: 'column',
      gap: '0.5rem',
      position: 'relative',
      overflow: 'hidden'
    },
    methodCardActive: {
      borderColor: 'var(--accent-purple)',
      background: 'rgba(88, 58, 255, 0.1)',
      boxShadow: '0 0 20px rgba(88, 58, 255, 0.1)'
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
    }
  };

  const headerStart = (
    <motion.div
      style={styles.backButton}
      whileHover={{ scale: 1.1, backgroundColor: 'rgba(26, 210, 255, 0.2)' }}
      onClick={() => navigate('/loja')}
    >
      <ArrowLeft size={20} />
    </motion.div>
  );

  if (success) {
    return (
      <DashboardLayout title="Compra Concluída" headerStart={headerStart}>
        <div style={{ textAlign: 'center', padding: '4rem 0' }}>
          <Check size={64} color="#22C55E" style={{ marginBottom: '1.5rem' }} />
          <h2 style={{ color: 'white', marginBottom: '1rem' }}>Sua compra foi realizada com sucesso!</h2>
          <button 
            onClick={() => navigate('/dashboard')}
            style={{ padding: '12px 32px', borderRadius: '8px', background: 'var(--gradient-primary)', border: 'none', color: 'white', fontWeight: 700, cursor: 'pointer' }}
          >
            Ir para o Dashboard
          </button>
        </div>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout title="Finalizar Compra" headerStart={headerStart}>
      <div style={styles.contentWrapper}>
        <div style={styles.mainGrid} className="checkout-mobile-grid">
          <div style={{ display: 'flex', flexDirection: 'column', gap: '3rem' }}>
            {/* ITENS NO CARRINHO */}
            <div>
              <h2 style={styles.sectionTitle}>Itens no Carrinho</h2>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                {cart.map((item) => (
                  <div key={item.id} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '1rem', background: 'rgba(255,255,255,0.03)', borderRadius: '12px', border: '1px solid rgba(255,255,255,0.05)' }}>
                    <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
                      <div style={{ width: '60px', height: '60px', background: 'var(--gradient-primary)', borderRadius: '8px', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '24px' }}>
                        {item.category === 'Plan' ? '💎' : '📦'}
                      </div>
                      <div>
                        <div style={{ color: 'white', fontWeight: 600 }}>{item.name}</div>
                        <div style={{ color: '#888', fontSize: '0.85rem' }}>{item.category} x {item.quantity}</div>
                      </div>
                    </div>
                    <div style={{ color: 'white', fontWeight: 700 }}>{formatPrice(item.price * item.quantity)}</div>
                  </div>
                ))}
              </div>
            </div>

            {/* MÉTODO DE PAGAMENTO */}
            <div>
              <h2 style={styles.sectionTitle}>Método de Pagamento</h2>
              <div style={{ display: 'flex', gap: '1rem' }} className="payment-methods-flex">
                <motion.div 
                  style={{ ...styles.methodCard, ...(paymentMethod === 'balance' ? styles.methodCardActive : {}) }}
                  onClick={() => setPaymentMethod('balance')}
                  whileHover={{ scale: 1.02 }}
                >
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <Wallet size={24} color={paymentMethod === 'balance' ? 'var(--accent-purple)' : '#888'} />
                    {paymentMethod === 'balance' && <Check size={20} color="var(--accent-purple)" />}
                  </div>
                  <div style={{ color: 'white', fontWeight: 700, marginTop: '0.5rem' }}>Saldo da Carteira</div>
                  <div style={{ color: '#888', fontSize: '0.85rem' }}>Seu saldo atual: {formatPrice(user.balance || 0)}</div>
                </motion.div>

                <motion.div 
                  style={{ ...styles.methodCard, ...(paymentMethod === 'direct' ? styles.methodCardActive : {}) }}
                  onClick={() => setPaymentMethod('direct')}
                  whileHover={{ scale: 1.02 }}
                >
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <QrCode size={24} color={paymentMethod === 'direct' ? 'var(--accent-purple)' : '#888'} />
                    {paymentMethod === 'direct' && <Check size={20} color="var(--accent-purple)" />}
                  </div>
                  <div style={{ color: 'white', fontWeight: 700, marginTop: '0.5rem' }}>PIX / Mercado Pago</div>
                  <div style={{ color: '#888', fontSize: '0.85rem' }}>Pague agora sem precisar depositar</div>
                </motion.div>
              </div>
            </div>
          </div>

          {/* RESUMO E BOTÃO */}
          <div style={styles.rightColumn}>
            <h2 style={styles.sectionTitle}>Resumo do Pedido</h2>
            
            {/* CUPOM */}
            <div style={{ marginBottom: '2rem' }}>
              {!appliedCoupon ? (
                <div style={{ display: 'flex', gap: '0.5rem' }}>
                  <input
                    type="text"
                    placeholder="Cupom"
                    value={couponCode}
                    onChange={(e) => setCouponCode(e.target.value.toUpperCase())}
                    style={{ flex: 1, padding: '12px', background: 'rgba(255,255,255,0.05)', border: '1px solid rgba(255,255,255,0.1)', borderRadius: '8px', color: 'white' }}
                  />
                  <button 
                    onClick={applyCoupon}
                    disabled={validating}
                    style={{ padding: '0 16px', background: 'rgba(224, 26, 79, 0.2)', border: '1px solid var(--accent-pink)', borderRadius: '8px', color: 'var(--accent-pink)', fontWeight: 600, cursor: 'pointer' }}
                  >
                    {validating ? '...' : 'Ok'}
                  </button>
                </div>
              ) : (
                <div style={{ padding: '12px', background: 'rgba(34, 197, 94, 0.1)', border: '1px solid #22c55e', borderRadius: '8px', color: '#22c55e', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <span>Cupom aplicado!</span>
                  <X size={16} style={{ cursor: 'pointer' }} onClick={() => { setAppliedCoupon(null); setDiscount(0); }} />
                </div>
              )}
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem', marginBottom: '2rem' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', color: '#888' }}>
                <span>Subtotal:</span>
                <span>{formatPrice(cartTotal)}</span>
              </div>
              {discount > 0 && (
                <div style={{ display: 'flex', justifyContent: 'space-between', color: '#22c55e' }}>
                  <span>Desconto:</span>
                  <span>-{formatPrice(discount)}</span>
                </div>
              )}
              <div style={{ display: 'flex', justifyContent: 'space-between', color: 'white', fontWeight: 700, fontSize: '1.25rem', paddingTop: '1rem', borderTop: '1px solid rgba(255,255,255,0.1)' }}>
                <span>Total:</span>
                <span style={{ color: 'var(--accent-pink)' }}>{formatPrice(finalPrice)}</span>
              </div>
            </div>

            <motion.button
              style={styles.checkoutButton}
              whileHover={{ scale: 1.02 }}
              whileTap={{ scale: 0.98 }}
              onClick={handleCheckout}
              disabled={loading}
            >
              {loading ? <Loader2 size={24} className="animate-spin" /> : <><CreditCard size={20} /> FINALIZAR COMPRA</>}
            </motion.button>

            {error && (
              <div style={{ marginTop: '1rem', color: '#ef4444', textAlign: 'center', fontSize: '0.9rem', background: 'rgba(239, 68, 68, 0.1)', padding: '10px', borderRadius: '8px' }}>
                {error}
              </div>
            )}
          </div>
        </div>
      </div>
      
      <style>{`
        @keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
        .animate-spin { animation: spin 1s linear infinite; }
        @media (max-width: 900px) {
          .checkout-mobile-grid { grid-template-columns: 1fr !important; }
          .payment-methods-flex { flex-direction: column; }
        }
      `}</style>
    </DashboardLayout>
  );
}

export default Checkout;
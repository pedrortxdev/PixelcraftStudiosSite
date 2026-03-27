// src/components/shop/FloatingCart.jsx
import { motion, AnimatePresence } from 'framer-motion';
import { ShoppingCart, X, Trash2, Plus, Minus } from 'lucide-react';
import { useCart } from '../../context/CartContext';
import { useNavigate } from 'react-router-dom';

function FloatingCart() {
  const {
    cart,
    cartCount,
    cartTotal,
    removeFromCart,
    updateQuantity,
    isCartOpen,
    setIsCartOpen
  } = useCart();

  const navigate = useNavigate();

  const formatPrice = (price) => {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(price);
  };

  const handleCheckout = () => {
    setIsCartOpen(false);
    navigate('/checkout');
  };

  const styles = {
    floatingButton: {
      position: 'fixed',
      bottom: '2rem',
      right: '2rem',
      padding: 'var(--btn-padding-md)',
      background: 'var(--gradient-primary)',
      borderRadius: '50%',
      border: 'none',
      cursor: 'pointer',
      boxShadow: '0 10px 40px rgba(88, 58, 255, 0.4)',
      zIndex: 1000,
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      width: '60px',
      height: '60px',
    },
    badge: {
      position: 'absolute',
      top: '-6px',
      right: '-6px',
      background: '#1AD2FF',
      color: '#0A0E1A',
      borderRadius: '50%',
      width: '22px',
      height: '22px',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      fontSize: '11px',
      fontWeight: 700,
    },
    cartPanel: {
      position: 'fixed',
      top: window.innerWidth <= 768 ? 'auto' : 0,
      bottom: 0,
      right: window.innerWidth <= 768 ? 0 : 0,
      left: window.innerWidth <= 768 ? 0 : 'auto',
      width: window.innerWidth <= 768 ? '100vw' : 'min(380px, 100vw)',
      height: window.innerWidth <= 768 ? '85vh' : '100vh',
      background: 'rgba(10, 14, 26, 0.98)',
      backdropFilter: 'blur(20px)',
      borderLeft: window.innerWidth <= 768 ? 'none' : '1px solid rgba(88, 58, 255, 0.3)',
      borderTop: window.innerWidth <= 768 ? '1px solid rgba(88, 58, 255, 0.3)' : 'none',
      borderTopLeftRadius: window.innerWidth <= 768 ? '24px' : '0',
      borderTopRightRadius: window.innerWidth <= 768 ? '24px' : '0',
      boxShadow: window.innerWidth <= 768 ? '0 -10px 50px rgba(0, 0, 0, 0.6)' : '-10px 0 50px rgba(0, 0, 0, 0.4)',
      zIndex: 2000,
      display: 'flex',
      flexDirection: 'column',
    },
    dragIndicator: {
      width: '40px',
      height: '4px',
      background: 'rgba(255, 255, 255, 0.2)',
      borderRadius: '2px',
      margin: '12px auto 0 auto',
      display: window.innerWidth <= 768 ? 'block' : 'none',
    },
    header: {
      padding: '2rem',
      borderBottom: '1px solid rgba(88, 58, 255, 0.2)',
      display: 'flex',
      justifyContent: 'space-between',
      alignItems: 'center',
    },
    title: {
      fontSize: 'var(--title-h4)',
      fontWeight: 700,
      color: '#F8F9FA',
    },
    closeButton: {
      background: 'transparent',
      border: 'none',
      color: '#B8BDC7',
      cursor: 'pointer',
      padding: '0.5rem',
    },
    cartItems: {
      flex: 1,
      overflowY: 'auto',
      padding: '1.5rem',
    },
    cartItem: {
      display: 'flex',
      gap: '1rem',
      padding: 'var(--btn-padding-md)',
      background: 'rgba(21, 26, 38, 0.6)',
      borderRadius: '12px',
      marginBottom: '1rem',
      border: '1px solid rgba(88, 58, 255, 0.15)',
    },
    itemImage: {
      width: '60px',
      height: '60px',
      borderRadius: '8px',
      overflow: 'hidden',
      border: '1px solid rgba(255,255,255,0.1)',
    },
    itemDetails: {
      flex: 1,
    },
    itemName: {
      fontSize: '14px',
      fontWeight: 600,
      color: '#F8F9FA',
      marginBottom: '0.25rem',
    },
    itemPrice: {
      fontSize: '14px',
      fontWeight: 700,
      background: 'var(--gradient-primary)',
      WebkitBackgroundClip: 'text',
      WebkitTextFillColor: 'transparent',
    },
    quantityControls: {
      display: 'flex',
      alignItems: 'center',
      gap: '0.5rem',
      marginTop: '0.5rem',
    },
    quantityButton: {
      width: '24px',
      height: '24px',
      borderRadius: '6px',
      border: '1px solid rgba(255, 255, 255, 0.15)',
      background: 'rgba(255,255,255,0.05)',
      color: '#F8F9FA',
      cursor: 'pointer',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
    },
    quantity: {
      fontSize: '14px',
      fontWeight: 600,
      color: '#F8F9FA',
      minWidth: '20px',
      textAlign: 'center',
    },
    removeButton: {
      background: 'transparent',
      border: 'none',
      color: '#EF4444',
      cursor: 'pointer',
      padding: '0.5rem',
    },
    emptyCart: {
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      height: '100%',
      gap: '1rem',
      color: '#B8BDC7',
    },
    footer: {
      padding: '2rem',
      borderTop: '1px solid rgba(88, 58, 255, 0.2)',
    },
    total: {
      display: 'flex',
      justifyContent: 'space-between',
      alignItems: 'center',
      marginBottom: '1.5rem',
    },
    totalLabel: {
      fontSize: '1.125rem',
      fontWeight: 600,
      color: '#B8BDC7',
    },
    totalValue: {
      fontSize: 'var(--title-h4)',
      fontWeight: 900,
      background: 'linear-gradient(135deg, #F8F9FA 0%, #583AFF 50%, #1AD2FF 100%)',
      WebkitBackgroundClip: 'text',
      WebkitTextFillColor: 'transparent',
    },
    checkoutButton: {
      width: '100%',
      padding: 'var(--btn-padding-md)',
      background: 'var(--gradient-primary)',
      border: 'none',
      borderRadius: '0.75rem',
      color: 'white',
      fontSize: '1rem',
      fontWeight: 700,
      cursor: 'pointer',
    },
  };

  return (
    <>
      {/* Floating Cart Button */}
      <motion.button
        style={styles.floatingButton}
        whileHover={{ scale: 1.1, boxShadow: '0 15px 50px rgba(88, 58, 255, 0.6)' }}
        whileTap={{ scale: 0.95 }}
        onClick={() => setIsCartOpen(true)}
      >
        <ShoppingCart size={24} color="white" />
        {cartCount > 0 && (
          <motion.div
            style={styles.badge}
            initial={{ scale: 0 }}
            animate={{ scale: 1 }}
            transition={{ type: 'spring', stiffness: 500, damping: 15 }}
          >
            {cartCount > 9 ? '9+' : cartCount}
          </motion.div>
        )}
      </motion.button>

      {/* Cart Panel */}
      <AnimatePresence>
        {isCartOpen && (
          <>
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              onClick={() => setIsCartOpen(false)}
              style={{
                position: 'fixed',
                top: 0,
                left: 0,
                right: 0,
                bottom: 0,
                background: 'rgba(0, 0, 0, 0.7)',
                zIndex: 1999,
              }}
            />

            <motion.div
              style={styles.cartPanel}
              initial={window.innerWidth <= 768 ? { y: '100%' } : { x: '100%' }}
              animate={window.innerWidth <= 768 ? { y: 0 } : { x: 0 }}
              exit={window.innerWidth <= 768 ? { y: '100%' } : { x: '100%' }}
              transition={{ type: 'spring', damping: 25, stiffness: 200 }}
              drag={window.innerWidth <= 768 ? "y" : false}
              dragConstraints={{ top: 0, bottom: 0 }}
              dragElastic={0.2}
              onDragEnd={(e, { offset, velocity }) => {
                const swipe = offset.y;
                if (swipe > 100 || velocity.y > 20) {
                  setIsCartOpen(false);
                }
              }}
            >
              <div style={styles.dragIndicator} />
              <div style={styles.header}>
                <h2 style={styles.title}>
                  Carrinho ({cartCount} {cartCount === 1 ? 'item' : 'itens'})
                </h2>
                <button style={styles.closeButton} onClick={() => setIsCartOpen(false)}>
                  <X size={24} />
                </button>
              </div>

              <div style={styles.cartItems}>
                {cart.length === 0 ? (
                  <div style={styles.emptyCart}>
                    <ShoppingCart size={64} opacity={0.3} />
                    <p>Seu carrinho está vazio</p>
                  </div>
                ) : (
                  cart.map((item) => (
                    <motion.div
                      key={item.id}
                      style={styles.cartItem}
                      initial={{ opacity: 0, x: 20 }}
                      animate={{ opacity: 1, x: 0 }}
                      exit={{ opacity: 0, x: -20 }}
                    >
                      <div style={styles.itemImage}>
                        {item.image_url ? (
                          <img
                            src={item.image_url}
                            alt={item.name}
                            style={{ width: '100%', height: '100%', objectFit: 'cover', borderRadius: '8px' }}
                          />
                        ) : (
                          <div style={{ width: '100%', height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#583AFF', fontSize: '24px' }}>
                            📦
                          </div>
                        )}
                      </div>
                      <div style={styles.itemDetails}>
                        <div style={styles.itemName}>{item.name}</div>
                        <div style={styles.itemPrice}>{formatPrice(item.price)}</div>
                        <div style={styles.quantityControls}>
                          <button
                            style={styles.quantityButton}
                            onClick={() => updateQuantity(item.id, item.quantity - 1)}
                          >
                            <Minus size={12} />
                          </button>
                          <span style={styles.quantity}>{item.quantity}</span>
                          <button
                            style={styles.quantityButton}
                            onClick={() => updateQuantity(item.id, item.quantity + 1)}
                          >
                            <Plus size={12} />
                          </button>
                        </div>
                      </div>
                      <button
                        style={styles.removeButton}
                        onClick={() => removeFromCart(item.id)}
                      >
                        <Trash2 size={18} />
                      </button>
                    </motion.div>
                  ))
                )}
              </div>

              {cart.length > 0 && (
                <div style={styles.footer}>
                  <div style={styles.total}>
                    <span style={styles.totalLabel}>Total:</span>
                    <span style={styles.totalValue}>{formatPrice(cartTotal)}</span>
                  </div>
                  <motion.button
                    style={styles.checkoutButton}
                    whileHover={{ boxShadow: '0 10px 40px rgba(88, 58, 255, 0.5)' }}
                    whileTap={{ scale: 0.98 }}
                    onClick={handleCheckout}
                  >
                    Finalizar Compra
                  </motion.button>
                </div>
              )}
            </motion.div>
          </>
        )}
      </AnimatePresence>
    </>
  );
}

export default FloatingCart;
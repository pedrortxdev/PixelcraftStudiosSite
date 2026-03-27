// src/components/shop/ProductCard.jsx
import { motion } from 'framer-motion';
import { ShoppingCart, Sparkles, Crown, Star } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { useCart } from '../../context/CartContext';

function ProductCard({ product, index }) {
  const navigate = useNavigate();
  const { addToCart, isInCart } = useCart();
  const inCart = isInCart(product.id);

  const handleViewDetails = () => {
    navigate(`/loja/produto/${product.id}`, { state: { product } });
  };

  const handleAddToCart = (_e) => {
    _e.stopPropagation();
    addToCart(product);
  };

  // Mapeamento visual
  const typeConfig = {
    PLUGIN: { label: 'Plugin', emoji: '⚙️', color: '#583AFF' },
    MOD: { label: 'Mod', emoji: '🔧', color: '#1AD2FF' },
    MAP: { label: 'Mapa', emoji: '🗺️', color: '#80FFEA' },
    TEXTUREPACK: { label: 'Textura', emoji: '🎨', color: '#FF6BFF' },
    SERVER_TEMPLATE: { label: 'Servidor', emoji: '🖥️', color: '#583AFF' },
  };

  const config = typeConfig[product.type] || { label: 'Item', emoji: '📦', color: '#B8BDC7' };
  const formattedPrice = new Intl.NumberFormat('pt-BR', {
    style: 'currency',
    currency: 'BRL',
  }).format(product.price);

  // Animação de entrada suave
  const cardVariants = {
    hidden: { opacity: 0, y: 20, scale: 0.98 },
    visible: {
      opacity: 1,
      y: 0,
      scale: 1,
      transition: {
        duration: 0.5,
        ease: [0.34, 1.56, 0.64, 1],
        delay: index * 0.07,
      },
    },
  };

  return (
    <motion.div
      variants={cardVariants}
      initial="hidden"
      whileInView="visible"
      viewport={{ once: true, margin: "-50px" }}
      style={{
        background: 'rgba(15, 18, 25, 0.4)',
        backdropFilter: 'blur(16px)',
        WebkitBackdropFilter: 'blur(16px)',
        border: '1px solid rgba(255, 255, 255, 0.06)',
        borderRadius: '20px',
        overflow: 'hidden',
        cursor: 'pointer',
        position: 'relative',
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        transition: 'all 0.4s cubic-bezier(0.25, 0.8, 0.25, 1)',
        boxShadow: '0 10px 30px rgba(0, 0, 0, 0.25)',
      }}
      whileHover={{
        y: -10,
        border: '1px solid rgba(88, 58, 255, 0.35)',
        boxShadow: '0 20px 50px rgba(88, 58, 255, 0.2), 0 0 0 1px rgba(26, 210, 255, 0.2)',
        transition: { duration: 0.35 },
      }}
      onClick={handleViewDetails}
    >
      {/* Efeito de brilho interno no hover */}
      <div
        style={{
          position: 'absolute',
          top: 0,
          left: 0,
          right: 0,
          height: '4px',
          background: 'linear-gradient(90deg, transparent, #583AFF, #1AD2FF, transparent)',
          opacity: 0,
          transition: 'opacity 0.4s ease',
          pointerEvents: 'none',
        }}
        className="card-glow"
      />

      {/* Imagem ou placeholder */}
      <div
        style={{
          height: '190px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          position: 'relative',
          overflow: 'hidden',
          background: 'linear-gradient(135deg, rgba(88, 58, 255, 0.04) 0%, rgba(26, 210, 255, 0.04) 100%)',
        }}
      >
        {/* Overlay sutil + Vignette forte para padronização premium */}
        <div
          style={{
            position: 'absolute',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            background: 'radial-gradient(circle at center, rgba(0,0,0,0) 20%, rgba(0,0,0,0.6) 100%), linear-gradient(to bottom, rgba(0,0,0,0.2) 0%, rgba(0,0,0,0) 50%, rgba(0,0,0,0.8) 100%)',
            pointerEvents: 'none',
            zIndex: 2,
          }}
        />

        {/* Badges */}
        {product.is_exclusive && (
          <motion.div
            initial={{ opacity: 0, scale: 0.8 }}
            animate={{ opacity: 1, scale: 1 }}
            style={{
              position: 'absolute',
              top: '12px',
              right: '12px',
              background: 'var(--gradient-primary)',
              color: 'white',
              padding: '4px 12px',
              borderRadius: '30px',
              fontSize: '11px',
              fontWeight: 800,
              display: 'flex',
              alignItems: 'center',
              gap: '4px',
              zIndex: 3,
              boxShadow: '0 4px 12px rgba(88, 58, 255, 0.4)',
              backdropFilter: 'blur(8px)',
            }}
          >
            <Crown size={12} style={{ strokeWidth: 2.5 }} />
            EXCLUSIVO
          </motion.div>
        )}

        {product.stock_quantity !== null && product.stock_quantity < 10 && (
          <motion.div
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: 0.1 }}
            style={{
              position: 'absolute',
              top: product.is_exclusive ? '48px' : '12px',
              right: '12px',
              background: 'linear-gradient(135deg, #1AD2FF, #80FFEA)',
              color: '#0A0E1A',
              padding: '4px 12px',
              borderRadius: '30px',
              fontSize: '11px',
              fontWeight: 800,
              display: 'flex',
              alignItems: 'center',
              gap: '4px',
              zIndex: 3,
              boxShadow: '0 4px 12px rgba(26, 210, 255, 0.4)',
              backdropFilter: 'blur(8px)',
            }}
          >
            <Sparkles size={12} style={{ strokeWidth: 2.5 }} />
            {product.stock_quantity} rest.
          </motion.div>
        )}

        {/* Imagem ou ícone */}
        {product.image_url ? (
          <img
            src={product.image_url}
            alt={product.name}
            style={{
              width: '100%',
              height: '100%',
              objectFit: 'cover',
              transition: 'transform 0.5s ease',
            }}
            className="product-image"
          />
        ) : (
          <div
            style={{
              fontSize: '60px',
              opacity: 0.6,
              color: config.color,
              filter: 'drop-shadow(0 2px 8px rgba(0,0,0,0.2))',
            }}
          >
            {config.emoji}
          </div>
        )}
      </div>

      {/* Corpo do card */}
      <div style={{ padding: '20px', flex: 1, display: 'flex', flexDirection: 'column' }}>
        {/* Categoria com destaque */}
        <div
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: '6px',
            marginBottom: '10px',
            padding: '4px 12px',
            background: 'rgba(88, 58, 255, 0.1)',
            border: '1px solid rgba(88, 58, 255, 0.2)',
            borderRadius: '12px',
            width: 'fit-content',
          }}
        >
          <div
            style={{
              width: '8px',
              height: '8px',
              borderRadius: '50%',
              background: config.color,
            }}
          />
          <span
            style={{
              fontSize: '12px',
              fontWeight: 700,
              color: config.color,
              textTransform: 'uppercase',
              letterSpacing: '0.5px',
            }}
          >
            {config.label}
          </span>
        </div>

        {/* Nome do produto */}
        <h3
          style={{
            fontSize: '18px',
            fontWeight: 800,
            color: '#F8F9FA',
            marginBottom: '12px',
            lineHeight: 1.3,
            letterSpacing: '-0.02em',
          }}
        >
          {product.name}
        </h3>

        {/* Descrição curta */}
        <p
          style={{
            fontSize: '13px',
            color: '#A0A7B8',
            lineHeight: 1.5,
            marginBottom: '16px',
            flex: 1,
            overflow: 'hidden',
            display: '-webkit-box',
            WebkitLineClamp: 2,
            WebkitBoxOrient: 'vertical',
          }}
        >
          {product.short_description ||
            product.description?.substring(0, 100) + (product.description?.length > 100 ? '...' : '') ||
            'Produto premium da Pixelcraft'}
        </p>

        {/* Rodapé: preço + ações */}
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            paddingTop: '14px',
            borderTop: '1px solid rgba(255, 255, 255, 0.06)',
          }}
        >
          <div
            style={{
              fontSize: '20px',
              fontWeight: 900,
              background: 'linear-gradient(135deg, #F8F9FA 0%, #583AFF 50%, #1AD2FF 100%)',
              WebkitBackgroundClip: 'text',
              WebkitTextFillColor: 'transparent',
              letterSpacing: '-0.02em',
            }}
          >
            {formattedPrice}
          </div>

          <div style={{ display: 'flex', gap: '10px' }}>
            {!inCart ? (
              <motion.button
                style={{
                  background: 'var(--gradient-primary)',
                  color: 'white',
                  border: 'none',
                  borderRadius: '12px',
                  padding: '8px 18px',
                  fontSize: '13px',
                  fontWeight: 700,
                  display: 'flex',
                  alignItems: 'center',
                  gap: '6px',
                  cursor: 'pointer',
                  boxShadow: '0 4px 16px rgba(88, 58, 255, 0.3)',
                }}
                whileHover={{ scale: 1.05, boxShadow: '0 6px 20px rgba(88, 58, 255, 0.5)' }}
                whileTap={{ scale: 0.95 }}
                onClick={handleAddToCart}
              >
                <ShoppingCart size={14} />
                Adicionar
              </motion.button>
            ) : (
              <motion.div
                style={{
                  background: 'rgba(88, 58, 255, 0.15)',
                  color: '#583AFF',
                  border: '1px solid rgba(88, 58, 255, 0.3)',
                  borderRadius: '12px',
                  padding: '8px 18px',
                  fontSize: '13px',
                  fontWeight: 700,
                  display: 'flex',
                  alignItems: 'center',
                  gap: '6px',
                }}
                whileHover={{ background: 'rgba(88, 58, 255, 0.25)' }}
              >
                <Star size={14} fill="#583AFF" style={{ strokeWidth: 0 }} />
                No carrinho
              </motion.div>
            )}
          </div>
        </div>
      </div>
    </motion.div>
  );
}

export default ProductCard;
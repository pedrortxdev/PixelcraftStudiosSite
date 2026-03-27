// src/context/CartContext.js
import { createContext, useContext, useState, useEffect, useMemo } from 'react';

const CartContext = createContext();

/**
 * Garante que o produto tenha os campos necessários para o carrinho
 */
const sanitizeProductForCart = (product) => {
  if (!product || !product.id) {
    throw new Error('Produto inválido: precisa ter "id"');
  }

  return {
    id: product.id,
    name: product.name || 'Produto sem nome',
    price: typeof product.price === 'number' ? product.price : parseFloat(product.price) || 0,
    image_url: product.image_url || null,
    type: product.type || 'UNKNOWN',
    // Não armazena objetos grandes (ex: descrição longa, metadados)
  };
};

export function CartProvider({ children }) {
  const [cart, setCart] = useState([]);
  const [isCartOpen, setIsCartOpen] = useState(false);

  // Carrega do localStorage ao iniciar
  useEffect(() => {
    const saved = localStorage.getItem('pixelcraft_cart');
    if (saved) {
      try {
        const parsed = JSON.parse(saved);
        if (Array.isArray(parsed)) {
          setCart(
            parsed.map((item) => ({
              ...item,
              price: typeof item.price === 'number' ? item.price : parseFloat(item.price) || 0,
            }))
          );
        }
      } catch (err) {
        console.error('Erro ao carregar carrinho:', err);
        localStorage.removeItem('pixelcraft_cart');
      }
    }
  }, []);

  // Salva no localStorage sempre que muda com debounce
  useEffect(() => {
    const timeoutId = setTimeout(() => {
      localStorage.setItem('pixelcraft_cart', JSON.stringify(cart));
    }, 500);

    return () => clearTimeout(timeoutId);
  }, [cart]);

  const addToCart = (product, quantity = 1) => {
    if (quantity <= 0) return;
    const cleanProduct = sanitizeProductForCart(product);

    setCart((prev) => {
      const existing = prev.find((item) => item.id === cleanProduct.id);
      if (existing) {
        return prev.map((item) =>
          item.id === cleanProduct.id
            ? { ...item, quantity: item.quantity + quantity }
            : item
        );
      } else {
        return [...prev, { ...cleanProduct, quantity }];
      }
    });

    setIsCartOpen(true);
    setTimeout(() => setIsCartOpen(false), 2500);
  };

  const removeFromCart = (productId) => {
    setCart((prev) => prev.filter((item) => item.id !== productId));
  };

  const updateQuantity = (productId, quantity) => {
    if (quantity <= 0) {
      removeFromCart(productId);
      return;
    }
    setCart((prev) =>
      prev.map((item) =>
        item.id === productId ? { ...item, quantity: Math.max(1, quantity) } : item
      )
    );
  };

  const clearCart = () => setCart([]);

  const cartCount = useMemo(() => cart.reduce((sum, item) => sum + item.quantity, 0), [cart]);

  const cartTotal = useMemo(() => cart.reduce((sum, item) => sum + item.price * item.quantity, 0), [cart]);

  const isInCart = (productId) => cart.some((item) => item.id === productId);

  const getProductQuantity = (productId) => {
    const item = cart.find((i) => i.id === productId);
    return item ? item.quantity : 0;
  };

  return (
    <CartContext.Provider
      value={{
        cart,
        addToCart,
        removeFromCart,
        updateQuantity,
        clearCart,
        cartCount,
        cartTotal,
        isInCart,
        getProductQuantity,
        isCartOpen,
        setIsCartOpen,
      }}
    >
      {children}
    </CartContext.Provider>
  );
}

export const useCart = () => {
  const context = useContext(CartContext);
  if (!context) {
    throw new Error('useCart deve ser usado dentro de um CartProvider');
  }
  return context;
};
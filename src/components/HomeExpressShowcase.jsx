import React, { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import { ShoppingCart, Zap } from 'lucide-react';
import { productsAPI } from '../services/api';
import { useCart } from '../context/CartContext';
import { useNavigate } from 'react-router-dom';

function HomeExpressShowcase() {
    const [products, setProducts] = useState([]);
    const [loading, setLoading] = useState(true);
    const { addToCart, setIsCartOpen } = useCart();
    const navigate = useNavigate();

    useEffect(() => {
        const fetchBestSellers = async () => {
            try {
                const data = await productsAPI.getAll({ limit: 4 }); // Get 4 quick items
                setProducts(data.data || []);
            } catch (error) {
                console.error("Failed to load express products:", error);
            } finally {
                setLoading(false);
            }
        };
        fetchBestSellers();
    }, []);

    const formatPrice = (price) => {
        return new Intl.NumberFormat('pt-BR', {
            style: 'currency',
            currency: 'BRL',
        }).format(price);
    };

    const handleQuickAdd = (e, product) => {
        e.stopPropagation(); // Prevent card click
        addToCart(product);
        setIsCartOpen(true);
    };

    if (loading || products.length === 0) return null;

    return (
        <section style={{ padding: '4rem 0 6rem 0', background: 'var(--bg-primary)' }}>
            <div style={{ maxWidth: '1400px', margin: '0 auto', padding: '0 2rem' }}>

                <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', marginBottom: '2rem' }}>
                    <Zap size={28} color="var(--accent-red)" />
                    <h2 style={{
                        fontFamily: 'var(--font-display)',
                        fontSize: '3rem',
                        margin: 0,
                        lineHeight: 1,
                        color: 'var(--text-primary)'
                    }}>
                        MAIS VENDIDOS
                    </h2>
                </div>

                <div style={{
                    display: 'grid',
                    gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))',
                    gap: '1.5rem'
                }}>
                    {products.slice(0, 4).map(product => (
                        <motion.div
                            key={product.id}
                            onClick={() => navigate(`/shop`)} // Redirect to shop to see the rest
                            whileHover={{ y: -5 }}
                            style={{
                                background: 'rgba(255, 255, 255, 0.02)',
                                border: '1px solid rgba(255, 255, 255, 0.05)',
                                position: 'relative',
                                overflow: 'hidden',
                                cursor: 'pointer',
                                display: 'flex',
                                flexDirection: 'column'
                            }}
                        >
                            {/* Image Container with Vignette */}
                            <div style={{ position: 'relative', height: '160px', overflow: 'hidden' }}>
                                {product.image_url ? (
                                    <img src={product.image_url} alt={product.name} style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
                                ) : (
                                    <div style={{ width: '100%', height: '100%', background: 'rgba(0,0,0,0.5)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                                        📦
                                    </div>
                                )}
                                {/* Vignette Overlay */}
                                <div style={{
                                    position: 'absolute', top: 0, left: 0, right: 0, bottom: 0,
                                    background: 'radial-gradient(circle at center, transparent 30%, rgba(10, 14, 26, 0.9) 100%)',
                                    pointerEvents: 'none'
                                }} />
                            </div>

                            {/* Content Box */}
                            <div style={{ padding: '1.5rem', display: 'flex', flexDirection: 'column', flex: 1, gap: '1rem' }}>
                                <div>
                                    <h4 style={{
                                        fontFamily: 'var(--font-display)',
                                        fontSize: '1.8rem',
                                        margin: '0 0 0.25rem 0',
                                        lineHeight: 1.1,
                                        textTransform: 'uppercase'
                                    }}>
                                        {product.name}
                                    </h4>
                                    <span style={{ fontFamily: 'var(--font-mono)', fontSize: '1.2rem', color: 'var(--accent-gold)' }}>
                                        {formatPrice(product.price)}
                                    </span>
                                </div>

                                <button
                                    onClick={(e) => handleQuickAdd(e, product)}
                                    style={{
                                        marginTop: 'auto',
                                        width: '100%',
                                        padding: '0.75rem',
                                        background: 'var(--text-primary)',
                                        color: 'var(--bg-primary)',
                                        fontFamily: 'var(--font-mono)',
                                        fontWeight: 800,
                                        textTransform: 'uppercase',
                                        display: 'flex',
                                        alignItems: 'center',
                                        justifyContent: 'center',
                                        gap: '0.5rem',
                                        border: 'none',
                                        clipPath: 'polygon(8px 0, 100% 0, 100% calc(100% - 8px), calc(100% - 8px) 100%, 0 100%, 0 8px)'
                                    }}
                                >
                                    <ShoppingCart size={18} />
                                    Adicionar
                                </button>
                            </div>
                        </motion.div>
                    ))}
                </div>
            </div>
        </section>
    );
}

export default HomeExpressShowcase;

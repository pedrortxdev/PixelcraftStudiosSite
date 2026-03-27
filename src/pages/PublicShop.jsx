import React, { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import { useNavigate } from 'react-router-dom';
import { productsAPI, gamesAPI } from '../services/api';
import PublicLayout from '../components/PublicLayout';
import { 
    Package, 
    ArrowRight, 
    ShoppingCart, 
    ChevronRight,
    Star,
    Zap,
    Shield
} from 'lucide-react';

function PublicShop() {
    const navigate = useNavigate();
    const [games, setGames] = useState([]);
    const [products, setProducts] = useState({});
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchData = async () => {
            try {
                const [gamesData, productsData] = await Promise.all([
                    gamesAPI.getAllWithCategories(),
                    productsAPI.getAll({ limit: 100 }) // Get enough to filter
                ]);

                setGames(gamesData);
                
                // Group products by game and limit to 2
                const grouped = {};
                gamesData.forEach(game => {
                    grouped[game.id] = productsData.filter(p => p.game_id === game.id).slice(0, 2);
                });
                setProducts(grouped);
            } catch (error) {
                console.error('Failed to fetch shop data:', error);
            } finally {
                setLoading(false);
            }
        };
        fetchData();
    }, []);

    const styles = {
        container: {
            maxWidth: '1400px',
            margin: '0 auto',
            padding: '8rem 2rem 4rem',
        },
        header: {
            textAlign: 'center',
            marginBottom: '4rem',
        },
        title: {
            fontSize: 'clamp(2.5rem, 5vw, 4rem)',
            fontWeight: 900,
            marginBottom: '1rem',
            background: 'linear-gradient(135deg, #F8F9FA 0%, #583AFF 100%)',
            WebkitBackgroundClip: 'text',
            WebkitTextFillColor: 'transparent',
        },
        gameSection: {
            marginBottom: '5rem',
        },
        gameHeader: {
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            marginBottom: '2rem',
            borderBottom: '1px solid rgba(255,255,255,0.05)',
            paddingBottom: '1rem',
        },
        gameTitle: {
            display: 'flex',
            alignItems: 'center',
            gap: '1rem',
            fontSize: '1.75rem',
            fontWeight: 800,
            color: '#F8F9FA',
        },
        productGrid: {
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fill, minmax(350px, 1fr))',
            gap: '2rem',
        },
        card: {
            background: 'rgba(15, 18, 25, 0.6)',
            backdropFilter: 'blur(20px)',
            border: '1px solid rgba(255, 255, 255, 0.08)',
            borderRadius: '16px',
            overflow: 'hidden',
            transition: 'all 0.3s ease',
            position: 'relative',
        },
        cardImage: {
            width: '100%',
            height: '200px',
            objectFit: 'cover',
            borderBottom: '1px solid rgba(255, 255, 255, 0.05)',
        },
        cardBody: {
            padding: '1.5rem',
        },
        productName: {
            fontSize: '1.25rem',
            fontWeight: 700,
            marginBottom: '0.5rem',
            color: '#F8F9FA',
        },
        price: {
            fontSize: '1.5rem',
            fontWeight: 900,
            color: '#583AFF',
            fontFamily: 'monospace',
        },
        seeMoreCard: {
            background: 'rgba(88, 58, 255, 0.05)',
            border: '2px dashed rgba(88, 58, 255, 0.2)',
            borderRadius: '16px',
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            padding: '2rem',
            cursor: 'pointer',
            textAlign: 'center',
            gap: '1rem',
            transition: 'all 0.3s ease',
        },
        badge: {
            padding: '0.25rem 0.75rem',
            borderRadius: '20px',
            fontSize: '0.75rem',
            fontWeight: 700,
            background: 'rgba(88, 58, 255, 0.1)',
            color: '#583AFF',
            border: '1px solid rgba(88, 58, 255, 0.2)',
            marginBottom: '1rem',
            display: 'inline-block'
        }
    };

    return (
        <PublicLayout>
            <div style={styles.container}>
                <header style={styles.header}>
                    <motion.div
                        initial={{ opacity: 0, y: -20 }}
                        animate={{ opacity: 1, y: 0 }}
                    >
                        <h1 style={styles.title}>Nossa Vitrine</h1>
                        <p style={{ color: '#B8BDC7', fontSize: '1.2rem', maxWidth: '700px', margin: '0 auto' }}>
                            Explore uma prévia do nosso catálogo exclusivo de plugins, scripts e mapas.
                        </p>
                    </motion.div>
                </header>

                {loading ? (
                    <div style={{ textAlign: 'center', padding: '4rem', color: '#B8BDC7' }}>Carregando catálogo...</div>
                ) : (
                    games.map(game => {
                        const gameProducts = products[game.id] || [];
                        if (gameProducts.length === 0) return null;

                        return (
                            <section key={game.id} style={styles.gameSection}>
                                <div style={styles.gameHeader}>
                                    <h2 style={styles.gameTitle}>
                                        <Package size={28} color="#583AFF" />
                                        {game.name}
                                    </h2>
                                    <span style={{ color: '#6C727F', fontSize: '0.9rem' }}>
                                        {gameProducts.length} itens em destaque
                                    </span>
                                </div>

                                <div style={styles.productGrid}>
                                    {gameProducts.map(product => (
                                        <motion.div 
                                            key={product.id}
                                            style={styles.card}
                                            whileHover={{ y: -10, borderColor: 'rgba(88, 58, 255, 0.4)' }}
                                        >
                                            <img 
                                                src={product.image_url || '/principal.png'} 
                                                alt={product.name} 
                                                style={styles.cardImage}
                                            />
                                            <div style={styles.cardBody}>
                                                <div style={styles.badge}>{product.type}</div>
                                                <h3 style={styles.productName}>{product.name}</h3>
                                                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: '1.5rem' }}>
                                                    <span style={styles.price}>
                                                        {new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(product.price)}
                                                    </span>
                                                    <button 
                                                        onClick={() => navigate('/register')}
                                                        style={{ 
                                                            padding: '0.6rem 1.2rem', 
                                                            background: 'var(--gradient-primary)', 
                                                            border: 'none', 
                                                            borderRadius: '8px',
                                                            color: '#white',
                                                            fontWeight: 700,
                                                            cursor: 'pointer',
                                                            display: 'flex',
                                                            alignItems: 'center',
                                                            gap: '0.5rem'
                                                        }}
                                                    >
                                                        Comprar <ShoppingCart size={16} />
                                                    </button>
                                                </div>
                                            </div>
                                        </motion.div>
                                    ))}

                                    {/* Ver Mais Card */}
                                    <motion.div 
                                        style={styles.seeMoreCard}
                                        whileHover={{ background: 'rgba(88, 58, 255, 0.1)', borderColor: 'rgba(88, 58, 255, 0.5)' }}
                                        onClick={() => navigate('/register')}
                                    >
                                        <div style={{ 
                                            width: '60px', 
                                            height: '60px', 
                                            borderRadius: '50%', 
                                            background: 'rgba(88, 58, 255, 0.1)',
                                            display: 'flex',
                                            alignItems: 'center',
                                            justifyContent: 'center',
                                            marginBottom: '1rem'
                                        }}>
                                            <ArrowRight size={30} color="#583AFF" />
                                        </div>
                                        <h3 style={{ color: '#F8F9FA', fontSize: '1.25rem', fontWeight: 700 }}>Ver Mais</h3>
                                        <p style={{ color: '#B8BDC7', fontSize: '0.9rem' }}>
                                            Crie uma conta para visualizar todos os produtos desta categoria.
                                        </p>
                                    </motion.div>
                                </div>
                            </section>
                        );
                    })
                )}

                {/* Promotional Banner */}
                <motion.div 
                    initial={{ opacity: 0 }}
                    whileInView={{ opacity: 1 }}
                    style={{
                        background: 'linear-gradient(90deg, rgba(88, 58, 255, 0.1) 0%, rgba(26, 210, 255, 0.1) 100%)',
                        border: '1px solid rgba(88, 58, 255, 0.2)',
                        borderRadius: '24px',
                        padding: '4rem',
                        textAlign: 'center',
                        marginTop: '4rem'
                    }}
                >
                    <h2 style={{ fontSize: '2.5rem', fontWeight: 900, marginBottom: '1.5rem' }}>Pronto para levar seu servidor ao próximo nível?</h2>
                    <p style={{ color: '#B8BDC7', fontSize: '1.2rem', marginBottom: '2.5rem', maxWidth: '800px', margin: '0 auto 2.5rem' }}>
                        Junte-se a mais de 500 proprietários de servidores que confiam na Pixelcraft Studio.
                    </p>
                    <button 
                        onClick={() => navigate('/register')}
                        style={{ 
                            padding: '1.25rem 3rem', 
                            background: 'var(--gradient-primary)', 
                            border: 'none', 
                            borderRadius: '12px',
                            color: 'white',
                            fontSize: '1.1rem',
                            fontWeight: 800,
                            cursor: 'pointer',
                            boxShadow: '0 10px 40px rgba(88, 58, 255, 0.3)'
                        }}
                    >
                        CRIAR CONTA AGORA
                    </button>
                </motion.div>
            </div>
        </PublicLayout>
    );
}

export default PublicShop;

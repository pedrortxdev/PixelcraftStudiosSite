import { motion } from 'framer-motion';
import LoadingSpinner from '../components/shared/LoadingSpinner';
import Skeleton from '../components/shared/Skeleton';
import {
  Package,
  Crown,
  Loader2,
  Sparkles
} from 'lucide-react';
import { useState, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

import api from '../services/api';
import ProductCard from '../components/shop/ProductCard';
import FloatingCart from '../components/shop/FloatingCart';
import DashboardLayout from '../components/DashboardLayout';

function Shop() {
  const [searchParams] = useSearchParams();
  const initialView = searchParams.get('view') === 'subscriptions' ? 'subscriptions' : 'products';
  // Check if routed directly to a game via ?game=id
  const initialGame = searchParams.get('game');

  const [viewMode, setViewMode] = useState(initialView); // 'products' | 'subscriptions'
  const [games, setGames] = useState([]);
  const [selectedGame, setSelectedGame] = useState(initialGame || null);
  const [categories, setCategories] = useState([]);
  const [selectedCategory, setSelectedCategory] = useState(null);
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const { user } = useAuth(); // Keep if needed for logic


  // Fetch games on mount
  useEffect(() => {
    const fetchGames = async () => {
      try {
        const data = await api.games.getWithCategories();
        setGames(data || []);
      } catch (err) {
        console.error('Failed to fetch games:', err);
      }
    };
    fetchGames();
  }, []);

  // Update categories when game changes
  useEffect(() => {
    if (selectedGame) {
      const game = games.find(g => g.game.id === selectedGame);
      setCategories(game?.categories || []);
      setSelectedCategory(null);
    } else {
      setCategories([]);
      setSelectedCategory(null);
    }
  }, [selectedGame, games]);

  // Fetch items when filters change
  useEffect(() => {
    const fetchItems = async () => {
      setLoading(true);
      setError(null);

      try {
        if (viewMode === 'products') {
          const params = {
            page: currentPage,
            page_size: 200, // Fetch more for Netflix view
          };

          if (selectedGame) {
            params.game_id = selectedGame;
          }
          if (selectedCategory) {
            params.category_id = selectedCategory;
          }

          const response = await api.products.getAll(params);
          setItems(response.products || []);
          // Pagination logic might need override on netflix view. Let's keep total_pages for filtered mode
          setTotalPages(response.total_pages || 1);

        } else {
          // LÓGICA DE PLANOS
          try {
            const data = await api.plans.getAll();

            const adaptedPlans = (data || []).map(plan => {
              let features = [];
              try {
                features = plan.features ? JSON.parse(plan.features) : [];
              } catch (e) {
                features = [];
              }

              return {
                id: plan.id,
                name: plan.name,
                price: Number(plan.price),
                description: plan.description,
                features: features,
                category: 'Plano',
                image_url: plan.imageUrl || null,
                type: 'PLAN'
              };
            });

            setItems(adaptedPlans);
            setTotalPages(1);
          } catch (_error) {
            console.error(_error); setError('Não foi possível carregar os planos disponíveis.');
          }
        }

      } catch (_error) {
        console.error('Failed to fetch items:', _error);
        setError('Erro ao carregar itens. Tente novamente.');
      } finally {
        setLoading(false);
      }
    };

    fetchItems();
  }, [currentPage, selectedGame, selectedCategory, viewMode]);

  // --- Estilos Específicos ---
  const styles = {
    // Replicating pageTitle style for custom header
    pageTitle: {
      fontSize: 'var(--title-h2)', fontWeight: 900,
      background: 'linear-gradient(135deg, #F8F9FA 0%, #583AFF 50%, #1AD2FF 100%)',
      WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent', letterSpacing: '-0.03em',
      marginRight: '1rem',
      lineHeight: 1 // Fix alignment
    },
    viewToggleContainer: {
      display: 'flex',
      background: 'rgba(255, 255, 255, 0.05)',
      padding: '4px',
      borderRadius: '12px',
      border: '1px solid rgba(255, 255, 255, 0.1)',
    },
    viewToggleButton: {
      display: 'flex',
      alignItems: 'center',
      gap: '8px',
      padding: '8px 16px',
      borderRadius: '8px',
      border: 'none',
      cursor: 'pointer',
      fontSize: '0.9rem',
      fontWeight: 600,
      transition: 'all 0.3s ease',
      fontFamily: 'inherit',
    },
    viewToggleButtonActive: {
      background: 'rgba(88, 58, 255, 0.2)',
      color: '#F8F9FA',
      boxShadow: '0 0 15px rgba(88, 58, 255, 0.15)',
      border: '1px solid rgba(88, 58, 255, 0.3)',
    },
    viewToggleButtonInactive: {
      background: 'transparent',
      color: '#B8BDC7',
    },
    filtersBar: { display: 'flex', gap: '12px', marginBottom: '2.5rem', flexWrap: 'wrap' },
    filterPill: { padding: '10px 24px', background: 'transparent', border: '1px solid rgba(255, 255, 255, 0.12)', borderRadius: '24px', color: '#E8E9EB', fontSize: '14px', fontWeight: 600, cursor: 'pointer', transition: 'all 0.3s ease' },
    filterPillActive: { background: 'var(--gradient-primary)', borderColor: '#583AFF', color: '#FFFFFF', boxShadow: '0 0 20px rgba(88, 58, 255, 0.4)' },
    productsGrid: { display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))', gap: '24px' },
    netflixRow: {
      display: 'flex',
      gap: '24px',
      overflowX: 'auto',
      paddingBottom: '20px',
      WebkitOverflowScrolling: 'touch',
      msOverflowStyle: 'none',  /* IE and Edge */
      scrollbarWidth: 'none',  /* Firefox */
    },
    netflixSection: {
      marginBottom: '4rem'
    },
    netflixTitle: {
      fontSize: 'var(--title-h3)',
      fontWeight: 800,
      color: 'var(--text-primary)',
      marginBottom: '1rem',
      letterSpacing: '-1px'
    },
    muted: { color: '#B8BDC7', fontSize: '0.9rem' },
  };

  // Custom Header containing Title and Toggle
  const customHeader = (
    <div style={{ display: 'flex', alignItems: 'center', gap: '2rem' }}>
      <h1 style={styles.pageTitle}>Loja</h1>
      <div style={styles.viewToggleContainer}>
        <button
          style={{
            ...styles.viewToggleButton,
            ...(viewMode === 'products' ? styles.viewToggleButtonActive : styles.viewToggleButtonInactive)
          }}
          onClick={() => setViewMode('products')}
        >
          <Package size={16} />
          Produtos
        </button>
        <button
          style={{
            ...styles.viewToggleButton,
            ...(viewMode === 'subscriptions' ? styles.viewToggleButtonActive : styles.viewToggleButtonInactive)
          }}
          onClick={() => setViewMode('subscriptions')}
        >
          <Crown size={16} />
          Assinaturas
        </button>
      </div>
    </div>
  );

  return (
    <DashboardLayout headerStart={customHeader} title="">
      {/* Game Filters (Only shown for Products) */}
      {viewMode === 'products' && (
        <>
          {/* Game Toggle Bar */}
          <div style={styles.filtersBar}>
            {/* REMOVED: Aba Todos */}
            {games.map((gameItem) => (
              <motion.button
                key={gameItem.game.id}
                style={{
                  ...styles.filterPill,
                  padding: '10px 24px',
                  borderRadius: '50px',
                  fontSize: '0.95rem',
                  border: selectedGame === gameItem.game.id
                    ? '1px solid rgba(88, 58, 255, 0.5)'
                    : '1px solid rgba(255, 255, 255, 0.1)',
                  background: selectedGame === gameItem.game.id
                    ? 'linear-gradient(135deg, rgba(88, 58, 255, 0.4) 0%, rgba(26, 210, 255, 0.4) 100%)'
                    : 'rgba(255, 255, 255, 0.03)',
                  color: selectedGame === gameItem.game.id ? '#FFFFFF' : '#B8BDC7',
                  boxShadow: selectedGame === gameItem.game.id ? '0 0 20px rgba(88, 58, 255, 0.25)' : 'none',
                  fontWeight: selectedGame === gameItem.game.id ? 700 : 500,
                }}
                whileHover={{
                  scale: 1.05,
                  backgroundColor: 'rgba(255, 255, 255, 0.08)',
                  borderColor: 'rgba(255, 255, 255, 0.3)',
                  color: '#FFF'
                }}
                whileTap={{ scale: 0.95 }}
                onClick={() => setSelectedGame(gameItem.game.id)}
              >
                {gameItem.game.name}
              </motion.button>
            ))}
          </div>

          {/* Category Sub-filter (Only when a game is selected) */}
          {selectedGame && categories.length > 0 && (
            <div style={{ ...styles.filtersBar, marginTop: '-1.5rem', marginBottom: '2.5rem', gap: '8px' }}>
              <motion.button
                key="all-cats"
                style={{
                  ...styles.filterPill,
                  padding: '6px 16px',
                  fontSize: '0.85rem',
                  borderRadius: '50px',
                  border: selectedCategory === null
                    ? '1px solid rgba(26, 210, 255, 0.5)'
                    : '1px solid rgba(255, 255, 255, 0.1)',
                  background: selectedCategory === null
                    ? 'linear-gradient(90deg, rgba(26, 210, 255, 0.2) 0%, rgba(88, 58, 255, 0.2) 100%)'
                    : 'rgba(255, 255, 255, 0.03)',
                  color: selectedCategory === null ? '#1AD2FF' : '#B8BDC7',
                  boxShadow: selectedCategory === null ? '0 0 15px rgba(26, 210, 255, 0.15)' : 'none',
                }}
                whileHover={{
                  scale: 1.05,
                  backgroundColor: 'rgba(255, 255, 255, 0.08)',
                  borderColor: 'rgba(255, 255, 255, 0.3)'
                }}
                whileTap={{ scale: 0.95 }}
                onClick={() => setSelectedCategory(null)}
              >
                Todas
              </motion.button>
              {categories.map((cat) => (
                <motion.button
                  key={cat.id}
                  style={{
                    ...styles.filterPill,
                    padding: '6px 16px',
                    fontSize: '0.85rem',
                    borderRadius: '50px',
                    border: selectedCategory === cat.id
                      ? '1px solid rgba(26, 210, 255, 0.5)'
                      : '1px solid rgba(255, 255, 255, 0.1)',
                    background: selectedCategory === cat.id
                      ? 'linear-gradient(90deg, rgba(26, 210, 255, 0.2) 0%, rgba(88, 58, 255, 0.2) 100%)'
                      : 'rgba(255, 255, 255, 0.03)',
                    color: selectedCategory === cat.id ? '#1AD2FF' : '#B8BDC7',
                    boxShadow: selectedCategory === cat.id ? '0 0 15px rgba(26, 210, 255, 0.15)' : 'none',
                  }}
                  whileHover={{
                    scale: 1.05,
                    backgroundColor: 'rgba(255, 255, 255, 0.08)',
                    borderColor: 'rgba(255, 255, 255, 0.3)'
                  }}
                  whileTap={{ scale: 0.95 }}
                  onClick={() => setSelectedCategory(cat.id)}
                >
                  {cat.name}
                </motion.button>
              ))}
            </div>
          )}
        </>
      )}

      {/* Conteúdo principal da loja */}
      {loading ? (
        <div style={styles.productsGrid}>
          {Array.from({ length: 8 }).map((_, i) => (
            <div key={i} style={{ height: '400px', display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              <Skeleton height="190px" borderRadius="16px" />
              <div style={{ padding: '0 1rem' }}>
                <Skeleton height="24px" width="70%" style={{ marginBottom: '1rem' }} />
                <Skeleton height="14px" width="100%" style={{ marginBottom: '0.5rem' }} />
                <Skeleton height="14px" width="90%" style={{ marginBottom: '2rem' }} />
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <Skeleton height="30px" width="40%" />
                  <Skeleton height="30px" width="30%" borderRadius="12px" />
                </div>
              </div>
            </div>
          ))}
        </div>
      ) : error ? (
        <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '400px', flexDirection: 'column', gap: '1rem', padding: '2rem', background: 'rgba(239, 68, 68, 0.1)', border: '1px solid rgba(239, 68, 68, 0.3)', borderRadius: '1rem', margin: '2rem 0' }}>
          <p style={{ color: '#EF4444', fontSize: '1.1rem' }}>{error}</p>
          <button
            onClick={() => window.location.reload()}
            style={{ padding: 'var(--btn-padding-md)', background: 'var(--gradient-primary)', border: 'none', borderRadius: '0.5rem', color: 'white', fontWeight: 600, cursor: 'pointer' }}
          >
            Tentar Novamente
          </button>
        </div>
      ) : items.length === 0 ? (
        <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '400px', flexDirection: 'column', gap: '1rem' }}>
          <Sparkles size={48} style={{ color: '#B8BDC7' }} />
          <p style={{ color: '#B8BDC7', fontSize: '1.1rem' }}>
            {viewMode === 'products' ? 'Nenhum produto encontrado' : 'Nenhuma assinatura disponível'}
          </p>
        </div>
      ) : (
        <>
          {viewMode === 'products' && !selectedGame ? (
            /* Netflix Style Layout rendering shelves per game */
            <div style={{ display: 'flex', flexDirection: 'column' }}>
              <style dangerouslySetInnerHTML={{
                __html: `
                .hide-scroll::-webkit-scrollbar {
                  display: none;
                }
              `}} />
              {games.map(gameObj => {
                const gameItems = items.filter(i => i.game_id === gameObj.game.id);
                if (gameItems.length === 0) return null;
                return (
                  <div key={gameObj.game.id} style={styles.netflixSection}>
                    <h3 style={styles.netflixTitle}>Destaques <span style={{ color: 'white' }}>{gameObj.game.name}</span></h3>
                    <div style={styles.netflixRow} className="hide-scroll">
                      {gameItems.slice(0, 10).map((item) => (
                        <div key={item.id} style={{ minWidth: '320px', maxWidth: '320px' }}>
                          <ProductCard product={item} />
                        </div>
                      ))}
                    </div>
                  </div>
                )
              })}
            </div>
          ) : (
            /* Standard Grid Layout if filtered by Game or Categories */
            <div style={styles.productsGrid}>
              {items.map((item) => (
                <ProductCard key={item.id} product={item} />
              ))}
            </div>
          )}

          {viewMode === 'products' && selectedGame && totalPages > 1 && (
            <div style={{ display: 'flex', justifyContent: 'center', gap: '0.5rem', marginTop: '3rem', padding: '2rem 0' }}>
              <button
                onClick={() => setCurrentPage(prev => Math.max(1, prev - 1))}
                disabled={currentPage === 1}
                style={{ padding: 'var(--btn-padding-md)', background: currentPage === 1 ? 'rgba(255, 255, 255, 0.05)' : 'var(--gradient-primary)', border: 'none', borderRadius: '0.5rem', color: 'white', fontWeight: 600, cursor: currentPage === 1 ? 'not-allowed' : 'pointer', opacity: currentPage === 1 ? 0.5 : 1 }}
              >
                Anterior
              </button>

              <span style={{ display: 'flex', alignItems: 'center', padding: '0 1rem', color: '#F8F9FA', fontWeight: 600 }}>
                Página {currentPage} de {totalPages}
              </span>

              <button
                onClick={() => setCurrentPage(prev => Math.min(totalPages, prev + 1))}
                disabled={currentPage === totalPages}
                style={{ padding: 'var(--btn-padding-md)', background: currentPage === totalPages ? 'rgba(255, 255, 255, 0.05)' : 'var(--gradient-primary)', border: 'none', borderRadius: '0.5rem', color: 'white', fontWeight: 600, cursor: currentPage === totalPages ? 'not-allowed' : 'pointer', opacity: currentPage === totalPages ? 0.5 : 1 }}
              >
                Próxima
              </button>
            </div>
          )}
        </>
      )}

      <FloatingCart />
    </DashboardLayout>
  );
}

export default Shop;
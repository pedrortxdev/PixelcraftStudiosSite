import { motion } from 'framer-motion';
import LoadingSpinner from '../components/shared/LoadingSpinner';
import Skeleton from '../components/shared/Skeleton';
import {
  Package,
  Crown,
  Loader2,
  Sparkles,
  Search
} from 'lucide-react';
import { useState, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

import api from '../services/api';
import ProductCard from '../components/shop/ProductCard';
import FloatingCart from '../components/shop/FloatingCart';
import DashboardLayout from '../components/DashboardLayout';

// Componente para seção de jogo com grade que quebra linha
const GameSection = ({ gameName, items }) => {
  const sectionStyles = {
    section: { marginBottom: '4rem' },
    title: {
      fontSize: 'var(--title-h3)',
      fontWeight: 800,
      color: 'var(--text-primary)',
      marginBottom: '1.5rem',
      letterSpacing: '-1px',
      display: 'flex',
      alignItems: 'center',
      gap: '12px'
    },
    titleLine: {
      flex: 1,
      height: '1px',
      background: 'linear-gradient(90deg, rgba(88, 58, 255, 0.3), transparent)',
    },
    grid: { 
      display: 'grid', 
      gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))', 
      gap: '24px' 
    }
  };

  return (
    <div style={sectionStyles.section}>
      <div style={sectionStyles.title}>
        <span style={{ color: 'var(--accent-cyan)' }}>{gameName}</span>
        <div style={sectionStyles.titleLine} />
      </div>
      
      <div style={sectionStyles.grid}>
        {items.map((item) => (
          <ProductCard key={item.id} product={item} />
        ))}
      </div>
    </div>
  );
};

function Shop() {
  const [searchParams] = useSearchParams();
  const initialView = searchParams.get('view') === 'subscriptions' ? 'subscriptions' : 'products';
  const initialGame = searchParams.get('game');

  const [viewMode, setViewMode] = useState(initialView);
  const [games, setGames] = useState([]);
  const [selectedGame, setSelectedGame] = useState(initialGame || null);
  const [categories, setCategories] = useState([]);
  const [selectedCategory, setSelectedCategory] = useState(null);
  const [items, setItems] = useState([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const { user } = useAuth();

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
            page_size: 200, 
          };

          if (selectedGame) params.game_id = selectedGame;
          if (selectedCategory) params.category_id = selectedCategory;

          const response = await api.products.getAll(params);
          setItems(response.products || []);
          setTotalPages(response.total_pages || 1);

        } else {
          try {
            const data = await api.plans.getAll();
            const adaptedPlans = (data || []).map(plan => {
              let features = [];
              try {
                features = plan.features ? JSON.parse(plan.features) : [];
              } catch (e) { features = []; }

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

  // Lógica de busca aproximada (Fuzzy-like)
  const filteredItems = items.filter(item => {
    if (!searchQuery) return true;
    
    const normalize = (str) => 
      str.toLowerCase().normalize("NFD").replace(/[\u0300-\u036f]/g, "");
    
    const searchLower = normalize(searchQuery);
    const nameLower = normalize(item.name || "");
    const descLower = normalize(item.description || "");
    const categoryLower = normalize(item.category || "");

    return nameLower.includes(searchLower) || 
           descLower.includes(searchLower) || 
           categoryLower.includes(searchLower);
  });

  const styles = {
    pageTitle: {
      fontSize: 'var(--title-h2)', fontWeight: 900,
      background: 'linear-gradient(135deg, #F8F9FA 0%, #583AFF 50%, #1AD2FF 100%)',
      WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent', letterSpacing: '-0.03em',
      marginRight: '1rem',
      lineHeight: 1
    },
    searchContainer: {
      position: 'relative',
      flex: 1,
      maxWidth: '400px',
      marginLeft: '1rem'
    },
    searchInput: {
      width: '100%',
      padding: '10px 16px 10px 44px',
      background: 'rgba(255, 255, 255, 0.05)',
      border: '1px solid rgba(88, 58, 255, 0.2)',
      borderRadius: '12px',
      color: '#F8F9FA',
      fontSize: '0.95rem',
      outline: 'none',
      transition: 'all 0.3s ease',
    },
    searchIcon: {
      position: 'absolute',
      left: '14px',
      top: '50%',
      transform: 'translateY(-50%)',
      color: '#6C727F'
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
    productsGrid: { 
      display: 'grid', 
      gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))', 
      gap: '24px' 
    },
    muted: { color: '#B8BDC7', fontSize: '0.9rem' },
  };

  const customHeader = (
    <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', flexWrap: 'wrap', width: '100%' }}>
      <h1 style={styles.pageTitle}>Loja</h1>
      
      <div style={styles.viewToggleContainer}>
        <button
          style={{
            ...styles.viewToggleButton,
            ...(viewMode === 'products' ? styles.viewToggleButtonActive : styles.viewToggleButtonInactive)
          }}
          onClick={() => { setViewMode('products'); setSelectedGame(null); }}
        >
          <Package size={16} />
          Produtos
        </button>
        <button
          style={{
            ...styles.viewToggleButton,
            ...(viewMode === 'subscriptions' ? styles.viewToggleButtonActive : styles.viewToggleButtonInactive)
          }}
          onClick={() => { setViewMode('subscriptions'); setSelectedGame(null); }}
        >
          <Crown size={16} />
          Assinaturas
        </button>
      </div>

      <div style={styles.searchContainer}>
        <Search size={18} style={styles.searchIcon} />
        <input 
          type="text" 
          placeholder="Pesquisar..." 
          style={styles.searchInput}
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          onFocus={(e) => e.target.style.borderColor = 'var(--accent-purple)'}
          onBlur={(e) => e.target.style.borderColor = 'rgba(88, 58, 255, 0.2)'}
        />
      </div>
    </div>
  );

  return (
    <DashboardLayout headerStart={customHeader} title="">
      <style dangerouslySetInnerHTML={{
        __html: `
        .searchInput:focus {
          border-color: #583AFF !important;
          box-shadow: 0 0 15px rgba(88, 58, 255, 0.2);
        }
      `}} />

      {viewMode === 'products' && (
        <>
          <div style={styles.filtersBar}>
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
                onClick={() => {
                  setSelectedGame(selectedGame === gameItem.game.id ? null : gameItem.game.id);
                  setCurrentPage(1);
                }}
              >
                {gameItem.game.name}
              </motion.button>
            ))}
          </div>

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
      ) : filteredItems.length === 0 ? (
        <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '400px', flexDirection: 'column', gap: '1rem' }}>
          <Sparkles size={48} style={{ color: '#B8BDC7' }} />
          <p style={{ color: '#B8BDC7', fontSize: '1.1rem' }}>
            {searchQuery ? `Nenhum resultado para "${searchQuery}"` : (viewMode === 'products' ? 'Nenhum produto encontrado' : 'Nenhuma assinatura disponível')}
          </p>
          {searchQuery && (
            <button 
              onClick={() => setSearchQuery('')}
              style={{ color: 'var(--accent-cyan)', background: 'none', border: 'none', cursor: 'pointer', fontWeight: 600 }}
            >
              Limpar busca
            </button>
          )}
        </div>
      ) : (
        <>
          {viewMode === 'products' && !selectedGame && !searchQuery ? (
            <div style={{ display: 'flex', flexDirection: 'column' }}>
              {games.map(gameObj => {
                const gameItems = items.filter(i => i.game_id === gameObj.game.id);
                if (gameItems.length === 0) return null;
                return (
                  <GameSection 
                    key={gameObj.game.id} 
                    gameName={gameObj.game.name} 
                    items={gameItems.slice(0, 15)} 
                  />
                )
              })}
            </div>
          ) : (
            <div style={styles.productsGrid}>
              {filteredItems.map((item) => (
                <ProductCard key={item.id} product={item} />
              ))}
            </div>
          )}

          {viewMode === 'products' && selectedGame && !searchQuery && totalPages > 1 && (
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
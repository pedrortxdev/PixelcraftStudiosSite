import React, { useEffect, useState } from 'react';
import {
  Download,
  Search,
  Filter,
  X,
  ChevronRight,
  AlertCircle,
  Loader2,
  Gamepad2
} from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { libraryAPI, gamesAPI } from '../services/api';
import { extractFilename } from '../utils/fileExtract';
import DashboardLayout from '../components/DashboardLayout';
import { useToast } from '../context/ToastContext';

function Downloads() {
  const navigate = useNavigate();
  const { user } = useAuth();
  const toast = useToast();
  const [products, setProducts] = useState([]);
  const [games, setGames] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [query, setQuery] = useState('');
  const [category, setCategory] = useState('all');
  const [selectedGame, setSelectedGame] = useState('all');

  // === Estilos Específicos ===
  const styles = {
    toolsBar: {
      display: 'flex',
      gap: '1rem',
      marginBottom: '2rem',
      alignItems: 'center',
      flexWrap: 'wrap',
    },
    searchBox: {
      position: 'relative',
      flex: 1,
      minWidth: '200px',
    },
    searchInput: {
      width: '100%', padding: '0.875rem 0.875rem 0.875rem 2.5rem',
      background: 'var(--bg-card)',
      backdropFilter: 'blur(10px)',
      border: '1px solid var(--border-card)',
      borderRadius: 'var(--radius-md)',
      color: 'var(--text-primary)', fontSize: '0.95rem',
    },
    searchIcon: {
      position: 'absolute', left: '0.875rem', top: '50%', transform: 'translateY(-50%)',
      color: '#B8BDC7',
    },
    selectInput: {
      padding: '0.875rem 1.25rem',
      background: 'var(--bg-card)',
      backdropFilter: 'blur(10px)',
      border: '1px solid var(--border-card)',
      borderRadius: 'var(--radius-md)',
      color: 'var(--text-primary)', fontSize: '0.95rem', cursor: 'pointer',
      outline: 'none',
      minWidth: '180px',
    },
    filterButton: {
      padding: '0.875rem 1.25rem',
      background: 'var(--bg-card)',
      backdropFilter: 'blur(10px)',
      border: '1px solid var(--border-card)',
      borderRadius: 'var(--radius-md)',
      color: 'var(--text-primary)', fontSize: '0.95rem', cursor: 'pointer',
      display: 'flex', alignItems: 'center', gap: '0.5rem',
    },
    clearButton: {
      padding: '0.875rem 1.25rem',
      background: 'var(--bg-card)',
      backdropFilter: 'blur(10px)',
      border: '1px solid rgba(239, 68, 68, 0.2)',
      borderRadius: 'var(--radius-md)',
      color: '#EF4444', fontSize: '0.95rem', cursor: 'pointer',
      display: 'flex', alignItems: 'center', gap: '0.5rem',
    },
    grid: {
      display: 'grid',
      gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))',
      gap: '1.5rem',
    },
    card: {
      background: 'var(--bg-card)',
      backdropFilter: 'blur(10px)',
      border: '1px solid var(--border-card)',
      borderRadius: 'var(--radius-lg)',
      padding: '1.25rem',
      boxShadow: 'var(--shadow-card)',
      textDecoration: 'none',
      color: '#E2E7F1',
      transition: 'all var(--transition-normal)',
      cursor: 'pointer',
    },
    cardHeader: {
      display: 'flex',
      alignItems: 'center',
      gap: '0.6rem',
      color: '#BFC7DB',
      fontWeight: 600,
      marginBottom: '0.5rem',
    },
    cardDescription: {
      color: '#E6EAF0',
      fontSize: '0.95rem',
      marginBottom: '0.75rem',
    },
    cardFooter: {
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'space-between',
      color: '#AEB7CD',
      fontSize: '0.875rem',
    },
    loadingContainer: {
      display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '400px', flexDirection: 'column', gap: '1rem',
    },
    errorContainer: {
      display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '400px', flexDirection: 'column', gap: '1rem', textAlign: 'center',
    },
    retryButton: {
      padding: 'var(--btn-padding-md)', background: 'var(--gradient-primary)', border: 'none', borderRadius: '0.5rem',
      color: 'white', fontWeight: 600, cursor: 'pointer',
    },
    categoryList: {
      background: 'var(--bg-card)',
      backdropFilter: 'blur(10px)',
      border: '1px solid var(--border-card)',
      borderRadius: 'var(--radius-lg)',
      padding: '1.25rem',
      boxShadow: 'var(--shadow-card)',
      height: 'fit-content'
    },
    categoryItem: {
      display: 'flex', alignItems: 'center', gap: '0.75rem', padding: '0.75rem 1rem', borderRadius: '0.75rem', cursor: 'pointer',
      color: '#B8BDC7', fontSize: '0.95rem',
    },
    categoryItemActive: {
      background: 'rgba(88, 58, 255, 0.15)',
      color: '#F8F9FA',
      border: '1px solid rgba(88, 58, 255, 0.3)',
    },
    sidebarTitle: {
      color: '#B8BDC7', fontSize: '0.9rem', fontWeight: 600, marginBottom: '1rem',
    },
  };

  useEffect(() => {
    const load = async () => {
      try {
        setLoading(true);
        setError(null);

        // Fetch library and games in parallel
        const [libraryData, gamesData] = await Promise.all([
          libraryAPI.getMyLibrary(),
          gamesAPI.getAll()
        ]);

        // Map backend ProductType to frontend category labels
        const categoryMapping = {
          'PLUGIN': 'plugins',
          'MOD': 'mods',
          'MAP': 'mapas',
          'TEXTUREPACK': 'texturas',
          'SERVER_TEMPLATE': 'servidores'
        };

        // Flatten data
        const flattenedData = Array.isArray(libraryData) ? libraryData.map(item => ({
          ...item.product,
          id: item.product?.id || item.purchase?.product_id, // Explicitly map the ID
          purchased_at: item.purchase?.purchased_at,
          category: categoryMapping[item.product.type] || item.product.type?.toLowerCase() || 'other',
          // Ensure game_id is preserved
          game_id: item.product.game_id
        })) : [];

        setProducts(flattenedData);
        setGames(Array.isArray(gamesData) ? gamesData : []);
      } catch (err) {
        console.error('Erro ao carregar dados:', err);
        setError('Erro ao carregar o catálogo de downloads.');
      } finally {
        setLoading(false);
      }
    };
    load();
  }, []);

  const filtered = products.filter((p) => {
    const matchesQuery = !query || p.name.toLowerCase().includes(query.toLowerCase());
    const matchesCategory = category === 'all' || p.category === category;
    const matchesGame = selectedGame === 'all' || p.game_id === selectedGame;
    return matchesQuery && matchesCategory && matchesGame;
  });

  const getCategoryLabel = (cat) => {
    const labels = {
      plugins: 'Plugins',
      mods: 'Mods',
      mapas: 'Mapas',
      servidores: 'Servidores',
      texturas: 'Texturas',
    };
    return labels[cat] || cat;
  };

  const validCategories = ['plugins', 'mods', 'mapas', 'servidores', 'texturas'];

  const [downloadingId, setDownloadingId] = useState(null);

  // ... styles ...
  const handleDownload = async (p) => {
    if (downloadingId) return;

    try {
      setDownloadingId(p.id);
      const response = await libraryAPI.downloadFile(p.id);

      const contentType = response.headers.get('content-type') || '';

      if (contentType.includes('application/json')) {
        const data = await response.json();
        if (data && data.download_url) {
          // On mobile, direct assignment is often better than window.open
          window.location.href = data.download_url;
        } else {
          toast.error(`O link de download para "${p.name}" não foi encontrado ou é inválido.`);
        }
        return; // Always exit after handling JSON
      }

      // If not JSON, it must be a file blob
      const contentDisposition = response.headers.get('Content-Disposition');
      let filename = extractFilename(contentDisposition, p.name);

      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();

      // Cleanup
      setTimeout(() => {
        window.URL.revokeObjectURL(url);
        document.body.removeChild(a);
      }, 100);
    } catch (err) {
      console.error('Download failed:', err);
      toast.error(`Erro ao baixar "${p.name}": ${err.message}`);
    } finally {
      setDownloadingId(null);
    }
  };

  return (
    <DashboardLayout title="Catálogo de Downloads">
      {/* Barra de busca e ações */}
      <div style={styles.toolsBar}>
        <div style={styles.searchBox}>
          <Search size={20} style={styles.searchIcon} />
          <input
            type="text"
            placeholder="Buscar downloads..."
            style={styles.searchInput}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
        </div>

        {/* Filtro de Jogo */}
        <select
          style={styles.selectInput}
          value={selectedGame}
          onChange={(e) => setSelectedGame(e.target.value)}
        >
          <option value="all">Todos os Jogos</option>
          {games.map(game => (
            <option key={game.id} value={game.id}>{game.name}</option>
          ))}
        </select>

        <button
          style={styles.clearButton}
          onClick={() => {
            setQuery('');
            setCategory('all');
            setSelectedGame('all');
          }}
        >
          <X size={18} /> Limpar
        </button>
      </div>

      <div className="downloads-grid-layout" style={{ gap: '2rem' }}>
        {/* Sidebar de Categorias */}
        <div style={styles.categoryList}>
          <div style={styles.sidebarTitle}>Categorias</div>
          <div
            style={{
              ...styles.categoryItem,
              ...(category === 'all' ? styles.categoryItemActive : {}),
            }}
            onClick={() => setCategory('all')}
          >
            Todos
          </div>
          {validCategories.map((cat) => (
            <div
              key={cat}
              style={{
                ...styles.categoryItem,
                ...(category === cat ? styles.categoryItemActive : {}),
              }}
              onClick={() => setCategory(cat)}
            >
              {getCategoryLabel(cat)}
            </div>
          ))}
        </div>

        {/* Lista de Produtos */}
        <div>
          {loading ? (
            <div style={styles.loadingContainer}>
              <Loader2 size={48} style={{ color: '#583AFF', animation: 'spin 1s linear infinite' }} />
              <p style={{ color: '#B8BDC7', fontSize: '1.1rem' }}>Carregando downloads...</p>
            </div>
          ) : error ? (
            <div style={styles.errorContainer}>
              <AlertCircle size={48} style={{ color: '#EF4444' }} />
              <p style={{ color: '#EF4444', fontSize: '1.1rem' }}>{error}</p>
              <button onClick={() => window.location.reload()} style={styles.retryButton}>Tentar Novamente</button>
            </div>
          ) : filtered.length > 0 ? (
            <div style={styles.grid}>
              {filtered.map((p) => (
                <div
                  key={p.id}
                  style={{
                    ...styles.card,
                    opacity: downloadingId && downloadingId !== p.id ? 0.6 : 1,
                    pointerEvents: downloadingId ? 'none' : 'auto'
                  }}
                  onClick={() => handleDownload(p)}
                  role="button"
                  tabIndex="0"
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') handleDownload(p);
                  }}
                >
                  <div style={styles.cardHeader}>
                    {downloadingId === p.id ? (
                      <Loader2 size={18} className="animate-spin" style={{ color: '#1AD2FF' }} />
                    ) : (
                      <Download size={18} />
                    )}
                    {p.name}
                  </div>
                  <div style={styles.cardFooter}>
                    <span>{getCategoryLabel(p.category)}</span>
                    <span style={{ display: 'inline-flex', alignItems: 'center', gap: '0.3rem', color: '#1AD2FF' }}>
                      {downloadingId === p.id ? 'Iniciando...' : 'Baixar'}
                      {downloadingId !== p.id && <ChevronRight size={16} />}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div style={{ textAlign: 'center', padding: '2rem', color: '#B8BDC7', fontStyle: 'italic', background: 'rgba(255,255,255,0.03)', borderRadius: '0.75rem' }}>
              Nenhum item encontrado.
            </div>
          )}
        </div>
      </div>
    </DashboardLayout>
  );
}

export default Downloads;
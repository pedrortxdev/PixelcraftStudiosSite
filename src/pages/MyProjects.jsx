import { motion } from 'framer-motion';
import {
  Search, Filter, Loader2, AlertCircle, ArrowLeft
} from 'lucide-react';
import { useState, useEffect, useMemo, useRef, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import ProjectCard from '../components/dashboard/ProjectCard';
import { subscriptionsAPI } from '../services/api';
import DashboardLayout from '../components/DashboardLayout';

function MyProjects() {
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState('all');
  const [showFilters, setShowFilters] = useState(false);

  const [projects, setProjects] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const navigate = useNavigate();
  const { user } = useAuth();
  const filterRef = useRef(null);

  // Close filter dropdown on outside click or Escape key
  useEffect(() => {
    if (!showFilters) return;
    const handleClick = (e) => {
      if (filterRef.current && !filterRef.current.contains(e.target)) {
        setShowFilters(false);
      }
    };
    const handleKey = (e) => {
      if (e.key === 'Escape') setShowFilters(false);
    };
    document.addEventListener('mousedown', handleClick);
    document.addEventListener('keydown', handleKey);
    return () => {
      document.removeEventListener('mousedown', handleClick);
      document.removeEventListener('keydown', handleKey);
    };
  }, [showFilters]);

  const mapStatus = (backendStatus) => {
    switch (backendStatus?.toUpperCase()) {
      case 'ACTIVE': return 'desenvolvimento';
      case 'COMPLETED': return 'concluido';
      case 'CANCELED': return 'cancelado';
      default: return 'desenvolvimento';
    }
  };

  const mapStageToNumber = (stageName) => {
    const stages = ['Planejamento', 'Desenvolvimento', 'Otimização', 'Testes', 'Entrega'];
    const index = stages.indexOf(stageName);
    return index >= 0 ? index + 1 : 1;
  };

  const formatDate = (isoDate) => {
    if (!isoDate) return 'N/A';
    return new Date(isoDate).toLocaleDateString('pt-BR');
  };

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        setError(null);

        const data = await subscriptionsAPI.getMySubscriptions();

        const formattedProjects = (data || []).map(sub => ({
          id: sub.id,
          name: `Projeto - ${sub.planName || 'Customizado'}`,
          status: mapStatus(sub.status),
          plan: sub.planName || 'Plano Desconhecido',
          startDate: formatDate(sub.startedAt),
          nextBilling: formatDate(sub.nextBillingDate),
          currentStage: mapStageToNumber(sub.projectStage),
          stages: ['Planejamento', 'Desenvolvimento', 'Otimização', 'Testes', 'Entrega'],
          logs: sub.logs || []
        }));

        setProjects(formattedProjects);
      } catch (err) {
        console.error('Failed to fetch projects:', err);
        setError('Erro ao carregar seus projetos. Tente novamente.');
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, []);

  const filteredProjects = useMemo(() => {
    return projects.filter(project => {
      const matchesSearch = project.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        project.plan.toLowerCase().includes(searchTerm.toLowerCase());
      const matchesStatus = statusFilter === 'all' || project.status === statusFilter;
      return matchesSearch && matchesStatus;
    });
  }, [projects, searchTerm, statusFilter]);

  const projectsByStatus = useMemo(() => {
    const grouped = {};
    filteredProjects.forEach(project => {
      if (!grouped[project.status]) {
        grouped[project.status] = [];
      }
      grouped[project.status].push(project);
    });
    return grouped;
  }, [filteredProjects]);

  const getStatusLabel = (status) => {
    switch (status) {
      case 'desenvolvimento': return 'Em Desenvolvimento';
      case 'concluido': return 'Concluídos';
      case 'cancelado': return 'Cancelados';
      default: return status;
    }
  };

  const getStatusColor = (status) => {
    switch (status) {
      case 'desenvolvimento': return '#583AFF';
      case 'concluido': return '#1AD2FF';
      case 'cancelado': return '#EF4444';
      default: return '#6C7384';
    }
  };

  const stats = {
    total: projects.length,
    active: projects.filter(p => p.status === 'desenvolvimento').length,
    completed: projects.filter(p => p.status === 'concluido').length,
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
    searchContainer: {
      display: 'flex', gap: '1rem', marginBottom: '2rem', alignItems: 'center',
    },
    searchBox: { position: 'relative', flex: 1 },
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
    filterButton: {
      padding: '0.875rem 1.25rem',
      background: 'var(--bg-card)',
      backdropFilter: 'blur(10px)',
      border: '1px solid var(--border-card)',
      borderRadius: 'var(--radius-md)',
      color: 'var(--text-primary)', fontSize: '0.95rem', cursor: 'pointer',
      display: 'flex', alignItems: 'center', gap: '0.5rem',
    },
    filterDropdown: {
      position: 'absolute', top: '100%', right: 0,
      background: 'var(--bg-secondary)',
      backdropFilter: 'blur(10px)',
      border: '1px solid var(--border-card)',
      borderRadius: 'var(--radius-md)',
      padding: '0.75rem', zIndex: 1000, minWidth: '180px', marginTop: '0.5rem',
    },
    filterOption: {
      padding: '0.625rem', borderRadius: '0.5rem', cursor: 'pointer', color: '#B8BDC7', fontSize: '0.9rem',
    },
    filterOptionActive: {
      background: 'rgba(88, 58, 255, 0.15)', color: '#F8F9FA',
    },
    statsBar: {
      display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '1.5rem', marginBottom: '2.5rem',
    },
    statCard: {
      background: 'var(--bg-card)',
      backdropFilter: 'blur(10px)',
      border: '1px solid var(--border-card)',
      borderRadius: 'var(--radius-lg)',
      padding: '1.5rem',
      boxShadow: 'var(--shadow-card)',
      textAlign: 'center',
    },
    statLabel: { color: 'var(--text-secondary)', fontSize: '0.9rem', fontWeight: 500, marginBottom: '0.5rem' },
    statValue: {
      fontSize: '1.75rem',
      fontWeight: 800,
      background: 'linear-gradient(135deg, #F8F9FA 0%, #583AFF 50%, #1AD2FF 100%)',
      WebkitBackgroundClip: 'text',
      WebkitTextFillColor: 'transparent'
    },
    statusSection: { marginBottom: '2.5rem' },
    statusHeader: { display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '1.25rem' },
    statusBadge: {
      padding: '0.25rem 0.75rem', borderRadius: '20px', fontSize: '0.8rem', fontWeight: 700,
    },
    sectionTitle: {
      fontSize: '1.25rem',
      fontWeight: 700,
      color: '#F8F9FA',
    },
    projectsGrid: { display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(480px, 1fr))', gap: '1.5rem' },
    noProjects: {
      textAlign: 'center', padding: '2rem', color: '#B8BDC7', fontSize: '1.1rem', fontStyle: 'italic',
      background: 'rgba(255,255,255,0.03)', borderRadius: '0.75rem'
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
      onClick={() => navigate('/dashboard')}
      title="Voltar ao Dashboard"
    >
      <ArrowLeft size={20} />
    </motion.div>
  );

  if (loading) {
    return (
      <DashboardLayout title="Meus Projetos" headerStart={headerStart}>
        <div style={styles.loadingContainer}>
          <Loader2 size={48} style={{ color: '#583AFF', animation: 'spin 1s linear infinite' }} />
          <p style={{ color: '#B8BDC7', fontSize: '1.1rem' }}>Carregando seus projetos...</p>
        </div>
      </DashboardLayout>
    );
  }

  if (error) {
    return (
      <DashboardLayout title="Meus Projetos" headerStart={headerStart}>
        <div style={styles.errorContainer}>
          <AlertCircle size={48} style={{ color: '#EF4444' }} />
          <p style={{ color: '#EF4444', fontSize: '1.1rem' }}>{error}</p>
          <button onClick={() => window.location.reload()} style={styles.retryButton}>Tentar Novamente</button>
        </div>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout title="Meus Projetos" headerStart={headerStart}>
      <div style={styles.searchContainer}>
        <div style={styles.searchBox}>
          <Search size={20} style={styles.searchIcon} />
          <input
            type="text"
            placeholder="Buscar projetos..."
            style={styles.searchInput}
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
        <div style={{ position: 'relative' }} ref={filterRef}>
          <motion.div
            style={styles.filterButton}
            whileHover={{ background: 'rgba(88, 58, 255, 0.15)', borderColor: 'rgba(88, 58, 255, 0.4)' }}
            onClick={() => setShowFilters(!showFilters)}
          >
            <Filter size={20} />
            Filtros
          </motion.div>
          {showFilters && (
            <motion.div
              style={styles.filterDropdown}
              initial={{ opacity: 0, y: -10 }}
              animate={{ opacity: 1, y: 0 }}
            >
              {['all', 'desenvolvimento', 'concluido', 'cancelado'].map((filter) => (
                <div
                  key={filter}
                  style={{
                    ...styles.filterOption,
                    ...(statusFilter === filter ? styles.filterOptionActive : {}),
                  }}
                  onClick={() => {
                    setStatusFilter(filter);
                    setShowFilters(false);
                  }}
                >
                  {filter === 'all' ? 'Todos os projetos' : getStatusLabel(filter)}
                </div>
              ))}
            </motion.div>
          )}
        </div>
      </div>

      <div style={styles.statsBar} className="projects-stats-grid">
        <motion.div style={styles.statCard} whileHover={{ transform: 'translateY(-4px)', boxShadow: '0 12px 40px rgba(88, 58, 255, 0.25)', borderColor: 'rgba(88, 58, 255, 0.4)' }}>
          <div style={styles.statLabel}>Total de Projetos</div>
          <div style={styles.statValue}>{stats.total}</div>
        </motion.div>
        <motion.div style={styles.statCard} whileHover={{ transform: 'translateY(-4px)', boxShadow: '0 12px 40px rgba(88, 58, 255, 0.25)', borderColor: 'rgba(88, 58, 255, 0.4)' }}>
          <div style={styles.statLabel}>Em Desenvolvimento</div>
          <div style={styles.statValue}>{stats.active}</div>
        </motion.div>
        <motion.div style={styles.statCard} whileHover={{ transform: 'translateY(-4px)', boxShadow: '0 12px 40px rgba(26, 210, 255, 0.25)', borderColor: 'rgba(26, 210, 255, 0.4)' }}>
          <div style={styles.statLabel}>Concluídos</div>
          <div style={styles.statValue}>{stats.completed}</div>
        </motion.div>
      </div>

      {Object.keys(projectsByStatus).length > 0 ? (
        Object.entries(projectsByStatus).map(([status, projects]) => (
          <div key={status} style={styles.statusSection}>
            <div style={styles.statusHeader}>
              <div
                style={{
                  ...styles.statusBadge,
                  background: `rgba(${parseInt(getStatusColor(status).slice(1, 3), 16)}, ${parseInt(getStatusColor(status).slice(3, 5), 16)}, ${parseInt(getStatusColor(status).slice(5, 7), 16)}, 0.1)`,
                  color: getStatusColor(status),
                  border: `1px solid ${getStatusColor(status)}`,
                }}
              >
                {getStatusLabel(status)}
              </div>
              <h2 style={styles.sectionTitle}>
                {projects.length} {projects.length === 1 ? 'Projeto' : 'Projetos'}
              </h2>
            </div>
            <div style={styles.projectsGrid}>
              {projects.map((project) => (
                <ProjectCard key={project.id} project={project} />
              ))}
            </div>
          </div>
        ))
      ) : (
        <div style={styles.noProjects}>
          {projects.length > 0
            ? 'Nenhum projeto encontrado com os filtros aplicados.'
            : 'Você não possui projetos em desenvolvimento no momento.'}
        </div>
      )}
    </DashboardLayout>
  );
}

export default MyProjects;
import { useState, useEffect } from 'react';
import {
    Cpu,
    MemoryStick,
    HardDrive,
    RefreshCw,
    Server,
    Activity,
    Clock,
    Zap,
    ShieldAlert,
    Download,
    Upload,
    Wifi
} from 'lucide-react';
import { adminAPI } from '../../services/api';
import { useAuth } from '../../context/AuthContext';

function AdminSystemResources() {
    const { user } = useAuth();
    const [metrics, setMetrics] = useState(null);
    const [loading, setLoading] = useState(true);
    const [refreshing, setRefreshing] = useState(false);
    const [error, setError] = useState(null);
    const [ping, setPing] = useState(0);

    // Check if user has SYSTEM VIEW permission
    const hasSystemPermission = () => {
        if (user?.is_admin) return true;
        const allowedRoles = ['DEVELOPMENT', 'ENGINEERING', 'DIRECTION'];
        const userRoles = user?.roles || [];
        return userRoles.some(role => allowedRoles.includes(role));
    };

    const canViewSystem = hasSystemPermission();

    const fetchMetrics = async () => {
        if (!canViewSystem) {
            setError('Você não tem permissão para visualizar os recursos do sistema');
            setLoading(false);
            return;
        }

        const start = performance.now();
        try {
            const data = await adminAPI.getSystemMetrics();
            const end = performance.now();
            setPing(Math.round(end - start));
            setMetrics(data);
            setError(null);
        } catch (err) {
            setError('Erro ao carregar métricas do sistema');
            console.error('Error fetching system metrics:', err);
        } finally {
            setLoading(false);
            setRefreshing(false);
        }
    };

    useEffect(() => {
        fetchMetrics();
        const interval = setInterval(fetchMetrics, 5000); // 5s for better real-time feel
        return () => clearInterval(interval);
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    const handleRefresh = async () => {
        setRefreshing(true);
        await fetchMetrics();
    };

    const formatBytes = (mb) => {
        if (mb >= 1024) return `${(mb / 1024).toFixed(2)} GB`;
        return `${mb.toFixed(0)} MB`;
    };

    const formatSpeed = (kbps) => {
        if (kbps >= 1024) return `${(kbps / 1024).toFixed(2)} Mbps`;
        return `${kbps.toFixed(1)} Kbps`;
    };

    const getUsageColor = (percent) => {
        if (percent >= 90) return '#EF4444';
        if (percent >= 70) return '#F59E0B';
        if (percent >= 50) return '#10B981';
        return '#3B82F6';
    };

    const getUsageGradient = (percent) => {
        const color = getUsageColor(percent);
        return `linear-gradient(90deg, ${color} 0%, ${color}80 100%)`;
    };

    const styles = {
        container: { color: '#F8F9FA' },
        header: { marginBottom: '2rem', display: 'flex', justifyContent: 'space-between', alignItems: 'center' },
        title: {
            fontSize: 'var(--title-h3)',
            fontWeight: 800,
            background: 'linear-gradient(135deg, #F8F9FA 0%, #583AFF 100%)',
            WebkitBackgroundClip: 'text',
            WebkitTextFillColor: 'transparent',
        },
        refreshButton: {
            display: 'flex', alignItems: 'center', gap: '0.5rem', padding: '0.6rem 1rem',
            background: 'rgba(88, 58, 255, 0.1)', border: '1px solid rgba(88, 58, 255, 0.3)',
            borderRadius: '0.5rem', color: '#583AFF', fontSize: '0.875rem', fontWeight: 600, cursor: 'pointer',
        },
        statsGrid: {
            display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))',
            gap: '1.5rem', marginBottom: '2rem',
        },
        statCard: {
            background: 'rgba(15, 18, 25, 0.6)', backdropFilter: 'blur(20px)',
            border: '1px solid rgba(255, 255, 255, 0.1)', borderRadius: '16px',
            padding: '1.5rem', position: 'relative', overflow: 'hidden',
        },
        statHeader: { display: 'flex', alignItems: 'center', gap: '1rem', marginBottom: '1.5rem' },
        statIcon: { width: '48px', height: '48px', borderRadius: '12px', display: 'flex', alignItems: 'center', justifyContent: 'center' },
        statTitle: { fontSize: '1rem', fontWeight: 600, color: '#B8BDC7' },
        statValue: { fontSize: '2rem', fontWeight: 700, color: '#F8F9FA', marginBottom: '0.5rem' },
        statLabel: { fontSize: '0.875rem', color: '#6C7384' },
        progressBar: { width: '100%', height: '8px', background: 'rgba(255, 255, 255, 0.1)', borderRadius: '4px', overflow: 'hidden', marginTop: '1rem' },
        progressFill: { height: '100%', borderRadius: '4px', transition: 'width 0.5s ease' },
        detailGrid: { display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', gap: '1.5rem' },
        detailCard: { background: 'rgba(15, 18, 25, 0.6)', backdropFilter: 'blur(20px)', border: '1px solid rgba(255, 255, 255, 0.1)', borderRadius: '16px', padding: '1.5rem' },
        detailTitle: { fontSize: '1.1rem', fontWeight: 700, color: '#F8F9FA', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.5rem' },
        detailItem: { display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '0.75rem 0', borderBottom: '1px solid rgba(255, 255, 255, 0.05)' },
        detailLabel: { color: '#B8BDC7', fontSize: '0.875rem' },
        detailValue: { color: '#F8F9FA', fontSize: '0.875rem', fontWeight: 600 },
        coreGrid: { display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(60px, 1fr))', gap: '0.5rem', marginTop: '1rem' },
        coreBadge: { background: 'rgba(255, 255, 255, 0.05)', border: '1px solid rgba(255, 255, 255, 0.1)', borderRadius: '8px', padding: '0.5rem', textAlign: 'center' },
        coreLabel: { fontSize: '0.7rem', color: '#6C7384', marginBottom: '0.25rem' },
        coreValue: { fontSize: '0.875rem', fontWeight: 600, color: '#F8F9FA' },
        loadingText: { textAlign: 'center', padding: '4rem', color: '#B8BDC7' },
        errorText: { textAlign: 'center', padding: '2rem', color: '#EF4444', background: 'rgba(239, 68, 68, 0.1)', borderRadius: '12px', border: '1px solid rgba(239, 68, 68, 0.3)' },
        timestamp: { fontSize: '0.75rem', color: '#6C7384', marginTop: '0.5rem', display: 'flex', alignItems: 'center', gap: '0.25rem' },
        ddosNormal: { background: 'rgba(16, 185, 129, 0.1)', color: '#10B981', border: '1px solid rgba(16, 185, 129, 0.2)', padding: '2px 8px', borderRadius: '4px', fontSize: '0.75rem' },
        ddosAttack: { background: 'rgba(239, 68, 68, 0.1)', color: '#EF4444', border: '1px solid rgba(239, 68, 68, 0.2)', padding: '2px 8px', borderRadius: '4px', fontSize: '0.75rem', animation: 'pulse 2s infinite' },
    };

    if (loading) return (
        <div style={styles.container}>
            <div style={styles.header}><h1 style={styles.title}>Recursos do Sistema</h1></div>
            <div style={styles.loadingText}>Carregando métricas do sistema...</div>
        </div>
    );

    if (error || !canViewSystem) return (
        <div style={styles.container}>
            <div style={styles.header}><h1 style={styles.title}>Recursos do Sistema</h1></div>
            <div style={error ? styles.errorText : { textAlign: 'center', padding: '4rem' }}>
                {!canViewSystem && <ShieldAlert size={64} color="#EF4444" style={{ margin: '0 auto 2rem' }} />}
                {error || 'Você não tem permissão para acessar esta área.'}
            </div>
        </div>
    );

    return (
        <div style={styles.container}>
            <style>{`
                @keyframes pulse {
                    0% { opacity: 1; }
                    50% { opacity: 0.5; }
                    100% { opacity: 1; }
                }
            `}</style>
            <div style={styles.header}>
                <h1 style={styles.title}>Recursos do Sistema</h1>
                <button style={styles.refreshButton} onClick={handleRefresh} disabled={refreshing}>
                    <RefreshCw size={16} style={{ animation: refreshing ? 'spin 1s linear infinite' : 'none' }} />
                    {refreshing ? 'Atualizando...' : 'Atualizar'}
                </button>
            </div>

            {metrics && (
                <>
                    <div style={styles.statsGrid}>
                        <div style={styles.statCard}>
                            <div style={styles.statHeader}>
                                <div style={{ ...styles.statIcon, background: 'rgba(224, 26, 79, 0.1)' }}><Cpu size={24} color="#E01A4F" /></div>
                                <div style={styles.statTitle}>CPU</div>
                            </div>
                            <div style={styles.statValue}>{metrics.cpu?.usage_percent?.toFixed(1)}%</div>
                            <div style={styles.statLabel}>Uso Atual ({metrics.cpu?.total_cores || 0} núcleos)</div>
                            <div style={styles.progressBar}>
                                <div style={{ ...styles.progressFill, width: `${metrics.cpu?.usage_percent || 0}%`, background: getUsageGradient(metrics.cpu?.usage_percent || 0) }} />
                            </div>
                        </div>

                        <div style={styles.statCard}>
                            <div style={styles.statHeader}>
                                <div style={{ ...styles.statIcon, background: 'rgba(88, 58, 255, 0.1)' }}><MemoryStick size={24} color="#583AFF" /></div>
                                <div style={styles.statTitle}>Memória RAM</div>
                            </div>
                            <div style={styles.statValue}>{formatBytes(metrics.memory?.used_mb || 0)}</div>
                            <div style={styles.statLabel}>de {formatBytes(metrics.memory?.total_mb || 0)} ({metrics.memory?.usage_percent?.toFixed(1)}%)</div>
                            <div style={styles.progressBar}>
                                <div style={{ ...styles.progressFill, width: `${metrics.memory?.usage_percent || 0}%`, background: getUsageGradient(metrics.memory?.usage_percent || 0) }} />
                            </div>
                        </div>

                        <div style={styles.statCard}>
                            <div style={styles.statHeader}>
                                <div style={{ ...styles.statIcon, background: 'rgba(16, 185, 129, 0.1)' }}><Wifi size={24} color="#10B981" /></div>
                                <div style={styles.statTitle}>Rede & DDoS</div>
                            </div>
                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-end', marginBottom: '0.5rem' }}>
                                <div>
                                    <div style={{ ...styles.statValue, fontSize: '1.5rem', marginBottom: 0 }}>{formatSpeed(metrics.network?.download_kbps || 0)}</div>
                                    <div style={styles.statLabel}><Download size={12} style={{ verticalAlign: 'middle', marginRight: '4px' }} /> Download</div>
                                </div>
                                <div style={{ textAlign: 'right' }}>
                                    <div style={{ ...styles.statValue, fontSize: '1.2rem', marginBottom: 0 }}>{formatSpeed(metrics.network?.upload_kbps || 0)}</div>
                                    <div style={styles.statLabel}><Upload size={12} style={{ verticalAlign: 'middle', marginRight: '4px' }} /> Upload</div>
                                </div>
                            </div>
                            <div style={{ marginTop: '1rem', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                <span style={styles.statLabel}>Ping API: <span style={{ color: ping > 200 ? '#EF4444' : '#10B981', fontWeight: 700 }}>{ping}ms</span></span>
                                <span style={metrics.ddos?.status === 'Normal' ? styles.ddosNormal : styles.ddosAttack}>
                                    {metrics.ddos?.status === 'Normal' ? 'Seguro' : metrics.ddos?.status}
                                </span>
                            </div>
                        </div>

                        <div style={styles.statCard}>
                            <div style={styles.statHeader}>
                                <div style={{ ...styles.statIcon, background: 'rgba(255, 107, 53, 0.1)' }}><Server size={24} color="#FF6B35" /></div>
                                <div style={styles.statTitle}>Sistema</div>
                            </div>
                            <div style={{ ...styles.statValue, fontSize: '1.5rem' }}>{metrics.goroutines || 0}</div>
                            <div style={styles.statLabel}>Goroutines Ativas</div>
                            <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: '1rem', fontSize: '0.8rem', color: '#6C7384' }}>
                                <span>Cores: {metrics.num_cpu || 0}</span>
                                <span>Uptime: {Math.floor(metrics.uptime_seconds / 3600)}h {Math.floor((metrics.uptime_seconds % 3600) / 60)}m</span>
                            </div>
                        </div>
                    </div>

                    <div style={styles.detailGrid}>
                        <div style={styles.detailCard}>
                            <div style={styles.detailTitle}><Activity size={20} color="#E01A4F" /> Uso por Núcleo</div>
                            <div style={styles.coreGrid}>
                                {(metrics.cpu?.per_core_usage || []).map((usage, idx) => (
                                    <div key={idx} style={styles.coreBadge}>
                                        <div style={styles.coreLabel}>Core {idx + 1}</div>
                                        <div style={{ ...styles.coreValue, color: getUsageColor(usage) }}>{usage.toFixed(1)}%</div>
                                    </div>
                                ))}
                            </div>
                        </div>

                        <div style={styles.detailCard}>
                            <div style={styles.detailTitle}><Zap size={20} color="#583AFF" /> Status de Ataque</div>
                            <div style={styles.detailItem}>
                                <span style={styles.detailLabel}>Estado</span>
                                <span style={metrics.ddos?.status === 'Normal' ? styles.ddosNormal : styles.ddosAttack}>{metrics.ddos?.status}</span>
                            </div>
                            <div style={styles.detailItem}>
                                <span style={styles.detailLabel}>Tráfego Ingress</span>
                                <span style={styles.detailValue}>{metrics.ddos?.ingress_rate_mbps} Mbps</span>
                            </div>
                            <div style={styles.detailItem}>
                                <span style={styles.detailLabel}>Motivo</span>
                                <span style={{ ...styles.detailValue, fontSize: '0.75rem' }}>{metrics.ddos?.detection_reason || 'Nenhuma ameaça'}</span>
                            </div>
                        </div>

                        <div style={styles.detailCard}>
                            <div style={styles.detailTitle}><HardDrive size={20} color="#FF6B35" /> Disco Principal</div>
                            <div style={styles.detailItem}><span style={styles.detailLabel}>Total</span><span style={styles.detailValue}>{metrics.disk?.total_gb?.toFixed(2)} GB</span></div>
                            <div style={styles.detailItem}><span style={styles.detailLabel}>Em Uso</span><span style={{ ...styles.detailValue, color: getUsageColor(metrics.disk?.usage_percent || 0) }}>{metrics.disk?.used_gb?.toFixed(2)} GB</span></div>
                            <div style={styles.detailItem}><span style={styles.detailLabel}>Livre</span><span style={styles.detailValue}>{metrics.disk?.free_gb?.toFixed(2)} GB</span></div>
                        </div>
                    </div>
                </>
            )}
        </div>
    );
}

export default AdminSystemResources;

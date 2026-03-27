import React, { useState, useRef, useEffect, useMemo } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { getAvatarUrl } from '../utils/formatAvatarUrl';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import RoleBadge from './RoleBadge';
import {
    LayoutDashboard,
    Server,
    ShoppingBag,
    Download,
    History,
    DollarSign,
    FileText,
    Settings,
    Bell,
    User,
    LogOut,
    Shield,
    CreditCard,
    ChevronDown,
    Headphones,
    Wallet as WalletIcon,
    Menu,
    X
} from 'lucide-react';
import BottomNavigation from './BottomNavigation';

const DashboardLayout = ({ children, title, headerStart }) => {
    const { user, logout } = useAuth();
    const navigate = useNavigate();
    const location = useLocation();
    const [isDropdownOpen, setIsDropdownOpen] = useState(false);
    const [isNotificationOpen, setIsNotificationOpen] = useState(false);
    const [isMobileSidebarOpen, setIsMobileSidebarOpen] = useState(false);
    const dropdownRef = useRef(null);
    const notificationRef = useRef(null);

    const styles = useMemo(() => ({
        container: {
            display: 'flex',
            minHeight: '100vh',
            background: 'linear-gradient(180deg, #0A0E1A 0%, #12182A 50%, #0A0E1A 100%)',
            position: 'relative',
            overflow: 'hidden',
        },
        backgroundTexture: {
            position: 'absolute',
            top: 0, left: 0, right: 0, bottom: 0,
            background: 'radial-gradient(circle at 20% 30%, rgba(88, 58, 255, 0.08) 0%, transparent 50%), radial-gradient(circle at 80% 70%, rgba(26, 210, 255, 0.08) 0%, transparent 50%)',
            pointerEvents: 'none',
        },
        sidebar: {
            width: '260px',
            background: 'rgba(10, 14, 26, 0.8)',
            backdropFilter: 'blur(20px)',
            borderRight: '1px solid rgba(88, 58, 255, 0.2)',
            padding: '2rem 0',
            display: 'flex',
            flexDirection: 'column',
            position: 'fixed',
            height: '100vh',
            zIndex: 100,
            transition: 'transform 0.3s ease',
        },
        mobileOverlay: {
            position: 'fixed',
            top: 0, left: 0, right: 0, bottom: 0,
            background: 'rgba(0, 0, 0, 0.6)',
            zIndex: 99,
        },
        hamburger: {
            display: 'none',
            position: 'fixed',
            top: '1.5rem',
            left: '1rem',
            zIndex: 200,
            width: '44px',
            height: '44px',
            borderRadius: 'var(--radius-md)',
            background: 'var(--bg-card)',
            border: '1px solid var(--border-card)',
            color: 'var(--text-primary)',
            alignItems: 'center',
            justifyContent: 'center',
            cursor: 'pointer',
            boxShadow: 'var(--shadow-card)',
        },
        logo: { display: 'flex', alignItems: 'center', gap: '0.75rem', padding: '0 1.5rem', marginBottom: '3rem' },
        logoImage: { width: '32px', height: '32px' },
        logoText: { fontSize: '1.25rem', fontWeight: 900, color: '#F8F9FA', letterSpacing: '-0.02em' },
        menuItem: {
            display: 'flex', alignItems: 'center', gap: '1rem', padding: '0.875rem 1.5rem',
            color: '#B8BDC7', textDecoration: 'none', transition: 'all 0.3s', cursor: 'pointer',
            fontSize: '0.95rem', fontWeight: 500, borderLeft: '3px solid transparent',
            background: 'transparent', border: 'none', borderRight: 'none', borderTop: 'none', borderBottom: 'none',
            width: '100%', textAlign: 'left', fontFamily: 'inherit', outline: 'none'
        },
        menuItemActive: {
            color: '#F8F9FA',
            background: 'rgba(88, 58, 255, 0.15)',
            borderLeft: '3px solid #583AFF',
            boxShadow: '0 0 20px rgba(88, 58, 255, 0.25)'
        },
        mainContent: { marginLeft: '260px', flex: 1, padding: '2.5rem 3rem', position: 'relative', zIndex: 1 },
        header: {
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: '2.5rem',
            paddingBottom: '1.5rem',
            borderBottom: '1px solid rgba(88, 58, 255, 0.2)',
            position: 'relative',
            zIndex: 50
        },
        headerLeft: { display: 'flex', alignItems: 'center', gap: '1rem' },
        pageTitle: {
            fontSize: 'var(--title-h2)', fontWeight: 900,
            background: 'linear-gradient(135deg, #F8F9FA 0%, #583AFF 50%, #1AD2FF 100%)',
            WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent', letterSpacing: '-0.03em',
        },
        headerRight: { display: 'flex', alignItems: 'center', gap: '1.5rem' },
        notificationIcon: {
            width: '48px', height: '48px', borderRadius: '50%',
            background: 'rgba(21, 26, 38, 0.6)',
            backdropFilter: 'blur(10px)',
            border: '1px solid rgba(26, 210, 255, 0.35)',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            color: '#F8F9FA', cursor: 'pointer', transition: 'all 0.3s',
            boxShadow: '0 4px 20px rgba(0, 0, 0, 0.3)',
        },
        userProfile: {
            display: 'flex', alignItems: 'center', gap: '0.75rem',
            padding: '0.5rem 1.25rem',
            background: 'rgba(21, 26, 38, 0.6)',
            backdropFilter: 'blur(10px)',
            border: '1px solid rgba(88, 58, 255, 0.2)',
            borderRadius: '2rem', cursor: 'pointer',
            boxShadow: '0 4px 20px rgba(0, 0, 0, 0.3)',
            position: 'relative',
            transition: 'all 0.3s ease',
        },
        avatar: {
            width: '36px', height: '36px', borderRadius: '50%',
            background: 'linear-gradient(135deg, #583AFF 0%, #1AD2FF 50%, #80FFEA 100%)',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            color: 'white', fontWeight: 700, fontSize: '0.875rem',
            boxShadow: '0 0 20px rgba(88, 58, 255, 0.4)',
            overflow: 'hidden',
        },
        avatarImg: {
            width: '100%', height: '100%', objectFit: 'cover'
        },
        dropdownMenu: {
            position: 'absolute',
            top: '120%',
            right: 0,
            width: '220px',
            background: 'rgba(15, 20, 35, 0.95)',
            backdropFilter: 'blur(20px)',
            border: '1px solid rgba(88, 58, 255, 0.2)',
            borderRadius: '1rem',
            padding: '0.5rem',
            boxShadow: '0 8px 32px rgba(0, 0, 0, 0.5)',
            zIndex: 1000,
            overflow: 'hidden',
        },
        dropdownItem: {
            display: 'flex',
            alignItems: 'center',
            gap: '0.75rem',
            padding: '0.75rem 1rem',
            color: '#B8BDC7',
            fontSize: '0.9rem',
            fontWeight: 500,
            borderRadius: '0.5rem',
            cursor: 'pointer',
            transition: 'all 0.2s',
            textDecoration: 'none',
        },
        dropdownSeparator: {
            height: '1px',
            background: 'rgba(255, 255, 255, 0.08)',
            margin: '0.5rem 0',
        }
    }), []);

    const menuItems = useMemo(() => [
        { id: 'dashboard', icon: LayoutDashboard, label: 'Dashboard', route: '/dashboard' },
        { id: 'projects', icon: Server, label: 'Meus Projetos', route: '/projetos' },
        { id: 'downloads', icon: Download, label: 'Downloads', route: '/downloads' },
        { id: 'history', icon: History, label: 'Histórico', route: '/history' },
        { id: 'wallet', icon: DollarSign, label: 'Carteira', route: '/carteira' },
        { id: 'billing', icon: FileText, label: 'Faturas', route: '/faturas' },
        { id: 'shop', icon: ShoppingBag, label: 'Loja', route: '/loja' },
        { id: 'support', icon: Headphones, label: 'Suporte', route: '/suporte' },
        { id: 'settings', icon: Settings, label: 'Configurações', route: '/configuracoes' },
    ], []);

    // Check for admin access via roles or legacy is_admin
    const adminRoles = ['SUPPORT', 'ADMIN', 'DEVELOPMENT', 'ENGINEERING', 'DIRECTION'];
    const hasAdminAccess = user?.is_admin || user?.roles?.some(role => adminRoles.includes(role));
    const adminMenuItem = hasAdminAccess ? { id: 'admin', icon: Shield, label: 'Admin', route: '/admin' } : null;

    const handleLogout = () => {
        logout();
        navigate('/login');
    };

    // Close dropdown when clicking outside
    useEffect(() => {
        const handleClickOutside = (event) => {
            if (dropdownRef.current && !dropdownRef.current.contains(event.target)) {
                setIsDropdownOpen(false);
            }
            if (notificationRef.current && !notificationRef.current.contains(event.target)) {
                setIsNotificationOpen(false);
            }
        };

        const handleKeyDown = (event) => {
            if (event.key === 'Escape') {
                setIsDropdownOpen(false);
                setIsNotificationOpen(false);
            }
        };

        document.addEventListener('mousedown', handleClickOutside);
        document.addEventListener('keydown', handleKeyDown);
        return () => {
            document.removeEventListener('mousedown', handleClickOutside);
            document.removeEventListener('keydown', handleKeyDown);
        };
    }, []);



    return (
        <div style={styles.container}>
            <div style={styles.backgroundTexture} />

            {/* Mobile hamburger */}
            <button
                className="dashboard-hamburger"
                style={styles.hamburger}
                onClick={() => setIsMobileSidebarOpen(!isMobileSidebarOpen)}
                aria-label="Toggle menu"
            >
                {isMobileSidebarOpen ? <X size={22} /> : <Menu size={22} />}
            </button>

            {/* Mobile overlay */}
            {isMobileSidebarOpen && (
                <div style={styles.mobileOverlay} className="dashboard-mobile-overlay" onClick={() => setIsMobileSidebarOpen(false)} />
            )}

            {/* Sidebar */}
            <aside className="dashboard-sidebar" style={{ ...styles.sidebar, ...(isMobileSidebarOpen ? { transform: 'translateX(0)' } : {}) }}>
                <div style={styles.logo}>
                    <img src="/pixelcraft-logo.png" alt="Pixelcraft" style={styles.logoImage} />
                    <div style={styles.logoText}>Pixelcraft</div>
                </div>

                <nav style={{ flex: 1 }}>
                    {menuItems.map((item) => {
                        const isActive = location.pathname === item.route;
                        return (
                            <motion.button
                                key={item.id}
                                style={{
                                    ...styles.menuItem,
                                    ...(isActive ? styles.menuItemActive : {}),
                                }}
                                whileHover={{
                                    background: 'rgba(88, 58, 255, 0.12)',
                                    color: '#F8F9FA',
                                    boxShadow: 'inset 0 0 20px rgba(88, 58, 255, 0.2)',
                                }}
                                onClick={() => {
                                    navigate(item.route);
                                    setIsMobileSidebarOpen(false);
                                }}
                                aria-current={isActive ? 'page' : undefined}
                            >
                                <item.icon size={20} />
                                {item.label}
                            </motion.button>
                        );
                    })}
                    {adminMenuItem && (
                        <motion.button
                            key={adminMenuItem.id}
                            style={{
                                ...styles.menuItem,
                                ...(location.pathname.startsWith('/admin') ? styles.menuItemActive : {}),
                            }}
                            whileHover={{
                                background: 'rgba(88, 58, 255, 0.12)',
                                color: '#F8F9FA',
                                boxShadow: 'inset 0 0 20px rgba(88, 58, 255, 0.2)',
                            }}
                            onClick={() => navigate(adminMenuItem.route)}
                        >
                            <adminMenuItem.icon size={20} />
                            {adminMenuItem.label}
                        </motion.button>
                    )}
                </nav>

                <div style={{ borderTop: '1px solid rgba(248, 249, 250, 0.05)', paddingTop: '1rem' }}>
                    <motion.button
                        style={styles.menuItem}
                        whileHover={{ background: 'rgba(239, 68, 68, 0.15)', color: '#EF4444' }}
                        onClick={handleLogout}
                    >
                        <LogOut size={20} />
                        Sair
                    </motion.button>
                </div>
            </aside>

            {/* Main Content */}
            <main className="dashboard-main" style={styles.mainContent}>
                <header style={styles.header}>
                    <div style={styles.headerLeft}>
                        {headerStart}
                        {title && <h1 style={styles.pageTitle}>{title}</h1>}
                    </div>

                    <div style={{ ...styles.headerRight, gap: '1rem', position: 'relative' }}>
                        {/* Notification Bell */}
                        <div ref={notificationRef} style={{ position: 'relative' }}>
                            <motion.div
                                role="button"
                                tabIndex={0}
                                style={{ ...styles.notificationIcon, cursor: 'pointer', outline: 'none' }}
                                onClick={() => setIsNotificationOpen(!isNotificationOpen)}
                                onKeyDown={(e) => {
                                    if (e.key === 'Enter' || e.key === ' ') {
                                        e.preventDefault();
                                        setIsNotificationOpen(!isNotificationOpen);
                                    }
                                }}
                                whileHover={{
                                    background: 'rgba(26, 210, 255, 0.2)',
                                    borderColor: '#1AD2FF',
                                    color: '#1AD2FF',
                                    boxShadow: '0 0 20px rgba(26, 210, 255, 0.4)',
                                }}
                            >
                                <Bell size={20} />
                                {/* Optional: Red dot indicator */}
                                {/* <div style={{ position: 'absolute', top: '10px', right: '10px', width: '8px', height: '8px', background: '#EF4444', borderRadius: '50%' }} /> */}
                            </motion.div>

                            <AnimatePresence>
                                {isNotificationOpen && (
                                    <motion.div
                                        style={{
                                            ...styles.dropdownMenu,
                                            width: '320px',
                                            right: '-60px'
                                        }}
                                        initial={{ opacity: 0, y: 10, scale: 0.95 }}
                                        animate={{ opacity: 1, y: 0, scale: 1 }}
                                        exit={{ opacity: 0, y: 10, scale: 0.95 }}
                                        transition={{ duration: 0.2 }}
                                    >
                                        <div style={{ padding: 'var(--btn-padding-md)', borderBottom: '1px solid rgba(255,255,255,0.05)', fontWeight: 600, color: '#F8F9FA' }}>
                                            Notificações
                                        </div>
                                        <div style={{ padding: '2rem 1rem', textAlign: 'center', color: '#6C7384', fontSize: '0.9rem' }}>
                                            <Bell size={24} style={{ margin: '0 auto 0.5rem', opacity: 0.5 }} />
                                            Nenhuma notificação no momento.
                                        </div>
                                    </motion.div>
                                )}
                            </AnimatePresence>
                        </div>

                        {/* Profile Dropdown */}
                        <div
                            ref={dropdownRef}
                            role="button"
                            tabIndex={0}
                            style={{ ...styles.userProfile, outline: 'none' }}
                            onClick={() => setIsDropdownOpen(!isDropdownOpen)}
                            onKeyDown={(e) => {
                                if (e.key === 'Enter' || e.key === ' ') {
                                    e.preventDefault();
                                    setIsDropdownOpen(!isDropdownOpen);
                                }
                            }}
                        >
                            <div style={styles.avatar}>
                                {user?.avatar_url ? (
                                    <img
                                        src={getAvatarUrl(user.avatar_url)}
                                        alt="Profile"
                                        style={styles.avatarImg}
                                    />
                                ) : (
                                    user?.full_name?.charAt(0).toUpperCase() || 'U'
                                )}
                            </div>
                            <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-start' }}>
                                <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                                    <span style={{ fontSize: '0.875rem', fontWeight: 600, color: '#F8F9FA' }}>
                                        {user?.full_name || 'Usuário'}
                                    </span>
                                    {user?.highest_role && <RoleBadge role={user.highest_role} size="small" />}
                                </div>
                                <div style={{ fontSize: '0.75rem', color: '#B8BDC7' }}>
                                    {user?.referral_code || ''}
                                </div>
                            </div>
                            <ChevronDown size={14} color="#B8BDC7" style={{ marginLeft: '4px' }} />

                            <AnimatePresence>
                                {isDropdownOpen && (
                                    <motion.div
                                        style={styles.dropdownMenu}
                                        initial={{ opacity: 0, y: 10, scale: 0.95 }}
                                        animate={{ opacity: 1, y: 0, scale: 1 }}
                                        exit={{ opacity: 0, y: 10, scale: 0.95 }}
                                        transition={{ duration: 0.2 }}
                                    >
                                        <div
                                            style={styles.dropdownItem}
                                            onClick={(e) => { e.stopPropagation(); navigate('/configuracoes'); setIsDropdownOpen(false); }}
                                            onMouseEnter={(e) => e.currentTarget.style.background = 'rgba(255, 255, 255, 0.05)'}
                                            onMouseLeave={(e) => e.currentTarget.style.background = 'transparent'}
                                        >
                                            <Settings size={16} /> Configurações
                                        </div>

                                        <div
                                            style={styles.dropdownItem}
                                            onClick={(e) => { e.stopPropagation(); navigate('/carteira'); setIsDropdownOpen(false); }}
                                            onMouseEnter={(e) => e.currentTarget.style.background = 'rgba(255, 255, 255, 0.05)'}
                                            onMouseLeave={(e) => e.currentTarget.style.background = 'transparent'}
                                        >
                                            <WalletIcon size={16} /> Carteira
                                        </div>

                                        <div style={styles.dropdownSeparator} />

                                        <div
                                            style={{ ...styles.dropdownItem, color: '#EF4444' }}
                                            onClick={(e) => { e.stopPropagation(); handleLogout(); }}
                                            onMouseEnter={(e) => e.currentTarget.style.background = 'rgba(239, 68, 68, 0.1)'}
                                            onMouseLeave={(e) => e.currentTarget.style.background = 'transparent'}
                                        >
                                            <LogOut size={16} /> Sair
                                        </div>
                                    </motion.div>
                                )}
                            </AnimatePresence>
                        </div>
                    </div>
                </header>

                {children}
            </main>

            {/* Native App Bottom Navigation (Visible on Mobile) */}
            <BottomNavigation />
        </div>
    );
};

// Also export WalletIcon alias if needed by consumers, but usually not.
// To handle the icon naming conflict in the file (Wallet vs WalletIcon), I imported Wallet from lucide.
// BUT in my dropdown logic I used WalletIcon.
// Let's fix imports:
// DollarSign from lucide is not "Wallet". "Wallet" is standard.
// My imports: DollarSign, CreditCard.
// The "Wallet" icon in lucide is called "Wallet".
// But my file also needs "Wallet" from Lucide for the menu icon?
// Ah, the icons I imported at top: DollarSign (for Carteira sidebar item).
// The user asked for "Carteira" in dropdown. I can use Wallet icon or DollarSign.
// I'll add "Wallet as WalletIcon" to imports.

export default DashboardLayout;

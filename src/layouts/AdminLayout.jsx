import React from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { motion } from 'framer-motion';
import {
    LayoutDashboard,
    ShoppingBag,
    LogOut,
    User,
    Shield,
    Package,
    MessageSquare,
    FileText,
    UserCog,
    Server,
    Ticket,
    Mail
} from 'lucide-react';
import { useAuth } from '../context/AuthContext';

const AdminLayout = () => {
    const { logout, user } = useAuth();
    const navigate = useNavigate();
    const location = useLocation();

    const handleLogout = () => {
        logout();
        navigate('/login');
    };

    const allMenuItems = [
        { id: 'admin-dashboard', icon: LayoutDashboard, label: 'Admin Dashboard', route: '/admin' },
        { id: 'users', icon: User, label: 'Usuários', route: '/admin/users' },
        { id: 'roles', icon: UserCog, label: 'Cargos', route: '/admin/roles' },
        { id: 'orders', icon: ShoppingBag, label: 'Pedidos', route: '/admin/orders' },
        { id: 'discounts', icon: Ticket, label: 'Descontos', route: '/admin/discounts' },
        { id: 'catalog', icon: Package, label: 'Catálogo', route: '/admin/catalog' },
        { id: 'files', icon: FileText, label: 'Arquivos', route: '/admin/catalog/files' },
        { id: 'tickets', icon: MessageSquare, label: 'Tickets', route: '/admin/support' },
        { id: 'email-settings', icon: Mail, label: 'E-mail', route: '/admin/email' },
        { id: 'system', icon: Server, label: 'Recursos do Sistema', route: '/admin/system' },
        { id: 'client-dashboard', icon: LayoutDashboard, label: 'Voltar ao Dashboard', route: '/dashboard' },
    ];

    // Filter menu items based on role
    const menuItems = React.useMemo(() => {
        if (!user) return [];

        // Legacy admin bypass
        if (user.is_admin) return allMenuItems;

        const role = user.highest_role;

        // Support role only sees Tickets
        if (role === 'SUPPORT') {
            return allMenuItems.filter(item => ['tickets', 'client-dashboard'].includes(item.id));
        }

        // ADMIN role cannot see system resources or email settings
        if (role === 'ADMIN') {
            return allMenuItems.filter(item => !['system', 'email-settings'].includes(item.id));
        }

        // Other admin roles see everything
        return allMenuItems;
    }, [user]);

    const styles = React.useMemo(() => ({
        container: {
            display: 'flex',
            minHeight: '100vh',
            background: 'linear-gradient(180deg, #0A0E1A 0%, #151924 50%, #0A0E1A 100%)',
            position: 'relative',
            overflow: 'hidden',
        },
        backgroundTexture: {
            position: 'absolute',
            top: 0, left: 0, right: 0, bottom: 0,
            background: 'radial-gradient(circle at 20% 30%, rgba(224, 26, 79, 0.05) 0%, transparent 50%), radial-gradient(circle at 80% 70%, rgba(255, 107, 53, 0.05) 0%, transparent 50%)',
            pointerEvents: 'none',
        },
        sidebar: {
            width: '260px',
            background: 'rgba(10, 14, 26, 0.8)',
            backdropFilter: 'blur(20px)',
            borderRight: '1px solid rgba(255, 255, 255, 0.1)',
            padding: '2rem 0',
            display: 'flex',
            flexDirection: 'column',
            position: 'fixed',
            height: '100vh',
            zIndex: 100,
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
            background: 'rgba(224, 26, 79, 0.15)',
            borderLeft: '3px solid #E01A4F',
            boxShadow: '0 0 20px rgba(224, 26, 79, 0.25)'
        },
        mainContent: { marginLeft: '260px', flex: 1, padding: '2.5rem 3rem', position: 'relative', zIndex: 1 },
        userProfile: {
            display: 'flex', alignItems: 'center', gap: '0.75rem', padding: '0.5rem 1.25rem',
            background: 'rgba(21, 26, 38, 0.6)', backdropFilter: 'blur(10px)', border: '1px solid rgba(255, 255, 255, 0.1)',
            borderRadius: '2rem', cursor: 'pointer', boxShadow: '0 4px 20px rgba(0, 0, 0, 0.3)',
            marginTop: 'auto', marginBottom: '1rem', margin: 'auto 1.5rem 1rem 1.5rem'
        },
        avatar: {
            width: '36px', height: '36px', borderRadius: '50%',
            background: 'var(--gradient-cta)',
            display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'white', fontWeight: 700, fontSize: '0.875rem',
            boxShadow: '0 0 20px rgba(224, 26, 79, 0.4)',
        },
    }), []);

    return (
        <div style={styles.container}>
            <div style={styles.backgroundTexture} />

            <aside style={styles.sidebar}>
                <div style={styles.logo}>
                    <img src="/pixelcraft-logo.png" alt="Pixelcraft" style={styles.logoImage} />
                    <div style={styles.logoText}>Pixelcraft <span style={{ fontSize: '0.8rem', color: '#E01A4F', marginLeft: '5px' }}>Admin</span></div>
                </div>

                <nav style={{ flex: 1, display: 'flex', flexDirection: 'column' }}>
                    {menuItems.map((item) => {
                        const isActive = location.pathname === item.route || (item.route !== '/admin' && location.pathname.startsWith(item.route));
                        return (
                            <motion.button
                                key={item.id}
                                style={{
                                    ...styles.menuItem,
                                    ...(isActive ? styles.menuItemActive : {}),
                                }}
                                whileHover={{
                                    background: 'rgba(224, 26, 79, 0.12)',
                                    color: '#F8F9FA',
                                    boxShadow: 'inset 0 0 20px rgba(224, 26, 79, 0.2)',
                                }}
                                onClick={() => navigate(item.route)}
                                aria-current={isActive ? 'page' : undefined}
                            >
                                <item.icon size={20} />
                                {item.label}
                            </motion.button>
                        );
                    })}
                </nav>

                <div style={{ borderTop: '1px solid rgba(248, 249, 250, 0.05)', paddingTop: '1rem' }}>
                    <div style={styles.userProfile}>
                        <div style={styles.avatar}>A</div>
                        <div>
                            <div style={{ fontSize: '0.875rem', fontWeight: 600, color: '#F8F9FA' }}>
                                {user?.full_name || 'Admin'}
                            </div>
                            <div style={{ fontSize: '0.75rem', color: '#B8BDC7' }}>
                                Administrador
                            </div>
                        </div>
                    </div>

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

            <main style={styles.mainContent}>
                <Outlet />
            </main>
        </div>
    );
};

export default AdminLayout;

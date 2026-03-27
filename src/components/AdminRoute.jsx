import { Navigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { hasAdminAccess } from './RoleBadge';

/**
 * AdminRoute - Protected route that requires user to be authenticated AND have admin access
 * Checks both legacy is_admin and new role-based system
 * Redirects non-admins to dashboard, unauthenticated users to login
 */
function AdminRoute({ children }) {
    const { user, loading } = useAuth();

    if (loading) {
        return (
            <div style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                minHeight: '100vh',
                background: 'var(--bg-primary)',
                color: 'var(--text-primary)',
            }}>
                <div style={{ textAlign: 'center' }}>
                    <div style={{
                        fontSize: 'var(--title-h3)',
                        marginBottom: '1rem',
                        background: 'var(--gradient-primary)',
                        WebkitBackgroundClip: 'text',
                        backgroundClip: 'text',
                        WebkitTextFillColor: 'transparent',
                    }}>
                        Carregando...
                    </div>
                </div>
            </div>
        );
    }

    // Not authenticated - redirect to login
    if (!user) {
        return <Navigate to="/login" replace />;
    }

    // Check for admin access via roles or legacy is_admin
    const canAccessAdmin = user.is_admin || hasAdminAccess(user.roles);

    if (!canAccessAdmin) {
        return <Navigate to="/dashboard" replace />;
    }

    return children;
}

export default AdminRoute;

import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { adminAPI } from '../../services/api'; // Ensure this matches your export
import { getAvatarUrl } from '../../utils/formatAvatarUrl';
import RoleBadge from '../../components/RoleBadge';
import { usePermissions } from '../../hooks/usePermissions';
import { maskCpf } from '../../utils/maskSensitiveData';
import {
    Users,
    Search,
    Edit,
    MoreHorizontal,
    ChevronLeft,
    ChevronRight,
    User,
    Shield
} from 'lucide-react';

const UsersPage = () => {
    const navigate = useNavigate();
    const { hasPermission } = usePermissions();
    const canViewCpf = hasPermission('view_cpf');
    const [users, setUsers] = useState([]);
    const [loading, setLoading] = useState(true);
    const [page, setPage] = useState(1);
    const [limit] = useState(20);
    const [total, setTotal] = useState(0);
    const [search, setSearch] = useState('');
    const [debouncedSearch, setDebouncedSearch] = useState('');

    // Debounce search
    useEffect(() => {
        const timer = setTimeout(() => {
            setDebouncedSearch(search);
            setPage(1); // Reset to page 1 on search change
        }, 500);
        return () => clearTimeout(timer);
    }, [search]);

    const fetchUsers = useCallback(async () => {
        try {
            setLoading(true);
            const response = await adminAPI.getUsers(page, limit, debouncedSearch);
            // Response format: { data: [], total: N, page: N, limit: N }
            setUsers(response.data || []);
            setTotal(response.total || 0);
        } catch (error) {
            console.error("Failed to fetch users:", error);
        } finally {
            setLoading(false);
        }
    }, [page, limit, debouncedSearch]);

    useEffect(() => {
        fetchUsers();
    }, [fetchUsers]);

    const handleRowClick = (id) => {
        navigate(`/admin/users/${id}`);
    };

    const totalPages = Math.max(1, Math.ceil(total / limit));

    const styles = {
        container: {
            padding: '2rem',
            color: '#F8F9FA',
            maxWidth: '1600px',
            margin: '0 auto',
        },
        header: {
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: '2rem',
        },
        title: {
            fontSize: '1.8rem',
            fontWeight: 700,
            display: 'flex',
            alignItems: 'center',
            gap: '0.75rem',
        },
        searchContainer: {
            position: 'relative',
            width: '300px',
        },
        searchInput: {
            width: '100%',
            background: 'rgba(255, 255, 255, 0.05)',
            border: '1px solid rgba(255, 255, 255, 0.1)',
            borderRadius: '0.5rem',
            padding: '0.75rem 1rem 0.75rem 2.5rem',
            color: '#fff',
            outline: 'none',
        },
        searchIcon: {
            position: 'absolute',
            left: '0.75rem',
            top: '50%',
            transform: 'translateY(-50%)',
            color: 'rgba(255, 255, 255, 0.4)',
        },
        tableContainer: {
            background: 'rgba(21, 26, 38, 0.6)',
            backdropFilter: 'blur(10px)',
            border: '1px solid rgba(255, 255, 255, 0.1)',
            borderRadius: '1rem',
            overflow: 'hidden',
        },
        table: {
            width: '100%',
            borderCollapse: 'collapse',
        },
        th: {
            textAlign: 'left',
            padding: '1rem 1.5rem',
            borderBottom: '1px solid rgba(255, 255, 255, 0.1)',
            color: 'rgba(255, 255, 255, 0.4)',
            fontSize: '0.85rem',
            textTransform: 'uppercase',
            letterSpacing: '0.05em',
        },
        td: {
            padding: '1rem 1.5rem',
            borderBottom: '1px solid rgba(255, 255, 255, 0.05)',
            color: '#F8F9FA',
            verticalAlign: 'middle',
        },
        tr: {
            cursor: 'pointer',
            transition: 'background 0.2s',
        },
        trHover: {
            background: 'rgba(255, 255, 255, 0.02)',
        },
        userInfo: {
            display: 'flex',
            alignItems: 'center',
            gap: '1rem',
        },
        avatar: {
            width: '32px',
            height: '32px',
            borderRadius: '50%',
            background: '#2D3748',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: '#E01A4F',
            fontSize: '0.9rem',
            fontWeight: 'bold',
            overflow: 'hidden',
        },
        avatarImg: {
            width: '100%',
            height: '100%',
            objectFit: 'cover',
        },
        badge: (type) => ({
            padding: '0.25rem 0.75rem',
            borderRadius: '9999px',
            fontSize: '0.75rem',
            fontWeight: 500,
            background: type === 'balance' ? 'rgba(34, 197, 94, 0.1)' : 'rgba(255, 255, 255, 0.1)',
            color: type === 'balance' ? '#22C55E' : '#B8BDC7',
        }),
        pagination: {
            display: 'flex',
            justifyContent: 'flex-end',
            alignItems: 'center',
            padding: '1.5rem',
            gap: '1rem',
        },
        pageButton: (disabled) => ({
            background: disabled ? 'transparent' : 'rgba(255, 255, 255, 0.05)',
            border: '1px solid rgba(255, 255, 255, 0.1)',
            borderRadius: '0.5rem',
            width: '32px',
            height: '32px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: disabled ? 'rgba(255, 255, 255, 0.2)' : '#fff',
            cursor: disabled ? 'not-allowed' : 'pointer',
        }),
    };



    const getInitials = (name) => {
        if (!name) return 'U';
        return name.split(' ').map(n => n[0]).join('').substring(0, 2).toUpperCase();
    };

    return (
        <div style={styles.container}>
            <div style={styles.header}>
                <h1 style={styles.title}>
                    <Users size={32} color="#E01A4F" />
                    Gerenciar Usuários
                </h1>
                <div style={styles.searchContainer}>
                    <Search size={16} style={styles.searchIcon} />
                    <input
                        type="text"
                        placeholder="Buscar por nome, email ou username..."
                        style={styles.searchInput}
                        value={search}
                        onChange={(e) => setSearch(e.target.value)}
                    />
                </div>
            </div>

            <div style={styles.tableContainer}>
                {loading ? (
                    <div style={{ padding: '3rem', textAlign: 'center', color: 'rgba(255,255,255,0.4)' }}>
                        Carregando usuários...
                    </div>
                ) : (
                    <>
                        <table style={styles.table} className="mobile-stacked-table">
                            <thead>
                                <tr>
                                    <th style={styles.th}>Usuário</th>
                                    <th style={styles.th}>Cargo</th>
                                    <th style={styles.th}>Email</th>
                                    {canViewCpf && <th style={styles.th}>CPF</th>}
                                    <th style={styles.th}>Saldo</th>
                                    <th style={styles.th}>Cadastro</th>
                                    <th style={styles.th}></th>
                                </tr>
                            </thead>
                            <tbody>
                                {users.length > 0 ? (
                                    users.map((user) => (
                                        <motion.tr
                                            key={user.id}
                                            style={styles.tr}
                                            whileHover={{ backgroundColor: 'rgba(255, 255, 255, 0.03)' }}
                                            onClick={() => handleRowClick(user.id)}
                                        >
                                            <td style={styles.td} data-label="Usuário">
                                                <div style={styles.userInfo} className="userInfo">
                                                    <div style={styles.avatar}>
                                                        {user.avatar_url ? (
                                                            <img
                                                                src={getAvatarUrl(user.avatar_url)}
                                                                alt={user.full_name}
                                                                style={styles.avatarImg}
                                                                onError={(e) => { e.target.style.display = 'none'; e.target.parentElement.innerText = getInitials(user.full_name); }}
                                                            />
                                                        ) : (
                                                            getInitials(user.full_name)
                                                        )}
                                                    </div>
                                                    <div>
                                                        <span
                                                            style={{ fontWeight: 500, color: '#F8F9FA' }}
                                                        >
                                                            {user.full_name || 'Sem nome'}
                                                        </span>
                                                        <div style={{ fontSize: '0.75rem', color: 'rgba(255,255,255,0.5)', marginTop: '2px' }}>
                                                            @{user.username || 'N/A'}
                                                        </div>
                                                    </div>
                                                </div>
                                            </td>
                                            <td style={styles.td} data-label="Cargo">
                                                {user.highest_role ? (
                                                    <RoleBadge role={user.highest_role} size="small" />
                                                ) : user.is_admin ? (
                                                    <span style={{ display: 'inline-flex', alignItems: 'center', gap: '4px', padding: '2px 8px', borderRadius: '4px', background: 'rgba(255, 63, 0, 0.15)', border: '1px solid rgba(255, 63, 0, 0.3)', color: '#ff3f00', fontSize: '0.75rem' }}>
                                                        <Shield size={12} /> Admin
                                                    </span>
                                                ) : (
                                                    <span style={{ color: '#666', fontSize: '0.75rem' }}>Sem cargo</span>
                                                )}
                                            </td>
                                            <td style={styles.td} data-label="Email">{user.email}</td>
                                            {canViewCpf && (
                                                <td style={{ ...styles.td, color: 'rgba(255,255,255,0.6)', fontFamily: 'monospace' }} data-label="CPF">
                                                    {user.cpf ? maskCpf(user.cpf, true) : 'Não consta'}
                                                </td>
                                            )}
                                            <td style={styles.td} data-label="Saldo">
                                                <span style={styles.badge('balance')}>
                                                    {new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(user.balance || 0)}
                                                </span>
                                            </td>
                                            <td style={styles.td} data-label="Data Cadastro">
                                                {new Date(user.created_at).toLocaleDateString('pt-BR')}
                                            </td>
                                            <td style={styles.td} data-label="Opções">
                                                <div style={{ display: 'flex', justifyContent: 'flex-end', paddingRight: '1rem' }}>
                                                    <Edit size={16} color="rgba(255,255,255,0.4)" />
                                                </div>
                                            </td>
                                        </motion.tr>
                                    ))
                                ) : (
                                    <tr>
                                        <td colSpan="6" style={{ ...styles.td, textAlign: 'center', padding: '3rem' }}>
                                            Nenhum usuário encontrado.
                                        </td>
                                    </tr>
                                )}
                            </tbody>
                        </table>

                        <div style={styles.pagination}>
                            <div style={{ color: 'rgba(255,255,255,0.4)', fontSize: '0.9rem' }}>
                                Página {page} de {totalPages} • Total: {total} usuários
                            </div>
                            <div style={{ display: 'flex', gap: '0.5rem' }}>
                                <motion.button
                                    style={styles.pageButton(page === 1)}
                                    disabled={page === 1}
                                    onClick={() => setPage(p => Math.max(1, p - 1))}
                                    whileHover={page !== 1 ? { scale: 1.05 } : {}}
                                    whileTap={page !== 1 ? { scale: 0.95 } : {}}
                                >
                                    <ChevronLeft size={16} />
                                </motion.button>
                                <motion.button
                                    style={styles.pageButton(page === totalPages)}
                                    disabled={page === totalPages}
                                    onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                                    whileHover={page !== totalPages ? { scale: 1.05 } : {}}
                                    whileTap={page !== totalPages ? { scale: 0.95 } : {}}
                                >
                                    <ChevronRight size={16} />
                                </motion.button>
                            </div>
                        </div>
                    </>
                )}
            </div>
        </div>
    );
};

export default UsersPage;

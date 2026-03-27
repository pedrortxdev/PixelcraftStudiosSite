import React, { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { adminAPI } from '../../services/api';
import { getAvatarUrl } from '../../utils/formatAvatarUrl';
import RoleBadge, { getAllRolesConfig } from '../../components/RoleBadge';
import {
    ArrowLeft,
    User,
    Mail,
    Phone,
    Globe,
    CreditCard,
    ShoppingBag,
    Clock,
    Shield,
    Save,
    Lock,
    AlertTriangle,
    CheckCircle,
    X,
    Package,
    Sparkles,
    Wand2,
    Loader2,
    Plus,
    Trash2,
    BadgeCheck
} from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import { usePermissions } from '../../hooks/usePermissions';
import { maskCpf } from '../../utils/maskSensitiveData';

const UserDetailPage = () => {
    const { id } = useParams();
    const navigate = useNavigate();
    const { generateAIAvatar, refreshUser, user: currentUser } = useAuth();
    const { hasPermission } = usePermissions();
    const canViewCpf = hasPermission('view_cpf');

    const [user, setUser] = useState(null);
    const [transactions, setTransactions] = useState([]);
    const [subscriptions, setSubscriptions] = useState([]);
    const [library, setLibrary] = useState([]);
    const [loading, setLoading] = useState(true);
    const [submitting, setSubmitting] = useState(false);
    const [error, setError] = useState(null);
    const [successMsg, setSuccessMsg] = useState('');
    const [generatingAI, setGeneratingAI] = useState(false);
    const [userRoles, setUserRoles] = useState([]);
    const [addingRole, setAddingRole] = useState(false);
    const [selectedNewRole, setSelectedNewRole] = useState('');

    const allRoles = getAllRolesConfig();

    // Edit Form State
    const [formData, setFormData] = useState({
        full_name: '',
        email: '',
        username: '',
        discord_handle: '',
        whatsapp_phone: '',
        cpf: '',
        is_admin: false,
        balance: 0,
        adjustment_type: 'Pix Direto'
    });

    // Password Reset State
    const [newPassword, setNewPassword] = useState('');

    const fetchDetails = useCallback(async () => {
        try {
            setLoading(true);
            const data = await adminAPI.getUserDetails(id);
            // data structure: { user, balance, transactions, subscriptions, library }
            setUser(data.user);
            setTransactions(data.transactions || []);
            setSubscriptions(data.subscriptions || []);
            setLibrary(data.library || []);

            setFormData({
                full_name: data.user.full_name || '',
                email: data.user.email || '',
                username: data.user.username || '',
                discord_handle: data.user.discord_handle || '',
                whatsapp_phone: data.user.whatsapp_phone || '',
                cpf: '',
                is_admin: data.user.is_admin || false,
                balance: data.user.balance || 0,
                adjustment_type: 'Pix Direto'
            });
            setUserRoles(data.user.roles || []);
        } catch (err) {
            console.error('Error fetching details:', err);
            setError('Falha ao carregar detalhes do usuário.');
        } finally {
            setLoading(false);
        }
    }, [id]);

    useEffect(() => {
        fetchDetails();
    }, [fetchDetails]);

    const handleUpdateProfile = async (e) => {
        e.preventDefault();
        setSubmitting(true);
        setError(null);
        setSuccessMsg('');

        try {
            // Filter out empty strings if needed, or send as is
            const payload = { ...formData };
            if (!payload.cpf) delete payload.cpf; // Only send CPF if updated

            await adminAPI.updateUser(id, payload);
            setSuccessMsg('Perfil atualizado com sucesso!');
            fetchDetails(); // Refresh
        } catch (err) {
            console.error('Update error:', err);
            setError('Falha ao atualizar perfil.');
        } finally {
            setSubmitting(false);
        }
    };

    const handleResetPassword = async (e) => {
        e.preventDefault();
        if (!newPassword || newPassword.length < 6) {
            setError('A senha deve ter pelo menos 6 caracteres.');
            return;
        }

        setSubmitting(true);
        setError(null);
        setSuccessMsg('');

        try {
            await adminAPI.updateUserPassword(id, newPassword);
            setSuccessMsg('Senha alterada com sucesso!');
            setNewPassword('');
        } catch (err) {
            console.error('Password reset error:', err);
            setError('Falha ao alterar senha.');
        } finally {
            setSubmitting(false);
        }
    };

    const handleGenerateAI = async () => {
        const prompt = window.prompt('Descreva como você quer o novo avatar para este usuário (ex: "Mago pixelizado"):');
        if (!prompt) return;

        try {
            setGeneratingAI(true);
            setError(null);
            setSuccessMsg('');
            await generateAIAvatar(prompt, id);
            setSuccessMsg('Avatar gerado com IA e aplicado com sucesso!');
            fetchDetails(); // Refresh view
        } catch (err) {
            console.error('AI generation error:', err);
            setError('Falha ao gerar avatar com IA.');
        } finally {
            setGeneratingAI(false);
        }
    };



    const getInitials = (name) => {
        if (!name) return 'U';
        return name.split(' ').map(n => n[0]).join('').substring(0, 2).toUpperCase();
    };

    const styles = {
        container: {
            padding: '2rem',
            color: '#F8F9FA',
            maxWidth: '1600px',
            margin: '0 auto',
        },
        header: {
            display: 'flex',
            alignItems: 'center',
            gap: '1rem',
            marginBottom: '2rem',
        },
        avatar: {
            width: '64px',
            height: '64px',
            borderRadius: '50%',
            background: '#2D3748',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: '#E01A4F',
            fontSize: 'var(--title-h4)',
            fontWeight: 'bold',
            overflow: 'hidden',
        },
        avatarImg: {
            width: '100%',
            height: '100%',
            objectFit: 'cover',
        },
        backButton: {
            background: 'rgba(255, 255, 255, 0.05)',
            border: '1px solid rgba(255, 255, 255, 0.1)',
            borderRadius: '0.5rem',
            padding: '0.5rem',
            color: '#B8BDC7',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
        },
        title: {
            fontSize: 'var(--title-h4)',
            fontWeight: 700,
        },
        grid: {
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(350px, 1fr))',
            gap: '2rem',
            alignItems: 'start',
        },
        card: {
            background: 'rgba(21, 26, 38, 0.6)',
            backdropFilter: 'blur(10px)',
            border: '1px solid rgba(255, 255, 255, 0.1)',
            borderRadius: '1rem',
            padding: '1.5rem',
            display: 'flex',
            flexDirection: 'column',
            gap: '1.5rem',
        },
        sectionTitle: {
            fontSize: '1.1rem',
            fontWeight: 600,
            color: '#E01A4F',
            display: 'flex',
            alignItems: 'center',
            gap: '0.5rem',
            marginBottom: '0.5rem',
            borderBottom: '1px solid rgba(255,255,255,0.1)',
            paddingBottom: '0.75rem',
        },
        formGroup: {
            display: 'flex',
            flexDirection: 'column',
            gap: '0.5rem',
        },
        label: {
            fontSize: '0.85rem',
            color: '#B8BDC7',
        },
        input: {
            background: 'rgba(0, 0, 0, 0.2)',
            border: '1px solid rgba(255, 255, 255, 0.1)',
            borderRadius: '0.5rem',
            padding: '0.75rem',
            color: '#F8F9FA',
            outline: 'none',
            fontFamily: 'inherit',
        },
        button: {
            background: 'var(--gradient-cta)',
            border: 'none',
            borderRadius: '0.5rem',
            padding: '0.75rem',
            color: 'white',
            fontWeight: 600,
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: '0.5rem',
            marginTop: '0.5rem',
        },
        secondaryButton: {
            background: 'rgba(255, 255, 255, 0.05)',
            border: '1px solid rgba(255, 255, 255, 0.1)',
            // ... similar to button
            padding: '0.75rem',
            borderRadius: '0.5rem',
            color: '#fff',
            cursor: 'pointer',
            textAlign: 'center'
        },
        statRow: {
            display: 'flex',
            justifyContent: 'space-between',
            padding: '0.75rem 0',
            borderBottom: '1px solid rgba(255,255,255,0.05)',
        },
        table: { width: '100%', borderCollapse: 'collapse', fontSize: '0.9rem' },
        th: { textAlign: 'left', padding: '0.75rem', color: 'rgba(255,255,255,0.4)', borderBottom: '1px solid rgba(255,255,255,0.1)' },
        td: { padding: '0.75rem', borderBottom: '1px solid rgba(255,255,255,0.05)' },
        badge: (status) => ({
            padding: '0.2rem 0.6rem',
            borderRadius: '4px',
            fontSize: '0.75rem',
            background: status === 'completed' || status === 'ACTIVE' ? 'rgba(34, 197, 94, 0.1)' : 'rgba(255, 255, 255, 0.1)',
            color: status === 'completed' || status === 'ACTIVE' ? '#22C55E' : '#B8BDC7',
        }),
    };

    if (loading) return <div style={{ padding: '4rem', textAlign: 'center', color: '#B8BDC7' }}>Carregando detalhes do usuário...</div>;
    if (!user) return <div style={{ padding: '4rem', textAlign: 'center', color: '#EF4444' }}>Usuário não encontrado.</div>;

    return (
        <div style={styles.container}>
            <div style={styles.header}>
                <motion.button
                    style={styles.backButton}
                    whileHover={{ scale: 1.05 }}
                    whileTap={{ scale: 0.95 }}
                    onClick={() => navigate('/admin/users')}
                >
                    <ArrowLeft size={20} />
                </motion.button>
                <h1 style={styles.title}>Detalhes do Usuário</h1>
            </div>

            {error && (
                <div style={{ padding: 'var(--btn-padding-md)', background: 'rgba(239, 68, 68, 0.1)', color: '#EF4444', borderRadius: '0.5rem', marginBottom: '1rem', display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                    <AlertTriangle size={20} />
                    {error}
                </div>
            )}

            {successMsg && (
                <div style={{ padding: 'var(--btn-padding-md)', background: 'rgba(34, 197, 94, 0.1)', color: '#22C55E', borderRadius: '0.5rem', marginBottom: '1rem', display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                    <CheckCircle size={20} />
                    {successMsg}
                </div>
            )}

            <div style={styles.grid}>
                {/* LEFT COLUMN: Profile & Security */}
                <div style={{ display: 'flex', flexDirection: 'column', gap: '2rem' }}>

                    {/* EDIT PROFILE */}
                    <motion.div style={styles.card} initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', marginBottom: '0.5rem' }}>
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
                            <h2 style={{ ...styles.sectionTitle, borderBottom: 'none', marginBottom: 0, paddingBottom: 0 }}><User size={18} /> Dados Pessoais</h2>
                        </div>

                        <motion.button
                            type="button"
                            onClick={handleGenerateAI}
                            disabled={generatingAI}
                            style={{
                                width: '100%',
                                padding: '0.6rem',
                                background: 'rgba(88, 58, 255, 0.1)',
                                border: '1px dashed rgba(88, 58, 255, 0.3)',
                                borderRadius: '0.5rem',
                                color: '#1AD2FF',
                                fontWeight: 600,
                                fontSize: '0.85rem',
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                                gap: '0.5rem',
                                cursor: generatingAI ? 'not-allowed' : 'pointer',
                                marginBottom: '0.5rem'
                            }}
                            whileHover={!generatingAI ? { background: 'rgba(88, 58, 255, 0.2)' } : {}}
                        >
                            {generatingAI ? (
                                <Loader2 size={16} className="animate-spin" />
                            ) : (
                                <Sparkles size={16} />
                            )}
                            {generatingAI ? 'Gerando...' : 'Gerar Foto com IA'}
                        </motion.button>
                        <form onSubmit={handleUpdateProfile} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                            <div style={styles.formGroup}>
                                <label style={styles.label}>Nome Completo</label>
                                <input
                                    type="text"
                                    style={styles.input}
                                    value={formData.full_name}
                                    onChange={e => setFormData({ ...formData, full_name: e.target.value })}
                                />
                            </div>
                            <div style={styles.formGroup}>
                                <label style={styles.label}>Email</label>
                                <input
                                    type="email"
                                    style={styles.input}
                                    value={formData.email}
                                    onChange={e => setFormData({ ...formData, email: e.target.value })}
                                />
                            </div>
                            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                                <div style={styles.formGroup}>
                                    <label style={styles.label}>Username</label>
                                    <input
                                        type="text"
                                        style={styles.input}
                                        value={formData.username}
                                        onChange={e => setFormData({ ...formData, username: e.target.value })}
                                    />
                                </div>
                                <div style={styles.formGroup}>
                                    <label style={styles.label}>WhatsApp</label>
                                    <input
                                        type="text"
                                        style={styles.input}
                                        value={formData.whatsapp_phone}
                                        onChange={e => setFormData({ ...formData, whatsapp_phone: e.target.value })}
                                    />
                                </div>
                            </div>
                            <div style={styles.formGroup}>
                                <label style={styles.label}>Discord</label>
                                <input
                                    type="text"
                                    style={styles.input}
                                    value={formData.discord_handle}
                                    onChange={e => setFormData({ ...formData, discord_handle: e.target.value })}
                                />
                            </div>

                            {canViewCpf && (
                                <div style={styles.formGroup}>
                                    <label style={styles.label}>CPF (Deixe em branco para manter)</label>
                                    <input
                                        type="text"
                                        style={styles.input}
                                        value={formData.cpf}
                                        placeholder="Apenas números"
                                        onChange={e => setFormData({ ...formData, cpf: e.target.value })}
                                    />
                                </div>
                            )}

                            <div style={styles.formGroup}>
                                <label style={styles.label}>Saldo (R$) - <span style={{ color: '#E01A4F' }}>Ajuste Administrativo</span></label>
                                <div style={{ display: 'flex', gap: '1rem' }}>
                                    <input
                                        type="number"
                                        step="0.01"
                                        style={{ ...styles.input, flex: 1 }}
                                        value={formData.balance}
                                        onChange={e => setFormData({ ...formData, balance: parseFloat(e.target.value) })}
                                    />
                                    <select
                                        style={{ ...styles.input, width: '150px' }}
                                        value={formData.adjustment_type}
                                        onChange={e => setFormData({ ...formData, adjustment_type: e.target.value })}
                                    >
                                        <option value="Pix Direto">Pix Direto</option>
                                        <option value="Teste">Teste</option>
                                    </select>
                                </div>
                                <div style={{ fontSize: '0.75rem', color: 'rgba(255,255,255,0.4)' }}>
                                    Alterar o saldo gerará um log de transação. "Pix Direto" conta na receita/vendas, "Teste" é ignorado.
                                </div>
                            </div>

                            {/* ROLE MANAGEMENT SECTION */}
                            <div style={{ marginTop: '1rem', padding: 'var(--btn-padding-md)', background: 'rgba(88, 58, 255, 0.05)', borderRadius: '0.75rem', border: '1px solid rgba(88, 58, 255, 0.2)' }}>
                                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
                                    <label style={{ ...styles.label, fontWeight: 600, fontSize: '0.95rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                                        <BadgeCheck size={18} color="#583AFF" /> Cargos do Usuário
                                    </label>
                                </div>

                                {/* Current Roles */}
                                <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.5rem', marginBottom: '1rem' }}>
                                    {userRoles.length > 0 ? (
                                        userRoles.map(role => (
                                            <div key={role} style={{ display: 'inline-flex', alignItems: 'center', gap: '6px' }}>
                                                <RoleBadge role={role} size="normal" />
                                                <motion.button
                                                    type="button"
                                                    style={{ background: 'rgba(239, 68, 68, 0.1)', border: '1px solid rgba(239, 68, 68, 0.3)', borderRadius: '4px', padding: '2px 4px', cursor: 'pointer', display: 'flex', alignItems: 'center' }}
                                                    whileHover={{ background: 'rgba(239, 68, 68, 0.2)' }}
                                                    onClick={async () => {
                                                        try {
                                                            await adminAPI.removeUserRole(id, role);
                                                            setUserRoles(userRoles.filter(r => r !== role));
                                                            setSuccessMsg(`Cargo ${allRoles[role]?.label || role} removido!`);
                                                            // Refresh current user if they modified their own roles
                                                            if (id === currentUser?.id) {
                                                                await refreshUser();
                                                            }
                                                        } catch (err) {
                                                            setError('Falha ao remover cargo.');
                                                        }
                                                    }}
                                                >
                                                    <X size={12} color="#ef4444" />
                                                </motion.button>
                                            </div>
                                        ))
                                    ) : (
                                        <span style={{ color: '#666', fontSize: '0.85rem' }}>Sem cargos atribuídos</span>
                                    )}
                                </div>

                                {/* Add Role */}
                                <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                                    <select
                                        style={{ ...styles.input, flex: 1, fontSize: '0.9rem' }}
                                        value={selectedNewRole}
                                        onChange={(e) => setSelectedNewRole(e.target.value)}
                                    >
                                        <option value="">Selecionar cargo...</option>
                                        {Object.entries(allRoles).filter(([key]) => !userRoles.includes(key)).map(([key, config]) => (
                                            <option key={key} value={key}>{config.label}</option>
                                        ))}
                                    </select>
                                    <motion.button
                                        type="button"
                                        disabled={!selectedNewRole || addingRole}
                                        style={{ padding: '0.6rem 1rem', background: 'rgba(88, 58, 255, 0.2)', border: '1px solid rgba(88, 58, 255, 0.4)', borderRadius: '0.5rem', color: '#F8F9FA', cursor: selectedNewRole ? 'pointer' : 'not-allowed', display: 'flex', alignItems: 'center', gap: '4px', fontSize: '0.85rem' }}
                                        whileHover={selectedNewRole ? { background: 'rgba(88, 58, 255, 0.3)' } : {}}
                                        onClick={async () => {
                                            if (!selectedNewRole) return;
                                            setAddingRole(true);
                                            try {
                                                await adminAPI.addUserRole(id, selectedNewRole);
                                                setUserRoles([...userRoles, selectedNewRole]);
                                                setSuccessMsg(`Cargo ${allRoles[selectedNewRole]?.label} adicionado!`);
                                                setSelectedNewRole('');
                                                // Refresh current user if they modified their own roles
                                                if (id === currentUser?.id) {
                                                    await refreshUser();
                                                }
                                            } catch (err) {
                                                setError('Falha ao adicionar cargo.');
                                            } finally {
                                                setAddingRole(false);
                                            }
                                        }}
                                    >
                                        <Plus size={14} /> Adicionar
                                    </motion.button>
                                </div>

                                {/* Legacy admin checkbox */}
                                <div style={{ marginTop: '1rem', paddingTop: '1rem', borderTop: '1px solid rgba(255,255,255,0.1)' }}>
                                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                                        <input
                                            type="checkbox"
                                            id="is_admin"
                                            checked={formData.is_admin}
                                            onChange={e => setFormData({ ...formData, is_admin: e.target.checked })}
                                            style={{ width: '16px', height: '16px', accentColor: '#E01A4F', cursor: 'pointer' }}
                                        />
                                        <label htmlFor="is_admin" style={{ fontSize: '0.8rem', color: '#888', cursor: 'pointer' }}>
                                            Admin legado (is_admin) - {formData.is_admin ? 'Ativo' : 'Inativo'}
                                        </label>
                                    </div>
                                </div>
                            </div>

                            <motion.button
                                type="submit"
                                style={styles.button}
                                whileHover={{ scale: 1.02 }}
                                whileTap={{ scale: 0.98 }}
                                disabled={submitting}
                            >
                                <Save size={18} /> Salvar Alterações
                            </motion.button>
                        </form>
                    </motion.div>

                    {/* SECURITY */}
                    <motion.div style={styles.card} initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.1 }}>
                        <h2 style={styles.sectionTitle}><Lock size={18} /> Segurança</h2>
                        <form onSubmit={handleResetPassword} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                            <div style={styles.formGroup}>
                                <label style={styles.label}>Nova Senha</label>
                                <input
                                    type="text"
                                    style={styles.input}
                                    value={newPassword}
                                    onChange={e => setNewPassword(e.target.value)}
                                    placeholder="Mínimo 6 caracteres"
                                />
                            </div>
                            <motion.button
                                type="submit"
                                style={{ ...styles.button, background: 'rgba(239, 68, 68, 0.1)', color: '#EF4444', border: '1px solid rgba(239, 68, 68, 0.2)' }}
                                whileHover={{ scale: 1.02, background: 'rgba(239, 68, 68, 0.2)' }}
                                whileTap={{ scale: 0.98 }}
                                disabled={submitting}
                            >
                                Redefinir Senha
                            </motion.button>
                        </form>
                    </motion.div>
                </div>

                {/* RIGHT COLUMN: Activity & History */}
                <div style={{ display: 'flex', flexDirection: 'column', gap: '2rem' }}>

                    {/* SUMMARY STATS */}
                    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                        <div style={{ ...styles.card, padding: 'var(--btn-padding-md)' }}>
                            <span style={styles.label}>Saldo Atual</span>
                            <span style={{ fontSize: 'var(--title-h4)', fontWeight: 700, color: '#22C55E' }}>
                                {new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(user.balance || 0)}
                            </span>
                        </div>
                        <div style={{ ...styles.card, padding: 'var(--btn-padding-md)' }}>
                            <span style={styles.label}>Produtos Comprados</span>
                            <span style={{ fontSize: 'var(--title-h4)', fontWeight: 700, color: '#F8F9FA' }}>
                                {library.length}
                            </span>
                        </div>
                    </div>

                    {/* SUBSCRIPTIONS */}
                    <motion.div style={styles.card} initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.2 }}>
                        <h2 style={styles.sectionTitle}><Shield size={18} /> Assinaturas</h2>
                        {subscriptions.length > 0 ? (
                            <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                                {subscriptions.map(sub => (
                                    <div key={sub.id} style={{
                                        padding: 'var(--btn-padding-md)',
                                        background: 'rgba(255,255,255,0.03)',
                                        borderRadius: '0.5rem',
                                        cursor: 'pointer',
                                    }} onClick={() => navigate(`/admin/subscriptions/${sub.id}`)}>
                                        <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem' }}>
                                            <span style={{ fontWeight: 600 }}>{sub.plan_name || 'Plano Personalizado'}</span>
                                            <span style={styles.badge(sub.status)}>{sub.status}</span>
                                        </div>
                                        <div style={{ fontSize: '0.85rem', color: 'rgba(255,255,255,0.5)' }}>
                                            ID: {sub.id.substring(0, 8)}...
                                        </div>
                                    </div>
                                ))}
                            </div>
                        ) : (
                            <div style={{ color: 'rgba(255,255,255,0.3)', textAlign: 'center', padding: 'var(--btn-padding-md)' }}>Nenhuma assinatura.</div>
                        )}
                    </motion.div>

                    {/* PURCHASED PRODUCTS */}
                    <motion.div style={styles.card} initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.3 }}>
                        <h2 style={styles.sectionTitle}><Package size={18} /> Produtos (Biblioteca)</h2>
                        {library.length > 0 ? (
                            <div style={{ maxHeight: '300px', overflowY: 'auto' }}>
                                <table style={styles.table}>
                                    <thead>
                                        <tr>
                                            <th style={styles.th}>Produto</th>
                                            <th style={styles.th}>Data</th>
                                            <th style={styles.th}>Preço</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {library.map((item, idx) => (
                                            <tr key={idx}>
                                                <td style={styles.td}>{item.product?.name || 'Desconhecido'}</td>
                                                <td style={styles.td}>{new Date(item.purchase?.purchased_at).toLocaleDateString('pt-BR')}</td>
                                                <td style={styles.td}>
                                                    {new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(item.purchase?.purchase_price || 0)}
                                                </td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            </div>
                        ) : (
                            <div style={{ color: 'rgba(255,255,255,0.3)', textAlign: 'center', padding: 'var(--btn-padding-md)' }}>Nenhum produto comprado.</div>
                        )}
                    </motion.div>

                    {/* TRANSACTIONS */}
                    <motion.div style={styles.card} initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.4 }}>
                        <h2 style={styles.sectionTitle}><Clock size={18} /> Histórico de Transações</h2>
                        {transactions.length > 0 ? (
                            <div style={{ maxHeight: '300px', overflowY: 'auto' }}>
                                <table style={styles.table}>
                                    <thead>
                                        <tr>
                                            <th style={styles.th}>Data</th>
                                            <th style={styles.th}>Tipo</th>
                                            <th style={styles.th}>Valor</th>
                                            <th style={styles.th}>Status</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {transactions.map(tx => (
                                            <tr key={tx.id}>
                                                <td style={styles.td}>{new Date(tx.created_at).toLocaleDateString('pt-BR')}</td>
                                                <td style={styles.td}>{tx.type}</td>
                                                <td style={styles.td}>
                                                    {new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(tx.amount || 0)}
                                                </td>
                                                <td style={styles.td}>
                                                    <span style={styles.badge(tx.status)}>{tx.status}</span>
                                                </td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            </div>
                        ) : (
                            <div style={{ color: 'rgba(255,255,255,0.3)', textAlign: 'center', padding: 'var(--btn-padding-md)' }}>Nenhuma transação.</div>
                        )}
                    </motion.div>

                </div>
            </div>
        </div>
    );
};

export default UserDetailPage;

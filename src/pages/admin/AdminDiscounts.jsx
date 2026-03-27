import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
    Ticket,
    Plus,
    Trash2,
    Edit,
    Search,
    Filter,
    Calendar,
    CheckCircle,
    XCircle,
    Percent,
    DollarSign,
    Save,
    X,
    ChevronDown,
    Gamepad2,
    Package,
    LayoutGrid
} from 'lucide-react';
import { adminAPI, gamesAPI, productsAPI } from '../../services/api';

const AdminDiscounts = () => {
    const [discounts, setDiscounts] = useState([]);
    const [loading, setLoading] = useState(true);
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [editingDiscount, setEditingDiscount] = useState(null);
    const [search, setSearch] = useState('');

    // Form data for new/edit discount
    const [formData, setFormData] = useState({
        code: '',
        type: 'PERCENTAGE',
        value: 0,
        restriction_type: 'ALL',
        target_ids: [],
        expires_at: '',
        max_uses: '',
        is_active: true
    });

    // Reference data for restrictions
    const [games, setGames] = useState([]);
    const [categories, setCategories] = useState([]);
    const [products, setProducts] = useState([]);

    useEffect(() => {
        fetchDiscounts();
        fetchReferenceData();
    }, []);

    const fetchDiscounts = async () => {
        try {
            setLoading(true);
            const data = await adminAPI.getDiscounts();
            setDiscounts(data || []);
        } catch (error) {
            console.error('Error fetching discounts:', error);
        } finally {
            setLoading(false);
        }
    };

    const fetchReferenceData = async () => {
        try {
            const [gamesData, productsData] = await Promise.all([
                gamesAPI.getWithCategories(),
                productsAPI.getAll({ limit: 100 })
            ]);
            
            setGames(gamesData || []);
            
            // Extract all categories from games
            const allCats = [];
            gamesData?.forEach(g => {
                if (g.categories) {
                    g.categories.forEach(c => allCats.push({ ...c, gameName: g.game.name }));
                }
            });
            setCategories(allCats);
            setProducts(productsData.products || []);
        } catch (error) {
            console.error('Error fetching reference data:', error);
        }
    };

    const handleSave = async (e) => {
        e.preventDefault();
        try {
            const payload = {
                ...formData,
                value: parseFloat(formData.value),
                max_uses: formData.max_uses ? parseInt(formData.max_uses) : null,
                expires_at: formData.expires_at ? new Date(formData.expires_at).toISOString() : null
            };

            if (editingDiscount) {
                await adminAPI.updateDiscount(editingDiscount.id, payload);
            } else {
                await adminAPI.createDiscount(payload);
            }
            
            setIsModalOpen(false);
            setEditingDiscount(null);
            resetForm();
            fetchDiscounts();
        } catch (error) {
            alert('Erro ao salvar cupom: ' + error.message);
        }
    };

    const handleDelete = async (id) => {
        if (!window.confirm('Tem certeza que deseja excluir este cupom?')) return;
        try {
            await adminAPI.deleteDiscount(id);
            fetchDiscounts();
        } catch (error) {
            alert('Erro ao excluir cupom');
        }
    };

    const openModal = (discount = null) => {
        if (discount) {
            setEditingDiscount(discount);
            setFormData({
                code: discount.code,
                type: discount.type,
                value: discount.value,
                restriction_type: discount.restriction_type || 'ALL',
                target_ids: discount.target_ids || [],
                expires_at: discount.expires_at ? new Date(discount.expires_at).toISOString().slice(0, 16) : '',
                max_uses: discount.max_uses || '',
                is_active: discount.is_active
            });
        } else {
            resetForm();
        }
        setIsModalOpen(true);
    };

    const resetForm = () => {
        setFormData({
            code: '',
            type: 'PERCENTAGE',
            value: 0,
            restriction_type: 'ALL',
            target_ids: [],
            expires_at: '',
            max_uses: '',
            is_active: true
        });
    };

    const toggleTarget = (id) => {
        setFormData(prev => {
            const ids = [...prev.target_ids];
            const index = ids.indexOf(id);
            if (index > -1) ids.splice(index, 1);
            else ids.push(id);
            return { ...prev, target_ids: ids };
        });
    };

    const filteredDiscounts = discounts.filter(d => 
        d.code.toLowerCase().includes(search.toLowerCase())
    );

    const styles = {
        container: { padding: '2rem' },
        header: { display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' },
        title: { fontSize: '1.8rem', fontWeight: 700, display: 'flex', alignItems: 'center', gap: '0.75rem' },
        addButton: { background: 'var(--gradient-cta)', border: 'none', borderRadius: '0.5rem', padding: '0.75rem 1.5rem', color: 'white', fontWeight: 600, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '0.5rem' },
        grid: { display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))', gap: '1.5rem' },
        card: { background: 'rgba(21, 26, 38, 0.6)', backdropFilter: 'blur(10px)', border: '1px solid rgba(255, 255, 255, 0.1)', borderRadius: '1rem', padding: '1.5rem', position: 'relative' },
        code: { fontSize: '1.2rem', fontWeight: 800, color: '#E01A4F', letterSpacing: '0.05em' },
        badge: (active) => ({ padding: '0.2rem 0.6rem', borderRadius: '4px', fontSize: '0.75rem', background: active ? 'rgba(34, 197, 94, 0.1)' : 'rgba(239, 68, 68, 0.1)', color: active ? '#22C55E' : '#EF4444' }),
        modalOverlay: { position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, background: 'rgba(0,0,0,0.8)', backdropFilter: 'blur(10px)', zIndex: 1000, display: 'flex', alignItems: 'center', justifyContent: 'center' },
        modal: { background: '#151924', border: '1px solid rgba(255,255,255,0.1)', borderRadius: '1.5rem', width: '100%', maxWidth: '600px', maxHeight: '90vh', overflowY: 'auto', padding: '2rem' },
        input: { background: 'rgba(0,0,0,0.2)', border: '1px solid rgba(255,255,255,0.1)', borderRadius: '0.5rem', padding: '0.75rem', color: 'white', width: '100%', outline: 'none' },
        label: { fontSize: '0.85rem', color: '#B8BDC7', marginBottom: '0.5rem', display: 'block' },
        selectList: { maxHeight: '200px', overflowY: 'auto', border: '1px solid rgba(255,255,255,0.1)', borderRadius: '0.5rem', padding: '0.5rem', marginTop: '0.5rem' },
        selectItem: (selected) => ({ padding: '0.5rem', borderRadius: '0.3rem', cursor: 'pointer', background: selected ? 'rgba(224, 26, 79, 0.2)' : 'transparent', border: selected ? '1px solid #E01A4F' : '1px solid transparent', fontSize: '0.85rem', marginBottom: '0.2rem' })
    };

    return (
        <div style={styles.container}>
            <div style={styles.header}>
                <h1 style={styles.title}><Ticket size={32} color="#E01A4F" /> Gerenciar Cupons</h1>
                <button style={styles.addButton} onClick={() => openModal()}><Plus size={20} /> Novo Cupom</button>
            </div>

            <div style={{ position: 'relative', marginBottom: '2rem' }}>
                <Search style={{ position: 'absolute', left: '1rem', top: '50%', transform: 'translateY(-50%)', color: 'rgba(255,255,255,0.4)' }} size={18} />
                <input 
                    type="text" 
                    placeholder="Buscar cupons..." 
                    style={{ ...styles.input, paddingLeft: '3rem' }}
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                />
            </div>

            {loading ? (
                <div style={{ textAlign: 'center', padding: '4rem' }}>Carregando cupons...</div>
            ) : (
                <div style={styles.grid}>
                    {filteredDiscounts.map(d => (
                        <motion.div key={d.id} style={styles.card} initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }}>
                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '1rem' }}>
                                <span style={styles.code}>{d.code}</span>
                                <span style={styles.badge(d.is_active)}>{d.is_active ? 'Ativo' : 'Inativo'}</span>
                            </div>
                            
                            <div style={{ display: 'flex', gap: '1rem', marginBottom: '1rem' }}>
                                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: '#22C55E', fontWeight: 700 }}>
                                    {d.type === 'PERCENTAGE' ? <Percent size={16} /> : <DollarSign size={16} />}
                                    {d.value}{d.type === 'PERCENTAGE' ? '%' : ' BRL'}
                                </div>
                                <div style={{ fontSize: '0.85rem', color: '#B8BDC7' }}>
                                    Usos: {d.current_uses} / {d.max_uses || '∞'}
                                </div>
                            </div>

                            <div style={{ fontSize: '0.8rem', color: '#6C7384', marginBottom: '1.5rem' }}>
                                <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', marginBottom: '0.3rem' }}>
                                    <Filter size={14} />
                                    {d.restriction_type === 'ALL' && 'Válido para tudo'}
                                    {d.restriction_type === 'GAME' && 'Restrito a jogos específicos'}
                                    {d.restriction_type === 'ITEM_CATEGORY' && 'Restrito a categorias'}
                                    {d.restriction_type === 'PRODUCT' && 'Restrito a itens específicos'}
                                </div>
                                {d.expires_at && (
                                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
                                        <Calendar size={14} /> Expirado em {new Date(d.expires_at).toLocaleDateString()}
                                    </div>
                                )}
                            </div>

                            <div style={{ display: 'flex', gap: '0.5rem', marginTop: 'auto' }}>
                                <button style={{ ...styles.addButton, background: 'rgba(88, 58, 255, 0.1)', color: '#583AFF', flex: 1, padding: '0.5rem' }} onClick={() => openModal(d)}><Edit size={16} /></button>
                                <button style={{ ...styles.addButton, background: 'rgba(239, 68, 68, 0.1)', color: '#EF4444', flex: 1, padding: '0.5rem' }} onClick={() => handleDelete(d.id)}><Trash2 size={16} /></button>
                            </div>
                        </motion.div>
                    ))}
                </div>
            )}

            <AnimatePresence>
                {isModalOpen && (
                    <div style={styles.modalOverlay} onClick={() => setIsModalOpen(false)}>
                        <motion.div 
                            style={styles.modal} 
                            onClick={e => e.stopPropagation()}
                            initial={{ scale: 0.9, opacity: 0 }}
                            animate={{ scale: 1, opacity: 1 }}
                            exit={{ scale: 0.9, opacity: 0 }}
                        >
                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' }}>
                                <h2 style={{ margin: 0 }}>{editingDiscount ? 'Editar Cupom' : 'Criar Novo Cupom'}</h2>
                                <button onClick={() => setIsModalOpen(false)} style={{ background: 'none', border: 'none', color: '#B8BDC7', cursor: 'pointer' }}><X size={24} /></button>
                            </div>

                            <form onSubmit={handleSave}>
                                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem', marginBottom: '1.5rem' }}>
                                    <div>
                                        <label style={styles.label}>Código do Cupom</label>
                                        <input 
                                            type="text" 
                                            style={styles.input} 
                                            placeholder="Ex: PROMO2024"
                                            value={formData.code}
                                            onChange={e => setFormData({...formData, code: e.target.value.toUpperCase()})}
                                            required
                                        />
                                    </div>
                                    <div>
                                        <label style={styles.label}>Tipo</label>
                                        <select 
                                            style={styles.input}
                                            value={formData.type}
                                            onChange={e => setFormData({...formData, type: e.target.value})}
                                        >
                                            <option value="PERCENTAGE">Percentual (%)</option>
                                            <option value="FIXED_AMOUNT">Valor Fixo (R$)</option>
                                        </select>
                                    </div>
                                </div>

                                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem', marginBottom: '1.5rem' }}>
                                    <div>
                                        <label style={styles.label}>Valor do Desconto</label>
                                        <input 
                                            type="number" 
                                            style={styles.input} 
                                            value={formData.value}
                                            onChange={e => setFormData({...formData, value: e.target.value})}
                                            required
                                        />
                                    </div>
                                    <div>
                                        <label style={styles.label}>Limite de Usos (Opcional)</label>
                                        <input 
                                            type="number" 
                                            style={styles.input} 
                                            value={formData.max_uses}
                                            onChange={e => setFormData({...formData, max_uses: e.target.value})}
                                        />
                                    </div>
                                </div>

                                <div style={{ marginBottom: '1.5rem' }}>
                                    <label style={styles.label}>Data de Expiração (Opcional)</label>
                                    <input 
                                        type="datetime-local" 
                                        style={styles.input} 
                                        value={formData.expires_at}
                                        onChange={e => setFormData({...formData, expires_at: e.target.value})}
                                    />
                                </div>

                                <div style={{ marginBottom: '1.5rem' }}>
                                    <label style={styles.label}>Restrição de Uso</label>
                                    <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
                                        {[
                                            { id: 'ALL', label: 'Tudo', icon: LayoutGrid },
                                            { id: 'GAME', label: 'Por Jogo', icon: Gamepad2 },
                                            { id: 'ITEM_CATEGORY', label: 'Por Categoria', icon: Filter },
                                            { id: 'PRODUCT', label: 'Itens Específicos', icon: Package },
                                        ].map(opt => (
                                            <button
                                                key={opt.id}
                                                type="button"
                                                onClick={() => setFormData({...formData, restriction_type: opt.id, target_ids: []})}
                                                style={{
                                                    display: 'flex', alignItems: 'center', gap: '0.5rem',
                                                    padding: '0.5rem 1rem', borderRadius: '0.5rem', fontSize: '0.85rem',
                                                    background: formData.restriction_type === opt.id ? '#E01A4F' : 'rgba(255,255,255,0.05)',
                                                    border: '1px solid ' + (formData.restriction_type === opt.id ? '#E01A4F' : 'rgba(255,255,255,0.1)'),
                                                    color: 'white', cursor: 'pointer'
                                                }}
                                            >
                                                <opt.icon size={14} /> {opt.label}
                                            </button>
                                        ))}
                                    </div>

                                    {formData.restriction_type !== 'ALL' && (
                                        <div style={styles.selectList}>
                                            {formData.restriction_type === 'GAME' && games.map(g => (
                                                <div 
                                                    key={g.game.id} 
                                                    style={styles.selectItem(formData.target_ids.includes(g.game.id))}
                                                    onClick={() => toggleTarget(g.game.id)}
                                                >
                                                    {g.game.name}
                                                </div>
                                            ))}
                                            {formData.restriction_type === 'ITEM_CATEGORY' && categories.map(c => (
                                                <div 
                                                    key={c.id} 
                                                    style={styles.selectItem(formData.target_ids.includes(c.id))}
                                                    onClick={() => toggleTarget(c.id)}
                                                >
                                                    [{c.gameName}] {c.name}
                                                </div>
                                            ))}
                                            {formData.restriction_type === 'PRODUCT' && products.map(p => (
                                                <div 
                                                    key={p.id} 
                                                    style={styles.selectItem(formData.target_ids.includes(p.id))}
                                                    onClick={() => toggleTarget(p.id)}
                                                >
                                                    {p.name} - R$ {p.price}
                                                </div>
                                            ))}
                                        </div>
                                    )}
                                </div>

                                <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '2rem' }}>
                                    <input 
                                        type="checkbox" 
                                        id="active" 
                                        checked={formData.is_active}
                                        onChange={e => setFormData({...formData, is_active: e.target.checked})}
                                    />
                                    <label htmlFor="active" style={{ fontSize: '0.85rem', cursor: 'pointer' }}>Cupom Ativo</label>
                                </div>

                                <button type="submit" style={{ ...styles.addButton, width: '100%', justifyContent: 'center' }}>
                                    <Save size={20} /> Salvar Cupom
                                </button>
                            </form>
                        </motion.div>
                    </div>
                )}
            </AnimatePresence>
        </div>
    );
};

export default AdminDiscounts;

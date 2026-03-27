import React, { useState, useEffect, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Loader2, Plus, Edit2, Trash2, X, Check, Package, Layers, Upload, Gamepad2, Tag } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { adminAPI, productsAPI, filesAPI, gamesAPI } from '../../services/api';

const AdminCatalog = () => {
    const navigate = useNavigate();
    const [activeTab, setActiveTab] = useState('products'); // 'products', 'plans', or 'games'
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    // Data States
    const [products, setProducts] = useState([]);
    const [plans, setPlans] = useState([]);
    const [games, setGames] = useState([]);

    // Modal States
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [editingProduct, setEditingProduct] = useState(null);
    const [modalFormData, setModalFormData] = useState({
        name: '',
        description: '',
        price: '',
        type: 'PLUGIN',
        game_id: '',
        category_id: '',
        image_url: '',
        stock_quantity: 0,
        download_url: '',
        is_exclusive: false
    });

    // Plan Modal States
    const [isPlanModalOpen, setIsPlanModalOpen] = useState(false);
    const [editingPlan, setEditingPlan] = useState(null);
    const [planData, setPlanData] = useState({
        name: '',
        description: '',
        price: '',
        features: [],
        imageUrl: ''
    });
    const [uploadingPlanImage, setUploadingPlanImage] = useState(false);

    // Game Modal States
    const [isGameModalOpen, setIsGameModalOpen] = useState(false);
    const [editingGame, setEditingGame] = useState(null);
    const [gameFormData, setGameFormData] = useState({
        name: '',
        slug: '',
        icon_url: '',
        display_order: 0
    });

    // Category Modal States
    const [isCategoryModalOpen, setIsCategoryModalOpen] = useState(false);
    const [editingCategory, setEditingCategory] = useState(null);
    const [categoryFormData, setCategoryFormData] = useState({
        game_id: '',
        name: '',
        slug: '',
        display_order: 0
    });

    // File Upload States
    const [fileUpload, setFileUpload] = useState({
        file: null,
        name: '',
        isUploading: false,
        progress: 0,
        uploadedFile: null, // Store uploaded file info
    });

    // File Selection for Product States
    const [availableFiles, setAvailableFiles] = useState([]);
    const [selectedFileId, setSelectedFileId] = useState(null);

    // Fetch Data
    const fetchData = useCallback(async () => {
        try {
            setLoading(true);
            const [productsData, plansData, gamesData] = await Promise.all([
                productsAPI.getAll({ page_size: 100 }),
                adminAPI.getPlans(),
                gamesAPI.getWithCategories()
            ]);

            // Fix: Ensure products is always an array to prevent .map() crash
            const productsArray = Array.isArray(productsData)
                ? productsData
                : (productsData?.products || productsData?.data || []);

            setProducts(Array.isArray(productsArray) ? productsArray : []);
            setPlans(Array.isArray(plansData) ? plansData : []);
            setGames(Array.isArray(gamesData) ? gamesData : []);
        } catch (_error) {
            console.error('Failed to fetch catalog data:', _error);
            setError('Erro ao carregar catálogo.');
        } finally {
            setLoading(false);
        }
    }, [setError, setLoading]);

    // Fetch available files for product assignment
    const fetchAvailableFiles = async () => {
        try {
            // Use admin endpoint to get all files
            const response = await filesAPI.listAllAdmin({ page: 1, page_size: 100 });
            setAvailableFiles(response.files || []);
        } catch (error) {
            console.error('Failed to fetch available files:', error);
            setAvailableFiles([]);
        }
    };

    useEffect(() => {
        fetchData();
    }, [fetchData]);

    // Handle file selection for upload
    const handleFileChange = (e) => {
        const file = e.target.files[0];
        if (file) {
            setFileUpload({
                ...fileUpload,
                file: file,
                name: file.name.replace(/\.[^/.]+$/, ""), // Remove extension for default name
                uploadedFile: null
            });
        }
    };

    // Upload file to server
    const handleFileUpload = async () => {
        if (!fileUpload.file) return;

        setFileUpload({
            ...fileUpload,
            isUploading: true,
            progress: 0
        });

        try {
            const response = await filesAPI.upload(fileUpload.file, fileUpload.name);
            setFileUpload({
                ...fileUpload,
                isUploading: false,
                uploadedFile: response,
                file: null,
                name: ''
            });
            // Fetch available files again to include the new one
            await fetchAvailableFiles();
            alert('Arquivo enviado com sucesso!');
        } catch (error) {
            console.error('Upload failed:', error);
            setFileUpload({
                ...fileUpload,
                isUploading: false
            });
            alert('Falha ao enviar arquivo: ' + error.message);
        }
    };

    // Handle product save with file upload
    const handleSaveProduct = async (e) => {
        e.preventDefault();
        try {
            const payload = {
                ...modalFormData,
                price: parseFloat(modalFormData.price) || 0,
                stock_quantity: parseInt(modalFormData.stock_quantity) || 0,
                is_exclusive: modalFormData.is_exclusive,
                download_url: selectedFileId ? '' : modalFormData.download_url,
                file_id: selectedFileId,
                // Ensure optional fields are handled or undefined if empty
                description: modalFormData.description || '',
                image_url: modalFormData.image_url || '',
                game_id: modalFormData.game_id || null,
                category_id: modalFormData.category_id || null
            };

            if (editingProduct) {
                await adminAPI.updateProduct(editingProduct.id, payload);
            } else {
                await adminAPI.createProduct(payload);
            }

            handleCloseModal();
            fetchData(); // Refresh list
        } catch (error) {
            console.error('Failed to save product:', error);
            alert('Erro ao salvar produto: ' + error.message);
        }
    };

    // Close modal and reset file states
    const handleCloseModal = () => {
        setIsModalOpen(false);
        setEditingProduct(null);
        setFileUpload({
            file: null,
            name: '',
            isUploading: false,
            progress: 0,
            uploadedFile: null,
        });
        setSelectedFileId(null);
    };

    // Product Actions
    const handleOpenModal = (product = null) => {
        if (product) {
            setEditingProduct(product);
            setModalFormData({
                name: product.name,
                description: product.description,
                price: product.price,
                type: product.type,
                game_id: product.game_id || '',
                category_id: product.category_id || '',
                // Set download_url or handle file_id appropriately
                download_url: product.download_url || '',
                image_url: product.image_url,
                stock_quantity: product.stock_quantity,
                is_exclusive: product.is_exclusive
            });
            // Set the selected file ID if the product was created with a file
            setSelectedFileId(product.file_id || null);
        } else {
            setEditingProduct(null);
            setModalFormData({
                name: '',
                description: '',
                price: '',
                type: 'PLUGIN',
                game_id: '',
                category_id: '',
                download_url: '',
                image_url: '',
                stock_quantity: 0,
                is_exclusive: false
            });
            setSelectedFileId(null);
        }
        setIsModalOpen(true);

        // Fetch available files for selection - always fetch to show current files
        fetchAvailableFiles();
    };

    const handleDeleteProduct = async (id) => {
        if (window.confirm('Tem certeza que deseja excluir este produto?')) {
            try {
                await adminAPI.deleteProduct(id);
                fetchData();
            } catch (error) {
                console.error('Failed to delete product:', error);
                alert('Erro ao excluir produto: ' + error.message);
            }
        }
    };

    // Plan Actions
    const handleOpenPlanModal = (plan = null) => {
        if (plan) {
            setEditingPlan(plan);
            // Parse features to string for textarea
            let featuresArray = [];
            try {
                if (Array.isArray(plan.features)) {
                    featuresArray = plan.features;
                } else if (typeof plan.features === 'string') {
                    const parsed = JSON.parse(plan.features);
                    if (Array.isArray(parsed)) featuresArray = parsed;
                }
            } catch (_error) {
                console.warn('Failed to parse plan features for edit:', _error);
            }

            setPlanData({
                name: plan.name,
                description: plan.description,
                price: plan.price.toString(),
                features: featuresArray,
                imageUrl: plan.imageUrl || ''
            });
        } else {
            setEditingPlan(null);
            setPlanData({
                name: '',
                description: '',
                price: '',
                features: [],
                imageUrl: ''
            });
        }
        setIsPlanModalOpen(true);
    };

    const handleClosePlanModal = () => {
        setIsPlanModalOpen(false);
        setEditingPlan(null);
        setPlanData({
            name: '',
            description: '',
            price: '',
            features: [],
            imageUrl: ''
        });
        setUploadingPlanImage(false);
    };

    const handlePlanImageUpload = async (e) => {
        const file = e.target.files?.[0];
        if (!file) return;

        try {
            setUploadingPlanImage(true);
            const data = await filesAPI.upload(file, file.name);
            setPlanData(prev => ({ ...prev, imageUrl: data.url }));
        } catch (err) {
            console.error('Failed to upload plan image:', err);
            alert('Falha ao enviar imagem do plano: ' + err.message);
        } finally {
            setUploadingPlanImage(false);
        }
    };

    const handleSavePlan = async (e) => {
        e.preventDefault();
        try {
            // Convert features array to string for textarea
            const featuresArray = planData.features
                .map(f => f.trim())
                .filter(f => f.length > 0);

            const payload = {
                name: planData.name,
                description: planData.description,
                price: parseFloat(planData.price) || 0,
                features: featuresArray,
                imageUrl: planData.imageUrl || null
            };

            if (editingPlan) {
                await adminAPI.updatePlan(editingPlan.id, payload);
            } else {
                await adminAPI.createPlan(payload);
            }

            handleClosePlanModal();
            fetchData();
        } catch (error) {
            console.error('Failed to save plan:', error);
            alert('Erro ao salvar plano: ' + error.message);
        }
    };



    const handleDeletePlan = async (id) => {
        if (window.confirm('Tem certeza que deseja excluir este plano?')) {
            try {
                await adminAPI.deletePlan(id);
                fetchData();
            } catch (error) {
                console.error('Failed to delete plan:', error);
                alert('Erro ao excluir plano: ' + error.message);
            }
        }
    };

    // Game Actions
    const handleOpenGameModal = (game = null) => {
        if (game) {
            setEditingGame(game);
            setGameFormData({
                name: game.name,
                slug: game.slug,
                icon_url: game.icon_url || '',
                display_order: game.display_order || 0
            });
        } else {
            setEditingGame(null);
            setGameFormData({
                name: '',
                slug: '',
                icon_url: '',
                display_order: 0
            });
        }
        setIsGameModalOpen(true);
    };

    const handleCloseGameModal = () => {
        setIsGameModalOpen(false);
        setEditingGame(null);
    };

    const handleSaveGame = async (e) => {
        e.preventDefault();
        try {
            const payload = {
                ...gameFormData,
                display_order: parseInt(gameFormData.display_order) || 0
            };

            if (editingGame) {
                await adminAPI.updateGame(editingGame.id, payload);
            } else {
                await adminAPI.createGame(payload);
            }

            handleCloseGameModal();
            fetchData();
        } catch (error) {
            console.error('Failed to save game:', error);
            alert('Erro ao salvar jogo: ' + error.message);
        }
    };

    const handleDeleteGame = async (id) => {
        if (window.confirm('Tem certeza que deseja excluir este jogo? Isso também excluirá todas as categorias associadas!')) {
            try {
                await adminAPI.deleteGame(id);
                fetchData();
            } catch (error) {
                console.error('Failed to delete game:', error);
                alert('Erro ao excluir jogo: ' + error.message);
            }
        }
    };

    // Category Actions
    const handleOpenCategoryModal = (category = null, gameId = null) => {
        if (category) {
            setEditingCategory(category);
            setCategoryFormData({
                game_id: category.game_id,
                name: category.name,
                slug: category.slug,
                display_order: category.display_order || 0
            });
        } else {
            setEditingCategory(null);
            setCategoryFormData({
                game_id: gameId || (games.length > 0 ? games[0].game.id : ''),
                name: '',
                slug: '',
                display_order: 0
            });
        }
        setIsCategoryModalOpen(true);
    };

    const handleCloseCategoryModal = () => {
        setIsCategoryModalOpen(false);
        setEditingCategory(null);
    };

    const handleSaveCategory = async (e) => {
        e.preventDefault();
        try {
            const payload = {
                ...categoryFormData,
                display_order: parseInt(categoryFormData.display_order) || 0
            };

            if (editingCategory) {
                await adminAPI.updateCategory(editingCategory.id, payload);
            } else {
                await adminAPI.createCategory(payload);
            }

            handleCloseCategoryModal();
            fetchData();
        } catch (error) {
            console.error('Failed to save category:', error);
            alert('Erro ao salvar categoria: ' + error.message);
        }
    };

    const handleDeleteCategory = async (id) => {
        if (window.confirm('Tem certeza que deseja excluir esta categoria?')) {
            try {
                await adminAPI.deleteCategory(id);
                fetchData();
            } catch (error) {
                console.error('Failed to delete category:', error);
                alert('Erro ao excluir categoria: ' + error.message);
            }
        }
    };

    // Styles
    const styles = {
        container: {
            color: '#F8F9FA'
        },
        header: {
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: '2rem',
        },
        pageTitle: {
            fontSize: 'var(--title-h3)',
            fontWeight: 800,
            background: 'linear-gradient(135deg, #F8F9FA 0%, #583AFF 50%, #1AD2FF 100%)',
            WebkitBackgroundClip: 'text',
            WebkitTextFillColor: 'transparent',
        },
        tabs: {
            display: 'flex',
            gap: '1rem',
            marginBottom: '2rem',
            borderBottom: '1px solid rgba(255, 255, 255, 0.1)',
            paddingBottom: '1rem',
        },
        tabButton: (isActive) => ({
            padding: 'var(--btn-padding-md)',
            borderRadius: '8px',
            background: isActive ? 'rgba(88, 58, 255, 0.2)' : 'transparent',
            color: isActive ? '#fff' : '#B8BDC7',
            border: isActive ? '1px solid rgba(88, 58, 255, 0.5)' : '1px solid transparent',
            cursor: 'pointer',
            fontWeight: 600,
            display: 'flex',
            alignItems: 'center',
            gap: '0.5rem',
            transition: 'all 0.2s',
        }),
        card: {
            background: 'rgba(255, 255, 255, 0.03)',
            border: '1px solid rgba(255, 255, 255, 0.05)',
            borderRadius: '12px',
            padding: '1.5rem',
            display: 'flex',
            flexDirection: 'column',
            gap: '1rem',
            transition: 'transform 0.2s, box-shadow 0.2s',
        },
        grid: {
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))',
            gap: '1.5rem',
        },
        badge: (type) => {
            const colors = {
                PLUGIN: '#583AFF',
                MOD: '#E01A4F',
                MAP: '#1AD2FF',
                TEXTUREPACK: '#F59E0B',
                SERVER_TEMPLATE: '#10B981'
            };
            const color = colors[type] || '#6C7384';
            return {
                padding: '0.25rem 0.75rem',
                borderRadius: '20px',
                fontSize: '0.75rem',
                fontWeight: 700,
                background: `${color}20`,
                color: color,
                border: `1px solid ${color}40`,
                display: 'inline-block',
            };
        },
        actionButton: {
            padding: '0.5rem',
            borderRadius: '6px',
            border: 'none',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            transition: 'background 0.2s',
        },
        modalOverlay: {
            position: 'fixed',
            top: 0, left: 0, right: 0, bottom: 0,
            background: 'rgba(0, 0, 0, 0.7)',
            backdropFilter: 'blur(5px)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            zIndex: 1000,
        },
        modalContent: {
            background: '#151A26',
            border: '1px solid rgba(88, 58, 255, 0.2)',
            borderRadius: '16px',
            padding: '2rem',
            width: '600px',
            maxWidth: '90vw',
            maxHeight: '90vh',
            overflowY: 'auto',
            boxShadow: '0 20px 50px rgba(0, 0, 0, 0.5)',
        },
        inputGroup: {
            marginBottom: '1rem',
        },
        label: {
            display: 'block',
            marginBottom: '0.5rem',
            color: '#B8BDC7',
            fontSize: '0.9rem',
        },
        input: {
            width: '100%',
            padding: '0.75rem',
            background: 'rgba(0, 0, 0, 0.2)',
            border: '1px solid rgba(255, 255, 255, 0.1)',
            borderRadius: '8px',
            color: '#fff',
            fontSize: '1rem',
            outline: 'none',
        },
        select: {
            width: '100%',
            padding: '0.75rem',
            background: '#0A0E1A',
            border: '1px solid rgba(255, 255, 255, 0.1)',
            borderRadius: '8px',
            color: '#fff',
            fontSize: '1rem',
            outline: 'none',
        },
        primaryButton: {
            background: 'var(--gradient-primary)',
            color: 'white',
            border: 'none',
            padding: 'var(--btn-padding-md)',
            borderRadius: '8px',
            fontWeight: 600,
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            gap: '0.5rem',
        }
    };

    return (
        <div style={styles.container}>
            <header style={styles.header}>
                <div>
                    <h1 style={styles.pageTitle}>Gerenciamento de Catálogo</h1>
                    <p style={{ color: '#B8BDC7', marginTop: '0.5rem' }}>Gerencie produtos e visualize planos disponíveis.</p>
                </div>
                {activeTab === 'products' && (
                    <button
                        style={styles.primaryButton}
                        onClick={() => handleOpenModal()}
                    >
                        <Plus size={20} /> Novo Produto
                    </button>
                )}
                {activeTab === 'plans' && (
                    <button
                        style={styles.primaryButton}
                        onClick={() => handleOpenPlanModal()}
                    >
                        <Plus size={20} /> Novo Plano
                    </button>
                )}
                {activeTab === 'games' && (
                    <button
                        style={styles.primaryButton}
                        onClick={() => handleOpenGameModal()}
                    >
                        <Plus size={20} /> Novo Jogo
                    </button>
                )}
            </header>

            {error && (
                <div style={{ padding: 'var(--btn-padding-md)', background: 'rgba(239, 68, 68, 0.1)', border: '1px solid rgba(239, 68, 68, 0.3)', borderRadius: '0.5rem', marginBottom: '1.5rem', color: '#EF4444' }}>
                    <p style={{ margin: 0 }}>{error}</p>
                </div>
            )}

            {/* Tabs */}
            <div style={styles.tabs}>
                <button
                    style={styles.tabButton(activeTab === 'products')}
                    onClick={() => setActiveTab('products')}
                >
                    <Package size={18} /> Produtos
                </button>
                <button
                    style={styles.tabButton(activeTab === 'plans')}
                    onClick={() => setActiveTab('plans')}
                >
                    <Layers size={18} /> Planos
                </button>
                <button
                    style={styles.tabButton(activeTab === 'games')}
                    onClick={() => setActiveTab('games')}
                >
                    <Gamepad2 size={18} /> Jogos e Categorias
                </button>
            </div>

            {/* Content */}
            {loading ? (
                <div style={{ display: 'flex', justifyContent: 'center', padding: '4rem' }}>
                    <Loader2 size={40} style={{ animation: 'spin 1s linear infinite', color: '#583AFF' }} />
                </div>
            ) : (
                <>
                    {activeTab === 'products' && (
                        <div style={styles.grid}>
                            {products.map(product => (
                                <motion.div
                                    key={product.id}
                                    style={styles.card}
                                    initial={{ opacity: 0, y: 20 }}
                                    animate={{ opacity: 1, y: 0 }}
                                    whileHover={{ y: -5, boxShadow: '0 10px 30px rgba(0,0,0,0.3)' }}
                                >
                                    <div style={{ height: '160px', borderRadius: '8px', overflow: 'hidden', background: '#000' }}>
                                        <img
                                            src={product.image_url || 'https://via.placeholder.com/300'}
                                            alt={product.name}
                                            style={{ width: '100%', height: '100%', objectFit: 'cover' }}
                                        />
                                    </div>
                                    <div>
                                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'start', marginBottom: '0.5rem' }}>
                                            <h3 style={{ fontSize: '1.1rem', fontWeight: 700 }}>{product.name}</h3>
                                            <span style={styles.badge(product.type)}>{product.type}</span>
                                        </div>
                                        <p style={{ color: '#B8BDC7', fontSize: '0.9rem', marginBottom: '1rem', height: '40px', overflow: 'hidden' }}>
                                            {product.description}
                                        </p>
                                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                            <span style={{ fontSize: '1.25rem', fontWeight: 700, color: '#1AD2FF' }}>
                                                R$ {parseFloat(product.price).toFixed(2)}
                                            </span>
                                            <div style={{ display: 'flex', gap: '0.5rem' }}>
                                                <button
                                                    style={{ ...styles.actionButton, background: 'rgba(88, 58, 255, 0.1)', color: '#583AFF' }}
                                                    onClick={() => handleOpenModal(product)}
                                                >
                                                    <Edit2 size={18} />
                                                </button>
                                                <button
                                                    style={{ ...styles.actionButton, background: 'rgba(239, 68, 68, 0.1)', color: '#EF4444' }}
                                                    onClick={() => handleDeleteProduct(product.id)}
                                                >
                                                    <Trash2 size={18} />
                                                </button>
                                            </div>
                                        </div>
                                    </div>
                                </motion.div>
                            ))}
                        </div>
                    )}

                    {activeTab === 'plans' && (
                        <div style={styles.grid}>
                            {plans.map(plan => (
                                <motion.div
                                    key={plan.id}
                                    style={{ ...styles.card, border: '1px solid rgba(26, 210, 255, 0.3)' }}
                                    initial={{ opacity: 0, scale: 0.95 }}
                                    animate={{ opacity: 1, scale: 1 }}
                                >
                                    <h3 style={{ fontSize: 'var(--title-h4)', fontWeight: 800, color: '#fff' }}>{plan.name}</h3>
                                    <div style={{ fontSize: 'var(--title-h3)', fontWeight: 900, color: '#1AD2FF' }}>
                                        R$ {parseFloat(plan.price).toFixed(2)}
                                        <span style={{ fontSize: '1rem', color: '#B8BDC7', fontWeight: 400 }}>/{plan.duration_months} meses</span>
                                    </div>
                                    <p style={{ color: '#B8BDC7' }}>{plan.description}</p>
                                    <div style={{ marginTop: '1rem' }}>
                                        {(() => {
                                            let features = [];
                                            try {
                                                if (typeof plan.features === 'string') {
                                                    const parsed = JSON.parse(plan.features);
                                                    if (Array.isArray(parsed)) features = parsed;
                                                } else if (Array.isArray(plan.features)) {
                                                    features = plan.features;
                                                }
                                            } catch (e) {
                                                console.warn('Failed to parse plan features:', e);
                                            }

                                            return features.map((feature, idx) => (
                                                <div key={idx} style={{ display: 'flex', gap: '0.5rem', alignItems: 'center', marginBottom: '0.5rem', color: '#B8BDC7' }}>
                                                    <Check size={16} color="#10B981" /> {feature}
                                                </div>
                                            ));
                                        })()}
                                    </div>
                                    <div style={{ display: 'flex', gap: '0.5rem', marginTop: '1rem' }}>
                                        <button
                                            style={{ ...styles.actionButton, background: 'rgba(88, 58, 255, 0.1)', color: '#583AFF' }}
                                            onClick={() => handleOpenPlanModal(plan)}
                                        >
                                            <Edit2 size={18} />
                                        </button>
                                        <button
                                            style={{ ...styles.actionButton, background: 'rgba(239, 68, 68, 0.1)', color: '#EF4444' }}
                                            onClick={() => handleDeletePlan(plan.id)}
                                        >
                                            <Trash2 size={18} />
                                        </button>
                                    </div>
                                </motion.div>
                            ))}
                        </div>
                    )}

                    {activeTab === 'games' && (
                        <div style={styles.grid}>
                            {games.map(gameItem => (
                                <motion.div
                                    key={gameItem.game.id}
                                    style={{ ...styles.card, border: '1px solid rgba(88, 58, 255, 0.3)' }}
                                    initial={{ opacity: 0, scale: 0.95 }}
                                    animate={{ opacity: 1, scale: 1 }}
                                >
                                    <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', marginBottom: '1rem' }}>
                                        {gameItem.game.icon_url ? (
                                            <img src={gameItem.game.icon_url} alt={gameItem.game.name} style={{ width: '48px', height: '48px', objectFit: 'cover', borderRadius: '8px' }} />
                                        ) : (
                                            <div style={{ width: '48px', height: '48px', background: 'rgba(88, 58, 255, 0.2)', borderRadius: '8px', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                                                <Gamepad2 size={24} color="#583AFF" />
                                            </div>
                                        )}
                                        <div>
                                            <h3 style={{ fontSize: '1.25rem', fontWeight: 800, color: '#fff', margin: 0 }}>{gameItem.game.name}</h3>
                                            <span style={{ fontSize: '0.8rem', color: '#B8BDC7' }}>/{gameItem.game.slug}</span>
                                        </div>
                                    </div>

                                    <div style={{ flex: 1 }}>
                                        <h4 style={{ fontSize: '0.9rem', color: '#B8BDC7', marginBottom: '0.75rem', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                                            Categorias
                                            <button
                                                style={{
                                                    background: 'rgba(26, 210, 255, 0.1)',
                                                    border: '1px solid rgba(26, 210, 255, 0.3)',
                                                    color: '#1AD2FF',
                                                    cursor: 'pointer',
                                                    fontSize: '0.75rem',
                                                    fontWeight: 600,
                                                    display: 'flex',
                                                    alignItems: 'center',
                                                    gap: '0.35rem',
                                                    padding: '0.35rem 0.75rem',
                                                    borderRadius: '20px'
                                                }}
                                                onClick={() => handleOpenCategoryModal(null, gameItem.game.id)}
                                            >
                                                <Plus size={14} /> Nova Categoria
                                            </button>
                                        </h4>
                                        <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.75rem' }}>
                                            {gameItem.categories && gameItem.categories.length > 0 ? (
                                                gameItem.categories.map(cat => (
                                                    <span key={cat.id} style={{
                                                        background: 'linear-gradient(135deg, rgba(88, 58, 255, 0.1) 0%, rgba(26, 210, 255, 0.1) 100%)',
                                                        border: '1px solid rgba(88, 58, 255, 0.2)',
                                                        color: '#E2E8F0',
                                                        display: 'inline-flex',
                                                        alignItems: 'center',
                                                        gap: '0.75rem',
                                                        padding: '0.4rem 0.85rem',
                                                        borderRadius: '10px',
                                                        fontSize: '0.85rem',
                                                        fontWeight: 500,
                                                        boxShadow: '0 2px 4px rgba(0,0,0,0.1)'
                                                    }}>
                                                        {cat.name}
                                                        <div style={{ display: 'flex', gap: '0.5rem', paddingLeft: '0.75rem', borderLeft: '1px solid rgba(255,255,255,0.1)' }}>
                                                            <Edit2
                                                                size={14}
                                                                style={{ cursor: 'pointer', color: '#1AD2FF', opacity: 0.8 }}
                                                                onClick={() => handleOpenCategoryModal(cat)}
                                                                onMouseOver={(e) => e.target.style.opacity = 1}
                                                                onMouseOut={(e) => e.target.style.opacity = 0.8}
                                                            />
                                                            <Trash2
                                                                size={14}
                                                                style={{ cursor: 'pointer', color: '#EF4444', opacity: 0.8 }}
                                                                onClick={() => handleDeleteCategory(cat.id)}
                                                                onMouseOver={(e) => e.target.style.opacity = 1}
                                                                onMouseOut={(e) => e.target.style.opacity = 0.8}
                                                            />
                                                        </div>
                                                    </span>
                                                ))
                                            ) : (
                                                <span style={{ fontSize: '0.85rem', color: '#6C7384', fontStyle: 'italic', padding: '0.5rem' }}>Nenhuma categoria cadastrada.</span>
                                            )}
                                        </div>
                                    </div>

                                    <div style={{ display: 'flex', gap: '0.5rem', marginTop: '1rem', paddingTop: '1rem', borderTop: '1px solid rgba(255, 255, 255, 0.05)' }}>
                                        <button
                                            style={{ ...styles.actionButton, background: 'rgba(88, 58, 255, 0.1)', color: '#583AFF', flex: 1 }}
                                            onClick={() => handleOpenGameModal(gameItem.game)}
                                        >
                                            <Edit2 size={18} /> Editar Jogo
                                        </button>
                                        <button
                                            style={{ ...styles.actionButton, background: 'rgba(239, 68, 68, 0.1)', color: '#EF4444' }}
                                            onClick={() => handleDeleteGame(gameItem.game.id)}
                                        >
                                            <Trash2 size={18} />
                                        </button>
                                    </div>
                                </motion.div>
                            ))}
                        </div>
                    )}
                </>
            )}

            {/* Modal */}
            <AnimatePresence>
                {isModalOpen && (
                    <motion.div
                        style={styles.modalOverlay}
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        exit={{ opacity: 0 }}
                    >
                        <motion.div
                            style={styles.modalContent}
                            initial={{ scale: 0.9, y: 20 }}
                            animate={{ scale: 1, y: 0 }}
                            exit={{ scale: 0.9, y: 20 }}
                        >
                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' }}>
                                <h2 style={{ fontSize: 'var(--title-h4)', fontWeight: 700 }}>
                                    {editingProduct ? 'Editar Produto' : 'Novo Produto'}
                                </h2>
                                <button onClick={handleCloseModal} style={{ background: 'transparent', border: 'none', color: '#B8BDC7', cursor: 'pointer' }} aria-label="Fechar">
                                    <X size={24} />
                                </button>
                            </div>

                            <form onSubmit={handleSaveProduct}>
                                <div style={styles.inputGroup}>
                                    <label style={styles.label}>Nome do Produto</label>
                                    <input
                                        style={styles.input}
                                        value={modalFormData.name}
                                        onChange={e => setModalFormData({ ...modalFormData, name: e.target.value })}
                                        required
                                    />
                                </div>

                                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                                    <div style={styles.inputGroup}>
                                        <label style={styles.label}>Preço (R$)</label>
                                        <input
                                            type="number"
                                            step="0.01"
                                            style={styles.input}
                                            value={modalFormData.price}
                                            onChange={e => setModalFormData({ ...modalFormData, price: e.target.value })}
                                            required
                                        />
                                    </div>
                                    <div style={styles.inputGroup}>
                                        <label style={styles.label}>Tipo</label>
                                        <select
                                            style={styles.select}
                                            value={modalFormData.type}
                                            onChange={e => setModalFormData({ ...modalFormData, type: e.target.value })}
                                        >
                                            <option value="PLUGIN">Plugin</option>
                                            <option value="MOD">Mod</option>
                                            <option value="MAP">Mapa</option>
                                            <option value="TEXTUREPACK">Textura</option>
                                            <option value="SERVER_TEMPLATE">Template</option>
                                        </select>
                                    </div>
                                </div>

                                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                                    <div style={styles.inputGroup}>
                                        <label style={styles.label}>Jogo Associado</label>
                                        <select
                                            style={styles.select}
                                            value={modalFormData.game_id || ''}
                                            onChange={e => setModalFormData({ ...modalFormData, game_id: e.target.value, category_id: '' })}
                                        >
                                            <option value="">Nenhum (Genérico)</option>
                                            {games.map(gameItem => (
                                                <option key={gameItem.game.id} value={gameItem.game.id}>
                                                    {gameItem.game.name}
                                                </option>
                                            ))}
                                        </select>
                                    </div>
                                    <div style={styles.inputGroup}>
                                        <label style={styles.label}>Categoria do Jogo</label>
                                        <select
                                            style={styles.select}
                                            value={modalFormData.category_id || ''}
                                            onChange={e => setModalFormData({ ...modalFormData, category_id: e.target.value })}
                                            disabled={!modalFormData.game_id}
                                        >
                                            <option value="">Nenhuma</option>
                                            {modalFormData.game_id && games
                                                .find(g => g.game.id === modalFormData.game_id)
                                                ?.categories.map(cat => (
                                                    <option key={cat.id} value={cat.id}>
                                                        {cat.name}
                                                    </option>
                                                ))
                                            }
                                        </select>
                                    </div>
                                </div>

                                <div style={styles.inputGroup}>
                                    <label style={styles.label}>Descrição</label>
                                    <textarea
                                        style={{ ...styles.input, minHeight: '100px', resize: 'vertical' }}
                                        value={modalFormData.description}
                                        onChange={e => setModalFormData({ ...modalFormData, description: e.target.value })}
                                        required
                                    />
                                </div>

                                <div style={styles.inputGroup}>
                                    <label style={styles.label}>URL da Imagem</label>
                                    <input
                                        style={styles.input}
                                        value={modalFormData.image_url}
                                        onChange={e => setModalFormData({ ...modalFormData, image_url: e.target.value })}
                                        placeholder="https://..."
                                    />
                                </div>

                                {/* Download URL or File Selection */}
                                <div style={styles.inputGroup}>
                                    <label style={styles.label}>Download</label>
                                    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                                        {/* Option to use existing file */}
                                        <div style={{ marginBottom: '1rem' }}>
                                            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.5rem' }}>
                                                <input
                                                    type="radio"
                                                    id="useFile"
                                                    name="downloadMethod"
                                                    checked={selectedFileId !== null}
                                                    onChange={() => {
                                                        if (availableFiles.length > 0) {
                                                            setSelectedFileId(availableFiles[0]?.id || '');
                                                        }
                                                        setModalFormData({ ...modalFormData, download_url: '' });
                                                    }}
                                                />
                                                <label htmlFor="useFile" style={{ ...styles.label, cursor: 'pointer', marginBottom: 0 }}>Arquivo existente</label>
                                            </div>
                                            {selectedFileId !== null && availableFiles.length > 0 && (
                                                <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                                                    <select
                                                        style={{ ...styles.select, flex: 1 }}
                                                        value={selectedFileId || ''}
                                                        onChange={(e) => setSelectedFileId(e.target.value || null)}
                                                    >
                                                        <option value="">Selecione um arquivo</option>
                                                        {availableFiles.map(file => (
                                                            <option key={file.id} value={file.id}>
                                                                {file.name} ({file.file_type})
                                                            </option>
                                                        ))}
                                                    </select>
                                                </div>
                                            )}
                                            {availableFiles.length === 0 && (
                                                <p style={{ color: '#B8BDC7', fontSize: '0.9rem', marginTop: '0.5rem' }}>Nenhum arquivo disponível. Faça upload primeiro.</p>
                                            )}
                                        </div>

                                        {/* Option to use URL */}
                                        <div>
                                            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.5rem' }}>
                                                <input
                                                    type="radio"
                                                    id="useUrl"
                                                    name="downloadMethod"
                                                    checked={selectedFileId === null}
                                                    onChange={() => {
                                                        setSelectedFileId(null);
                                                    }}
                                                />
                                                <label htmlFor="useUrl" style={{ ...styles.label, cursor: 'pointer', marginBottom: 0 }}>URL externa</label>
                                            </div>
                                            {selectedFileId === null && (
                                                <input
                                                    style={styles.input}
                                                    value={modalFormData.download_url}
                                                    onChange={e => setModalFormData({ ...modalFormData, download_url: e.target.value })}
                                                    placeholder="https://..."
                                                />
                                            )}
                                        </div>

                                        {/* File Upload Section */}
                                        <div style={{ marginTop: '1.5rem', borderTop: '1px solid rgba(255,255,255,0.1)', paddingTop: '1.5rem' }}>
                                            <h4 style={{ ...styles.label, marginBottom: '0.75rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                                                <Upload size={16} /> Upload de Novo Arquivo
                                            </h4>
                                            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                                                <div style={{ display: 'flex', gap: '0.5rem' }}>
                                                    <input
                                                        type="file"
                                                        onChange={handleFileChange}
                                                        style={{ flex: 1, padding: '0.5rem', background: 'rgba(0, 0, 0, 0.2)', border: '1px solid rgba(255, 255, 255, 0.1)', borderRadius: '8px', color: '#fff' }}
                                                    />
                                                </div>
                                                {fileUpload.file && (
                                                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                                                        <input
                                                            type="text"
                                                            value={fileUpload.name}
                                                            onChange={(e) => setFileUpload({ ...fileUpload, name: e.target.value })}
                                                            style={{ flex: 1, padding: '0.5rem', background: 'rgba(0, 0, 0, 0.2)', border: '1px solid rgba(255, 255, 255, 0.1)', borderRadius: '8px', color: '#fff' }}
                                                            placeholder="Nome para o arquivo"
                                                        />
                                                        <button
                                                            type="button"
                                                            onClick={handleFileUpload}
                                                            disabled={fileUpload.isUploading}
                                                            style={{
                                                                background: 'var(--gradient-primary)',
                                                                color: 'white',
                                                                border: 'none',
                                                                padding: 'var(--btn-padding-sm)',
                                                                borderRadius: '0.5rem',
                                                                cursor: 'pointer'
                                                            }}
                                                        >
                                                            {fileUpload.isUploading ? 'Enviando...' : 'Upload'}
                                                        </button>
                                                    </div>
                                                )}
                                                {fileUpload.uploadedFile && (
                                                    <div style={{ padding: '0.75rem', background: 'rgba(34, 197, 94, 0.1)', borderRadius: '0.5rem', border: '1px solid rgba(34, 197, 94, 0.3)', color: '#22C55E', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                                                        <File size={16} /> Arquivo "{fileUpload.uploadedFile.name}" enviado com sucesso!
                                                    </div>
                                                )}
                                            </div>
                                        </div>
                                    </div>
                                </div>

                                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                                    <div style={styles.inputGroup}>
                                        <label style={styles.label}>Estoque</label>
                                        <input
                                            type="number"
                                            style={styles.input}
                                            value={modalFormData.stock_quantity}
                                            onChange={e => setModalFormData({ ...modalFormData, stock_quantity: e.target.value })}
                                        />
                                    </div>
                                    <div style={{ ...styles.inputGroup, display: 'flex', alignItems: 'center', gap: '0.5rem', paddingTop: '1.5rem' }}>
                                        <input
                                            type="checkbox"
                                            checked={modalFormData.is_exclusive}
                                            onChange={e => setModalFormData({ ...modalFormData, is_exclusive: e.target.checked })}
                                            style={{ width: '20px', height: '20px' }}
                                        />
                                        <label style={{ ...styles.label, marginBottom: 0 }}>Produto Exclusivo?</label>
                                    </div>
                                </div>

                                <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '1rem', marginTop: '2rem' }}>
                                    <button
                                        type="button"
                                        onClick={handleCloseModal}
                                        style={{ ...styles.primaryButton, background: 'transparent', border: '1px solid rgba(255,255,255,0.1)' }}
                                    >
                                        Cancelar
                                    </button>
                                    <button type="submit" style={styles.primaryButton}>
                                        {editingProduct ? 'Salvar Alterações' : 'Criar Produto'}
                                    </button>
                                </div>
                            </form>
                        </motion.div>
                    </motion.div>
                )}
            </AnimatePresence>

            {/* Plan Modal */}
            <AnimatePresence>
                {isPlanModalOpen && (
                    <motion.div
                        style={styles.modalOverlay}
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        exit={{ opacity: 0 }}
                    >
                        <motion.div
                            style={styles.modalContent}
                            initial={{ scale: 0.9, y: 20 }}
                            animate={{ scale: 1, y: 0 }}
                            exit={{ scale: 0.9, y: 20 }}
                        >
                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' }}>
                                <h2 style={{ fontSize: 'var(--title-h4)', fontWeight: 700 }}>
                                    {editingPlan ? 'Editar Plano' : 'Novo Plano'}
                                </h2>
                                <button onClick={handleClosePlanModal} style={{ background: 'transparent', border: 'none', color: '#B8BDC7', cursor: 'pointer' }} aria-label="Fechar">
                                    <X size={24} />
                                </button>
                            </div>

                            <form onSubmit={handleSavePlan}>
                                <div style={styles.inputGroup}>
                                    <label style={styles.label}>Nome do Plano</label>
                                    <input
                                        style={styles.input}
                                        value={planData.name}
                                        onChange={e => setPlanData({ ...planData, name: e.target.value })}
                                        required
                                    />
                                </div>

                                <div style={styles.inputGroup}>
                                    <label style={styles.label}>Imagem do Plano (Opcional)</label>
                                    <label
                                        style={{
                                            ...styles.input,
                                            cursor: 'pointer',
                                            display: 'flex',
                                            alignItems: 'center',
                                            gap: '0.5rem',
                                            justifyContent: 'center',
                                            border: '1px dashed rgba(255,255,255,0.2)'
                                        }}
                                    >
                                        <input
                                            type="file"
                                            accept="image/*"
                                            onChange={handlePlanImageUpload}
                                            style={{ display: 'none' }}
                                        />
                                        {uploadingPlanImage ? (
                                            <><Loader2 className="animate-spin" size={20} /> Uploading...</>
                                        ) : planData.imageUrl ? (
                                            <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', width: '100%' }}>
                                                <img
                                                    src={planData.imageUrl}
                                                    alt="Preview"
                                                    style={{ width: 40, height: 40, borderRadius: 4, objectFit: 'cover' }}
                                                />
                                                <span style={{ flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', fontSize: '0.9rem' }}>
                                                    Imagem carregada (Clique para alterar)
                                                </span>
                                            </div>
                                        ) : (
                                            <><Upload size={20} /> Carregar Imagem</>
                                        )}
                                    </label>
                                </div>

                                <div style={styles.inputGroup}>
                                    <label style={styles.label}>Preço (R$)</label>
                                    <input
                                        type="number"
                                        step="0.01"
                                        style={styles.input}
                                        value={planData.price}
                                        onChange={e => setPlanData({ ...planData, price: e.target.value })}
                                        required
                                    />
                                </div>

                                <div style={styles.inputGroup}>
                                    <label style={styles.label}>Descrição</label>
                                    <textarea
                                        style={{ ...styles.input, minHeight: '80px', resize: 'vertical' }}
                                        value={planData.description}
                                        onChange={e => setPlanData({ ...planData, description: e.target.value })}
                                        required
                                    />
                                </div>

                                <div style={styles.inputGroup}>
                                    <label style={styles.label}>Features (uma por linha)</label>
                                    <textarea
                                        style={{ ...styles.input, minHeight: '150px', resize: 'vertical', fontFamily: 'monospace' }}
                                        value={planData.features.join('\n')}
                                        onChange={e => setPlanData({ ...planData, features: e.target.value.split('\n') })}
                                        placeholder="Feature 1&#10;Feature 2&#10;Feature 3"
                                    />
                                </div>

                                <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '1rem', marginTop: '2rem' }}>
                                    <button
                                        type="button"
                                        onClick={handleClosePlanModal}
                                        style={{ ...styles.primaryButton, background: 'transparent', border: '1px solid rgba(255,255,255,0.1)' }}
                                    >
                                        Cancelar
                                    </button>
                                    <button type="submit" style={styles.primaryButton}>
                                        {editingPlan ? 'Salvar Alterações' : 'Criar Plano'}
                                    </button>
                                </div>
                            </form>
                        </motion.div>
                    </motion.div>
                )}
            </AnimatePresence>
            {/* Game Modal */}
            <AnimatePresence>
                {isGameModalOpen && (
                    <motion.div
                        style={styles.modalOverlay}
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        exit={{ opacity: 0 }}
                    >
                        <motion.div
                            style={styles.modalContent}
                            initial={{ scale: 0.9, y: 20 }}
                            animate={{ scale: 1, y: 0 }}
                            exit={{ scale: 0.9, y: 20 }}
                        >
                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' }}>
                                <h2 style={{ fontSize: 'var(--title-h4)', fontWeight: 700 }}>
                                    {editingGame ? 'Editar Jogo' : 'Novo Jogo'}
                                </h2>
                                <button onClick={handleCloseGameModal} style={{ background: 'transparent', border: 'none', color: '#B8BDC7', cursor: 'pointer' }} aria-label="Fechar">
                                    <X size={24} />
                                </button>
                            </div>

                            <form onSubmit={handleSaveGame}>
                                <div style={styles.inputGroup}>
                                    <label style={styles.label}>Nome do Jogo</label>
                                    <input
                                        style={styles.input}
                                        value={gameFormData.name}
                                        onChange={e => setGameFormData({ ...gameFormData, name: e.target.value })}
                                        required
                                    />
                                </div>
                                <div style={styles.inputGroup}>
                                    <label style={styles.label}>Slug (URL amigável)</label>
                                    <input
                                        style={styles.input}
                                        value={gameFormData.slug}
                                        onChange={e => setGameFormData({ ...gameFormData, slug: e.target.value })}
                                        required
                                    />
                                </div>
                                <div style={styles.inputGroup}>
                                    <label style={styles.label}>URL do Ícone (Opcional)</label>
                                    <input
                                        style={styles.input}
                                        value={gameFormData.icon_url}
                                        onChange={e => setGameFormData({ ...gameFormData, icon_url: e.target.value })}
                                        placeholder="https://..."
                                    />
                                </div>
                                <div style={styles.inputGroup}>
                                    <label style={styles.label}>Ordem de Exibição</label>
                                    <input
                                        type="number"
                                        style={styles.input}
                                        value={gameFormData.display_order}
                                        onChange={e => setGameFormData({ ...gameFormData, display_order: e.target.value })}
                                        required
                                    />
                                </div>

                                <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '1rem', marginTop: '2rem' }}>
                                    <button
                                        type="button"
                                        onClick={handleCloseGameModal}
                                        style={{ ...styles.primaryButton, background: 'transparent', border: '1px solid rgba(255,255,255,0.1)' }}
                                    >
                                        Cancelar
                                    </button>
                                    <button type="submit" style={styles.primaryButton}>
                                        {editingGame ? 'Salvar Alterações' : 'Criar Jogo'}
                                    </button>
                                </div>
                            </form>
                        </motion.div>
                    </motion.div>
                )}
            </AnimatePresence>

            {/* Category Modal */}
            <AnimatePresence>
                {isCategoryModalOpen && (
                    <motion.div
                        style={styles.modalOverlay}
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        exit={{ opacity: 0 }}
                    >
                        <motion.div
                            style={styles.modalContent}
                            initial={{ scale: 0.9, y: 20 }}
                            animate={{ scale: 1, y: 0 }}
                            exit={{ scale: 0.9, y: 20 }}
                        >
                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' }}>
                                <h2 style={{ fontSize: 'var(--title-h4)', fontWeight: 700 }}>
                                    {editingCategory ? 'Editar Categoria' : 'Nova Categoria'}
                                </h2>
                                <button onClick={handleCloseCategoryModal} style={{ background: 'transparent', border: 'none', color: '#B8BDC7', cursor: 'pointer' }} aria-label="Fechar">
                                    <X size={24} />
                                </button>
                            </div>

                            <form onSubmit={handleSaveCategory}>
                                <div style={styles.inputGroup}>
                                    <label style={styles.label}>Jogo Associado</label>
                                    <select
                                        style={styles.select}
                                        value={categoryFormData.game_id}
                                        onChange={e => setCategoryFormData({ ...categoryFormData, game_id: e.target.value })}
                                        disabled={!editingCategory && games.length === 0}
                                        required
                                    >
                                        {games.map(gameItem => (
                                            <option key={gameItem.game.id} value={gameItem.game.id}>
                                                {gameItem.game.name}
                                            </option>
                                        ))}
                                    </select>
                                </div>
                                <div style={styles.inputGroup}>
                                    <label style={styles.label}>Nome da Categoria</label>
                                    <input
                                        style={styles.input}
                                        value={categoryFormData.name}
                                        onChange={e => setCategoryFormData({ ...categoryFormData, name: e.target.value })}
                                        required
                                    />
                                </div>
                                <div style={styles.inputGroup}>
                                    <label style={styles.label}>Slug (URL amigável)</label>
                                    <input
                                        style={styles.input}
                                        value={categoryFormData.slug}
                                        onChange={e => setCategoryFormData({ ...categoryFormData, slug: e.target.value })}
                                        required
                                    />
                                </div>
                                <div style={styles.inputGroup}>
                                    <label style={styles.label}>Ordem de Exibição</label>
                                    <input
                                        type="number"
                                        style={styles.input}
                                        value={categoryFormData.display_order}
                                        onChange={e => setCategoryFormData({ ...categoryFormData, display_order: e.target.value })}
                                        required
                                    />
                                </div>

                                <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '1rem', marginTop: '2rem' }}>
                                    <button
                                        type="button"
                                        onClick={handleCloseCategoryModal}
                                        style={{ ...styles.primaryButton, background: 'transparent', border: '1px solid rgba(255,255,255,0.1)' }}
                                    >
                                        Cancelar
                                    </button>
                                    <button type="submit" style={styles.primaryButton}>
                                        {editingCategory ? 'Salvar Alterações' : 'Criar Categoria'}
                                    </button>
                                </div>
                            </form>
                        </motion.div>
                    </motion.div>
                )}
            </AnimatePresence>
        </div>
    );
};

export default AdminCatalog;

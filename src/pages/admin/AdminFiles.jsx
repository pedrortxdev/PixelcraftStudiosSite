import React, { useState, useEffect, useRef } from 'react';
import { motion } from 'framer-motion';
import {
    Loader2, Download, Trash2, Search, File, Shield, Users, Package,
    Link, Clock, RefreshCw, Eye, Copy, Check, X, Lock, Globe, Upload,
    Key
} from 'lucide-react';
import { filesAPI, adminAPI, productsAPI } from '../../services/api';

const AdminFiles = () => {
    const [files, setFiles] = useState([]);
    const [loading, setLoading] = useState(true);
    const [page, setPage] = useState(1);
    const [totalPages, setTotalPages] = useState(1);
    const [search, setSearch] = useState('');
    const [selectedFile, setSelectedFile] = useState(null);
    const [permissions, setPermissions] = useState(null);
    const [showPermissionsModal, setShowPermissionsModal] = useState(false);
    const [accessType, setAccessType] = useState('PRIVATE');
    const [selectedRoles, setSelectedRoles] = useState([]);
    const [selectedProducts, setSelectedProducts] = useState([]);
    const [products, setProducts] = useState([]);
    const [publicLink, setPublicLink] = useState('');
    const [copied, setCopied] = useState(false);
    const [maxDownloads, setMaxDownloads] = useState('');
    const [expiresAt, setExpiresAt] = useState('');
    const [uploading, setUploading] = useState(false);
    const [uploadFile, setUploadFile] = useState(null);
    const [uploadName, setUploadName] = useState('');
    const [showUploadModal, setShowUploadModal] = useState(false);
    const [showOneTimeLinkModal, setShowOneTimeLinkModal] = useState(false);
    const [oneTimeLinkData, setOneTimeLinkData] = useState(null);
    const [oneTimeExpiresIn, setOneTimeExpiresIn] = useState(15);
    const [oneTimeMaxDownloads, setOneTimeMaxDownloads] = useState(1);
    const fileInputRef = useRef(null);

    const ROLES = [
        { value: 'DIRECTION', label: 'Direction', color: '#E01A4F' },
        { value: 'ENGINEERING', label: 'Engineering', color: '#FF6B35' },
        { value: 'DEVELOPMENT', label: 'Development', color: '#583AFF' },
        { value: 'ADMIN', label: 'Admin', color: '#1AD2FF' },
        { value: 'SUPPORT', label: 'Support', color: '#22C55E' },
    ];

    const handleDownload = async (fileId) => {
        try {
            const token = localStorage.getItem('pixelcraft_token');
            const apiUrl = import.meta.env.VITE_API_URL || 'https://api.pixelcraft-studio.store/api/v1';
            const response = await fetch(`${apiUrl}/files/${fileId}/download`, {
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });

            if (!response.ok) {
                throw new Error('Download failed');
            }

            // Get filename from Content-Disposition header
            const contentDisposition = response.headers.get('Content-Disposition');
            let filename = 'download';
            
            if (contentDisposition) {
                // Try to extract filename from filename*=UTF-8'' format first (RFC 5987)
                const utf8Match = contentDisposition.match(/filename\*=UTF-8''([^;]+)/i);
                if (utf8Match) {
                    filename = decodeURIComponent(utf8Match[1]);
                } else {
                    // Fallback to regular filename="" format
                    const regularMatch = contentDisposition.match(/filename="([^"]+)"/i);
                    if (regularMatch) {
                        filename = regularMatch[1];
                    }
                }
            }

            const blob = await response.blob();
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = filename;
            document.body.appendChild(a);
            a.click();
            window.URL.revokeObjectURL(url);
            document.body.removeChild(a);
        } catch (error) {
            console.error('Download failed:', error);
            alert('Erro ao baixar arquivo');
        }
    };

    const fetchFiles = async () => {
        try {
            setLoading(true);
            const response = await filesAPI.listAllAdmin({
                page,
                page_size: 20,
                search
            });
            setFiles(response.files || []);
            setTotalPages(response.total_pages || 1);
        } catch (error) {
            console.error('Failed to fetch files:', error);
        } finally {
            setLoading(false);
        }
    };

    const fetchProducts = async () => {
        try {
            const response = await productsAPI.getAll();
            setProducts(response.products || []);
        } catch (error) {
            console.error('Failed to fetch products:', error);
        }
    };

    useEffect(() => {
        fetchFiles();
        fetchProducts();
    }, [page, search]);

    const handleDelete = async (id) => {
        if (window.confirm('Tem certeza que deseja excluir este arquivo?')) {
            try {
                await filesAPI.delete(id);
                fetchFiles();
            } catch (error) {
                console.error('Failed to delete file:', error);
                alert('Erro ao excluir arquivo');
            }
        }
    };

    const handleOpenPermissions = async (file) => {
        setSelectedFile(file);
        try {
            const perms = await filesAPI.getPermissions(file.id);
            setPermissions(perms);
            setAccessType(perms.access_type || 'PRIVATE');
            setSelectedRoles(perms.allowed_roles || []);
            setSelectedProducts(perms.allowed_product_ids || []);
            setPublicLink(perms.public_link_url || '');
            setMaxDownloads(perms.max_downloads?.toString() || '');
            setExpiresAt(perms.public_link_expires_at ? 
                new Date(perms.public_link_expires_at).toISOString().slice(0, 16) : '');
            setShowPermissionsModal(true);
        } catch (error) {
            console.error('Failed to fetch permissions:', error);
            alert('Erro ao carregar permissões');
        }
    };

    const handleUpload = async () => {
        if (!uploadFile || !uploadName) {
            alert('Por favor, selecione um arquivo e dê um nome');
            return;
        }

        try {
            setUploading(true);
            await filesAPI.upload(uploadFile, uploadName);
            alert('Arquivo enviado com sucesso!');
            setShowUploadModal(false);
            setUploadFile(null);
            setUploadName('');
            fetchFiles();
        } catch (error) {
            console.error('Upload failed:', error);
            alert('Erro ao fazer upload: ' + (error.message || 'Tente novamente'));
        } finally {
            setUploading(false);
        }
    };

    const handleFileChange = (e) => {
        const file = e.target.files[0];
        if (file) {
            setUploadFile(file);
            if (!uploadName) {
                setUploadName(file.name);
            }
        }
    };

    const handleSavePermissions = async () => {
        if (!selectedFile) return;

        try {
            const updateData = {
                access_type: accessType,
                allowed_roles: accessType === 'ROLE' ? selectedRoles : [],
                allowed_product_ids: accessType === 'PRIVATE' ? selectedProducts : [],
                max_downloads: maxDownloads ? parseInt(maxDownloads) : null,
                public_link_expires_at: expiresAt || null,
            };

            if (accessType === 'ROLE' && selectedRoles.length > 0) {
                updateData.required_role = selectedRoles[0];
            }

            await filesAPI.updatePermissions(selectedFile.id, updateData);
            
            // Regenerate public link if PUBLIC
            if (accessType === 'PUBLIC') {
                const linkResponse = await filesAPI.regeneratePublicLink(selectedFile.id);
                setPublicLink(linkResponse.public_link_url);
            }

            alert('Permissões atualizadas com sucesso!');
            setShowPermissionsModal(false);
        } catch (error) {
            console.error('Failed to update permissions:', error);
            alert('Erro ao atualizar permissões');
        }
    };

    const handleCopyLink = () => {
        navigator.clipboard.writeText(publicLink);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    const handleRegenerateLink = async () => {
        if (!selectedFile) return;
        try {
            const response = await filesAPI.regeneratePublicLink(selectedFile.id);
            setPublicLink(response.public_link_url);
            alert('Link público regenerado com sucesso!');
        } catch (error) {
            console.error('Failed to regenerate link:', error);
            alert('Erro ao regenerar link');
        }
    };

    const handleGenerateOneTimeLink = async () => {
        if (!selectedFile) return;
        try {
            const response = await filesAPI.generateOneTimeLink(selectedFile.id, {
                expires_in_minutes: oneTimeExpiresIn,
                max_downloads: oneTimeMaxDownloads
            });
            setOneTimeLinkData(response);
            alert('Link one-time gerado com sucesso!');
        } catch (error) {
            console.error('Failed to generate one-time link:', error);
            alert('Erro ao gerar link one-time');
        }
    };

    const handleCopyOneTimeLink = () => {
        if (!oneTimeLinkData?.download_url) return;
        navigator.clipboard.writeText(oneTimeLinkData.download_url);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    const toggleRole = (role) => {
        if (selectedRoles.includes(role)) {
            setSelectedRoles(selectedRoles.filter(r => r !== role));
        } else {
            setSelectedRoles([...selectedRoles, role]);
        }
    };

    const toggleProduct = (productId) => {
        if (selectedProducts.includes(productId)) {
            setSelectedProducts(selectedProducts.filter(p => p !== productId));
        } else {
            setSelectedProducts([...selectedProducts, productId]);
        }
    };

    const formatFileSize = (bytes) => {
        if (bytes < 1024) return bytes + ' B';
        if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(2) + ' KB';
        return (bytes / (1024 * 1024)).toFixed(2) + ' MB';
    };

    const getAccessTypeIcon = (type) => {
        switch (type) {
            case 'PUBLIC': return Globe;
            case 'ROLE': return Shield;
            case 'PRIVATE': return Lock;
            default: return Lock;
        }
    };

    const getAccessTypeColor = (type) => {
        switch (type) {
            case 'PUBLIC': return '#22C55E';
            case 'ROLE': return '#583AFF';
            case 'PRIVATE': return '#E01A4F';
            default: return '#6C7384';
        }
    };

    return (
        <div style={{ color: '#F8F9FA' }}>
            <header style={{ marginBottom: '2rem', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div>
                    <h1 style={{
                        fontSize: 'var(--title-h3)',
                        fontWeight: 800,
                        background: 'linear-gradient(135deg, #F8F9FA 0%, #583AFF 50%, #1AD2FF 100%)',
                        WebkitBackgroundClip: 'text',
                        WebkitTextFillColor: 'transparent'
                    }}>
                        Gerenciamento de Arquivos
                    </h1>
                    <p style={{ color: '#B8BDC7', marginTop: '0.5rem' }}>
                        Visualize e gerencie todos os arquivos com controle de acesso e permissões.
                    </p>
                </div>
                <button
                    onClick={() => setShowUploadModal(true)}
                    style={{
                        padding: '0.75rem 1.5rem',
                        borderRadius: '8px',
                        background: 'linear-gradient(135deg, #583AFF 0%, #1AD2FF 100%)',
                        color: '#fff',
                        border: 'none',
                        cursor: 'pointer',
                        fontWeight: 700,
                        display: 'flex',
                        alignItems: 'center',
                        gap: '0.5rem',
                        fontSize: '1rem'
                    }}
                >
                    <Upload size={20} />
                    Novo Upload
                </button>
            </header>

            {/* Search Bar */}
            <div style={{ marginBottom: '1.5rem' }}>
                <div style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '0.5rem',
                    background: 'rgba(255, 255, 255, 0.03)',
                    border: '1px solid rgba(255, 255, 255, 0.1)',
                    borderRadius: '8px',
                    padding: '0.75rem'
                }}>
                    <Search size={20} color="#B8BDC7" />
                    <input
                        type="text"
                        placeholder="Buscar por nome de arquivo..."
                        value={search}
                        onChange={(e) => setSearch(e.target.value)}
                        style={{
                            flex: 1,
                            background: 'transparent',
                            border: 'none',
                            color: '#fff',
                            outline: 'none',
                            fontSize: '1rem'
                        }}
                    />
                </div>
            </div>

            {/* Files Table */}
            {loading ? (
                <div style={{ display: 'flex', justifyContent: 'center', padding: '4rem' }}>
                    <Loader2 size={40} style={{ animation: 'spin 1s linear infinite', color: '#583AFF' }} />
                </div>
            ) : (
                <>
                    <div style={{
                        background: 'rgba(255, 255, 255, 0.03)',
                        border: '1px solid rgba(255, 255, 255, 0.05)',
                        borderRadius: '12px',
                        overflow: 'hidden'
                    }}>
                        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                            <thead>
                                <tr style={{ background: 'rgba(255, 255, 255, 0.05)' }}>
                                    <th style={{ padding: 'var(--btn-padding-md)', textAlign: 'left', color: '#B8BDC7' }}>Nome</th>
                                    <th style={{ padding: 'var(--btn-padding-md)', textAlign: 'left', color: '#B8BDC7' }}>Tipo</th>
                                    <th style={{ padding: 'var(--btn-padding-md)', textAlign: 'left', color: '#B8BDC7' }}>Tamanho</th>
                                    <th style={{ padding: 'var(--btn-padding-md)', textAlign: 'left', color: '#B8BDC7' }}>Acesso</th>
                                    <th style={{ padding: 'var(--btn-padding-md)', textAlign: 'left', color: '#B8BDC7' }}>Downloads</th>
                                    <th style={{ padding: 'var(--btn-padding-md)', textAlign: 'left', color: '#B8BDC7' }}>Usuário</th>
                                    <th style={{ padding: 'var(--btn-padding-md)', textAlign: 'center', color: '#B8BDC7' }}>Ações</th>
                                </tr>
                            </thead>
                            <tbody>
                                {files.map(file => {
                                    const AccessIcon = getAccessTypeIcon(file.access_type || 'PRIVATE');
                                    return (
                                        <motion.tr
                                            key={file.id}
                                            initial={{ opacity: 0 }}
                                            animate={{ opacity: 1 }}
                                            style={{ borderTop: '1px solid rgba(255, 255, 255, 0.05)' }}
                                        >
                                            <td style={{ padding: 'var(--btn-padding-md)' }}>
                                                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                                                    <File size={18} color="#583AFF" />
                                                    {file.name}
                                                </div>
                                            </td>
                                            <td style={{ padding: 'var(--btn-padding-md)' }}>
                                                <span style={{
                                                    padding: '0.25rem 0.75rem',
                                                    borderRadius: '20px',
                                                    fontSize: '0.75rem',
                                                    fontWeight: 700,
                                                    background: 'rgba(88, 58, 255, 0.2)',
                                                    color: '#583AFF',
                                                    border: '1px solid rgba(88, 58, 255, 0.4)'
                                                }}>
                                                    {file.file_type}
                                                </span>
                                            </td>
                                            <td style={{ padding: 'var(--btn-padding-md)', color: '#B8BDC7' }}>
                                                {formatFileSize(file.size)}
                                            </td>
                                            <td style={{ padding: 'var(--btn-padding-md)' }}>
                                                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                                                    <AccessIcon size={16} color={getAccessTypeColor(file.access_type || 'PRIVATE')} />
                                                    <span style={{ 
                                                        fontSize: '0.875rem',
                                                        color: getAccessTypeColor(file.access_type || 'PRIVATE')
                                                    }}>
                                                        {file.access_type || 'PRIVATE'}
                                                    </span>
                                                </div>
                                            </td>
                                            <td style={{ padding: 'var(--btn-padding-md)', color: '#B8BDC7' }}>
                                                {file.download_count || 0}
                                            </td>
                                            <td style={{ padding: 'var(--btn-padding-md)', color: '#B8BDC7' }}>
                                                {file.user_name || file.user_email}
                                            </td>
                                            <td style={{ padding: 'var(--btn-padding-md)' }}>
                                                <div style={{ display: 'flex', gap: '0.5rem', justifyContent: 'center' }}>
                                                    <button
                                                        onClick={() => handleOpenPermissions(file)}
                                                        style={{
                                                            padding: '0.5rem',
                                                            borderRadius: '6px',
                                                            background: 'rgba(88, 58, 255, 0.1)',
                                                            color: '#583AFF',
                                                            border: 'none',
                                                            cursor: 'pointer',
                                                            display: 'flex',
                                                            alignItems: 'center'
                                                        }}
                                                        title="Gerenciar Permissões"
                                                    >
                                                        <Shield size={18} />
                                                    </button>
                                                    <button
                                                        onClick={() => {
                                                            setSelectedFile(file);
                                                            setOneTimeLinkData(null);
                                                            setShowOneTimeLinkModal(true);
                                                        }}
                                                        style={{
                                                            padding: '0.5rem',
                                                            borderRadius: '6px',
                                                            background: 'rgba(26, 210, 255, 0.1)',
                                                            color: '#1AD2FF',
                                                            border: 'none',
                                                            cursor: 'pointer',
                                                            display: 'flex',
                                                            alignItems: 'center'
                                                        }}
                                                        title="Gerar Link One-Time"
                                                    >
                                                        <Key size={18} />
                                                    </button>
                                                    <button
                                                        onClick={() => handleDownload(file.id)}
                                                        style={{
                                                            padding: '0.5rem',
                                                            borderRadius: '6px',
                                                            background: 'rgba(26, 210, 255, 0.1)',
                                                            color: '#1AD2FF',
                                                            border: 'none',
                                                            cursor: 'pointer',
                                                            display: 'flex',
                                                            alignItems: 'center'
                                                        }}
                                                        title="Baixar"
                                                    >
                                                        <Download size={18} />
                                                    </button>
                                                    <button
                                                        onClick={() => handleDelete(file.id)}
                                                        style={{
                                                            padding: '0.5rem',
                                                            borderRadius: '6px',
                                                            background: 'rgba(239, 68, 68, 0.1)',
                                                            color: '#EF4444',
                                                            border: 'none',
                                                            cursor: 'pointer',
                                                            display: 'flex',
                                                            alignItems: 'center'
                                                        }}
                                                        title="Excluir"
                                                    >
                                                        <Trash2 size={18} />
                                                    </button>
                                                </div>
                                            </td>
                                        </motion.tr>
                                    );
                                })}
                            </tbody>
                        </table>
                    </div>

                    {/* Pagination */}
                    {totalPages > 1 && (
                        <div style={{
                            display: 'flex',
                            justifyContent: 'center',
                            gap: '0.5rem',
                            marginTop: '2rem'
                        }}>
                            <button
                                onClick={() => setPage(p => Math.max(1, p - 1))}
                                disabled={page === 1}
                                style={{
                                    padding: 'var(--btn-padding-sm)',
                                    borderRadius: '8px',
                                    background: page === 1 ? 'rgba(255, 255, 255, 0.05)' : 'rgba(88, 58, 255, 0.2)',
                                    color: page === 1 ? '#6C7384' : '#fff',
                                    border: 'none',
                                    cursor: page === 1 ? 'not-allowed' : 'pointer'
                                }}
                            >
                                Anterior
                            </button>
                            <span style={{
                                padding: 'var(--btn-padding-sm)',
                                color: '#B8BDC7',
                                display: 'flex',
                                alignItems: 'center'
                            }}>
                                Página {page} de {totalPages}
                            </span>
                            <button
                                onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                                disabled={page === totalPages}
                                style={{
                                    padding: 'var(--btn-padding-sm)',
                                    borderRadius: '8px',
                                    background: page === totalPages ? 'rgba(255, 255, 255, 0.05)' : 'rgba(88, 58, 255, 0.2)',
                                    color: page === totalPages ? '#6C7384' : '#fff',
                                    border: 'none',
                                    cursor: page === totalPages ? 'not-allowed' : 'pointer'
                                }}
                            >
                                Próxima
                            </button>
                        </div>
                    )}
                </>
            )}

            {/* Permissions Modal */}
            {showPermissionsModal && (
                <div style={{
                    position: 'fixed',
                    top: 0,
                    left: 0,
                    right: 0,
                    bottom: 0,
                    background: 'rgba(0, 0, 0, 0.8)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    zIndex: 1000
                }}>
                    <motion.div
                        initial={{ opacity: 0, scale: 0.95 }}
                        animate={{ opacity: 1, scale: 1 }}
                        style={{
                            background: 'linear-gradient(135deg, rgba(30, 34, 45, 0.95) 0%, rgba(15, 23, 42, 0.98) 100%)',
                            borderRadius: '16px',
                            padding: '2rem',
                            maxWidth: '800px',
                            width: '90%',
                            maxHeight: '90vh',
                            overflow: 'auto',
                            border: '1px solid rgba(255, 255, 255, 0.1)'
                        }}
                    >
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
                            <h2 style={{
                                fontSize: '1.5rem',
                                fontWeight: 700,
                                background: 'linear-gradient(135deg, #F8F9FA 0%, #583AFF 50%, #1AD2FF 100%)',
                                WebkitBackgroundClip: 'text',
                                WebkitTextFillColor: 'transparent'
                            }}>
                                Permissões de Arquivo
                            </h2>
                            <button
                                onClick={() => setShowPermissionsModal(false)}
                                style={{
                                    background: 'transparent',
                                    border: 'none',
                                    color: '#B8BDC7',
                                    cursor: 'pointer',
                                    padding: '0.5rem',
                                    display: 'flex',
                                    alignItems: 'center'
                                }}
                            >
                                <X size={24} />
                            </button>
                        </div>

                        <div style={{ marginBottom: '1.5rem' }}>
                            <p style={{ color: '#B8BDC7', marginBottom: '0.5rem' }}>
                                Arquivo: <strong style={{ color: '#F8F9FA' }}>{selectedFile?.name}</strong>
                            </p>
                            <p style={{ color: '#B8BDC7' }}>
                                Downloads: <strong style={{ color: '#F8F9FA' }}>{selectedFile?.download_count || 0}</strong>
                            </p>
                        </div>

                        {/* Access Type */}
                        <div style={{ marginBottom: '1.5rem' }}>
                            <label style={{ display: 'block', marginBottom: '0.5rem', color: '#B8BDC7' }}>
                                Tipo de Acesso
                            </label>
                            <div style={{ display: 'flex', gap: '0.5rem' }}>
                                <button
                                    onClick={() => setAccessType('PUBLIC')}
                                    style={{
                                        flex: 1,
                                        padding: '0.75rem',
                                        borderRadius: '8px',
                                        border: accessType === 'PUBLIC' ? '2px solid #22C55E' : '1px solid rgba(255, 255, 255, 0.1)',
                                        background: accessType === 'PUBLIC' ? 'rgba(34, 197, 94, 0.1)' : 'rgba(255, 255, 255, 0.03)',
                                        color: accessType === 'PUBLIC' ? '#22C55E' : '#B8BDC7',
                                        cursor: 'pointer',
                                        display: 'flex',
                                        alignItems: 'center',
                                        justifyContent: 'center',
                                        gap: '0.5rem'
                                    }}
                                >
                                    <Globe size={18} />
                                    Público
                                </button>
                                <button
                                    onClick={() => setAccessType('ROLE')}
                                    style={{
                                        flex: 1,
                                        padding: '0.75rem',
                                        borderRadius: '8px',
                                        border: accessType === 'ROLE' ? '2px solid #583AFF' : '1px solid rgba(255, 255, 255, 0.1)',
                                        background: accessType === 'ROLE' ? 'rgba(88, 58, 255, 0.1)' : 'rgba(255, 255, 255, 0.03)',
                                        color: accessType === 'ROLE' ? '#583AFF' : '#B8BDC7',
                                        cursor: 'pointer',
                                        display: 'flex',
                                        alignItems: 'center',
                                        justifyContent: 'center',
                                        gap: '0.5rem'
                                    }}
                                >
                                    <Shield size={18} />
                                    Cargos
                                </button>
                                <button
                                    onClick={() => setAccessType('PRIVATE')}
                                    style={{
                                        flex: 1,
                                        padding: '0.75rem',
                                        borderRadius: '8px',
                                        border: accessType === 'PRIVATE' ? '2px solid #E01A4F' : '1px solid rgba(255, 255, 255, 0.1)',
                                        background: accessType === 'PRIVATE' ? 'rgba(224, 26, 79, 0.1)' : 'rgba(255, 255, 255, 0.03)',
                                        color: accessType === 'PRIVATE' ? '#E01A4F' : '#B8BDC7',
                                        cursor: 'pointer',
                                        display: 'flex',
                                        alignItems: 'center',
                                        justifyContent: 'center',
                                        gap: '0.5rem'
                                    }}
                                >
                                    <Lock size={18} />
                                    Privado
                                </button>
                            </div>
                        </div>

                        {/* Role Selection */}
                        {accessType === 'ROLE' && (
                            <div style={{ marginBottom: '1.5rem' }}>
                                <label style={{ display: 'block', marginBottom: '0.5rem', color: '#B8BDC7' }}>
                                    Cargos Permitidos
                                </label>
                                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                                    {ROLES.map(role => (
                                        <button
                                            key={role.value}
                                            onClick={() => toggleRole(role.value)}
                                            style={{
                                                padding: '0.75rem',
                                                borderRadius: '8px',
                                                border: selectedRoles.includes(role.value) ? `2px solid ${role.color}` : '1px solid rgba(255, 255, 255, 0.1)',
                                                background: selectedRoles.includes(role.value) ? `${role.color}20` : 'rgba(255, 255, 255, 0.03)',
                                                color: selectedRoles.includes(role.value) ? role.color : '#B8BDC7',
                                                cursor: 'pointer',
                                                display: 'flex',
                                                alignItems: 'center',
                                                justifyContent: 'space-between'
                                            }}
                                        >
                                            <span>{role.label}</span>
                                            {selectedRoles.includes(role.value) && <Check size={18} />}
                                        </button>
                                    ))}
                                </div>
                            </div>
                        )}

                        {/* Product Selection */}
                        {accessType === 'PRIVATE' && (
                            <div style={{ marginBottom: '1.5rem' }}>
                                <label style={{ display: 'block', marginBottom: '0.5rem', color: '#B8BDC7' }}>
                                    Produtos Requeridos (opcional)
                                </label>
                                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', maxHeight: '200px', overflow: 'auto' }}>
                                    {products.map(product => (
                                        <button
                                            key={product.id}
                                            onClick={() => toggleProduct(product.id)}
                                            style={{
                                                padding: '0.75rem',
                                                borderRadius: '8px',
                                                border: selectedProducts.includes(product.id) ? '2px solid #1AD2FF' : '1px solid rgba(255, 255, 255, 0.1)',
                                                background: selectedProducts.includes(product.id) ? 'rgba(26, 210, 255, 0.1)' : 'rgba(255, 255, 255, 0.03)',
                                                color: selectedProducts.includes(product.id) ? '#1AD2FF' : '#B8BDC7',
                                                cursor: 'pointer',
                                                display: 'flex',
                                                alignItems: 'center',
                                                justifyContent: 'space-between'
                                            }}
                                        >
                                            <span>{product.name}</span>
                                            {selectedProducts.includes(product.id) && <Check size={18} />}
                                        </button>
                                    ))}
                                </div>
                            </div>
                        )}

                        {/* Public Link */}
                        {accessType === 'PUBLIC' && (
                            <div style={{ marginBottom: '1.5rem' }}>
                                <label style={{ display: 'block', marginBottom: '0.5rem', color: '#B8BDC7' }}>
                                    Link Público
                                </label>
                                {publicLink ? (
                                    <div style={{
                                        display: 'flex',
                                        gap: '0.5rem',
                                        background: 'rgba(255, 255, 255, 0.03)',
                                        padding: '0.75rem',
                                        borderRadius: '8px',
                                        border: '1px solid rgba(255, 255, 255, 0.1)'
                                    }}>
                                        <input
                                            type="text"
                                            value={publicLink}
                                            readOnly
                                            style={{
                                                flex: 1,
                                                background: 'transparent',
                                                border: 'none',
                                                color: '#F8F9FA',
                                                outline: 'none',
                                                fontSize: '0.875rem'
                                            }}
                                        />
                                        <button
                                            onClick={handleCopyLink}
                                            style={{
                                                padding: '0.5rem',
                                                borderRadius: '6px',
                                                background: copied ? 'rgba(34, 197, 94, 0.2)' : 'rgba(88, 58, 255, 0.2)',
                                                color: copied ? '#22C55E' : '#583AFF',
                                                border: 'none',
                                                cursor: 'pointer',
                                                display: 'flex',
                                                alignItems: 'center'
                                            }}
                                        >
                                            {copied ? <Check size={18} /> : <Copy size={18} />}
                                        </button>
                                        <button
                                            onClick={handleRegenerateLink}
                                            style={{
                                                padding: '0.5rem',
                                                borderRadius: '6px',
                                                background: 'rgba(26, 210, 255, 0.2)',
                                                color: '#1AD2FF',
                                                border: 'none',
                                                cursor: 'pointer',
                                                display: 'flex',
                                                alignItems: 'center'
                                            }}
                                        >
                                            <RefreshCw size={18} />
                                        </button>
                                    </div>
                                ) : (
                                    <p style={{ color: '#B8BDC7', fontStyle: 'italic' }}>
                                        Clique em Salvar para gerar o link público
                                    </p>
                                )}
                            </div>
                        )}

                        {/* Download Limits */}
                        <div style={{ marginBottom: '1.5rem' }}>
                            <label style={{ display: 'block', marginBottom: '0.5rem', color: '#B8BDC7' }}>
                                Limite de Downloads (opcional)
                            </label>
                            <input
                                type="number"
                                value={maxDownloads}
                                onChange={(e) => setMaxDownloads(e.target.value)}
                                placeholder="Deixe vazio para ilimitado"
                                style={{
                                    width: '100%',
                                    padding: '0.75rem',
                                    borderRadius: '8px',
                                    background: 'rgba(255, 255, 255, 0.03)',
                                    border: '1px solid rgba(255, 255, 255, 0.1)',
                                    color: '#F8F9FA',
                                    outline: 'none'
                                }}
                            />
                        </div>

                        {/* Expiration */}
                        <div style={{ marginBottom: '1.5rem' }}>
                            <label style={{ display: 'block', marginBottom: '0.5rem', color: '#B8BDC7' }}>
                                Expiração do Link (opcional)
                            </label>
                            <input
                                type="datetime-local"
                                value={expiresAt}
                                onChange={(e) => setExpiresAt(e.target.value)}
                                style={{
                                    width: '100%',
                                    padding: '0.75rem',
                                    borderRadius: '8px',
                                    background: 'rgba(255, 255, 255, 0.03)',
                                    border: '1px solid rgba(255, 255, 255, 0.1)',
                                    color: '#F8F9FA',
                                    outline: 'none'
                                }}
                            />
                        </div>

                        {/* Actions */}
                        <div style={{ display: 'flex', gap: '0.5rem', justifyContent: 'flex-end' }}>
                            <button
                                onClick={() => setShowPermissionsModal(false)}
                                style={{
                                    padding: '0.75rem 1.5rem',
                                    borderRadius: '8px',
                                    background: 'rgba(255, 255, 255, 0.05)',
                                    color: '#B8BDC7',
                                    border: 'none',
                                    cursor: 'pointer'
                                }}
                            >
                                Cancelar
                            </button>
                            <button
                                onClick={handleSavePermissions}
                                style={{
                                    padding: '0.75rem 1.5rem',
                                    borderRadius: '8px',
                                    background: 'linear-gradient(135deg, #583AFF 0%, #1AD2FF 100%)',
                                    color: '#fff',
                                    border: 'none',
                                    cursor: 'pointer',
                                    fontWeight: 700
                                }}
                            >
                                Salvar Permissões
                            </button>
                        </div>
                    </motion.div>
                </div>
            )}

            {/* Upload Modal */}
            {showUploadModal && (
                <div style={{
                    position: 'fixed',
                    top: 0,
                    left: 0,
                    right: 0,
                    bottom: 0,
                    background: 'rgba(0, 0, 0, 0.8)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    zIndex: 1000
                }}>
                    <motion.div
                        initial={{ opacity: 0, scale: 0.95 }}
                        animate={{ opacity: 1, scale: 1 }}
                        style={{
                            background: 'linear-gradient(135deg, rgba(30, 34, 45, 0.95) 0%, rgba(15, 23, 42, 0.98) 100%)',
                            borderRadius: '16px',
                            padding: '2rem',
                            maxWidth: '600px',
                            width: '90%',
                            border: '1px solid rgba(255, 255, 255, 0.1)'
                        }}
                    >
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
                            <h2 style={{
                                fontSize: '1.5rem',
                                fontWeight: 700,
                                background: 'linear-gradient(135deg, #F8F9FA 0%, #583AFF 50%, #1AD2FF 100%)',
                                WebkitBackgroundClip: 'text',
                                WebkitTextFillColor: 'transparent'
                            }}>
                                Upload de Arquivo
                            </h2>
                            <button
                                onClick={() => setShowUploadModal(false)}
                                style={{
                                    background: 'transparent',
                                    border: 'none',
                                    color: '#B8BDC7',
                                    cursor: 'pointer',
                                    padding: '0.5rem',
                                    display: 'flex',
                                    alignItems: 'center'
                                }}
                            >
                                <X size={24} />
                            </button>
                        </div>

                        <div style={{ marginBottom: '1.5rem' }}>
                            <label style={{ display: 'block', marginBottom: '0.5rem', color: '#B8BDC7' }}>
                                Nome do Arquivo
                            </label>
                            <input
                                type="text"
                                value={uploadName}
                                onChange={(e) => setUploadName(e.target.value)}
                                placeholder="Ex: Plugin Premium v1.0"
                                style={{
                                    width: '100%',
                                    padding: '0.75rem',
                                    borderRadius: '8px',
                                    background: 'rgba(255, 255, 255, 0.03)',
                                    border: '1px solid rgba(255, 255, 255, 0.1)',
                                    color: '#F8F9FA',
                                    outline: 'none'
                                }}
                            />
                        </div>

                        <div style={{ marginBottom: '1.5rem' }}>
                            <label style={{ display: 'block', marginBottom: '0.5rem', color: '#B8BDC7' }}>
                                Arquivo
                            </label>
                            <div
                                onClick={() => fileInputRef.current?.click()}
                                style={{
                                    padding: '2rem',
                                    borderRadius: '12px',
                                    border: '2px dashed rgba(88, 58, 255, 0.5)',
                                    background: 'rgba(88, 58, 255, 0.05)',
                                    cursor: 'pointer',
                                    textAlign: 'center',
                                    transition: 'all 0.3s ease'
                                }}
                            >
                                <input
                                    ref={fileInputRef}
                                    type="file"
                                    onChange={handleFileChange}
                                    style={{ display: 'none' }}
                                />
                                <Upload size={48} color="#583AFF" style={{ margin: '0 auto 1rem' }} />
                                {uploadFile ? (
                                    <div>
                                        <p style={{ color: '#F8F9FA', fontWeight: 700 }}>{uploadFile.name}</p>
                                        <p style={{ color: '#B8BDC7', fontSize: '0.875rem' }}>
                                            {(uploadFile.size / 1024 / 1024).toFixed(2)} MB
                                        </p>
                                    </div>
                                ) : (
                                    <div>
                                        <p style={{ color: '#B8BDC7' }}>
                                            Clique para selecionar ou arraste o arquivo aqui
                                        </p>
                                        <p style={{ color: '#6C7384', fontSize: '0.75rem', marginTop: '0.5rem' }}>
                                            Formatos: JAR, ZIP, EXE, PNG, JPG, PDF (Máx. 500MB)
                                        </p>
                                    </div>
                                )}
                            </div>
                        </div>

                        <div style={{ display: 'flex', gap: '0.5rem', justifyContent: 'flex-end' }}>
                            <button
                                onClick={() => setShowUploadModal(false)}
                                style={{
                                    padding: '0.75rem 1.5rem',
                                    borderRadius: '8px',
                                    background: 'rgba(255, 255, 255, 0.05)',
                                    color: '#B8BDC7',
                                    border: 'none',
                                    cursor: 'pointer'
                                }}
                            >
                                Cancelar
                            </button>
                            <button
                                onClick={handleUpload}
                                disabled={!uploadFile || !uploadName || uploading}
                                style={{
                                    padding: '0.75rem 1.5rem',
                                    borderRadius: '8px',
                                    background: uploading || !uploadFile || !uploadName
                                        ? 'rgba(88, 58, 255, 0.3)'
                                        : 'linear-gradient(135deg, #583AFF 0%, #1AD2FF 100%)',
                                    color: '#fff',
                                    border: 'none',
                                    cursor: uploading || !uploadFile || !uploadName ? 'not-allowed' : 'pointer',
                                    fontWeight: 700,
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: '0.5rem'
                                }}
                            >
                                {uploading && <Loader2 size={18} style={{ animation: 'spin 1s linear infinite' }} />}
                                {uploading ? 'Enviando...' : 'Enviar Arquivo'}
                            </button>
                        </div>
                    </motion.div>
                </div>
            )}

            {/* One-Time Link Modal */}
            {showOneTimeLinkModal && (
                <div style={{
                    position: 'fixed',
                    top: 0,
                    left: 0,
                    right: 0,
                    bottom: 0,
                    background: 'rgba(0, 0, 0, 0.8)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    zIndex: 1000
                }}>
                    <motion.div
                        initial={{ opacity: 0, scale: 0.95 }}
                        animate={{ opacity: 1, scale: 1 }}
                        style={{
                            background: 'linear-gradient(135deg, rgba(30, 34, 45, 0.95) 0%, rgba(15, 23, 42, 0.98) 100%)',
                            borderRadius: '16px',
                            padding: '2rem',
                            maxWidth: '600px',
                            width: '90%',
                            border: '1px solid rgba(255, 255, 255, 0.1)'
                        }}
                    >
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
                            <h2 style={{
                                fontSize: '1.5rem',
                                fontWeight: 700,
                                background: 'linear-gradient(135deg, #F8F9FA 0%, #583AFF 50%, #1AD2FF 100%)',
                                WebkitBackgroundClip: 'text',
                                WebkitTextFillColor: 'transparent'
                            }}>
                                Gerar Link One-Time
                            </h2>
                            <button
                                onClick={() => {
                                    setShowOneTimeLinkModal(false);
                                    setOneTimeLinkData(null);
                                }}
                                style={{
                                    background: 'transparent',
                                    border: 'none',
                                    color: '#B8BDC7',
                                    cursor: 'pointer',
                                    padding: '0.5rem',
                                    display: 'flex',
                                    alignItems: 'center'
                                }}
                            >
                                <X size={24} />
                            </button>
                        </div>

                        {selectedFile && (
                            <div style={{ marginBottom: '1.5rem' }}>
                                <p style={{ color: '#B8BDC7', marginBottom: '0.5rem' }}>
                                    Arquivo: <strong style={{ color: '#F8F9FA' }}>{selectedFile.name}</strong>
                                </p>
                                <p style={{ color: '#B8BDC7', fontSize: '0.875rem' }}>
                                    Este link é único e temporário. Após o uso ou expiração, ele será automaticamente invalidado.
                                </p>
                            </div>
                        )}

                        {!oneTimeLinkData ? (
                            <>
                                <div style={{ marginBottom: '1.5rem' }}>
                                    <label style={{ display: 'block', marginBottom: '0.5rem', color: '#B8BDC7' }}>
                                        Tempo de Expiração (minutos)
                                    </label>
                                    <input
                                        type="number"
                                        value={oneTimeExpiresIn}
                                        onChange={(e) => setOneTimeExpiresIn(parseInt(e.target.value) || 15)}
                                        min="1"
                                        max="1440"
                                        style={{
                                            width: '100%',
                                            padding: '0.75rem',
                                            borderRadius: '8px',
                                            background: 'rgba(255, 255, 255, 0.03)',
                                            border: '1px solid rgba(255, 255, 255, 0.1)',
                                            color: '#F8F9FA',
                                            outline: 'none'
                                        }}
                                    />
                                    <p style={{ color: '#6C7384', fontSize: '0.75rem', marginTop: '0.5rem' }}>
                                        Mín: 1 minuto, Máx: 1440 minutos (24 horas)
                                    </p>
                                </div>

                                <div style={{ marginBottom: '1.5rem' }}>
                                    <label style={{ display: 'block', marginBottom: '0.5rem', color: '#B8BDC7' }}>
                                        Máximo de Downloads
                                    </label>
                                    <input
                                        type="number"
                                        value={oneTimeMaxDownloads}
                                        onChange={(e) => setOneTimeMaxDownloads(parseInt(e.target.value) || 1)}
                                        min="1"
                                        max="100"
                                        style={{
                                            width: '100%',
                                            padding: '0.75rem',
                                            borderRadius: '8px',
                                            background: 'rgba(255, 255, 255, 0.03)',
                                            border: '1px solid rgba(255, 255, 255, 0.1)',
                                            color: '#F8F9FA',
                                            outline: 'none'
                                        }}
                                    />
                                    <p style={{ color: '#6C7384', fontSize: '0.75rem', marginTop: '0.5rem' }}>
                                        Máx: 100 downloads
                                    </p>
                                </div>

                                <div style={{ display: 'flex', gap: '0.5rem', justifyContent: 'flex-end' }}>
                                    <button
                                        onClick={() => {
                                            setShowOneTimeLinkModal(false);
                                            setOneTimeLinkData(null);
                                        }}
                                        style={{
                                            padding: '0.75rem 1.5rem',
                                            borderRadius: '8px',
                                            background: 'rgba(255, 255, 255, 0.05)',
                                            color: '#B8BDC7',
                                            border: 'none',
                                            cursor: 'pointer'
                                        }}
                                    >
                                        Cancelar
                                    </button>
                                    <button
                                        onClick={handleGenerateOneTimeLink}
                                        style={{
                                            padding: '0.75rem 1.5rem',
                                            borderRadius: '8px',
                                            background: 'linear-gradient(135deg, #583AFF 0%, #1AD2FF 100%)',
                                            color: '#fff',
                                            border: 'none',
                                            cursor: 'pointer',
                                            fontWeight: 700,
                                            display: 'flex',
                                            alignItems: 'center',
                                            gap: '0.5rem'
                                        }}
                                    >
                                        <Key size={18} />
                                        Gerar Link
                                    </button>
                                </div>
                            </>
                        ) : (
                            <div>
                                <div style={{
                                    padding: '1rem',
                                    borderRadius: '8px',
                                    background: 'rgba(34, 197, 94, 0.1)',
                                    border: '1px solid rgba(34, 197, 94, 0.3)',
                                    marginBottom: '1.5rem'
                                }}>
                                    <p style={{ color: '#22C55E', fontSize: '0.875rem', marginBottom: '0.5rem' }}>
                                        ✅ Link gerado com sucesso!
                                    </p>
                                    <div style={{
                                        display: 'flex',
                                        gap: '0.5rem',
                                        alignItems: 'center',
                                        background: 'rgba(0, 0, 0, 0.3)',
                                        padding: '0.75rem',
                                        borderRadius: '6px'
                                    }}>
                                        <code style={{
                                            flex: 1,
                                            color: '#F8F9FA',
                                            fontSize: '0.75rem',
                                            wordBreak: 'break-all'
                                        }}>
                                            {oneTimeLinkData.download_url}
                                        </code>
                                        <button
                                            onClick={handleCopyOneTimeLink}
                                            style={{
                                                padding: '0.5rem',
                                                borderRadius: '6px',
                                                background: copied ? 'rgba(34, 197, 94, 0.3)' : 'rgba(255, 255, 255, 0.1)',
                                                color: copied ? '#22C55E' : '#F8F9FA',
                                                border: 'none',
                                                cursor: 'pointer',
                                                display: 'flex',
                                                alignItems: 'center'
                                            }}
                                        >
                                            {copied ? <Check size={18} /> : <Copy size={18} />}
                                        </button>
                                    </div>
                                </div>

                                <div style={{
                                    display: 'grid',
                                    gridTemplateColumns: '1fr 1fr',
                                    gap: '1rem',
                                    marginBottom: '1.5rem'
                                }}>
                                    <div style={{
                                        padding: '1rem',
                                        borderRadius: '8px',
                                        background: 'rgba(255, 255, 255, 0.03)',
                                        border: '1px solid rgba(255, 255, 255, 0.1)'
                                    }}>
                                        <p style={{ color: '#B8BDC7', fontSize: '0.75rem' }}>Expira em</p>
                                        <p style={{ color: '#F8F9FA', fontWeight: 700 }}>
                                            {oneTimeExpiresIn} minutos
                                        </p>
                                    </div>
                                    <div style={{
                                        padding: '1rem',
                                        borderRadius: '8px',
                                        background: 'rgba(255, 255, 255, 0.03)',
                                        border: '1px solid rgba(255, 255, 255, 0.1)'
                                    }}>
                                        <p style={{ color: '#B8BDC7', fontSize: '0.75rem' }}>Max Downloads</p>
                                        <p style={{ color: '#F8F9FA', fontWeight: 700 }}>
                                            {oneTimeMaxDownloads}
                                        </p>
                                    </div>
                                </div>

                                <div style={{ display: 'flex', gap: '0.5rem', justifyContent: 'flex-end' }}>
                                    <button
                                        onClick={() => {
                                            setShowOneTimeLinkModal(false);
                                            setOneTimeLinkData(null);
                                        }}
                                        style={{
                                            padding: '0.75rem 1.5rem',
                                            borderRadius: '8px',
                                            background: 'linear-gradient(135deg, #583AFF 0%, #1AD2FF 100%)',
                                            color: '#fff',
                                            border: 'none',
                                            cursor: 'pointer',
                                            fontWeight: 700
                                        }}
                                    >
                                        Fechar
                                    </button>
                                </div>
                            </div>
                        )}
                    </motion.div>
                </div>
            )}
        </div>
    );
};

export default AdminFiles;

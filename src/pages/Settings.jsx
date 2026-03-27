import React, { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import { getAvatarUrl } from '../utils/formatAvatarUrl';
import {
  User,
  Shield,
  MessageCircle,
  Check,
  AlertCircle,
  Loader2,
  Crown,
  Lock,
  Camera,
  Sparkles,
  Wand2,
  Copy,
  Mail,
  Phone,
  Gamepad2
} from 'lucide-react';
import { maskCpf } from '../utils/maskSensitiveData';
import { copyToClipboard } from '../utils/clipboard';
import { useAuth } from '../context/AuthContext';
import DashboardLayout from '../components/DashboardLayout';
import RoleBadge, { getHighestRole, RoleBadgeList } from '../components/RoleBadge';

function Settings() {
  const { user, updateUser, uploadAvatar, generateAIAvatar, loading: isLoading } = useAuth();
  const fileInputRef = React.useRef(null);
  const [formData, setFormData] = useState({
    username: '',
    full_name: '',
    discord_handle: '',
    whatsapp_phone: '',
    preferences: {
      density: 'comfortable',
      font: 'modern',
      backgroundFilter: true
    }
  });
  const [saving, setSaving] = useState(false);
  const [generatingAI, setGeneratingAI] = useState(false);
  const [message, setMessage] = useState({ type: '', text: '' });
  const [initialData, setInitialData] = useState(null);
  const [isAvatarHovered, setIsAvatarHovered] = useState(false);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (user) {
      const userData = {
        username: user.username || '',
        full_name: user.full_name || '',
        discord_handle: user.discord_handle || '',
        whatsapp_phone: user.whatsapp_phone || '',
        preferences: user.preferences || {
          density: 'comfortable',
          font: 'modern',
          backgroundFilter: true
        }
      };
      setFormData(userData);
      setInitialData(userData);
    }
  }, [user]);

  const handleAvatarClick = () => {
    fileInputRef.current?.click();
  };

  const handleFileChange = async (e) => {
    const file = e.target.files?.[0];
    if (!file) return;

    try {
      setSaving(true);
      await uploadAvatar(file);
      setMessage({ type: 'success', text: 'Foto de perfil atualizada!' });
    } catch (err) {
      setMessage({ type: 'error', text: err.message || 'Erro ao atualizar foto.' });
    } finally {
      setSaving(false);
    }
  };

  const handleGenerateAI = async () => {
    const prompt = window.prompt('Descreva como você quer seu avatar (ex: "Um mago épico com túnica azul", "Um guerreiro samurai"):');
    if (!prompt) return;

    try {
      setGeneratingAI(true);
      setMessage({ type: '', text: '' });
      await generateAIAvatar(prompt);
      setMessage({ type: 'success', text: 'Avatar gerado com sucesso!' });
    } catch (err) {
      setMessage({ type: 'error', text: err.message || 'Erro ao gerar avatar.' });
    } finally {
      setGeneratingAI(false);
    }
  };

  const copyReferralCode = async () => {
    if (user?.referral_code) {
      const success = await copyToClipboard(user.referral_code);
      if (success) {
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
      }
    }
  };



  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const handlePreferenceChange = (key, value) => {
    setFormData(prev => ({
      ...prev,
      preferences: {
        ...prev.preferences,
        [key]: value
      }
    }));
  };

  const hasChanges = () => {
    if (!initialData) return false;
    return (
      formData.username !== initialData.username ||
      formData.full_name !== initialData.full_name ||
      formData.discord_handle !== initialData.discord_handle ||
      formData.whatsapp_phone !== initialData.whatsapp_phone ||
      JSON.stringify(formData.preferences) !== JSON.stringify(initialData.preferences)
    );
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!hasChanges()) return;

    setSaving(true);
    setMessage({ type: '', text: '' });

    const payload = {};
    if (formData.username !== initialData.username) payload.username = formData.username;
    if (formData.full_name !== initialData.full_name) payload.full_name = formData.full_name;
    if (formData.discord_handle !== initialData.discord_handle) payload.discord_handle = formData.discord_handle;
    if (formData.whatsapp_phone !== initialData.whatsapp_phone) payload.whatsapp_phone = formData.whatsapp_phone;
    if (JSON.stringify(formData.preferences) !== JSON.stringify(initialData.preferences)) payload.preferences = formData.preferences;

    try {
      await updateUser(payload);
      setMessage({ type: 'success', text: 'Alterações salvas!' });
      setInitialData({ ...formData });
    } catch (err) {
      setMessage({ type: 'error', text: err.message || 'Erro ao salvar.' });
    } finally {
      setSaving(false);
    }
  };

  const handleCancel = () => {
    setFormData({ ...initialData });
    setMessage({ type: '', text: '' });
  };

  if (isLoading) {
    return (
      <DashboardLayout title="Configurações">
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', minHeight: '400px', gap: '1rem' }}>
          <Loader2 size={48} style={{ color: '#583AFF', animation: 'spin 1s linear infinite' }} />
          <p style={{ color: '#B8BDC7' }}>Carregando...</p>
        </div>
      </DashboardLayout>
    );
  }

  const highestRole = user?.highest_role || getHighestRole(user?.roles);

  return (
    <DashboardLayout title="Configurações">
      <div className="settings-main-grid" style={{ gap: '2rem', alignItems: 'start' }}>

        {/* Profile Card - Left Column */}
        <motion.div
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ duration: 0.4 }}
          style={{
            background: 'var(--bg-card)',
            borderRadius: 'var(--radius-xl)',
            padding: '2rem',
            border: '1px solid var(--border-card)',
            position: 'sticky',
            top: '1rem',
          }}
        >
          {/* Avatar */}
          <div style={{ textAlign: 'center', marginBottom: '1.5rem' }}>
            <div
              onClick={handleAvatarClick}
              onMouseEnter={() => setIsAvatarHovered(true)}
              onMouseLeave={() => setIsAvatarHovered(false)}
              style={{
                width: '140px',
                height: '140px',
                borderRadius: '50%',
                margin: '0 auto 1rem',
                position: 'relative',
                cursor: 'pointer',
                background: user?.avatar_url ? 'transparent' : 'var(--gradient-primary)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontSize: '3.5rem',
                fontWeight: 800,
                color: 'white',
                boxShadow: '0 0 60px rgba(88, 58, 255, 0.3)',
                overflow: 'hidden',
              }}
            >
              {user?.avatar_url ? (
                <img src={getAvatarUrl(user.avatar_url)} alt="Avatar de Perfil do Usuário" style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
              ) : (
                user?.full_name?.charAt(0).toUpperCase() || 'U'
              )}

              {/* Hover overlay */}
              <div style={{
                position: 'absolute',
                inset: 0,
                background: 'rgba(0, 0, 0, 0.6)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                opacity: isAvatarHovered ? 1 : 0,
                transition: 'opacity 0.2s',
              }}>
                <Camera size={36} color="white" />
              </div>

              <input
                type="file"
                ref={fileInputRef}
                onChange={handleFileChange}
                accept="image/png, image/jpeg, image/gif, image/webp"
                style={{ display: 'none' }}
              />
            </div>

            {/* AI Generate Button */}
            <motion.button
              onClick={handleGenerateAI}
              disabled={generatingAI}
              whileHover={!generatingAI ? { scale: 1.02 } : {}}
              whileTap={!generatingAI ? { scale: 0.98 } : {}}
              style={{
                background: 'linear-gradient(135deg, rgba(88, 58, 255, 0.15) 0%, rgba(26, 210, 255, 0.15) 100%)',
                border: '1px dashed rgba(88, 58, 255, 0.4)',
                borderRadius: '0.75rem',
                padding: '0.625rem 1.25rem',
                color: '#B8BDC7',
                fontSize: '0.85rem',
                fontWeight: 500,
                display: 'inline-flex',
                alignItems: 'center',
                gap: '0.5rem',
                cursor: generatingAI ? 'not-allowed' : 'pointer',
                opacity: generatingAI ? 0.6 : 1,
              }}
            >
              {generatingAI ? (
                <><Loader2 size={16} style={{ animation: 'spin 1s linear infinite' }} /> Gerando...</>
              ) : (
                <><Sparkles size={16} /> Gerar com IA</>
              )}
            </motion.button>
          </div>

          {/* Name & Email */}
          <div style={{ textAlign: 'center', marginBottom: '1.5rem' }}>
            <h2 style={{ fontSize: '1.4rem', fontWeight: 700, color: '#F8F9FA', marginBottom: '0.25rem' }}>
              {user?.full_name || 'Usuário'}
            </h2>
            <p style={{ fontSize: '0.9rem', color: '#6C7384', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '0.5rem' }}>
              <Mail size={14} /> {user?.email}
            </p>
          </div>

          {/* Role Badge - Clean display without conflicting background */}
          <div style={{ display: 'flex', justifyContent: 'center', marginBottom: '1.5rem' }}>
            {user?.roles?.length > 0 || highestRole ? (
              <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '0.75rem' }}>
                <RoleBadge role={highestRole} size="large" />
                {user?.roles?.length > 1 && (
                  <div style={{ fontSize: '0.75rem', color: '#6C7384' }}>
                    +{user.roles.length - 1} outros cargos
                  </div>
                )}
              </div>
            ) : user?.is_admin ? (
              <div style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: '0.5rem',
                padding: 'var(--btn-padding-sm)',
                background: 'rgba(224, 26, 79, 0.15)',
                border: '1px solid rgba(224, 26, 79, 0.3)',
                borderRadius: '0.5rem',
                color: '#E01A4F',
                fontWeight: 600,
                fontSize: '0.85rem',
              }}>
                <Crown size={16} /> Admin Legado
              </div>
            ) : (
              <div style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: '0.5rem',
                padding: 'var(--btn-padding-sm)',
                background: 'rgba(108, 115, 132, 0.15)',
                border: '1px solid rgba(108, 115, 132, 0.3)',
                borderRadius: '0.5rem',
                color: '#6C7384',
                fontWeight: 500,
                fontSize: '0.85rem',
              }}>
                <User size={16} /> Membro
              </div>
            )}
          </div>

          {/* Referral Code */}
          <div style={{
            background: 'var(--bg-input)',
            borderRadius: 'var(--radius-md)',
            padding: 'var(--btn-padding-md)',
            border: '1px solid var(--border-card)',
          }}>
            <div style={{ fontSize: '0.75rem', color: '#6C7384', marginBottom: '0.5rem', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
              Código de Indicação
            </div>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <span style={{ fontSize: '1.1rem', fontWeight: 700, color: '#F8F9FA', fontFamily: 'monospace' }}>
                {user?.referral_code || '---'}
              </span>
              {user?.referral_code && (
                <motion.button
                  onClick={copyReferralCode}
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                  style={{
                    background: copied ? 'rgba(34, 197, 94, 0.2)' : 'rgba(88, 58, 255, 0.15)',
                    border: 'none',
                    borderRadius: '0.5rem',
                    padding: '0.5rem',
                    cursor: 'pointer',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                  }}
                >
                  {copied ? <Check size={16} color="#22C55E" /> : <Copy size={16} color="#B8BDC7" />}
                </motion.button>
              )}
            </div>
          </div>
        </motion.div>

        {/* Forms - Right Column */}
        <div>
          {/* Message Alert */}
          {message.text && (
            <motion.div
              initial={{ opacity: 0, y: -10 }}
              animate={{ opacity: 1, y: 0 }}
              style={{
                padding: 'var(--btn-padding-md)',
                borderRadius: '0.75rem',
                marginBottom: '1.5rem',
                display: 'flex',
                alignItems: 'center',
                gap: '0.75rem',
                background: message.type === 'success' ? 'rgba(34, 197, 94, 0.1)' : 'rgba(239, 68, 68, 0.1)',
                color: message.type === 'success' ? '#22C55E' : '#EF4444',
                border: `1px solid ${message.type === 'success' ? 'rgba(34, 197, 94, 0.3)' : 'rgba(239, 68, 68, 0.3)'}`,
              }}
            >
              {message.type === 'success' ? <Check size={20} /> : <AlertCircle size={20} />}
              {message.text}
            </motion.div>
          )}

          {/* Personal Info */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1 }}
            style={{
              background: 'var(--bg-card)',
              borderRadius: 'var(--radius-lg)',
              padding: '1.5rem',
              marginBottom: '1rem',
              border: '1px solid var(--border-card)',
            }}
          >
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '1.25rem' }}>
              <div style={{
                width: '36px', height: '36px', borderRadius: '0.5rem',
                background: 'linear-gradient(135deg, rgba(88, 58, 255, 0.2) 0%, rgba(26, 210, 255, 0.2) 100%)',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
              }}>
                <User size={18} color="#583AFF" />
              </div>
              <h3 style={{ fontSize: '1.1rem', fontWeight: 600, color: '#F8F9FA' }}>Informações Pessoais</h3>
            </div>

            <div className="settings-form-grid" style={{ gap: '1rem' }}>
              <div>
                <label htmlFor="username" style={{ display: 'block', fontSize: '0.85rem', color: '#6C7384', marginBottom: '0.5rem' }}>
                  Nome de Usuário
                </label>
                <input id="username"
                  type="text"
                  name="username"
                  value={formData.username}
                  onChange={handleChange}
                  placeholder="@username"
                  style={{
                    width: '100%',
                    padding: '0.875rem 1rem',
                    background: 'var(--bg-input)',
                    border: '1px solid var(--border-input)',
                    borderRadius: 'var(--radius-md)',
                    color: 'var(--text-primary)',
                    fontSize: '0.95rem',
                    outline: 'none',
                    transition: 'border-color 0.2s',
                  }}
                />
              </div>
              <div>
                <label htmlFor="full_name" style={{ display: 'block', fontSize: '0.85rem', color: '#6C7384', marginBottom: '0.5rem' }}>
                  Nome Completo
                </label>
                <input id="full_name"
                  type="text"
                  name="full_name"
                  value={formData.full_name}
                  onChange={handleChange}
                  placeholder="Seu nome"
                  style={{
                    width: '100%',
                    padding: '0.875rem 1rem',
                    background: 'var(--bg-input)',
                    border: '1px solid var(--border-input)',
                    borderRadius: 'var(--radius-md)',
                    color: 'var(--text-primary)',
                    fontSize: '0.95rem',
                    outline: 'none',
                  }}
                />
              </div>
              <div style={{ gridColumn: '1 / -1' }}>
                <label htmlFor="cpf" style={{ display: 'block', fontSize: '0.85rem', color: '#6C7384', marginBottom: '0.5rem' }}>
                  CPF (Documento de Identidade)
                </label>
                <input id="cpf"
                  type="text"
                  name="cpf"
                  value={user?.cpf ? maskCpf(user.cpf, true) : formData.cpf || ''}
                  onChange={handleChange}
                  disabled={!!user?.cpf}
                  placeholder={user?.cpf ? '***.***.***-**' : 'Apenas números'}
                  style={{
                    width: '100%',
                    padding: '0.875rem 1rem',
                    background: 'var(--bg-input)',
                    border: '1px solid var(--border-input)',
                    borderRadius: 'var(--radius-md)',
                    color: user?.cpf ? 'var(--text-muted)' : 'var(--text-primary)',
                    fontSize: '0.95rem',
                    outline: 'none',
                    cursor: user?.cpf ? 'not-allowed' : 'text',
                  }}
                />
              </div>
            </div>
          </motion.div>

          {/* Contact */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.2 }}
            style={{
              background: 'var(--bg-card)',
              borderRadius: 'var(--radius-lg)',
              padding: '1.5rem',
              marginBottom: '1rem',
              border: '1px solid var(--border-card)',
            }}
          >
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '1.25rem' }}>
              <div style={{
                width: '36px', height: '36px', borderRadius: '0.5rem',
                background: 'linear-gradient(135deg, rgba(88, 58, 255, 0.2) 0%, rgba(26, 210, 255, 0.2) 100%)',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
              }}>
                <MessageCircle size={18} color="#1AD2FF" />
              </div>
              <h3 style={{ fontSize: '1.1rem', fontWeight: 600, color: '#F8F9FA' }}>Contato</h3>
            </div>

            <div className="settings-form-grid" style={{ gap: '1rem' }}>
              <div>
                <label htmlFor="discord_handle" style={{ fontSize: '0.85rem', color: '#6C7384', marginBottom: '0.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  <Gamepad2 size={14} /> Discord
                </label>
                <input id="discord_handle"
                  type="text"
                  name="discord_handle"
                  value={formData.discord_handle}
                  onChange={handleChange}
                  placeholder="usuario#1234"
                  style={{
                    width: '100%',
                    padding: '0.875rem 1rem',
                    background: 'var(--bg-input)',
                    border: '1px solid var(--border-input)',
                    borderRadius: 'var(--radius-md)',
                    color: 'var(--text-primary)',
                    fontSize: '0.95rem',
                    outline: 'none',
                  }}
                />
              </div>
              <div>
                <label htmlFor="whatsapp_phone" style={{ fontSize: '0.85rem', color: '#6C7384', marginBottom: '0.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  <Phone size={14} /> WhatsApp
                </label>
                <input id="whatsapp_phone"
                  type="text"
                  name="whatsapp_phone"
                  value={formData.whatsapp_phone}
                  onChange={handleChange}
                  placeholder="+55 11 99999-9999"
                  style={{
                    width: '100%',
                    padding: '0.875rem 1rem',
                    background: 'var(--bg-input)',
                    border: '1px solid var(--border-input)',
                    borderRadius: 'var(--radius-md)',
                    color: 'var(--text-primary)',
                    fontSize: '0.95rem',
                    outline: 'none',
                  }}
                />
              </div>
            </div>
          </motion.div>

          {/* Security */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.3 }}
            style={{
              background: 'var(--bg-card)',
              borderRadius: 'var(--radius-lg)',
              padding: '1.5rem',
              marginBottom: '1rem',
              border: '1px solid var(--border-card)',
            }}
          >
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '1.25rem' }}>
              <div style={{
                width: '36px', height: '36px', borderRadius: '0.5rem',
                background: 'linear-gradient(135deg, rgba(224, 26, 79, 0.2) 0%, rgba(255, 107, 107, 0.2) 100%)',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
              }}>
                <Lock size={18} color="#E01A4F" />
              </div>
              <h3 style={{ fontSize: '1.1rem', fontWeight: 600, color: '#F8F9FA' }}>Segurança</h3>
            </div>

            <div>
              <label style={{ display: 'block', fontSize: '0.85rem', color: '#6C7384', marginBottom: '0.5rem' }}>
                E-mail
              </label>
              <input
                type="email"
                value={user?.email || ''}
                disabled
                style={{
                  width: '100%',
                  padding: '0.875rem 1rem',
                  background: 'rgba(10, 14, 26, 0.3)',
                  border: '1px solid rgba(255, 255, 255, 0.04)',
                  borderRadius: '0.625rem',
                  color: '#6C7384',
                  fontSize: '0.95rem',
                  cursor: 'not-allowed',
                }}
              />
              <p style={{ fontSize: '0.8rem', color: '#4A5060', marginTop: '0.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <Lock size={12} /> O e-mail não pode ser alterado por segurança
              </p>
            </div>
          </motion.div>

          {/* Interface Preferences */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.35 }}
            style={{
              background: 'var(--bg-card)',
              borderRadius: 'var(--radius-lg)',
              padding: '1.5rem',
              marginBottom: '1.5rem',
              border: '1px solid var(--border-card)',
            }}
          >
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '1.25rem' }}>
              <div style={{
                width: '36px', height: '36px', borderRadius: '0.5rem',
                background: 'linear-gradient(135deg, rgba(88, 58, 255, 0.2) 0%, rgba(26, 210, 255, 0.2) 100%)',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
              }}>
                <Wand2 size={18} color="#583AFF" />
              </div>
              <h3 style={{ fontSize: '1.1rem', fontWeight: 600, color: '#F8F9FA' }}>Preferências de Interface</h3>
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
              {/* Density */}
              <div>
                <label style={{ display: 'block', fontSize: '0.85rem', color: '#6C7384', marginBottom: '0.75rem' }}>
                  Densidade do Layout
                </label>
                <div style={{ display: 'flex', gap: '0.5rem', background: 'rgba(0,0,0,0.2)', padding: '0.25rem', borderRadius: '0.75rem', width: 'fit-content' }}>
                  <button 
                    onClick={() => handlePreferenceChange('density', 'minimalist')}
                    style={{
                      padding: '0.5rem 1rem',
                      borderRadius: '0.5rem',
                      border: 'none',
                      background: formData.preferences.density === 'minimalist' ? '#583AFF' : 'transparent',
                      color: formData.preferences.density === 'minimalist' ? 'white' : '#B8BDC7',
                      cursor: 'pointer',
                      fontSize: '0.85rem',
                      fontWeight: 600,
                      transition: 'all 0.2s'
                    }}
                  >Minimalista</button>
                  <button 
                    onClick={() => handlePreferenceChange('density', 'comfortable')}
                    style={{
                      padding: '0.5rem 1rem',
                      borderRadius: '0.5rem',
                      border: 'none',
                      background: formData.preferences.density === 'comfortable' ? '#583AFF' : 'transparent',
                      color: formData.preferences.density === 'comfortable' ? 'white' : '#B8BDC7',
                      cursor: 'pointer',
                      fontSize: '0.85rem',
                      fontWeight: 600,
                      transition: 'all 0.2s'
                    }}
                  >Confortável</button>
                </div>
              </div>

              {/* Font */}
              <div>
                <label style={{ display: 'block', fontSize: '0.85rem', color: '#6C7384', marginBottom: '0.75rem' }}>
                  Estilo da Fonte
                </label>
                <div style={{ display: 'flex', gap: '0.5rem', background: 'rgba(0,0,0,0.2)', padding: '0.25rem', borderRadius: '0.75rem', width: 'fit-content' }}>
                  <button 
                    onClick={() => handlePreferenceChange('font', 'modern')}
                    style={{
                      padding: '0.5rem 1rem',
                      borderRadius: '0.5rem',
                      border: 'none',
                      background: formData.preferences.font === 'modern' ? '#583AFF' : 'transparent',
                      color: formData.preferences.font === 'modern' ? 'white' : '#B8BDC7',
                      cursor: 'pointer',
                      fontSize: '0.85rem',
                      fontWeight: 600,
                      transition: 'all 0.2s'
                    }}
                  >Moderna</button>
                  <button 
                    onClick={() => handlePreferenceChange('font', 'gamer')}
                    style={{
                      padding: '0.5rem 1rem',
                      borderRadius: '0.5rem',
                      border: 'none',
                      background: formData.preferences.font === 'gamer' ? '#583AFF' : 'transparent',
                      color: formData.preferences.font === 'gamer' ? 'white' : '#B8BDC7',
                      cursor: 'pointer',
                      fontSize: '0.85rem',
                      fontWeight: 600,
                      transition: 'all 0.2s'
                    }}
                  >Gamer</button>
                  <button 
                    onClick={() => handlePreferenceChange('font', 'classic')}
                    style={{
                      padding: '0.5rem 1rem',
                      borderRadius: '0.5rem',
                      border: 'none',
                      background: formData.preferences.font === 'classic' ? '#583AFF' : 'transparent',
                      color: formData.preferences.font === 'classic' ? 'white' : '#B8BDC7',
                      cursor: 'pointer',
                      fontSize: '0.85rem',
                      fontWeight: 600,
                      transition: 'all 0.2s'
                    }}
                  >Pixel</button>
                </div>
              </div>

              {/* Background Filter */}
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <div>
                  <div style={{ fontSize: '0.95rem', fontWeight: 600, color: '#F8F9FA' }}>Filtro de Fundo</div>
                  <div style={{ fontSize: '0.8rem', color: '#6C7384' }}>Ativa o efeito pixelado/fosco no fundo</div>
                </div>
                <button 
                  onClick={() => handlePreferenceChange('backgroundFilter', !formData.preferences.backgroundFilter)}
                  style={{
                    width: '48px',
                    height: '24px',
                    borderRadius: '12px',
                    background: formData.preferences.backgroundFilter ? '#583AFF' : 'rgba(255,255,255,0.1)',
                    border: 'none',
                    position: 'relative',
                    cursor: 'pointer',
                    transition: 'all 0.3s'
                  }}
                >
                  <div style={{
                    width: '18px',
                    height: '18px',
                    borderRadius: '50%',
                    background: 'white',
                    position: 'absolute',
                    top: '3px',
                    left: formData.preferences.backgroundFilter ? '27px' : '3px',
                    transition: 'all 0.3s'
                  }} />
                </button>
              </div>
            </div>
          </motion.div>

          {/* Action Buttons */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.4 }}
            style={{ display: 'flex', gap: '1rem', justifyContent: 'flex-end' }}
          >
            <motion.button
              type="button"
              onClick={handleCancel}
              disabled={!hasChanges() || saving}
              whileHover={hasChanges() && !saving ? { background: 'rgba(255, 255, 255, 0.08)' } : {}}
              style={{
                padding: '0.875rem 1.5rem',
                background: 'transparent',
                border: '1px solid rgba(255, 255, 255, 0.1)',
                borderRadius: '0.625rem',
                color: '#B8BDC7',
                fontWeight: 500,
                fontSize: '0.95rem',
                cursor: hasChanges() && !saving ? 'pointer' : 'not-allowed',
                opacity: hasChanges() && !saving ? 1 : 0.4,
              }}
            >
              Cancelar
            </motion.button>
            <motion.button
              type="submit"
              onClick={handleSubmit}
              disabled={saving || !hasChanges()}
              whileHover={hasChanges() && !saving ? { scale: 1.02, boxShadow: '0 8px 30px rgba(88, 58, 255, 0.4)' } : {}}
              whileTap={hasChanges() && !saving ? { scale: 0.98 } : {}}
              style={{
                padding: '0.875rem 2rem',
                background: 'var(--gradient-primary)',
                border: 'none',
                borderRadius: '0.625rem',
                color: 'white',
                fontWeight: 600,
                fontSize: '0.95rem',
                cursor: hasChanges() && !saving ? 'pointer' : 'not-allowed',
                opacity: hasChanges() && !saving ? 1 : 0.5,
                display: 'inline-flex',
                alignItems: 'center',
                gap: '0.5rem',
                boxShadow: '0 4px 20px rgba(88, 58, 255, 0.25)',
              }}
            >
              {saving ? (
                <><Loader2 size={18} style={{ animation: 'spin 1s linear infinite' }} /> Salvando...</>
              ) : (
                <><Check size={18} /> Salvar Alterações</>
              )}
            </motion.button>
          </motion.div>
        </div>
      </div>
    </DashboardLayout >
  );
}

export default Settings;
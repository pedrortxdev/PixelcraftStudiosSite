import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { KeyRound, Shield, Eye, EyeOff, Lock, ArrowRight, CheckCircle2, AlertCircle, Loader2 } from 'lucide-react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { authAPI } from '../services/api';
import DashboardLayout from '../components/DashboardLayout';
import LoadingSpinner from '../components/shared/LoadingSpinner';

const ResetPassword = () => {
    const [searchParams] = useSearchParams();
    const token = searchParams.get('token');
    const navigate = useNavigate();

    const [step, setStep] = useState(token ? 2 : 1);
    const [email, setEmail] = useState('');
    const [validationCode, setValidationCode] = useState('');
    const [newPassword, setNewPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');

    const [showPassword, setShowPassword] = useState(false);
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [error, setError] = useState(null);
    const [success, setSuccess] = useState(null);

    // Send initialization request
    const handleRequestReset = async (e) => {
        e.preventDefault();
        setError(null);
        setSuccess(null);
        setIsSubmitting(true);

        try {
            await authAPI.forgotPassword(email);
            setSuccess('Instruções enviadas! Verifique sua caixa de entrada.');
            setStep(1.5); // Waiting for user to click link or enter code if we support it
        } catch (err) {
            setError(err.message || 'Erro ao solicitar redefinição. Verifique o email.');
        } finally {
            setIsSubmitting(false);
        }
    };

    // Submit new password
    const handleResetPassword = async (e) => {
        e.preventDefault();
        if (newPassword !== confirmPassword) {
            setError('As senhas não coincidem.');
            return;
        }
        setError(null);
        setSuccess(null);
        setIsSubmitting(true);

        try {
            const res = await fetch(`${import.meta.env.VITE_API_URL || 'https://api.pixelcraft-studio.store/api/v1'}/auth/reset-password`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    token: token,
                    code: validationCode,
                    new_password: newPassword
                })
            });

            if (!res.ok) {
                const data = await res.json();
                throw new Error(data.message || data.error || 'Erro ao redefinir senha.');
            }

            setSuccess('Senha redefinida com sucesso! Redirecionando...');
            setTimeout(() => navigate('/login'), 3000);
        } catch (err) {
            setError(err.message || 'Código inválido ou expirado.');
        } finally {
            setIsSubmitting(false);
        }
    };

    const pageVariants = {
        hidden: { opacity: 0, scale: 0.98 },
        visible: { opacity: 1, scale: 1, transition: { duration: 0.4, ease: "easeOut" } }
    };

    return (
        <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh', background: 'var(--bg-primary)' }}>
            <motion.div
                variants={pageVariants}
                initial="hidden"
                animate="visible"
                style={{ width: '100%', maxWidth: '440px', padding: '2rem' }}
                className="glass-card-element"
            >
                <div style={{ textAlign: 'center', marginBottom: '2rem' }}>
                    <div style={{
                        width: '64px', height: '64px', borderRadius: '50%', background: 'rgba(88, 58, 255, 0.1)',
                        display: 'flex', justifyContent: 'center', alignItems: 'center', margin: '0 auto 1.5rem',
                        border: '1px solid rgba(88, 58, 255, 0.2)'
                    }}>
                        <KeyRound size={32} color="#583AFF" />
                    </div>
                    <h2 style={{ fontSize: '1.75rem', fontWeight: 800, color: '#F8F9FA', marginBottom: '0.5rem' }}>
                        Recuperar Senha
                    </h2>
                    <p style={{ color: '#B8BDC7', fontSize: '0.95rem' }}>
                        {step === 1 ? 'Informe seu email para receber as instruções de recuperação.' : 'Redefina sua senha usando o código recebido.'}
                    </p>
                </div>

                {error && (
                    <div style={{ padding: '1rem', background: 'rgba(239, 68, 68, 0.1)', border: '1px solid rgba(239, 68, 68, 0.2)', borderRadius: '0.75rem', color: '#EF4444', display: 'flex', gap: '0.75rem', marginBottom: '1.5rem', alignItems: 'center' }}>
                        <AlertCircle size={20} /> <span style={{ fontSize: '0.9rem' }}>{error}</span>
                    </div>
                )}

                {success && (
                    <div style={{ padding: '1rem', background: 'rgba(34, 197, 94, 0.1)', border: '1px solid rgba(34, 197, 94, 0.2)', borderRadius: '0.75rem', color: '#22C55E', display: 'flex', gap: '0.75rem', marginBottom: '1.5rem', alignItems: 'center' }}>
                        <CheckCircle2 size={20} /> <span style={{ fontSize: '0.9rem' }}>{success}</span>
                    </div>
                )}

                {step === 1 || step === 1.5 ? (
                    <form onSubmit={handleRequestReset} style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
                        <div>
                            <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: 500, color: '#F8F9FA', marginBottom: '0.5rem' }}>E-mail Cadastrado</label>
                            <input
                                type="email"
                                required
                                value={email}
                                onChange={(e) => setEmail(e.target.value)}
                                placeholder="seu.email@exemplo.com"
                                style={{ width: '100%', padding: '0.875rem 1rem', background: 'rgba(10, 14, 26, 0.5)', border: '1px solid rgba(255, 255, 255, 0.1)', borderRadius: '0.75rem', color: '#F8F9FA', outline: 'none', transition: ' border-color 0.2s' }}
                                disabled={isSubmitting || step === 1.5}
                            />
                        </div>
                        {step === 1 && (
                            <button
                                type="submit"
                                disabled={isSubmitting}
                                style={{ width: '100%', padding: '1rem', background: 'var(--gradient-primary)', border: 'none', borderRadius: '0.75rem', color: '#FFF', fontWeight: 600, marginTop: '0.5rem', cursor: isSubmitting ? 'not-allowed' : 'pointer', display: 'flex', justifyContent: 'center', alignItems: 'center', gap: '0.5rem' }}
                            >
                                {isSubmitting ? <Loader2 size={20} style={{ animation: 'spin 1s linear infinite' }} /> : 'Continuar'}
                                {!isSubmitting && <ArrowRight size={20} />}
                            </button>
                        )}
                        <div style={{ textAlign: 'center', marginTop: '1rem' }}>
                            <button type="button" onClick={() => navigate('/login')} style={{ background: 'none', border: 'none', color: '#B8BDC7', fontSize: '0.9rem', cursor: 'pointer', textDecoration: 'underline' }}>
                                Voltar para o Login
                            </button>
                        </div>
                    </form>
                ) : (
                    <form onSubmit={handleResetPassword} style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
                        {/* Input para o Token numérico (caso o usuário não clique no link ou o código precise ser inserido) */}
                        <div>
                            <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: 500, color: '#F8F9FA', marginBottom: '0.5rem' }}>Código de 8 dígitos</label>
                            <input
                                type="text"
                                required
                                value={validationCode}
                                onChange={(e) => setValidationCode(e.target.value.toUpperCase())}
                                placeholder="Ex: A1B2C3D4"
                                maxLength={8}
                                style={{ width: '100%', padding: '0.875rem 1rem', background: 'rgba(10, 14, 26, 0.5)', border: '1px solid rgba(255, 255, 255, 0.1)', borderRadius: '0.75rem', color: '#F8F9FA', letterSpacing: '2px', textAlign: 'center', outline: 'none', textTransform: 'uppercase' }}
                                disabled={isSubmitting}
                            />
                        </div>

                        <div style={{ position: 'relative' }}>
                            <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: 500, color: '#F8F9FA', marginBottom: '0.5rem' }}>Nova Senha</label>
                            <div style={{ position: 'relative' }}>
                                <Lock size={18} color="#8A94A6" style={{ position: 'absolute', left: '1rem', top: '50%', transform: 'translateY(-50%)' }} />
                                <input
                                    type={showPassword ? 'text' : 'password'}
                                    required
                                    value={newPassword}
                                    onChange={(e) => setNewPassword(e.target.value)}
                                    placeholder="Mínimo 8 caracteres"
                                    style={{ width: '100%', padding: '0.875rem 1rem 0.875rem 2.8rem', background: 'rgba(10, 14, 26, 0.5)', border: '1px solid rgba(255, 255, 255, 0.1)', borderRadius: '0.75rem', color: '#F8F9FA', outline: 'none' }}
                                    disabled={isSubmitting}
                                />
                                <button
                                    type="button"
                                    onClick={() => setShowPassword(!showPassword)}
                                    style={{ position: 'absolute', right: '1rem', top: '50%', transform: 'translateY(-50%)', background: 'none', border: 'none', color: '#8A94A6', cursor: 'pointer', padding: 0 }}
                                >
                                    {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                                </button>
                            </div>
                        </div>

                        <div style={{ position: 'relative' }}>
                            <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: 500, color: '#F8F9FA', marginBottom: '0.5rem' }}>Confirmar Senha</label>
                            <div style={{ position: 'relative' }}>
                                <Lock size={18} color="#8A94A6" style={{ position: 'absolute', left: '1rem', top: '50%', transform: 'translateY(-50%)' }} />
                                <input
                                    type={showPassword ? 'text' : 'password'}
                                    required
                                    value={confirmPassword}
                                    onChange={(e) => setConfirmPassword(e.target.value)}
                                    placeholder="Repita a nova senha"
                                    style={{ width: '100%', padding: '0.875rem 1rem 0.875rem 2.8rem', background: 'rgba(10, 14, 26, 0.5)', border: '1px solid rgba(255, 255, 255, 0.1)', borderRadius: '0.75rem', color: '#F8F9FA', outline: 'none' }}
                                    disabled={isSubmitting}
                                />
                            </div>
                        </div>

                        <button
                            type="submit"
                            disabled={isSubmitting || newPassword.length < 6}
                            style={{ width: '100%', padding: '1rem', background: 'var(--gradient-primary)', border: 'none', borderRadius: '0.75rem', color: '#FFF', fontWeight: 600, marginTop: '0.5rem', cursor: isSubmitting ? 'not-allowed' : 'pointer', display: 'flex', justifyContent: 'center', alignItems: 'center', gap: '0.5rem' }}
                        >
                            {isSubmitting ? <Loader2 size={20} style={{ animation: 'spin 1s linear infinite' }} /> : 'Atualizar Senha'}
                            {!isSubmitting && <Shield size={20} />}
                        </button>
                    </form>
                )}
            </motion.div>
        </div>
    );
};

export default ResetPassword;

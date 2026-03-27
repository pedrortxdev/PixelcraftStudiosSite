import React from 'react';

/**
 * ErrorBoundary - Catches JavaScript errors in child component tree
 * Prevents entire app from crashing and shows fallback UI
 */
class ErrorBoundary extends React.Component {
    constructor(props) {
        super(props);
        this.state = { hasError: false, error: null };
    }

    static getDerivedStateFromError(error) {
        return { hasError: true, error };
    }

    componentDidCatch(error, errorInfo) {
        // Log error to monitoring service in production
        console.error('ErrorBoundary caught an error:', error, errorInfo);
    }

    render() {
        if (this.state.hasError) {
            return (
                <div style={{
                    display: 'flex',
                    flexDirection: 'column',
                    alignItems: 'center',
                    justifyContent: 'center',
                    minHeight: '100vh',
                    background: 'linear-gradient(180deg, #0A0E1A 0%, #12182A 100%)',
                    color: '#F8F9FA',
                    padding: '2rem',
                    textAlign: 'center',
                }}>
                    <h1 style={{
                        fontSize: 'var(--title-h3)',
                        marginBottom: '1rem',
                        background: 'var(--gradient-cta)',
                        WebkitBackgroundClip: 'text',
                        WebkitTextFillColor: 'transparent',
                    }}>
                        Ops! Algo deu errado
                    </h1>
                    <p style={{ color: '#B8BDC7', marginBottom: '2rem', maxWidth: '400px' }}>
                        Ocorreu um erro inesperado. Por favor, tente recarregar a página.
                    </p>
                    <button
                        onClick={() => window.location.reload()}
                        style={{
                            padding: 'var(--btn-padding-lg)',
                            background: 'var(--gradient-primary)',
                            border: 'none',
                            borderRadius: '0.5rem',
                            color: 'white',
                            fontWeight: 600,
                            cursor: 'pointer',
                            fontSize: '1rem',
                        }}
                    >
                        Recarregar Página
                    </button>
                </div>
            );
        }

        return this.props.children;
    }
}

export default ErrorBoundary;

import React from 'react';
import { Loader2 } from 'lucide-react';

const LoadingSpinner = ({ message = "Carregando...", size = 32, fullHeight = true }) => {
    return (
        <div style={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            minHeight: fullHeight ? '400px' : 'auto',
            padding: '2rem',
            color: 'var(--text-secondary)'
        }}>
            <Loader2 size={size} className="animate-spin" style={{ color: 'var(--accent-primary)', marginBottom: '1rem' }} />
            <p style={{ fontSize: '0.95rem', fontWeight: 500 }}>{message}</p>
        </div>
    );
};

export default LoadingSpinner;

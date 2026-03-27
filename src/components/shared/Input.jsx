import React from 'react';

const Input = ({ label, id, error, icon: Icon, fullWidth = true, ...props }) => {
    const styles = {
        container: {
            display: 'flex',
            flexDirection: 'column',
            gap: '0.5rem',
            width: fullWidth ? '100%' : 'auto',
            marginBottom: '1rem'
        },
        label: {
            fontSize: '0.85rem',
            color: '#6C7384',
            fontWeight: 500,
            display: 'flex',
            alignItems: 'center',
            gap: '0.5rem'
        },
        inputWrapper: {
            position: 'relative',
            display: 'flex',
            alignItems: 'center'
        },
        input: {
            width: '100%',
            background: 'var(--bg-secondary)',
            border: `1px solid ${error ? 'var(--accent-red)' : 'var(--border-card)'}`,
            borderRadius: 'var(--radius-md)',
            color: 'var(--text-primary)',
            padding: Icon ? '0.75rem 1rem 0.75rem 2.5rem' : '0.75rem 1rem',
            fontSize: '0.95rem',
            transition: 'all 0.3s ease',
            outline: 'none',
        },
        icon: {
            position: 'absolute',
            left: '1rem',
            color: '#6C7384',
            pointerEvents: 'none'
        },
        errorText: {
            color: 'var(--accent-red)',
            fontSize: '0.8rem',
            marginTop: '0.25rem'
        }
    };

    return (
        <div style={styles.container}>
            {label && (
                <label htmlFor={id} style={styles.label}>
                    {label}
                </label>
            )}
            <div style={styles.inputWrapper}>
                {Icon && <Icon size={16} style={styles.icon} />}
                <input
                    id={id}
                    style={styles.input}
                    {...props}
                />
            </div>
            {error && <span style={styles.errorText}>{error}</span>}
        </div>
    );
};

export default Input;

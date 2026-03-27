import React from 'react';
import { motion } from 'framer-motion';
import { useNavigate } from 'react-router-dom';
import { Server, Code, Box, Puzzle } from 'lucide-react';

function HomeBentoCategories() {
    const navigate = useNavigate();

    const categories = [
        {
            id: 'fivem',
            title: 'SERVIDORES PRONTOS',
            subtitle: 'O RP não pode parar',
            icon: <Server size={32} />,
            color: '#E01A4F',
            span: 'col-span-12 md:col-span-8',
            height: 'h-64'
        },
        {
            id: 'scripts',
            title: 'SCRIPTS FIVEM',
            subtitle: 'Sistemas exclusivos',
            icon: <Code size={32} />,
            color: '#583AFF',
            span: 'col-span-12 md:col-span-4',
            height: 'h-64'
        },
        {
            id: 'ragnarok',
            title: 'SISTEMAS RAGNAROK',
            subtitle: 'Classic & Renewal',
            icon: <Box size={32} />,
            color: '#FFD700',
            span: 'col-span-12 md:col-span-6',
            height: 'h-72'
        },
        {
            id: 'minecraft',
            title: 'PLUGINS E MAPAS',
            subtitle: 'Multiplayer Premium',
            icon: <Puzzle size={32} />,
            color: '#1AD2FF',
            span: 'col-span-12 md:col-span-6',
            height: 'h-72'
        }
    ];

    return (
        <section style={{ padding: '0 0 2rem 0', background: 'var(--bg-primary)' }}>
            <div style={{ maxWidth: '1400px', margin: '0 auto', padding: '0 2rem' }}>

                {/* CSS Grid fallback using inline styles since we don't use Tailwind classes directly */}
                <div style={{
                    display: 'grid',
                    gridTemplateColumns: 'repeat(12, 1fr)',
                    gap: '1.5rem',
                }}>
                    {categories.map((cat, index) => {
                        // Calculate spans based on our tailwind-like string
                        const isWide = cat.span.includes('md:col-span-8');
                        const isHalf = cat.span.includes('md:col-span-6');
                        const gridColumn = isWide ? 'span 8' : (isHalf ? 'span 6' : (cat.span.includes('md:col-span-4') ? 'span 4' : 'span 5'));
                        const mobileGridColumn = 'span 12'; // Fallback on mobile via media query handled in CSS if needed, but we'll use inline flex for simplicity if grid fails, but grid is supported.

                        return (
                            <motion.div
                                key={cat.id}
                                onClick={() => navigate(`/loja?game=${cat.id === 'scripts' ? 'fivem' : cat.id}`)}
                                whileHover={{ scale: 0.98, borderColor: cat.color }}
                                whileTap={{ scale: 0.95 }}
                                className="bento-card"
                                style={{
                                    gridColumn: `var(--bento-span, ${gridColumn})`,
                                    minHeight: cat.height === 'h-64' ? '250px' : '300px',
                                    background: 'rgba(255, 255, 255, 0.02)',
                                    border: '1px solid rgba(255, 255, 255, 0.05)',
                                    cursor: 'pointer',
                                    position: 'relative',
                                    overflow: 'hidden',
                                    display: 'flex',
                                    flexDirection: 'column',
                                    justifyContent: 'flex-end',
                                    padding: '2rem',
                                }}
                            >
                                {/* Background Glow */}
                                <div style={{
                                    position: 'absolute',
                                    top: '-50%',
                                    right: '-50%',
                                    width: '100%',
                                    height: '100%',
                                    background: `radial-gradient(circle, ${cat.color}20 0%, transparent 70%)`,
                                    zIndex: 0,
                                    pointerEvents: 'none',
                                }} />

                                {/* Content */}
                                <div style={{ position: 'relative', zIndex: 10, display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                                    <div style={{ color: cat.color, marginBottom: 'auto', alignSelf: 'flex-end', position: 'absolute', top: '-180px', right: '0', opacity: 0.5 }}>
                                        {cat.icon}
                                    </div>
                                    <span style={{
                                        fontFamily: 'var(--font-mono)',
                                        color: 'var(--text-muted)',
                                        fontSize: '0.875rem',
                                        letterSpacing: '1px',
                                        textTransform: 'uppercase'
                                    }}>
                                        {cat.subtitle}
                                    </span>
                                    <h3 style={{
                                        fontFamily: 'var(--font-display)',
                                        fontSize: '3.5rem',
                                        lineHeight: '0.9',
                                        margin: 0,
                                        color: 'var(--text-primary)',
                                        textTransform: 'uppercase'
                                    }}>
                                        {cat.title}
                                    </h3>
                                </div>
                            </motion.div>
                        );
                    })}
                </div>

                {/* Global style override for Bento grid mobile parsing */}
                <style dangerouslySetInnerHTML={{
                    __html: `
          @media (max-width: 900px) {
            .bento-card {
              grid-column: span 12 !important;
              min-height: 200px !important;
            }
          }
        `}} />
            </div>
        </section>
    );
}

export default HomeBentoCategories;

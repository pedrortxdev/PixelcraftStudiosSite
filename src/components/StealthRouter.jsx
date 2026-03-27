import { motion } from 'framer-motion';
import { useNavigate } from 'react-router-dom';
import { useState } from 'react';

function StealthRouter() {
    const navigate = useNavigate();
    const [hoveredNode, setHoveredNode] = useState(null);

    const routers = [
        {
            id: 'fivem',
            title: 'FiveM',
            color: '#E01A4F', // Brand Red
            image: '/downloads/fivem-bg.png', // Fallback aesthetic
        },
        {
            id: 'minecraft',
            title: 'Minecraft',
            color: '#1AD2FF', // Blue/Cyan
            image: '/downloads/minecraft-bg.png',
        },
        {
            id: 'ragnarok',
            title: 'Ragnarok',
            color: '#FFB800', // Yellow/Gold
            image: '/downloads/ragnarok-bg.png', // Assuming
        }
    ];

    const styles = {
        section: {
            padding: '4rem 0 8rem 0',
            background: 'var(--bg-primary)',
            position: 'relative',
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
        },
        title: {
            fontSize: '2rem',
            fontWeight: 800,
            color: 'var(--text-secondary)',
            marginBottom: '3rem',
            letterSpacing: '-0.02em',
            opacity: 0.6,
        },
        pillContainer: {
            display: 'flex',
            gap: '1.5rem',
            justifyContent: 'center',
            position: 'relative',
            zIndex: 10,
            overflowX: 'auto',
            paddingBottom: '1rem', // Space for invisible scrollbar
            WebkitOverflowScrolling: 'touch',
            msOverflowStyle: 'none',  /* IE and Edge */
            scrollbarWidth: 'none',  /* Firefox */
            width: '100%',
            padding: '0 1rem', // padding inside to allow swipe beyond edge
        },
        pill: {
            flexShrink: 0, // Prevent squishing on mobile
            background: 'rgba(255, 255, 255, 0.03)',
            border: '1px solid rgba(255, 255, 255, 0.05)',
            borderRadius: '999px',
            padding: '1rem 3rem',
            fontSize: '1.25rem',
            fontWeight: 600,
            color: 'var(--text-secondary)',
            cursor: 'pointer',
            transition: 'all 0.4s cubic-bezier(0.16, 1, 0.3, 1)',
            backdropFilter: 'blur(10px)',
        },
        bgOverlay: {
            position: 'absolute',
            top: 0, left: 0, right: 0, bottom: 0,
            pointerEvents: 'none',
            zIndex: 1,
            opacity: hoveredNode ? 0.08 : 0,
            transition: 'opacity 0.6s ease',
            backgroundSize: 'cover',
            backgroundPosition: 'center',
            /* The background image will be injected dynamically below */
        }
    };

    return (
        <section style={styles.section}>
            <div
                style={{
                    ...styles.bgOverlay,
                    backgroundImage: hoveredNode ? `url(${routers.find(r => r.id === hoveredNode)?.image})` : 'none',
                }}
            />

            <h2 style={styles.title}>Selecione seu Ecossistema</h2>

            <style dangerouslySetInnerHTML={{
                __html: `
              .hide-stealth-scroll::-webkit-scrollbar {
                display: none;
              }
            `}} />

            <div style={styles.pillContainer} className="hide-stealth-scroll">
                {routers.map((router) => {
                    const isHovered = hoveredNode === router.id;
                    return (
                        <motion.div
                            key={router.id}
                            onHoverStart={() => setHoveredNode(router.id)}
                            onHoverEnd={() => setHoveredNode(null)}
                            onClick={() => navigate(`/shop?game=${router.id}`)}
                            whileHover={{ scale: 1.05 }}
                            whileTap={{ scale: 0.98 }}
                            style={{
                                ...styles.pill,
                                color: isHovered ? '#fff' : 'var(--text-secondary)',
                                borderColor: isHovered ? router.color : 'rgba(255, 255, 255, 0.05)',
                                boxShadow: isHovered ? `0 0 30px ${router.color}40, inset 0 0 20px ${router.color}20` : 'none',
                                background: isHovered ? 'rgba(255, 255, 255, 0.08)' : 'rgba(255, 255, 255, 0.03)'
                            }}
                        >
                            {router.title}
                        </motion.div>
                    );
                })}
            </div>
        </section>
    );
}

export default StealthRouter;

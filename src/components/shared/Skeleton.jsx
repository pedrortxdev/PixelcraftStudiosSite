import React from 'react';
import { motion } from 'framer-motion';

function Skeleton({ width, height, borderRadius = '8px', style = {} }) {
    // Configuração da animação de shimmer (brilho pulsante) que corre pela da esquerda para direita
    const shimmerAnimation = {
        hidden: { x: '-100%' },
        visible: {
            x: '100%',
            transition: {
                repeat: Infinity,
                duration: 1.5,
                ease: 'linear',
            }
        }
    };

    return (
        <div
            style={{
                width: width || '100%',
                height: height || '20px',
                borderRadius: borderRadius,
                background: 'rgba(255, 255, 255, 0.05)',
                position: 'relative',
                overflow: 'hidden',
                border: '1px solid rgba(255, 255, 255, 0.02)',
                ...style
            }}
        >
            <motion.div
                variants={shimmerAnimation}
                initial="hidden"
                animate="visible"
                style={{
                    position: 'absolute',
                    top: 0,
                    left: 0,
                    right: 0,
                    bottom: 0,
                    background: 'linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.08), transparent)',
                    width: '50%',
                }}
            />
        </div>
    );
}

export default Skeleton;

import { motion } from 'framer-motion';
import { Sparkles, Zap, Cpu, Layers } from 'lucide-react';

function HeroIllustration() {
  return (
    <div style={{
      position: 'relative',
      width: '100%',
      height: '600px',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
    }}>
      {/* GLOWING FOUNDATION PLATFORM */}
      <motion.div
        initial={{ opacity: 0, scale: 0.8 }}
        animate={{
          opacity: 1,
          scale: 1,
          boxShadow: [
            '0 0 60px rgba(224, 26, 79, 0.4), 0 0 100px rgba(255, 107, 53, 0.3)',
            '0 0 80px rgba(255, 107, 53, 0.5), 0 0 120px rgba(224, 26, 79, 0.4)',
            '0 0 60px rgba(224, 26, 79, 0.4), 0 0 100px rgba(255, 107, 53, 0.3)',
          ]
        }}
        transition={{
          duration: 0.8,
          boxShadow: { duration: 3, repeat: Infinity, ease: 'easeInOut' }
        }}
        style={{
          position: 'absolute',
          bottom: '15%',
          width: '350px',
          height: '220px',
          background: 'linear-gradient(135deg, rgba(224, 26, 79, 0.15) 0%, rgba(255, 107, 53, 0.15) 50%, rgba(255, 215, 0, 0.15) 100%)',
          borderRadius: '1.5rem',
          border: '3px solid rgba(224, 26, 79, 0.4)',
          backdropFilter: 'blur(10px)',
        }}
      >
        {/* Grid pattern inside foundation */}
        <svg width="100%" height="100%" style={{ position: 'absolute', inset: 0, opacity: 0.3 }}>
          <defs>
            <pattern id="grid" width="30" height="30" patternUnits="userSpaceOnUse">
              <path d="M 30 0 L 0 0 0 30" fill="none" stroke="rgba(224, 26, 79, 0.4)" strokeWidth="1" />
            </pattern>
          </defs>
          <rect width="100%" height="100%" fill="url(#grid)" />
        </svg>
      </motion.div>

      {/* PIXELCRAFT LETTERS - Horizontal animated text */}
      <div style={{
        position: 'absolute',
        top: '65%',
        left: '50%',
        transform: 'translate(-50%, -50%)',
        display: 'flex',
        gap: '0.5rem',
        fontSize: 'var(--title-hero)',
        fontWeight: 900,
        color: '#FFFFFF',
        letterSpacing: '-0.05em',
        zIndex: 10,
      }}>
        {'PIXELCRAFT'.split('').map((letter, i) => (
          <motion.div
            key={i}
            initial={{ opacity: 0, y: -50 }}
            animate={{
              opacity: [0, 1, 1, 0],
              y: [-50, 0, 0, 30],
              scale: [0.8, 1, 1, 0.9],
            }}
            transition={{
              duration: 4,
              delay: i * 0.1,
              repeat: Infinity,
              repeatDelay: 2,
              ease: 'easeInOut',
            }}
            style={{
              background: i < 5
                ? 'var(--gradient-cta)'
                : i < 9
                  ? 'linear-gradient(135deg, #FF6B35 0%, #FFD700 100%)'
                  : 'linear-gradient(135deg, #FFD700 0%, #E01A4F 100%)',
              WebkitBackgroundClip: 'text',
              backgroundClip: 'text',
              WebkitTextFillColor: 'transparent',
              textShadow: '0 0 40px rgba(224, 26, 79, 0.5)',
              filter: 'drop-shadow(0 0 20px rgba(255, 107, 53, 0.6))',
            }}
          >
            {letter}
          </motion.div>
        ))}
      </div>

      {/* ENERGY LINES - Connecting tools to blocks */}
      {[0, 1, 2].map((i) => (
        <motion.div
          key={`line-${i}`}
          initial={{ opacity: 0, pathLength: 0 }}
          animate={{
            opacity: [0.3, 0.6, 0.3],
            pathLength: 1,
          }}
          transition={{
            opacity: { duration: 2, repeat: Infinity, delay: i * 0.3 },
            pathLength: { duration: 1.5, delay: 1.2 + i * 0.2 },
          }}
          style={{
            position: 'absolute',
            top: '40%',
            left: '50%',
            width: '2px',
            height: '150px',
            background: `linear-gradient(180deg, ${i === 0 ? '#E01A4F' : i === 1 ? '#FF6B35' : '#FFD700'}00, ${i === 0 ? '#E01A4F' : i === 1 ? '#FF6B35' : '#FFD700'})`,
            transform: `rotate(${i * 45}deg) translateX(${i * 30}px)`,
            filter: 'blur(1px)',
            boxShadow: `0 0 10px ${i === 0 ? '#E01A4F' : i === 1 ? '#FF6B35' : '#FFD700'}`,
          }}
        />
      ))}

      {/* ASCENDING PARTICLES - Magical construction effect */}
      {[...Array(20)].map((_, i) => (
        <motion.div
          key={`particle-${i}`}
          initial={{ opacity: 0 }}
          animate={{
            y: [0, -280 - i * 15],
            x: [(i % 2 === 0 ? -1 : 1) * (25 + i * 3), (i % 2 === 0 ? -1 : 1) * (35 + i * 3)],
            opacity: [0, 0.8, 0],
            scale: [0, 1.2, 0],
          }}
          transition={{
            duration: 2.5 + i * 0.15,
            repeat: Infinity,
            delay: i * 0.2,
            ease: 'easeOut',
          }}
          style={{
            position: 'absolute',
            bottom: '20%',
            left: '50%',
            width: i % 3 === 0 ? '6px' : '4px',
            height: i % 3 === 0 ? '6px' : '4px',
            background: i % 3 === 0 ? '#E01A4F' : i % 3 === 1 ? '#FF6B35' : '#FFD700',
            borderRadius: i % 2 === 0 ? '50%' : '20%',
            boxShadow: `0 0 15px ${i % 3 === 0 ? '#E01A4F' : i % 3 === 1 ? '#FF6B35' : '#FFD700'}`,
          }}
        />
      ))}

      {/* SPARKLE EFFECTS - Premium touches */}
      {[0, 1, 2, 3].map((i) => (
        <motion.div
          key={`sparkle-${i}`}
          initial={{ opacity: 0, scale: 0 }}
          animate={{
            opacity: [0, 1, 0],
            scale: [0, 1.5, 0],
            rotate: [0, 180, 360],
          }}
          transition={{
            duration: 2,
            repeat: Infinity,
            delay: 1.5 + i * 0.5,
          }}
          style={{
            position: 'absolute',
            top: `${30 + i * 15}%`,
            right: `${15 + i * 8}%`,
          }}
        >
          <Sparkles size={i % 2 === 0 ? 24 : 20} color="#FFD700" fill="#FFD700" />
        </motion.div>
      ))}

      {/* AMBIENT GLOW ORBS - Background depth */}
      <motion.div
        animate={{
          scale: [1, 1.2, 1],
          opacity: [0.2, 0.4, 0.2],
        }}
        transition={{ duration: 5, repeat: Infinity }}
        style={{
          position: 'absolute',
          top: '10%',
          right: '5%',
          width: '250px',
          height: '250px',
          background: 'radial-gradient(circle, rgba(224, 26, 79, 0.3) 0%, transparent 70%)',
          borderRadius: '50%',
          filter: 'blur(60px)',
          pointerEvents: 'none',
        }}
      />

      <motion.div
        animate={{
          scale: [1, 1.3, 1],
          opacity: [0.15, 0.35, 0.15],
        }}
        transition={{ duration: 6, repeat: Infinity }}
        style={{
          position: 'absolute',
          bottom: '5%',
          left: '10%',
          width: '200px',
          height: '200px',
          background: 'radial-gradient(circle, rgba(255, 215, 0, 0.3) 0%, transparent 70%)',
          borderRadius: '50%',
          filter: 'blur(60px)',
          pointerEvents: 'none',
        }}
      />
    </div>
  );
}

export default HeroIllustration;

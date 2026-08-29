'use client';

import React, { useState, useRef, useEffect } from 'react';
import { motion, useInView } from 'framer-motion';

/* ═══════════════════════════════════════════════════════════════
   SPOTLIGHT CARD — Cursor-tracking radial gradient (Motion Primitives style)
   ═══════════════════════════════════════════════════════════════ */
export function SpotlightCard({
  children,
  className = '',
  spotlightColor = 'rgba(139, 92, 246, 0.08)',
  onClick,
}: {
  children: React.ReactNode;
  className?: string;
  spotlightColor?: string;
  onClick?: () => void;
}) {
  const divRef = useRef<HTMLDivElement>(null);
  const [pos, setPos] = useState({ x: 0, y: 0 });
  const [opacity, setOpacity] = useState(0);

  const handleMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
    if (!divRef.current) return;
    const rect = divRef.current.getBoundingClientRect();
    setPos({ x: e.clientX - rect.left, y: e.clientY - rect.top });
  };

  return (
    <motion.div
      ref={divRef}
      onMouseMove={handleMouseMove}
      onMouseEnter={() => setOpacity(1)}
      onMouseLeave={() => setOpacity(0)}
      onClick={onClick}
      whileHover={{ y: -2, transition: { duration: 0.25, ease: 'easeOut' } }}
      style={{
        position: 'relative',
        borderRadius: 'var(--radius-lg)',
        background: 'var(--bg-card)',
        border: '1px solid var(--border-subtle)',
        overflow: 'hidden',
        cursor: onClick ? 'pointer' : 'default',
        transition: 'border-color 0.3s ease',
      }}
      className={className}
    >
      {/* Radial spotlight follower */}
      <div
        style={{
          position: 'absolute',
          inset: -1,
          borderRadius: 'inherit',
          opacity,
          transition: 'opacity 0.3s ease',
          pointerEvents: 'none',
          background: `radial-gradient(400px circle at ${pos.x}px ${pos.y}px, ${spotlightColor}, transparent 70%)`,
        }}
      />
      <div style={{ position: 'relative', zIndex: 1, padding: '28px' }}>
        {children}
      </div>
    </motion.div>
  );
}

/* ═══════════════════════════════════════════════════════════════
   BADGE — Minimal pill badge with pulsing dot
   ═══════════════════════════════════════════════════════════════ */
export function BadgePulse({
  text,
  variant = 'success',
}: {
  text: string;
  variant?: 'success' | 'warning' | 'danger' | 'info' | 'purple';
}) {
  const styles = {
    success: { bg: 'var(--color-emerald-bg)', border: 'var(--color-emerald-border)', text: 'var(--color-emerald)', dot: '#10b981' },
    warning: { bg: 'var(--color-amber-bg)', border: 'var(--color-amber-border)', text: 'var(--color-amber)', dot: '#f59e0b' },
    danger: { bg: 'var(--color-red-bg)', border: 'var(--color-red-border)', text: 'var(--color-red)', dot: '#ef4444' },
    info: { bg: 'var(--color-accent-bg)', border: 'var(--color-accent-border)', text: 'var(--color-accent-light)', dot: '#8b5cf6' },
    purple: { bg: 'var(--color-accent-bg)', border: 'var(--color-accent-border)', text: 'var(--color-accent-light)', dot: '#8b5cf6' },
  }[variant];

  return (
    <span
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: '6px',
        padding: '5px 12px',
        borderRadius: 'var(--radius-full)',
        background: styles.bg,
        border: `1px solid ${styles.border}`,
        color: styles.text,
        fontSize: '12px',
        fontWeight: 500,
        letterSpacing: '0.02em',
      }}
    >
      <span
        style={{
          width: '5px',
          height: '5px',
          borderRadius: '50%',
          backgroundColor: styles.dot,
          boxShadow: `0 0 6px ${styles.dot}`,
        }}
      />
      {text}
    </span>
  );
}

/* ═══════════════════════════════════════════════════════════════
   MOTION BUTTON — Pill-shaped with spring physics
   ═══════════════════════════════════════════════════════════════ */
export function MotionButton({
  children,
  onClick,
  variant = 'primary',
  disabled = false,
  className = '',
  icon,
}: {
  children: React.ReactNode;
  onClick?: () => void;
  variant?: 'primary' | 'secondary' | 'outline' | 'gradient';
  disabled?: boolean;
  className?: string;
  icon?: React.ReactNode;
}) {
  const getStyle = (): React.CSSProperties => {
    switch (variant) {
      case 'primary':
        return {
          background: 'var(--color-accent)',
          color: '#ffffff',
          border: 'none',
        };
      case 'gradient':
        return {
          background: 'linear-gradient(135deg, #8b5cf6 0%, #6366f1 50%, #10b981 100%)',
          color: '#ffffff',
          fontWeight: 700,
          border: 'none',
        };
      case 'outline':
        return {
          background: 'transparent',
          color: 'var(--text-primary)',
          border: '1px solid var(--border-default)',
        };
      default:
        return {
          background: 'rgba(255, 255, 255, 0.06)',
          color: 'var(--text-primary)',
          border: '1px solid var(--border-subtle)',
        };
    }
  };

  return (
    <motion.button
      whileHover={{ scale: disabled ? 1 : 1.02 }}
      whileTap={{ scale: disabled ? 1 : 0.97 }}
      onClick={onClick}
      disabled={disabled}
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        justifyContent: 'center',
        gap: '8px',
        padding: '11px 24px',
        borderRadius: 'var(--radius-full)',
        fontSize: '14px',
        fontWeight: 600,
        fontFamily: 'var(--font-sans)',
        cursor: disabled ? 'not-allowed' : 'pointer',
        opacity: disabled ? 0.5 : 1,
        transition: 'all 0.2s cubic-bezier(0.4, 0, 0.2, 1)',
        ...getStyle(),
      }}
      className={className}
    >
      {icon && <span style={{ display: 'flex' }}>{icon}</span>}
      {children}
    </motion.button>
  );
}

/* ═══════════════════════════════════════════════════════════════
   FADE IN — Scroll-triggered reveal wrapper
   ═══════════════════════════════════════════════════════════════ */
export function FadeIn({
  children,
  delay = 0,
  direction = 'up',
  className = '',
}: {
  children: React.ReactNode;
  delay?: number;
  direction?: 'up' | 'down' | 'left' | 'right' | 'none';
  className?: string;
}) {
  const ref = useRef<HTMLDivElement>(null);
  const isInView = useInView(ref, { once: true, margin: '-60px' });

  const directionOffset = {
    up: { y: 24 },
    down: { y: -24 },
    left: { x: 24 },
    right: { x: -24 },
    none: {},
  }[direction];

  return (
    <motion.div
      ref={ref}
      initial={{ opacity: 0, ...directionOffset }}
      animate={isInView ? { opacity: 1, x: 0, y: 0 } : {}}
      transition={{
        duration: 0.5,
        delay,
        ease: [0.25, 0.4, 0.25, 1],
      }}
      className={className}
    >
      {children}
    </motion.div>
  );
}

/* ═══════════════════════════════════════════════════════════════
   ANIMATED COUNTER — Spring-based number animation
   ═══════════════════════════════════════════════════════════════ */
export function AnimatedCounter({
  value,
  prefix = '',
  suffix = '',
  className = '',
}: {
  value: number;
  prefix?: string;
  suffix?: string;
  className?: string;
}) {
  const ref = useRef<HTMLSpanElement>(null);
  const isInView = useInView(ref, { once: true });
  const [display, setDisplay] = useState(0);

  useEffect(() => {
    if (!isInView) return;
    const duration = 1200;
    const start = Date.now();
    const startVal = 0;

    const tick = () => {
      const elapsed = Date.now() - start;
      const progress = Math.min(elapsed / duration, 1);
      // Ease out cubic
      const eased = 1 - Math.pow(1 - progress, 3);
      setDisplay(Math.round(startVal + (value - startVal) * eased));
      if (progress < 1) requestAnimationFrame(tick);
    };

    requestAnimationFrame(tick);
  }, [isInView, value]);

  return (
    <span ref={ref} className={className}>
      {prefix}{display.toLocaleString()}{suffix}
    </span>
  );
}

/* ═══════════════════════════════════════════════════════════════
   Legacy exports for backward compatibility
   ═══════════════════════════════════════════════════════════════ */
export const GlowCard = SpotlightCard;

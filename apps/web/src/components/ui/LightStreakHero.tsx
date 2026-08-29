'use client';

import React, { useEffect, useRef, useState } from 'react';
import Link from 'next/link';
import { motion } from 'framer-motion';
import {
  TrendingUp,
  ArrowRight,
  Play,
  ShieldCheck,
  Zap,
  Sparkles,
} from 'lucide-react';

export default function LightStreakHero() {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    let animationFrameId: number;
    let width = (canvas.width = window.innerWidth);
    let height = (canvas.height = window.innerHeight);

    const handleResize = () => {
      if (!canvas) return;
      width = canvas.width = window.innerWidth;
      height = canvas.height = window.innerHeight;
    };

    window.addEventListener('resize', handleResize);

    // Multi-layer curved neon energy streaks
    interface Streak {
      yOffset: number;
      speed: number;
      amplitude: number;
      frequency: number;
      phase: number;
      thickness: number;
      color: string;
      glowColor: string;
      blur: number;
      opacity: number;
    }

    const streaks: Streak[] = [
      {
        yOffset: 0.42,
        speed: 0.003,
        amplitude: 60,
        frequency: 0.0018,
        phase: 0,
        thickness: 3.5,
        color: '#ffffff',
        glowColor: '#38bdf8',
        blur: 16,
        opacity: 0.95,
      },
      {
        yOffset: 0.40,
        speed: 0.0025,
        amplitude: 80,
        frequency: 0.0015,
        phase: 1.2,
        thickness: 4,
        color: '#38bdf8',
        glowColor: '#0ea5e9',
        blur: 24,
        opacity: 0.85,
      },
      {
        yOffset: 0.45,
        speed: 0.0018,
        amplitude: 110,
        frequency: 0.0012,
        phase: 2.5,
        thickness: 6,
        color: '#818cf8',
        glowColor: '#6366f1',
        blur: 35,
        opacity: 0.75,
      },
      {
        yOffset: 0.35,
        speed: 0.002,
        amplitude: 95,
        frequency: 0.0014,
        phase: 3.8,
        thickness: 5,
        color: '#c084fc',
        glowColor: '#a855f7',
        blur: 30,
        opacity: 0.7,
      },
      {
        yOffset: 0.50,
        speed: 0.0015,
        amplitude: 130,
        frequency: 0.001,
        phase: 4.9,
        thickness: 7,
        color: '#ec4899',
        glowColor: '#f43f5e',
        blur: 40,
        opacity: 0.55,
      },
      {
        yOffset: 0.38,
        speed: 0.0012,
        amplitude: 150,
        frequency: 0.0008,
        phase: 0.8,
        thickness: 12,
        color: '#3b82f6',
        glowColor: '#1d4ed8',
        blur: 60,
        opacity: 0.4,
      },
    ];

    let t = 0;

    const render = () => {
      ctx.clearRect(0, 0, width, height);

      const isLight = document.documentElement.classList.contains('light');

      // Adaptive background gradient
      const bgGrad = ctx.createRadialGradient(
        width * 0.75,
        height * 0.4,
        50,
        width * 0.75,
        height * 0.4,
        width * 0.85
      );

      if (isLight) {
        bgGrad.addColorStop(0, 'rgba(219, 234, 254, 0.4)');
        bgGrad.addColorStop(0.5, 'rgba(243, 232, 255, 0.25)');
        bgGrad.addColorStop(1, 'rgba(248, 250, 252, 0)');
      } else {
        bgGrad.addColorStop(0, 'rgba(14, 25, 60, 0.45)');
        bgGrad.addColorStop(0.5, 'rgba(6, 12, 32, 0.3)');
        bgGrad.addColorStop(1, 'rgba(3, 7, 18, 0)');
      }

      ctx.fillStyle = bgGrad;
      ctx.fillRect(0, 0, width, height);

      // Render each neon energy streak with glowing bloom
      streaks.forEach((s) => {
        ctx.save();
        ctx.beginPath();

        const baseY = height * s.yOffset;
        ctx.moveTo(0, baseY + Math.sin(t * s.speed + s.phase) * s.amplitude);

        for (let x = 0; x <= width; x += 12) {
          const curveX = x / width;
          const arch = Math.pow(curveX, 2) * -120;
          const wave = Math.sin(x * s.frequency + t * s.speed + s.phase) * s.amplitude * (0.6 + curveX * 0.8);
          ctx.lineTo(x, baseY + arch + wave);
        }

        ctx.strokeStyle = isLight ? (s.color === '#ffffff' ? '#6366f1' : s.color) : s.color;
        ctx.lineWidth = s.thickness;
        ctx.shadowColor = s.glowColor;
        ctx.shadowBlur = isLight ? s.blur * 0.7 : s.blur;
        ctx.globalAlpha = isLight ? s.opacity * 0.75 : s.opacity;
        ctx.stroke();

        if (!isLight) {
          ctx.lineWidth = Math.max(1, s.thickness * 0.4);
          ctx.strokeStyle = '#ffffff';
          ctx.globalAlpha = s.opacity * 0.9;
          ctx.shadowBlur = s.blur * 0.5;
          ctx.stroke();
        }

        ctx.restore();
      });

      // Bright focal optical lens flare point on the main streak
      const flareProgress = (Math.sin(t * 0.0015) + 1) / 2;
      const flareX = width * (0.15 + flareProgress * 0.75);
      const flareY = height * 0.42 - Math.pow(flareX / width, 2) * 120 + Math.sin(flareX * 0.0018 + t * 0.003) * 60;

      ctx.save();
      const flareHalo = ctx.createRadialGradient(flareX, flareY, 0, flareX, flareY, 180);
      if (isLight) {
        flareHalo.addColorStop(0, 'rgba(147, 197, 253, 0.8)');
        flareHalo.addColorStop(0.3, 'rgba(192, 132, 252, 0.4)');
        flareHalo.addColorStop(1, 'rgba(255, 255, 255, 0)');
      } else {
        flareHalo.addColorStop(0, 'rgba(255, 255, 255, 0.95)');
        flareHalo.addColorStop(0.2, 'rgba(56, 189, 248, 0.8)');
        flareHalo.addColorStop(0.5, 'rgba(139, 92, 246, 0.35)');
        flareHalo.addColorStop(1, 'rgba(0, 0, 0, 0)');
      }
      ctx.fillStyle = flareHalo;
      ctx.beginPath();
      ctx.arc(flareX, flareY, 180, 0, Math.PI * 2);
      ctx.fill();

      // Horizontal optical anamorphic beam
      ctx.beginPath();
      ctx.ellipse(flareX, flareY, 320, 2.5, 0, 0, Math.PI * 2);
      ctx.fillStyle = isLight ? '#6366f1' : '#ffffff';
      ctx.shadowColor = '#38bdf8';
      ctx.shadowBlur = isLight ? 15 : 30;
      ctx.fill();

      ctx.restore();

      t += 1;
      animationFrameId = requestAnimationFrame(render);
    };

    render();

    return () => {
      window.removeEventListener('resize', handleResize);
      cancelAnimationFrame(animationFrameId);
    };
  }, []);

  return (
    <section
      style={{
        position: 'relative',
        minHeight: '94vh',
        width: '100%',
        display: 'flex',
        flexDirection: 'column',
        justifyContent: 'space-between',
        overflow: 'hidden',
        backgroundColor: 'var(--bg-primary)',
        paddingTop: '80px',
        transition: 'background-color 0.3s ease',
      }}
    >
      {/* 1. Animated Canvas Light Streaks & Energy Trails */}
      <canvas
        ref={canvasRef}
        style={{
          position: 'absolute',
          top: 0,
          left: 0,
          width: '100%',
          height: '100%',
          zIndex: 1,
          pointerEvents: 'none',
        }}
      />

      {/* 2. Top-Right Ambient Glow Orb */}
      <div
        style={{
          position: 'absolute',
          top: '10%',
          right: '5%',
          width: '600px',
          height: '600px',
          borderRadius: '50%',
          background: 'radial-gradient(circle, rgba(56, 189, 248, 0.15) 0%, rgba(139, 92, 246, 0.1) 40%, transparent 70%)',
          filter: 'blur(100px)',
          zIndex: 0,
          pointerEvents: 'none',
        }}
      />

      {/* 3. Main Hero Content Grid */}
      <div
        style={{
          position: 'relative',
          zIndex: 10,
          maxWidth: '1320px',
          width: '100%',
          margin: '0 auto',
          padding: '40px 32px',
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(360px, 1fr))',
          alignItems: 'center',
          gap: '48px',
          flex: 1,
        }}
      >
        {/* Left: Typography & Actions */}
        <motion.div
          initial={{ opacity: 0, x: -30 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ duration: 0.8, ease: [0.16, 1, 0.3, 1] }}
          style={{ maxWidth: '640px' }}
        >
          {/* Badge */}
          <div
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: '8px',
              padding: '6px 14px',
              borderRadius: '9999px',
              background: 'var(--color-accent-bg)',
              border: '1px solid var(--color-accent-border)',
              marginBottom: '24px',
            }}
          >
            <Sparkles size={14} color="var(--color-accent)" />
            <span style={{ fontSize: '12.5px', fontWeight: 700, color: 'var(--color-accent)', letterSpacing: '0.04em' }}>
              AUTONOMOUS AI PAYMENT RECOVERY • 68.4% CAPTURE RATE
            </span>
          </div>

          {/* Heading */}
          <h1
            style={{
              fontSize: 'clamp(40px, 6.2vw, 76px)',
              fontWeight: 800,
              lineHeight: 1.04,
              letterSpacing: '-2px',
              color: 'var(--text-primary)',
              marginBottom: '24px',
            }}
          >
            Turn Failed Payments into
            <br />
            <span
              style={{
                fontFamily: 'serif',
                fontStyle: 'italic',
                textDecoration: 'underline',
                textDecorationColor: '#38bdf8',
                background: 'linear-gradient(135deg, #8b5cf6, #ec4899)',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
              }}
            >
              Captured Revenue.
            </span>
          </h1>

          {/* Subtitle */}
          <p
            style={{
              fontSize: 'clamp(15px, 1.8vw, 18px)',
              color: 'var(--text-secondary)',
              lineHeight: 1.65,
              marginBottom: '36px',
              maxWidth: '540px',
            }}
          >
            ReviveOS eliminates involuntary churn with real-time DeepSeek-R1 intelligence. Predict recoverability, execute zero-touch retries, and dispatch smart recovery checkout links autonomously.
          </p>

          {/* CTA Buttons */}
          <div style={{ display: 'flex', alignItems: 'center', gap: '16px', flexWrap: 'wrap' }}>
            <a href="#simulator">
              <button
                className="btn-primary"
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: '8px',
                  padding: '13px 28px',
                  fontSize: '14.5px',
                  fontWeight: 700,
                }}
              >
                <Play size={15} /> Test Live Simulator
              </button>
            </a>

            <Link href="/login">
              <span
                style={{
                  color: 'var(--text-secondary)',
                  fontSize: '14.5px',
                  fontWeight: 600,
                  cursor: 'pointer',
                  padding: '10px 14px',
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: '6px',
                  transition: 'color 0.2s ease',
                }}
                onMouseEnter={(e) => (e.currentTarget.style.color = 'var(--text-primary)')}
                onMouseLeave={(e) => (e.currentTarget.style.color = 'var(--text-secondary)')}
              >
                Explore Portals <ArrowRight size={14} />
              </span>
            </Link>
          </div>
        </motion.div>

        {/* Right: Floating Revenue Metric Card (Theme Aware) */}
        <motion.div
          initial={{ opacity: 0, y: 30, scale: 0.94 }}
          animate={{ opacity: 1, y: 0, scale: 1 }}
          transition={{ duration: 0.9, delay: 0.2, ease: [0.16, 1, 0.3, 1] }}
          style={{
            display: 'flex',
            justifyContent: 'flex-end',
          }}
        >
          <div
            style={{
              width: '100%',
              maxWidth: '380px',
              borderRadius: '24px',
              background: 'var(--bg-surface)',
              backdropFilter: 'blur(24px)',
              WebkitBackdropFilter: 'blur(24px)',
              border: '1px solid var(--border-default)',
              padding: '28px',
              boxShadow: 'var(--card-shadow)',
              position: 'relative',
              overflow: 'hidden',
              transition: 'background-color 0.3s ease, border-color 0.3s ease',
            }}
          >
            {/* Top row */}
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                <TrendingUp size={18} color="var(--color-accent)" />
                <span style={{ fontSize: '15px', fontWeight: 600, color: 'var(--text-primary)' }}>
                  Recovered Revenue
                </span>
              </div>
              <span style={{ color: 'var(--text-muted)', fontSize: '16px', letterSpacing: '2px' }}>•••</span>
            </div>

            {/* Big Revenue Number */}
            <div style={{ display: 'flex', alignItems: 'baseline', gap: '10px', marginBottom: '20px' }}>
              <span
                className="mono"
                style={{
                  fontSize: '44px',
                  fontWeight: 900,
                  letterSpacing: '-1.5px',
                  color: 'var(--text-primary)',
                }}
              >
                ₹18,62,900
              </span>
              <span style={{ fontSize: '12.5px', color: 'var(--text-muted)' }}>Last 30 days</span>
            </div>

            {/* Growth Pill Badge */}
            <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
              <div
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: '5px',
                  padding: '5px 12px',
                  borderRadius: '9999px',
                  background: 'var(--color-emerald-bg)',
                  border: '1px solid var(--color-emerald-border)',
                  color: 'var(--color-emerald)',
                  fontSize: '13px',
                  fontWeight: 800,
                }}
              >
                +68.4%
              </div>
              <span style={{ fontSize: '12.5px', color: 'var(--text-secondary)' }}>
                Net autonomous capture
              </span>
            </div>

            {/* Mini Progress Spark Line */}
            <div
              style={{
                marginTop: '22px',
                paddingTop: '16px',
                borderTop: '1px solid var(--border-subtle)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                fontSize: '12px',
                color: 'var(--text-muted)',
              }}
            >
              <span>Zero-Touch Autopay</span>
              <span style={{ color: 'var(--color-emerald)', fontWeight: 700 }}>99.2% Settle</span>
            </div>
          </div>
        </motion.div>
      </div>

      {/* 4. Bottom Hero Bar (Theme Aware) */}
      <div
        style={{
          position: 'relative',
          zIndex: 10,
          borderTop: '1px solid var(--border-subtle)',
          padding: '18px 32px',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          flexWrap: 'wrap',
          gap: '16px',
          maxWidth: '1320px',
          width: '100%',
          margin: '0 auto',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <span style={{ fontSize: '11px', letterSpacing: '0.2em', fontWeight: 800, color: 'var(--text-muted)' }}>
            SCROLL
          </span>
          <span style={{ color: 'var(--color-accent)', fontSize: '14px' }}>↓</span>
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
          <span style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
            Trusted by 50+ payment switches & SaaS
          </span>
          <div
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: '6px',
              padding: '4px 12px',
              borderRadius: '9999px',
              background: 'var(--bg-card)',
              border: '1px solid var(--border-default)',
            }}
          >
            <ShieldCheck size={14} color="var(--color-emerald)" />
            <span style={{ fontSize: '12px', fontWeight: 700, color: 'var(--text-primary)' }}>PCI-DSS Ready</span>
          </div>
        </div>
      </div>
    </section>
  );
}

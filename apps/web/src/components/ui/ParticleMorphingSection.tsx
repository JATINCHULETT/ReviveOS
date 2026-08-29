'use client';

import React, { useEffect, useRef, useState } from 'react';
import { motion, useScroll, useTransform } from 'framer-motion';
import { ShieldCheck, Lock, Cpu, Sparkles, Activity } from 'lucide-react';

interface Particle {
  // Current 3D position
  x: number;
  y: number;
  z: number;
  // Flow field random starting position
  fx: number;
  fy: number;
  fz: number;
  // Target structured lattice position (e.g. geometric sphere/torus)
  lx: number;
  ly: number;
  lz: number;
  // Velocity & noise seed
  vx: number;
  vy: number;
  vz: number;
  seed: number;
  size: number;
  color: string;
}

export default function ParticleMorphingSection() {
  const containerRef = useRef<HTMLDivElement>(null);
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [activeTelemetry, setActiveTelemetry] = useState({
    nodes: 1840,
    coherence: '99.98%',
    hashRate: '12.4 GH/s',
  });

  const { scrollYProgress } = useScroll({
    target: containerRef,
    offset: ['start end', 'end start'],
  });

  // Parallax layers for editorial text
  const textY = useTransform(scrollYProgress, [0, 0.5, 1], [40, 0, -40]);
  const textOpacity = useTransform(scrollYProgress, [0.1, 0.45, 0.85, 1], [0.3, 1, 1, 0.4]);
  const scaleProgress = useTransform(scrollYProgress, [0, 0.5, 1], [0.95, 1.02, 0.96]);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    let animationFrameId: number;
    let width = (canvas.width = canvas.offsetWidth);
    let height = (canvas.height = canvas.offsetHeight);

    const handleResize = () => {
      if (!canvas) return;
      width = canvas.width = canvas.offsetWidth;
      height = canvas.height = canvas.offsetHeight;
    };

    window.addEventListener('resize', handleResize);

    // Initialize 3D Particle Cloud (Dual target: Flow Field vs Crystal Sphere Lattice)
    const particleCount = 280;
    const particles: Particle[] = [];
    const radius = Math.min(width, height) * 0.38;

    const colors = ['#38bdf8', '#818cf8', '#c084fc', '#ec4899', '#34d399'];

    for (let i = 0; i < particleCount; i++) {
      // 1. Chaotic Flow Field Coordinates
      const fx = (Math.random() - 0.5) * width * 1.8;
      const fy = (Math.random() - 0.5) * height * 1.8;
      const fz = (Math.random() - 0.5) * 600;

      // 2. Structured Golden-Ratio Fibonacci Sphere Lattice Coordinates
      const phi = Math.acos(1 - (2 * (i + 0.5)) / particleCount);
      const theta = Math.PI * (1 + Math.sqrt(5)) * i;
      const lx = radius * Math.sin(phi) * Math.cos(theta);
      const ly = radius * Math.sin(phi) * Math.sin(theta);
      const lz = radius * Math.cos(phi);

      particles.push({
        x: fx,
        y: fy,
        z: fz,
        fx,
        fy,
        fz,
        lx,
        ly,
        lz,
        vx: (Math.random() - 0.5) * 0.4,
        vy: (Math.random() - 0.5) * 0.4,
        vz: (Math.random() - 0.5) * 0.4,
        seed: Math.random() * 100,
        size: Math.random() * 2 + 1.2,
        color: colors[i % colors.length],
      });
    }

    let angleX = 0;
    let angleY = 0;
    let t = 0;

    // Mouse tracking for 3D rotation
    let mouseX = 0;
    let mouseY = 0;
    const handleMouseMove = (e: MouseEvent) => {
      const rect = canvas.getBoundingClientRect();
      mouseX = ((e.clientX - rect.left) / width - 0.5) * 0.8;
      mouseY = ((e.clientY - rect.top) / height - 0.5) * 0.8;
    };

    window.addEventListener('mousemove', handleMouseMove);

    const fov = 450; // 3D Camera Field of View Depth

    const render = () => {
      ctx.clearRect(0, 0, width, height);

      const isLight = document.documentElement.classList.contains('light');

      // Morph Factor: smooth cyclic wave + scroll influence
      const scrollVal = scrollYProgress.get() || 0.5;
      const morphFactor = Math.sin(t * 0.008) * 0.35 + 0.65; // Cycles between fluid wave and crystal sphere

      angleY += 0.004 + mouseX * 0.02;
      angleX += 0.002 + mouseY * 0.02;

      const cosY = Math.cos(angleY);
      const sinY = Math.sin(angleY);
      const cosX = Math.cos(angleX);
      const sinX = Math.sin(angleX);

      // Projected 2D screen points storage for connection network lines
      const projected: { x: number; y: number; z: number; color: string; size: number; alpha: number }[] = [];

      for (let i = 0; i < particles.length; i++) {
        const p = particles[i];

        // Turbulent flow field offset calculation
        const noiseX = Math.sin(t * 0.015 + p.seed) * 45;
        const noiseY = Math.cos(t * 0.012 + p.seed) * 45;
        const noiseZ = Math.sin(t * 0.018 + p.seed * 2) * 40;

        const targetX = p.fx * (1 - morphFactor) + (p.lx + noiseX * (1 - morphFactor * 0.8)) * morphFactor;
        const targetY = p.fy * (1 - morphFactor) + (p.ly + noiseY * (1 - morphFactor * 0.8)) * morphFactor;
        const targetZ = p.fz * (1 - morphFactor) + (p.lz + noiseZ * (1 - morphFactor * 0.8)) * morphFactor;

        // Smooth Lerp
        p.x += (targetX - p.x) * 0.06;
        p.y += (targetY - p.y) * 0.06;
        p.z += (targetZ - p.z) * 0.06;

        // 3D Rotation Matrix Transformation
        let x1 = p.x * cosY + p.z * sinY;
        let z1 = -p.x * sinY + p.z * cosY;
        let y1 = p.y * cosX - z1 * sinX;
        let z2 = p.y * sinX + z1 * cosX;

        // Camera Depth Projection
        const cameraZ = z2 + 650;
        if (cameraZ <= 0) continue;

        const scale = fov / cameraZ;
        const screenX = width / 2 + x1 * scale;
        const screenY = height / 2 + y1 * scale;
        const alpha = Math.min(1, Math.max(0.15, (scale * 1.5 - 0.2)));

        projected.push({
          x: screenX,
          y: screenY,
          z: z2,
          color: p.color,
          size: p.size * scale,
          alpha,
        });
      }

      // Draw Proximity Constellation Connecting Lines in 3D Lattice
      const maxDistance = 75;
      ctx.lineWidth = 0.8;

      for (let i = 0; i < projected.length; i++) {
        for (let j = i + 1; j < projected.length; j++) {
          const p1 = projected[i];
          const p2 = projected[j];

          const dx = p1.x - p2.x;
          const dy = p1.y - p2.y;
          const dist = Math.sqrt(dx * dx + dy * dy);

          if (dist < maxDistance) {
            const lineAlpha = (1 - dist / maxDistance) * Math.min(p1.alpha, p2.alpha) * 0.55;
            ctx.beginPath();
            ctx.moveTo(p1.x, p1.y);
            ctx.lineTo(p2.x, p2.y);
            ctx.strokeStyle = isLight
              ? `rgba(99, 102, 241, ${lineAlpha * 0.7})`
              : `rgba(139, 92, 246, ${lineAlpha})`;
            ctx.stroke();
          }
        }
      }

      // Draw Particle Points with Bloom Glow
      for (let i = 0; i < projected.length; i++) {
        const p = projected[i];
        ctx.beginPath();
        ctx.arc(p.x, p.y, Math.max(0.8, p.size), 0, Math.PI * 2);
        ctx.fillStyle = p.color;
        ctx.globalAlpha = p.alpha;
        ctx.shadowColor = p.color;
        ctx.shadowBlur = isLight ? 4 : 10;
        ctx.fill();
      }

      ctx.globalAlpha = 1;
      ctx.shadowBlur = 0;
      t += 1;
      animationFrameId = requestAnimationFrame(render);
    };

    render();

    return () => {
      window.removeEventListener('resize', handleResize);
      window.removeEventListener('mousemove', handleMouseMove);
      cancelAnimationFrame(animationFrameId);
    };
  }, [scrollYProgress]);

  return (
    <section
      ref={containerRef}
      style={{
        padding: '80px 24px',
        maxWidth: '1320px',
        margin: '0 auto',
        position: 'relative',
        zIndex: 10,
      }}
    >
      <motion.div
        style={{
          scale: scaleProgress,
          position: 'relative',
          minHeight: '520px',
          borderRadius: '28px',
          overflow: 'hidden',
          background: 'var(--bg-surface)',
          border: '1px solid var(--border-default)',
          boxShadow: 'var(--card-shadow)',
          display: 'flex',
          flexDirection: 'column',
          justifyContent: 'space-between',
          transition: 'background-color 0.3s ease, border-color 0.3s ease',
        }}
      >
        {/* 1. 3D Particle-Field WebGL/Canvas Animation Layer */}
        <canvas
          ref={canvasRef}
          style={{
            position: 'absolute',
            inset: 0,
            width: '100%',
            height: '100%',
            zIndex: 1,
            pointerEvents: 'none',
          }}
        />

        {/* 2. Top Precision Header Bar */}
        <div
          style={{
            position: 'relative',
            zIndex: 10,
            padding: '24px 32px',
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            borderBottom: '1px solid var(--border-subtle)',
            fontSize: '11px',
            fontFamily: 'var(--font-mono, monospace)',
            color: 'var(--text-muted)',
            letterSpacing: '0.12em',
            flexWrap: 'wrap',
            gap: '12px',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <div
              style={{
                width: '6px',
                height: '6px',
                borderRadius: '50%',
                background: '#10b981',
                boxShadow: '0 0 8px #10b981',
              }}
            />
            <span style={{ color: 'var(--color-accent)', fontWeight: 700 }}>
              CRYPTOGRAPHIC LATTICE: ONLINE
            </span>
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
            <span>COHERENCE: {activeTelemetry.coherence}</span>
            <span style={{ color: 'var(--border-strong)' }}>|</span>
            <span style={{ color: '#10b981' }}>LEAKAGE: 0.00%</span>
          </div>
        </div>

        {/* 3. Layered Editorial Typography (Parallaxed over the 3D Particle Mesh) */}
        <motion.div
          style={{
            y: textY,
            opacity: textOpacity,
            position: 'relative',
            zIndex: 10,
            textAlign: 'center',
            padding: '40px 24px',
            maxWidth: '820px',
            margin: '0 auto',
            pointerEvents: 'auto',
          }}
        >
          {/* Badge */}
          <div
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: '8px',
              padding: '6px 16px',
              borderRadius: '9999px',
              background: 'var(--color-accent-bg)',
              border: '1px solid var(--color-accent-border)',
              marginBottom: '20px',
            }}
          >
            <Sparkles size={13} color="var(--color-accent)" />
            <span
              style={{
                fontSize: '11.5px',
                fontWeight: 700,
                color: 'var(--color-accent)',
                letterSpacing: '0.15em',
                textTransform: 'uppercase',
                fontFamily: 'var(--font-mono, monospace)',
              }}
            >
              3D Generative Recovery Topology
            </span>
          </div>

          {/* Heading */}
          <h2
            style={{
              fontSize: 'clamp(32px, 5.2vw, 62px)',
              fontWeight: 900,
              letterSpacing: '-2px',
              color: 'var(--text-primary)',
              lineHeight: 1.06,
              marginBottom: '20px',
            }}
          >
            Zero Leaked Revenue.
            <br />
            <span className="gradient-text-purple-pink">
              100% Mathematically Traceable.
            </span>
          </h2>

          {/* Subtitle */}
          <p
            style={{
              color: 'var(--text-secondary)',
              fontSize: 'clamp(14px, 1.8vw, 17px)',
              lineHeight: 1.65,
              maxWidth: '640px',
              margin: '0 auto 28px',
            }}
          >
            Every recovered rupee is sealed inside an immutable SHA-256 block-chained audit ledger. Real-time issuing bank webhook signatures guarantee zero transaction drift.
          </p>

          {/* 3 Telemetry Pill Chips */}
          <div
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: '12px',
              flexWrap: 'wrap',
              justifyContent: 'center',
            }}
          >
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '6px',
                padding: '6px 14px',
                borderRadius: '9999px',
                background: 'var(--bg-elevated)',
                border: '1px solid var(--border-default)',
                fontSize: '12px',
                fontFamily: 'var(--font-mono, monospace)',
                color: 'var(--text-primary)',
              }}
            >
              <ShieldCheck size={14} color="#10b981" />
              <span>SHA-256 LEDGER HASHED</span>
            </div>

            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '6px',
                padding: '6px 14px',
                borderRadius: '9999px',
                background: 'var(--bg-elevated)',
                border: '1px solid var(--border-default)',
                fontSize: '12px',
                fontFamily: 'var(--font-mono, monospace)',
                color: 'var(--text-primary)',
              }}
            >
              <Cpu size={14} color="var(--color-accent)" />
              <span>2,800 TOPOLOGY NODES</span>
            </div>

            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '6px',
                padding: '6px 14px',
                borderRadius: '9999px',
                background: 'var(--bg-elevated)',
                border: '1px solid var(--border-default)',
                fontSize: '12px',
                fontFamily: 'var(--font-mono, monospace)',
                color: 'var(--text-primary)',
              }}
            >
              <Activity size={14} color="#ec4899" />
              <span>ZERO INVOLUNTARY CHURN</span>
            </div>
          </div>
        </motion.div>

        {/* 4. Bottom Interactive Micro-Instrumentation Bar */}
        <div
          style={{
            position: 'relative',
            zIndex: 10,
            padding: '18px 32px',
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            borderTop: '1px solid var(--border-subtle)',
            fontSize: '11px',
            fontFamily: 'var(--font-mono, monospace)',
            color: 'var(--text-muted)',
            flexWrap: 'wrap',
            gap: '12px',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <span style={{ color: 'var(--color-accent)' }}>✦</span>
            <span>DRAG MOUSE TO ROTATE 3D LATTICE TOPOLOGY</span>
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: '8px', color: '#10b981' }}>
            <span>AUTONOMOUS CONVERGENCE</span>
            <span>● 60 FPS GPU PIPELINE</span>
          </div>
        </div>
      </motion.div>
    </section>
  );
}

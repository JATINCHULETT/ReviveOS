'use client';

import React, { useRef, useState } from 'react';
import { motion, useScroll, useSpring, useTransform } from 'framer-motion';
import { ArrowRight, ArrowLeft, ChevronLeft, ChevronRight } from 'lucide-react';

/* ═══════════════════════════════════════════════════════════════
   1. SCROLL PROGRESS BAR (Fixed at Top)
   ═══════════════════════════════════════════════════════════════ */
export function ScrollProgressBar() {
  const { scrollYProgress } = useScroll();
  const scaleX = useSpring(scrollYProgress, {
    stiffness: 100,
    damping: 30,
    restDelta: 0.001,
  });

  return (
    <motion.div
      style={{
        scaleX,
        transformOrigin: '0%',
        position: 'fixed',
        top: 0,
        left: 0,
        right: 0,
        height: '3px',
        background: 'linear-gradient(90deg, #8b5cf6 0%, #ec4899 50%, #f43f5e 100%)',
        zIndex: 100,
        boxShadow: '0 0 12px rgba(236, 72, 153, 0.7)',
      }}
    />
  );
}

/* ═══════════════════════════════════════════════════════════════
   2. HERO SPOTLIGHT SVG (Signature Motion Primitives Glow)
   ═══════════════════════════════════════════════════════════════ */
export function HeroSpotlight() {
  return (
    <div style={{ position: 'absolute', inset: 0, overflow: 'hidden', pointerEvents: 'none', zIndex: 0 }}>
      {/* Dynamic Ambient Blur Spheres */}
      <div
        style={{
          position: 'absolute',
          top: '15%',
          left: '50%',
          transform: 'translate(-50%, -50%)',
          width: '700px',
          height: '700px',
          borderRadius: '50%',
          background: 'radial-gradient(circle, rgba(139, 92, 246, 0.18) 0%, rgba(236, 72, 153, 0.08) 45%, transparent 70%)',
          filter: 'blur(100px)',
        }}
      />
      <div
        style={{
          position: 'absolute',
          top: '45%',
          right: '10%',
          width: '500px',
          height: '500px',
          borderRadius: '50%',
          background: 'radial-gradient(circle, rgba(6, 182, 212, 0.12) 0%, transparent 70%)',
          filter: 'blur(120px)',
        }}
      />

      {/* SVG Elliptical Spotlight */}
      <svg
        style={{
          position: 'absolute',
          top: '-150px',
          left: '50%',
          transform: 'translateX(-50%)',
          width: '1200px',
          height: '900px',
          opacity: 0.7,
        }}
        viewBox="0 0 3787 2842"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <g filter="url(#spotlightFilter)">
          <ellipse
            cx="1924.71"
            cy="273.501"
            rx="1924.71"
            ry="273.501"
            transform="matrix(-0.822377 -0.568943 -0.568943 0.822377 3631.88 2291.09)"
            fill="rgba(168, 85, 247, 0.22)"
          />
        </g>
        <defs>
          <filter id="spotlightFilter" x="0" y="0" width="3785" height="2840" filterUnits="userSpaceOnUse">
            <feGaussianBlur stdDeviation="150" />
          </filter>
        </defs>
      </svg>
    </div>
  );
}

/* ═══════════════════════════════════════════════════════════════
   3. CONNECTED STAGE CARDS WITH ARROWS (All 6 visible & medium sized)
   ═══════════════════════════════════════════════════════════════ */
interface StageInfo {
  num: string;
  title: string;
  subtitle: string;
  tag: string;
  color: string;
  accentGradient: string;
  description: string;
}

export function ScrollOrchestrator({ stages }: { stages: StageInfo[] }) {
  return (
    <section style={{ padding: '60px 24px', maxWidth: '1280px', margin: '0 auto', position: 'relative', zIndex: 10 }}>
      {/* Section Header */}
      <div style={{ textAlign: 'center', marginBottom: '48px' }}>
        <span
          style={{
            fontSize: '12px',
            fontWeight: 700,
            letterSpacing: '0.15em',
            textTransform: 'uppercase',
            color: '#c084fc',
            padding: '5px 14px',
            borderRadius: '9999px',
            background: 'rgba(139, 92, 246, 0.15)',
            border: '1px solid rgba(139, 92, 246, 0.3)',
            display: 'inline-block',
            marginBottom: '14px',
          }}
        >
          6-Stage Orchestration Loop
        </span>
        <h2 style={{ fontSize: 'clamp(28px, 4vw, 42px)', fontWeight: 800, letterSpacing: '-1px', color: 'var(--text-primary)', marginBottom: '10px' }}>
          Deterministic AI Recovery Pipeline
        </h2>
        <p style={{ color: 'var(--text-secondary)', maxWidth: '580px', margin: '0 auto', fontSize: '15px' }}>
          Each stage executes autonomously with continuous cryptographic state reconciliation.
        </p>
      </div>

      {/* Grid of 6 Medium Cards Connected with Arrows */}
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))',
          gap: '20px',
          position: 'relative',
        }}
      >
        {stages.map((stage, idx) => (
          <motion.div
            key={stage.num}
            initial={{ opacity: 0, y: 16 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true, margin: '-30px' }}
            transition={{ duration: 0.35, delay: idx * 0.06 }}
            whileHover={{ y: -3, transition: { duration: 0.2 } }}
            style={{
              position: 'relative',
              background: 'var(--bg-card)',
              backdropFilter: 'blur(20px)',
              WebkitBackdropFilter: 'blur(20px)',
              border: '1px solid var(--border-subtle)',
              borderRadius: '20px',
              padding: '24px',
              display: 'flex',
              flexDirection: 'column',
              justifyContent: 'space-between',
              boxShadow: 'var(--card-shadow)',
              overflow: 'hidden',
              minHeight: '240px',
            }}
          >
            {/* Subtle stage ambient glow */}
            <div
              style={{
                position: 'absolute',
                top: '-20px',
                right: '-20px',
                width: '120px',
                height: '120px',
                borderRadius: '50%',
                background: `radial-gradient(circle, ${stage.color}25 0%, transparent 70%)`,
                filter: 'blur(25px)',
                pointerEvents: 'none',
              }}
            />

            <div>
              {/* Card Header: Stage pill & Number with Arrow */}
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '14px' }}>
                <span
                  style={{
                    fontSize: '11px',
                    fontWeight: 700,
                    letterSpacing: '0.08em',
                    textTransform: 'uppercase',
                    padding: '3px 10px',
                    borderRadius: '9999px',
                    background: `${stage.color}18`,
                    border: `1px solid ${stage.color}40`,
                    color: stage.color,
                  }}
                >
                  {stage.tag}
                </span>

                <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                  <span
                    className="mono"
                    style={{
                      fontSize: '16px',
                      fontWeight: 800,
                      color: 'var(--text-muted)',
                    }}
                  >
                    {stage.num}
                  </span>
                  {idx < stages.length - 1 && (
                    <div
                      style={{
                        width: '22px',
                        height: '22px',
                        borderRadius: '50%',
                        background: 'var(--bg-card-hover)',
                        border: '1px solid var(--border-subtle)',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        color: stage.color,
                      }}
                    >
                      <ArrowRight size={12} />
                    </div>
                  )}
                </div>
              </div>

              {/* Title & Subtitle */}
              <h3 style={{ fontSize: '18px', fontWeight: 800, color: 'var(--text-primary)', marginBottom: '3px', letterSpacing: '-0.3px' }}>
                {stage.title}
              </h3>
              <div
                style={{
                  fontSize: '12.5px',
                  fontWeight: 600,
                  marginBottom: '10px',
                  background: stage.accentGradient,
                  WebkitBackgroundClip: 'text',
                  WebkitTextFillColor: 'transparent',
                }}
              >
                {stage.subtitle}
              </div>

              {/* Description */}
              <p style={{ fontSize: '13px', color: 'var(--text-secondary)', lineHeight: 1.55 }}>
                {stage.description}
              </p>
            </div>

            {/* Bottom Progress Accent Bar */}
            <div
              style={{
                marginTop: '16px',
                width: '100%',
                height: '3px',
                borderRadius: '2px',
                background: 'var(--border-subtle)',
                overflow: 'hidden',
              }}
            >
              <div
                style={{
                  height: '100%',
                  background: stage.accentGradient,
                  width: `${((idx + 1) / stages.length) * 100}%`,
                }}
              />
            </div>
          </motion.div>
        ))}
      </div>
    </section>
  );
}

/* ═══════════════════════════════════════════════════════════════
   4. HORIZONTAL FEATURE CAROUSEL SLIDER (Smooth & Zero Black Void)
   ═══════════════════════════════════════════════════════════════ */
interface FeatureCardItem {
  id: string;
  num: string;
  title: string;
  desc: string;
  gradient: string;
  accent: string;
  icon?: React.ReactNode;
}

export function HorizontalScrollSection({
  cards,
  heading,
  subheading,
}: {
  cards: FeatureCardItem[];
  heading: string;
  subheading: string;
}) {
  const scrollContainerRef = useRef<HTMLDivElement>(null);
  const [canScrollLeft, setCanScrollLeft] = useState(false);
  const [canScrollRight, setCanScrollRight] = useState(true);

  const checkScroll = () => {
    if (!scrollContainerRef.current) return;
    const { scrollLeft, scrollWidth, clientWidth } = scrollContainerRef.current;
    setCanScrollLeft(scrollLeft > 10);
    setCanScrollRight(scrollLeft < scrollWidth - clientWidth - 10);
  };

  const scroll = (direction: 'left' | 'right') => {
    if (!scrollContainerRef.current) return;
    const scrollAmount = 340;
    scrollContainerRef.current.scrollBy({
      left: direction === 'left' ? -scrollAmount : scrollAmount,
      behavior: 'smooth',
    });
  };

  return (
    <section style={{ padding: '60px 24px', maxWidth: '1280px', margin: '0 auto', position: 'relative' }}>
      {/* Header with Navigation Arrows */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-end', marginBottom: '28px', flexWrap: 'wrap', gap: '16px' }}>
        <div>
          <span
            style={{
              fontSize: '12px',
              fontWeight: 700,
              letterSpacing: '0.15em',
              textTransform: 'uppercase',
              color: '#c084fc',
              display: 'block',
              marginBottom: '6px',
            }}
          >
            Capabilities & Architecture
          </span>
          <h2 style={{ fontSize: 'clamp(26px, 3.5vw, 40px)', fontWeight: 800, letterSpacing: '-1px', color: 'var(--text-primary)' }}>
            {heading}
          </h2>
          <p style={{ color: 'var(--text-secondary)', fontSize: '15px', marginTop: '4px' }}>
            {subheading}
          </p>
        </div>

        {/* Carousel controls */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <button
            type="button"
            onClick={() => scroll('left')}
            disabled={!canScrollLeft}
            title="Previous Cards"
            style={{
              width: '38px',
              height: '38px',
              borderRadius: '50%',
              background: 'var(--bg-card)',
              border: '1px solid var(--border-subtle)',
              color: canScrollLeft ? 'var(--text-primary)' : 'var(--text-muted)',
              cursor: canScrollLeft ? 'pointer' : 'default',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              opacity: canScrollLeft ? 1 : 0.4,
              transition: 'all 0.2s ease',
            }}
          >
            <ChevronLeft size={18} />
          </button>

          <button
            type="button"
            onClick={() => scroll('right')}
            disabled={!canScrollRight}
            title="Next Cards"
            style={{
              width: '38px',
              height: '38px',
              borderRadius: '50%',
              background: 'var(--bg-card)',
              border: '1px solid var(--border-subtle)',
              color: canScrollRight ? 'var(--text-primary)' : 'var(--text-muted)',
              cursor: canScrollRight ? 'pointer' : 'default',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              opacity: canScrollRight ? 1 : 0.4,
              transition: 'all 0.2s ease',
            }}
          >
            <ChevronRight size={18} />
          </button>
        </div>
      </div>

      {/* Scrollable Track */}
      <div
        ref={scrollContainerRef}
        onScroll={checkScroll}
        style={{
          display: 'flex',
          gap: '20px',
          overflowX: 'auto',
          scrollSnapType: 'x mandatory',
          scrollbarWidth: 'none',
          paddingBottom: '16px',
        }}
      >
        {cards.map((card) => (
          <div
            key={card.id}
            style={{
              flexShrink: 0,
              width: 'clamp(270px, 28vw, 340px)',
              height: '290px',
              borderRadius: '22px',
              background: 'var(--bg-surface)',
              border: '1px solid var(--border-default)',
              padding: '24px',
              display: 'flex',
              flexDirection: 'column',
              justifyContent: 'space-between',
              position: 'relative',
              overflow: 'hidden',
              boxShadow: 'var(--card-shadow)',
              scrollSnapAlign: 'start',
              transition: 'background-color 0.3s ease, border-color 0.3s ease',
            }}
          >
            {/* Subtle background number */}
            <div
              style={{
                position: 'absolute',
                right: '-6px',
                top: '-15px',
                fontSize: '110px',
                fontWeight: 900,
                color: 'var(--border-subtle)',
                lineHeight: 1,
                pointerEvents: 'none',
                userSelect: 'none',
              }}
            >
              {card.num}
            </div>

            {/* Top tag & Icon */}
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', position: 'relative', zIndex: 2 }}>
              <span className="mono" style={{ fontSize: '13px', color: 'var(--text-muted)', fontWeight: 700 }}>
                {card.num}
              </span>
              {card.icon && (
                <div
                  style={{
                    width: '34px',
                    height: '34px',
                    borderRadius: '10px',
                    background: 'var(--bg-elevated)',
                    border: '1px solid var(--border-subtle)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                  }}
                >
                  {card.icon}
                </div>
              )}
            </div>

            {/* Bottom text */}
            <div style={{ position: 'relative', zIndex: 2 }}>
              <h3 style={{ fontSize: '19px', fontWeight: 800, color: 'var(--text-primary)', marginBottom: '6px' }}>
                {card.title}
              </h3>
              <p style={{ fontSize: '13px', color: 'var(--text-secondary)', lineHeight: 1.55 }}>
                {card.desc}
              </p>
              <div
                style={{
                  width: '38px',
                  height: '3px',
                  borderRadius: '2px',
                  background: card.accent,
                  marginTop: '14px',
                }}
              />
            </div>
          </div>
        ))}
      </div>
    </section>
  );
}

/* ═══════════════════════════════════════════════════════════════
   5. WORD-BY-WORD TEXT SCRUB REVEAL
   ═══════════════════════════════════════════════════════════════ */
export function TextScrubReveal({ text }: { text: string }) {
  const containerRef = useRef<HTMLDivElement>(null);
  const { scrollYProgress } = useScroll({
    target: containerRef,
    offset: ['start 0.8', 'end 0.3'],
  });

  const words = text.split(' ');

  return (
    <section ref={containerRef} style={{ padding: '70px 24px', maxWidth: '1080px', margin: '0 auto' }}>
      <div
        style={{
          fontSize: 'clamp(24px, 3.6vw, 44px)',
          fontWeight: 800,
          letterSpacing: '-1.2px',
          lineHeight: 1.3,
          display: 'flex',
          flexWrap: 'wrap',
          gap: '8px 10px',
        }}
      >
        {words.map((word, idx) => {
          const start = idx / words.length;
          const end = (idx + 1) / words.length;
          // eslint-disable-next-line react-hooks/rules-of-hooks
          const opacity = useTransform(scrollYProgress, [start, end], [0.18, 1]);

          return (
            <motion.span
              key={idx}
              style={{
                opacity,
                color: idx % 4 === 0 ? 'var(--color-pink)' : idx % 3 === 0 ? 'var(--color-accent-light)' : 'var(--text-primary)',
                display: 'inline-block',
              }}
            >
              {word}
            </motion.span>
          );
        })}
      </div>
    </section>
  );
}

/* ═══════════════════════════════════════════════════════════════
   6. KINETIC MARQUEE (Massive outlined typographic ribbons)
   ═══════════════════════════════════════════════════════════════ */
export function KineticMarquee() {
  return (
    <div style={{ padding: '36px 0', overflow: 'hidden', userSelect: 'none', position: 'relative' }}>
      {/* Row 1 */}
      <div className="animate-marquee" style={{ marginBottom: '6px' }}>
        {[1, 2].map((k) => (
          <div key={k} style={{ display: 'flex', alignItems: 'center', whiteSpace: 'nowrap', gap: '24px', paddingRight: '24px' }}>
            <span
              style={{
                fontSize: 'clamp(36px, 6vw, 76px)',
                fontWeight: 900,
                letterSpacing: '-2px',
                WebkitTextStroke: '1px var(--border-default)',
                color: 'transparent',
              }}
            >
              AUTONOMOUS RECOVERY
            </span>
            <span style={{ fontSize: '22px', color: '#ec4899' }}>✦</span>
            <span
              style={{
                fontSize: 'clamp(36px, 6vw, 76px)',
                fontWeight: 900,
                letterSpacing: '-2px',
                background: 'linear-gradient(90deg, #f59e0b, #ef4444)',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
              }}
            >
              ML FRAUD DETECTION
            </span>
            <span style={{ fontSize: '22px', color: '#f59e0b' }}>✦</span>
            <span
              style={{
                fontSize: 'clamp(36px, 6vw, 76px)',
                fontWeight: 900,
                letterSpacing: '-2px',
                background: 'linear-gradient(90deg, #c084fc, #ec4899)',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
              }}
            >
              DEEPSEEK-R1 AI
            </span>
            <span style={{ fontSize: '22px', color: '#8b5cf6' }}>✦</span>
            <span
              style={{
                fontSize: 'clamp(36px, 6vw, 76px)',
                fontWeight: 900,
                letterSpacing: '-2px',
                WebkitTextStroke: '1px var(--border-default)',
                color: 'transparent',
              }}
            >
              ZERO CHURN
            </span>
            <span style={{ fontSize: '22px', color: '#10b981' }}>✦</span>
          </div>
        ))}
      </div>

      {/* Row 2 (Reverse direction) */}
      <div className="animate-marquee-reverse">
        {[1, 2].map((k) => (
          <div key={k} style={{ display: 'flex', alignItems: 'center', whiteSpace: 'nowrap', gap: '24px', paddingRight: '24px' }}>
            <span
              style={{
                fontSize: 'clamp(36px, 6vw, 76px)',
                fontWeight: 900,
                letterSpacing: '-2px',
                background: 'linear-gradient(90deg, #38bdf8, #818cf8)',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
              }}
            >
              68.4% CAPTURED
            </span>
            <span style={{ fontSize: '22px', color: '#38bdf8' }}>✦</span>
            <span
              style={{
                fontSize: 'clamp(36px, 6vw, 76px)',
                fontWeight: 900,
                letterSpacing: '-2px',
                background: 'linear-gradient(90deg, #f59e0b, #d97706)',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
              }}
            >
              RANDOM FOREST RISK GUARD
            </span>
            <span style={{ fontSize: '22px', color: '#f59e0b' }}>✦</span>
            <span
              style={{
                fontSize: 'clamp(36px, 6vw, 76px)',
                fontWeight: 900,
                letterSpacing: '-2px',
                WebkitTextStroke: '1px var(--border-default)',
                color: 'transparent',
              }}
            >
              ZERO-TOUCH RETRY
            </span>
            <span style={{ fontSize: '22px', color: '#ec4899' }}>✦</span>
            <span
              style={{
                fontSize: 'clamp(36px, 6vw, 76px)',
                fontWeight: 900,
                letterSpacing: '-2px',
                background: 'linear-gradient(90deg, #34d399, #10b981)',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
              }}
            >
              SHA-256 AUDITED
            </span>
            <span style={{ fontSize: '22px', color: '#10b981' }}>✦</span>
          </div>
        ))}
      </div>
    </div>
  );
}

/* ═══════════════════════════════════════════════════════════════
   7. 3D PERSPECTIVE DEPTH PARALLAX
   ═══════════════════════════════════════════════════════════════ */
export function DepthParallaxSection() {
  const containerRef = useRef<HTMLDivElement>(null);
  const { scrollYProgress } = useScroll({
    target: containerRef,
    offset: ['start end', 'end start'],
  });

  const layer1Y = useTransform(scrollYProgress, [0, 1], [-40, 40]);
  const layer2Y = useTransform(scrollYProgress, [0, 1], [30, -30]);
  const scale = useTransform(scrollYProgress, [0, 0.5, 1], [0.97, 1.01, 0.97]);

  return (
    <section ref={containerRef} style={{ padding: '60px 24px', maxWidth: '1200px', margin: '0 auto' }}>
      <div
        style={{
          position: 'relative',
          height: '380px',
          borderRadius: '24px',
          overflow: 'hidden',
          background: 'var(--bg-surface)',
          border: '1px solid var(--border-default)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          boxShadow: 'var(--card-shadow)',
          transition: 'background-color 0.3s ease, border-color 0.3s ease',
        }}
      >
        {/* Deep Z layer */}
        <motion.div style={{ y: layer1Y, position: 'absolute', inset: 0, pointerEvents: 'none' }}>
          <div
            style={{
              position: 'absolute',
              top: '20%',
              left: '15%',
              width: '140px',
              height: '140px',
              borderRadius: '20px',
              background: 'linear-gradient(135deg, rgba(139, 92, 246, 0.15), rgba(236, 72, 153, 0.08))',
              border: '1px solid var(--border-subtle)',
              transform: 'rotate(12deg)',
            }}
          />
          <div
            style={{
              position: 'absolute',
              bottom: '20%',
              right: '18%',
              width: '150px',
              height: '150px',
              borderRadius: '50%',
              background: 'radial-gradient(circle, rgba(6, 182, 212, 0.12), rgba(59, 130, 246, 0.04))',
              border: '1px solid var(--border-subtle)',
            }}
          />
        </motion.div>

        {/* Foreground Content */}
        <motion.div
          style={{ y: layer2Y, scale, position: 'relative', zIndex: 10, textAlign: 'center', padding: '24px' }}
        >
          <span
            style={{
              fontSize: '11px',
              fontWeight: 700,
              letterSpacing: '0.2em',
              textTransform: 'uppercase',
              color: '#38bdf8',
              marginBottom: '10px',
              display: 'inline-block',
            }}
          >
            Cryptographic Guarantee
          </span>
          <h3
            style={{
              fontSize: 'clamp(24px, 3.8vw, 46px)',
              fontWeight: 900,
              letterSpacing: '-1.2px',
              color: '#ffffff',
              lineHeight: 1.15,
              marginBottom: '12px',
            }}
          >
            Authoritative Ledger.
            <br />
            <span className="gradient-text-purple-pink">Zero False Positives.</span>
          </h3>
          <p
            style={{
              fontSize: '14.5px',
              color: 'var(--text-secondary)',
              maxWidth: '480px',
              margin: '0 auto',
              lineHeight: 1.6,
            }}
          >
            Every recovered transaction is cross-verified directly against gateway settlement webhooks before being permanently logged.
          </p>
        </motion.div>
      </div>
    </section>
  );
}

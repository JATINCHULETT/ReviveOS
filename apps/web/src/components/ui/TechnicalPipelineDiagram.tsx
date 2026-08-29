'use client';

import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
  Activity,
  TrendingUp,
  Cpu,
  Clock,
  Send,
  Lock,
} from 'lucide-react';

export default function TechnicalPipelineDiagram() {
  const [activeStage, setActiveStage] = useState<number>(0);
  const [evalPass, setEvalPass] = useState<number>(1);

  // Auto-advance the active stage on the circular dial
  useEffect(() => {
    const timer = setInterval(() => {
      setActiveStage((prev) => (prev + 1) % 6);
      setEvalPass((prev) => (prev >= 6 ? 1 : prev + 1));
    }, 3200);
    return () => clearInterval(timer);
  }, []);

  const stages = [
    {
      num: '01',
      category: 'CLASSIFICATION',
      title: 'Diagnose & Classify',
      subtitle: 'Deterministic Error Mapping',
      desc: 'ReviveOS intercepts payment gateway webhook errors across 50+ bank codes. In milliseconds, failures are mapped into high-probability recovery categories.',
      accent: '#0ea5e9', // cyan
      metric: '50+ Codes Decoded',
      icon: <Activity size={14} color="#0ea5e9" />,
    },
    {
      num: '02',
      category: 'PREDICTION',
      title: 'Predict Recoverability',
      subtitle: 'Machine Learning Probability Engine',
      desc: 'Our statistical inference model scores likelihood of successful recapture based on historical issuing bank reliability, past customer habits, and transaction size.',
      accent: '#8b5cf6', // violet
      metric: '76.4% Prob Score',
      icon: <TrendingUp size={14} color="#8b5cf6" />,
    },
    {
      num: '03',
      category: 'AI REASONING',
      title: 'Decide Strategy',
      subtitle: 'DeepSeek-R1 Chain-of-Thought Reasoning',
      desc: 'Instead of crude static retries, DeepSeek-R1 reasons through the root cause to choose between Zero-Touch Autopay Retries, Payment Method Update Links, or Fallback Channels.',
      accent: '#ec4899', // pink
      metric: 'Zero-Touch Route',
      icon: <Cpu size={14} color="#ec4899" />,
    },
    {
      num: '04',
      category: 'TIMING OPTIMIZATION',
      title: 'Dynamic Timing',
      subtitle: 'Payday & Gateway Health Alignment',
      desc: 'Retries are delayed and dynamically fired at the precise optimal window—avoiding bank maintenance hours and aligning with salary credit periods.',
      accent: '#f59e0b', // amber
      metric: 'Optimal Window Sync',
      icon: <Clock size={14} color="#f59e0b" />,
    },
    {
      num: '05',
      category: 'EXECUTION',
      title: 'Multi-Channel Execution',
      subtitle: 'Zero-Touch Tokens & Resend Notifications',
      desc: 'Executes zero-touch token charges or dispatches customized, high-converting payment links via Resend API and Razorpay Auto-Notify without disturbing the user.',
      accent: '#10b981', // emerald
      metric: 'Token + Smart Link',
      icon: <Send size={14} color="#10b981" />,
    },
    {
      num: '06',
      category: 'VERIFICATION',
      title: 'Authoritative Verification',
      subtitle: 'SHA-256 Cryptographic Audit Ledger',
      desc: 'Every recovered rupee is validated directly against gateway settlement webhooks and permanently recorded with cryptographic block hashing.',
      accent: '#a855f7', // purple
      metric: '100% Hash Audit',
      icon: <Lock size={14} color="#a855f7" />,
    },
  ];

  const current = stages[activeStage];

  return (
    <section
      style={{
        position: 'relative',
        zIndex: 10,
        padding: '80px 16px',
        maxWidth: '1320px',
        margin: '0 auto',
        fontFamily: "'JetBrains Mono', 'SF Mono', 'Roboto Mono', Menlo, monospace",
      }}
    >
      {/* Container Styled with Dynamic Theme Classes */}
      <div
        style={{
          position: 'relative',
          background: 'var(--bg-surface)',
          border: '1px solid var(--border-default)',
          borderRadius: '20px',
          padding: '36px 32px 30px',
          boxShadow: 'var(--card-shadow)',
          overflow: 'hidden',
          transition: 'background-color 0.3s ease, border-color 0.3s ease',
        }}
      >
        {/* Subtle grid background */}
        <div
          style={{
            position: 'absolute',
            inset: 0,
            backgroundImage: `
              linear-gradient(var(--border-subtle) 1px, transparent 1px),
              linear-gradient(90deg, var(--border-subtle) 1px, transparent 1px)
            `,
            backgroundSize: '36px 36px',
            pointerEvents: 'none',
            opacity: 0.6,
          }}
        />

        {/* ════ TOP SCHEMATIC HEADER ════ */}
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'flex-start',
            fontSize: '11px',
            letterSpacing: '0.18em',
            textTransform: 'uppercase',
            color: 'var(--text-muted)',
            marginBottom: '28px',
            position: 'relative',
            zIndex: 2,
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px', width: '220px' }}>
            <span style={{ width: '6px', height: '6px', borderRadius: '50%', background: '#10b981', boxShadow: '0 0 8px #10b981' }} />
            <span style={{ color: 'var(--color-accent)' }}>REVIVE-R1 PIPELINE</span>
          </div>

          {/* Center Title */}
          <div style={{ textAlign: 'center' }}>
            <div style={{ fontWeight: 800, letterSpacing: '0.22em', color: 'var(--text-primary)', fontSize: '13px' }}>
              AUTONOMOUS AI RECOVERY PIPELINE
            </div>
            <div style={{ fontSize: '10.5px', color: current.accent, marginTop: '3px', fontWeight: 600 }}>
              STAGE 0{evalPass}/06 • {current.category}
            </div>
          </div>

          {/* Right Header */}
          <div style={{ textAlign: 'right', width: '220px', letterSpacing: '0.16em', color: 'var(--color-pink)', fontWeight: 700 }}>
            MEASURED OUTCOME
          </div>
        </div>

        {/* ════ MAIN DIAGRAM SVG CANVAS & NODES ════ */}
        <div
          style={{
            position: 'relative',
            minHeight: '380px',
            display: 'grid',
            gridTemplateColumns: 'minmax(210px, 250px) 1fr minmax(220px, 260px)',
            alignItems: 'center',
            gap: '20px',
            zIndex: 2,
          }}
        >
          {/* ════ LEFT COLUMN: 4 INGESTION NODES WITH SVG WAVEFORMS ════ */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px', fontSize: '10px', color: 'var(--color-accent)' }}>
              <span>↑</span>
              <span style={{ letterSpacing: '0.12em' }}>INCOMING_FAILURE_STREAM</span>
            </div>

            {/* 00-1 Gateway Declines */}
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: '10px' }}>
              <div>
                <div style={{ fontSize: '9.5px', color: '#0ea5e9', fontWeight: 700 }}>00-1</div>
                <div style={{ fontSize: '12.5px', fontWeight: 700, color: 'var(--text-primary)' }}>Gateway declines</div>
              </div>
              <svg width="60" height="24" viewBox="0 0 60 24">
                <line x1="0" y1="4" x2="48" y2="4" stroke="#0ea5e9" strokeWidth="1.5" />
                <line x1="0" y1="10" x2="38" y2="10" stroke="#0ea5e9" strokeWidth="1.5" />
                <line x1="0" y1="16" x2="52" y2="16" stroke="#0ea5e9" strokeWidth="1.5" />
                <circle cx="56" cy="10" r="2" fill="#0ea5e9" />
              </svg>
            </div>

            {/* 00-2 Mandate Lapses */}
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: '10px' }}>
              <div>
                <div style={{ fontSize: '9.5px', color: '#8b5cf6', fontWeight: 700 }}>00-2</div>
                <div style={{ fontSize: '12.5px', fontWeight: 700, color: 'var(--text-primary)' }}>Mandate lapses</div>
              </div>
              <svg width="60" height="24" viewBox="0 0 60 24">
                <line x1="0" y1="6" x2="28" y2="6" stroke="#8b5cf6" strokeWidth="1" strokeDasharray="3 3" />
                <circle cx="34" cy="6" r="1.5" fill="#8b5cf6" />
                <circle cx="42" cy="6" r="1" fill="#8b5cf6" />
                <line x1="0" y1="14" x2="36" y2="14" stroke="#8b5cf6" strokeWidth="1" strokeDasharray="4 2" />
                <circle cx="48" cy="14" r="1.5" fill="#8b5cf6" />
                <circle cx="56" cy="10" r="2" fill="#8b5cf6" />
              </svg>
            </div>

            {/* 00-3 Bank Switch Drops */}
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: '10px' }}>
              <div>
                <div style={{ fontSize: '9.5px', color: '#ec4899', fontWeight: 700 }}>00-3</div>
                <div style={{ fontSize: '12.5px', fontWeight: 700, color: 'var(--text-primary)' }}>Bank switch drops</div>
              </div>
              <svg width="60" height="24" viewBox="0 0 60 24">
                <path
                  d="M 0 16 L 10 16 L 10 6 L 22 6 L 22 16 L 34 16 L 34 6 L 46 6 L 46 16 L 54 16"
                  fill="none"
                  stroke="#ec4899"
                  strokeWidth="1.2"
                />
                <circle cx="56" cy="16" r="2" fill="#ec4899" />
              </svg>
            </div>

            {/* 00-4 3DS Drop-offs */}
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: '10px' }}>
              <div>
                <div style={{ fontSize: '9.5px', color: '#10b981', fontWeight: 700 }}>00-4</div>
                <div style={{ fontSize: '12.5px', fontWeight: 700, color: 'var(--text-primary)' }}>3DS drop-offs</div>
              </div>
              <svg width="60" height="24" viewBox="0 0 60 24">
                <path
                  d="M 0 18 L 12 12 L 24 16 L 36 6 L 48 14 L 54 10"
                  fill="none"
                  stroke="#10b981"
                  strokeWidth="1.2"
                />
                <circle cx="12" cy="12" r="1.8" fill="#10b981" />
                <circle cx="24" cy="16" r="1.8" fill="#10b981" />
                <circle cx="36" cy="6" r="1.8" fill="#10b981" />
                <circle cx="48" cy="14" r="1.8" fill="#10b981" />
                <circle cx="56" cy="10" r="2" fill="#10b981" />
              </svg>
            </div>
          </div>

          {/* ════ CENTER COLUMN: CONVERGING BEZIER WIRES + PROCESSOR CHIP + CIRCULAR HUD RADAR DIAL ════ */}
          <div
            style={{
              position: 'relative',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              width: '100%',
              minHeight: '340px',
            }}
          >
            {/* SVG Wire Convergence into Center Chip */}
            <svg
              viewBox="0 0 540 340"
              style={{
                width: '100%',
                height: '100%',
                overflow: 'visible',
              }}
            >
              {/* 4 Curved Bezier Traces from Left inputs into the Central Chip */}
              <path
                d="M 10 50 C 70 50, 90 160, 140 160"
                fill="none"
                stroke="rgba(14, 165, 233, 0.6)"
                strokeWidth="1.2"
              />
              <path
                d="M 10 110 C 60 110, 90 165, 140 165"
                fill="none"
                stroke="rgba(139, 92, 246, 0.6)"
                strokeWidth="1.2"
              />
              <path
                d="M 10 220 C 60 220, 90 175, 140 175"
                fill="none"
                stroke="rgba(236, 72, 153, 0.6)"
                strokeWidth="1.2"
              />
              <path
                d="M 10 280 C 70 280, 90 180, 140 180"
                fill="none"
                stroke="rgba(16, 185, 129, 0.6)"
                strokeWidth="1.2"
              />

              {/* Central Ingestion Processor Chip Box */}
              <rect
                x="140"
                y="145"
                width="34"
                height="50"
                rx="6"
                fill="var(--bg-elevated)"
                stroke="var(--color-accent)"
                strokeWidth="1.2"
              />
              {/* Chip internal schematic lines */}
              <line x1="147" y1="158" x2="167" y2="158" stroke="var(--color-accent)" strokeWidth="1" />
              <circle cx="157" cy="170" r="3.5" fill="none" stroke="#ec4899" strokeWidth="1.2" />
              <line x1="147" y1="182" x2="167" y2="182" stroke="#8b5cf6" strokeWidth="1" />

              {/* Line from Chip into the Circular Radar Dial */}
              <line x1="174" y1="170" x2="225" y2="170" stroke="var(--border-strong)" strokeWidth="1.2" />

              {/* ════ CIRCULAR RADAR DIAL HUD ════ */}
              <g transform="translate(325, 170)">
                {/* Outer Tick Marks Circle */}
                <circle
                  cx="0"
                  cy="0"
                  r="98"
                  fill="none"
                  stroke="var(--border-default)"
                  strokeWidth="1.2"
                  strokeDasharray="2 6"
                />

                {/* Concentric Middle Ring */}
                <circle
                  cx="0"
                  cy="0"
                  r="78"
                  fill="none"
                  stroke="var(--border-subtle)"
                  strokeWidth="1"
                />

                {/* Inner Ring */}
                <circle
                  cx="0"
                  cy="0"
                  r="58"
                  fill="none"
                  stroke="var(--border-default)"
                  strokeWidth="1"
                />

                {/* Glowing Concentric Core Ring */}
                <circle
                  cx="0"
                  cy="0"
                  r="28"
                  fill="var(--bg-surface)"
                  stroke={current.accent}
                  strokeWidth="1.5"
                  style={{
                    filter: `drop-shadow(0 0 10px ${current.accent})`,
                    transition: 'stroke 0.3s ease, filter 0.3s ease',
                  }}
                />

                {/* Central Zap Logo Icon */}
                <path
                  d="M -2 -8 L 4 -8 L 0 0 L 5 0 L -3 9 L -1 2 L -5 2 Z"
                  fill="var(--text-primary)"
                />

                {/* 6 Stage Diamond Markers around the Orbit */}
                {stages.map((st, idx) => {
                  const angle = (idx / stages.length) * 360 - 90;
                  const rad = (angle * Math.PI) / 180;
                  const x = Math.cos(rad) * 78;
                  const y = Math.sin(rad) * 78;
                  const isCurrent = idx === activeStage;

                  return (
                    <g key={st.num} onClick={() => setActiveStage(idx)} style={{ cursor: 'pointer' }}>
                      <rect
                        x={x - 4}
                        y={y - 4}
                        width="8"
                        height="8"
                        fill={isCurrent ? 'var(--text-primary)' : 'var(--bg-elevated)'}
                        stroke={isCurrent ? 'var(--text-primary)' : st.accent}
                        strokeWidth="1.5"
                        transform={`rotate(45 ${x} ${y})`}
                        style={{
                          filter: isCurrent ? `drop-shadow(0 0 8px ${st.accent})` : 'none',
                          transition: 'all 0.3s ease',
                        }}
                      />
                    </g>
                  );
                })}

                {/* Arc Section on the Right ("SHIPS ONLY AFTER PROVEN UPLIFT") */}
                <path
                  d="M 70 -50 A 98 98 0 0 1 70 50"
                  fill="none"
                  stroke="var(--color-pink)"
                  strokeWidth="2.5"
                  style={{ filter: 'drop-shadow(0 0 8px var(--color-pink))' }}
                />
              </g>

              {/* Connecting Output Line from Dial to Right Outcome Box */}
              <line x1="423" y1="170" x2="520" y2="170" stroke="var(--border-strong)" strokeWidth="1.2" />
            </svg>

            {/* Dial Text Labels with ReviveOS Stages */}
            <div
              style={{
                position: 'absolute',
                top: '10px',
                left: '38%',
                fontSize: '10px',
                color: activeStage === 0 ? '#0ea5e9' : 'var(--text-muted)',
                fontWeight: 700,
                letterSpacing: '0.1em',
              }}
            >
              01 CLASSIFY
            </div>

            <div
              style={{
                position: 'absolute',
                top: '10px',
                right: '12%',
                fontSize: '10px',
                color: activeStage === 1 ? '#8b5cf6' : 'var(--text-muted)',
                fontWeight: 700,
                letterSpacing: '0.1em',
              }}
            >
              02 PREDICT
            </div>

            <div
              style={{
                position: 'absolute',
                bottom: '12px',
                left: '40%',
                fontSize: '10px',
                color: activeStage === 2 ? '#ec4899' : 'var(--text-muted)',
                fontWeight: 700,
                letterSpacing: '0.1em',
              }}
            >
              03 AI REASON
            </div>

            <div
              style={{
                position: 'absolute',
                bottom: '12px',
                right: '10%',
                fontSize: '10px',
                color: activeStage === 3 ? '#f59e0b' : 'var(--text-muted)',
                fontWeight: 700,
                letterSpacing: '0.1em',
              }}
            >
              04 TIMING
            </div>

            {/* "SHIPS ONLY AFTER PROVEN UPLIFT" Label along the Right Arc */}
            <div
              style={{
                position: 'absolute',
                right: '16px',
                top: '40%',
                fontSize: '8.5px',
                letterSpacing: '0.14em',
                color: 'var(--color-pink)',
                fontWeight: 800,
                maxWidth: '120px',
                textAlign: 'left',
              }}
            >
              SHIPS ONLY AFTER PROVEN UPLIFT
            </div>
          </div>

          {/* ════ RIGHT COLUMN: MEASURED OUTCOME / UPLIFT BOX WITH HATCHED PATTERN ════ */}
          <div style={{ display: 'flex', alignItems: 'center', gap: '14px' }}>
            {/* The Precision Hatched Box */}
            <div
              style={{
                width: '100%',
                minHeight: '230px',
                border: '1px solid var(--border-strong)',
                borderRadius: '10px',
                padding: '16px 14px',
                background: 'var(--bg-elevated)',
                position: 'relative',
                boxShadow: 'var(--card-shadow)',
              }}
            >
              {/* Corner crosshair brackets */}
              <div style={{ position: 'absolute', top: '-4px', left: '-4px', width: '8px', height: '8px', borderTop: '1.5px solid var(--color-pink)', borderLeft: '1.5px solid var(--color-pink)' }} />
              <div style={{ position: 'absolute', top: '-4px', right: '-4px', width: '8px', height: '8px', borderTop: '1.5px solid var(--color-pink)', borderRight: '1.5px solid var(--color-pink)' }} />
              <div style={{ position: 'absolute', bottom: '-4px', left: '-4px', width: '8px', height: '8px', borderBottom: '1.5px solid var(--color-pink)', borderLeft: '1.5px solid var(--color-pink)' }} />
              <div style={{ position: 'absolute', bottom: '-4px', right: '-4px', width: '8px', height: '8px', borderBottom: '1.5px solid var(--color-pink)', borderRight: '1.5px solid var(--color-pink)' }} />

              {/* Vertical Tick Ruler on Left Edge */}
              <div
                style={{
                  position: 'absolute',
                  left: '8px',
                  top: '20px',
                  bottom: '20px',
                  width: '6px',
                  borderLeft: '1px solid var(--border-default)',
                  display: 'flex',
                  flexDirection: 'column',
                  justifyContent: 'space-between',
                }}
              >
                {[...Array(6)].map((_, i) => (
                  <div key={i} style={{ width: '4px', height: '1px', background: 'var(--border-strong)' }} />
                ))}
              </div>

              {/* Top Label: POST-RECOVERY with highlighted endpoint node */}
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  fontSize: '10px',
                  color: 'var(--text-primary)',
                  letterSpacing: '0.14em',
                  fontWeight: 700,
                  marginBottom: '10px',
                  paddingLeft: '14px',
                }}
              >
                <span>POST-RECOVERY</span>
                <span
                  style={{
                    width: '6px',
                    height: '6px',
                    borderRadius: '50%',
                    background: '#10b981',
                    boxShadow: '0 0 8px #10b981',
                  }}
                />
              </div>

              {/* Hatched Graphic Fill Box in Purple/Pink Theme */}
              <div
                style={{
                  height: '110px',
                  marginLeft: '14px',
                  border: '1px dashed var(--border-strong)',
                  borderRadius: '6px',
                  background: 'var(--bg-surface)',
                  backgroundImage: `
                    repeating-linear-gradient(
                      45deg,
                      var(--color-accent-bg),
                      var(--color-accent-bg) 4px,
                      transparent 4px,
                      transparent 9px
                    )
                  `,
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: 'center',
                  justifyContent: 'center',
                  padding: '8px',
                }}
              >
                <div style={{ fontSize: '28px', fontWeight: 900, color: 'var(--text-primary)', letterSpacing: '-1px' }}>
                  +68.4%
                </div>
                <div style={{ fontSize: '9px', color: 'var(--color-accent)', letterSpacing: '0.12em', fontWeight: 700 }}>
                  CAPTURED REVENUE
                </div>
              </div>

              {/* Bottom Label: BASELINE (HELD-OUT TASKS) */}
              <div
                style={{
                  marginTop: '12px',
                  paddingLeft: '14px',
                  fontSize: '9.5px',
                  letterSpacing: '0.14em',
                  color: 'var(--text-secondary)',
                }}
              >
                <div style={{ fontWeight: 700, color: 'var(--text-primary)' }}>BASELINE</div>
                <div style={{ fontSize: '8.5px', color: 'var(--text-muted)' }}>9.2% INVOLUNTARY CHURN</div>
              </div>
            </div>

            {/* Big Pink/Purple Delta Symbol & "MEASURED UPLIFT" Label */}
            <div style={{ textAlign: 'center', minWidth: '75px' }}>
              <div
                style={{
                  fontSize: '32px',
                  fontWeight: 900,
                  color: 'var(--color-pink)',
                  lineHeight: 1,
                  filter: 'drop-shadow(0 0 10px rgba(236, 72, 153, 0.4))',
                }}
              >
                Δ
              </div>
              <div
                style={{
                  fontSize: '9.5px',
                  fontWeight: 800,
                  color: 'var(--color-pink)',
                  letterSpacing: '0.15em',
                  marginTop: '6px',
                  lineHeight: 1.3,
                }}
              >
                MEASURED
                <br />
                UPLIFT
              </div>
            </div>
          </div>
        </div>

        {/* ════ CURRENT ACTIVE STAGE TELEMETRY FOOTNOTE BAR ════ */}
        <AnimatePresence mode="wait">
          <motion.div
            key={current.num}
            initial={{ opacity: 0, y: 6 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -6 }}
            transition={{ duration: 0.2 }}
            style={{
              marginTop: '22px',
              padding: '12px 18px',
              borderRadius: '10px',
              background: 'var(--bg-elevated)',
              border: `1px solid ${current.accent}`,
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              fontSize: '11.5px',
              flexWrap: 'wrap',
              gap: '10px',
            }}
          >
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
              {current.icon}
              <span style={{ fontWeight: 800, color: 'var(--text-primary)' }}>
                STAGE {current.num} — {current.title}
              </span>
              <span style={{ color: current.accent }}>({current.subtitle})</span>
            </div>
            <div style={{ color: 'var(--text-secondary)', fontSize: '11px' }}>
              {current.desc}
            </div>
            <div style={{ color: '#10b981', fontWeight: 700 }}>
              {current.metric}
            </div>
          </motion.div>
        </AnimatePresence>

        {/* ════ BOTTOM CONTINUOUS FEEDBACK LOOP WIRE & LABEL ════ */}
        <div
          style={{
            marginTop: '22px',
            paddingTop: '16px',
            borderTop: '1px dashed var(--border-default)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontSize: '10px',
            letterSpacing: '0.16em',
            color: 'var(--text-secondary)',
            textTransform: 'uppercase',
            position: 'relative',
            zIndex: 2,
          }}
        >
          <span>RECONCILED TRANSACTIONS RETRAIN INFERENCE HEURISTICS → CONTINUOUS RECOVERY LOOP → ZERO CHURN</span>
        </div>
      </div>
    </section>
  );
}

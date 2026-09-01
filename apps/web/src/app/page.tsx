'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { motion } from 'framer-motion';
import {
  Zap,
  ShieldCheck,
  Cpu,
  ArrowRight,
  Sparkles,
  TrendingUp,
  RefreshCw,
  Mail,
  Lock,
  Activity,
  Play,
  CheckCircle2,
  ExternalLink,
  Code2,
} from 'lucide-react';
import {
  ScrollProgressBar,
  HeroSpotlight,
  ScrollOrchestrator,
  HorizontalScrollSection,
  TextScrubReveal,
  KineticMarquee,
  DepthParallaxSection,
} from '@/components/ui/MotionScrollComponents';
import { BadgePulse } from '@/components/ui/AnimatedComponents';
import { ThemeToggle, FloatingThemeToggle } from '@/components/ui/ThemeToggle';
import LightStreakHero from '@/components/ui/LightStreakHero';
import TechnicalPipelineDiagram from '@/components/ui/TechnicalPipelineDiagram';
import ParticleMorphingSection from '@/components/ui/ParticleMorphingSection';
import DeveloperDocsSection from '@/components/ui/DeveloperDocsSection';
import ReviveLogo from '@/components/ui/ReviveLogo';
import { createSandboxPaymentLink, getMerchants } from '@/lib/api';

export default function LandingPage() {
  // Simulator State
  const [selectedError, setSelectedError] = useState('INSUFFICIENT_FUNDS');
  const [simAmount, setSimAmount] = useState(2499);
  const [isSimulating, setIsSimulating] = useState(false);
  const [simStage, setSimStage] = useState(0);

  // ROI Calculator State
  const [monthlyGMV, setMonthlyGMV] = useState(2500000); // 25 Lakhs default

  const runSimulation = async () => {
    setIsSimulating(true);
    setSimStage(1);

    // Synchronously create a real simulated payment failure in PostgreSQL so it syncs to /workflows
    try {
      const merchants = await getMerchants().catch(() => []);
      const merchantId = merchants[0]?.id || '11111111-1111-1111-1111-111111111111';
      await createSandboxPaymentLink({
        merchant_id: merchantId,
        customer_email: `delivered@resend.dev`,
        amount: simAmount,
        trigger_failure_immediately: true,
        failure_code: selectedError,
        description: `Simulated ${selectedError} Recovery`,
      });
    } catch (err) {
      console.warn('Simulation sync notice:', err);
    }

    setTimeout(() => setSimStage(2), 700);
    setTimeout(() => setSimStage(3), 1500);
    setTimeout(() => setSimStage(4), 2300);
    setTimeout(() => {
      setSimStage(5);
      setIsSimulating(false);
    }, 3200);
  };

  // ROI Math
  const failureRate = 0.09;
  const lostRevenue = monthlyGMV * failureRate;
  const recoveredMonthly = lostRevenue * 0.684;
  const recoveredAnnual = recoveredMonthly * 12;

  const failureProfiles: Record<
    string,
    { title: string; desc: string; action: string; prob: string; strategy: string }
  > = {
    INSUFFICIENT_FUNDS: {
      title: 'Insufficient Funds / Low Balance',
      desc: 'Customer balance low at billing cycle time.',
      action: 'DELAYED_RETRY (Zero-Touch Token)',
      prob: '76.4%',
      strategy: 'AI schedules smart token retry on next payday / salary window.',
    },
    EXPIRED_CARD: {
      title: 'Expired Card / Mandate Lapsed',
      desc: 'Recurring card or e-mandate has expired.',
      action: 'PAYMENT_METHOD_UPDATE (Resend Link)',
      prob: '62.1%',
      strategy: 'Dispatches high-converting payment method update link via Resend email.',
    },
    BANK_DOWNTIME: {
      title: 'Temporary Bank Downtime / Issuer Offline',
      desc: 'Issuer switch timeout or temporary bank network drop.',
      action: 'EXPONENTIAL_BACKOFF_RETRY',
      prob: '94.2%',
      strategy: 'Predictive algorithm retries automatically when gateway switch recovers.',
    },
    LIMIT_EXCEEDED: {
      title: 'Velocity Limit Exceeded',
      desc: 'Daily or transaction banking volume limit reached.',
      action: 'PAYMENT_LINK (Smart Multi-Rail)',
      prob: '81.0%',
      strategy: 'Sends instant multi-rail smart checkout link with UPI and Netbanking fallbacks.',
    },
  };

  // Pipeline Stages
  const pipelineStages = [
    {
      num: '01',
      title: 'Deterministic Gateway Classifier',
      subtitle: 'Failure Taxonomy Engine',
      tag: 'Classification',
      color: '#38bdf8',
      accentGradient: 'linear-gradient(90deg, #38bdf8, #818cf8)',
      description:
        'Instant deterministic taxonomy mapping of raw gateway decline codes into actionable operational categories.',
    },
    {
      num: '02',
      title: 'Revenue Risk & ML Fraud Guard',
      subtitle: 'Random Forest & Gradient Boosting',
      tag: 'Risk Protection',
      color: '#f59e0b',
      accentGradient: 'linear-gradient(90deg, #f59e0b, #ef4444)',
      description:
        'Dual-model AI engine evaluates transaction anomaly scores, return risk, and expected financial loss to prevent fraud leakage.',
    },
    {
      num: '03',
      title: 'Recovery Probability Model',
      subtitle: 'Logistic Calibration Engine',
      tag: 'Prediction',
      color: '#c084fc',
      accentGradient: 'linear-gradient(90deg, #c084fc, #ec4899)',
      description:
        'Statistical machine learning model calculates exact recoverability probability calibrated against historical outcomes.',
    },
    {
      num: '04',
      title: 'Contextual AI Reasoning',
      subtitle: 'DeepSeek-R1 Strategy Engine',
      tag: 'Strategy',
      color: '#ec4899',
      accentGradient: 'linear-gradient(90deg, #ec4899, #f43f5e)',
      description:
        'Context-augmented LLM reasons over empirical recovery distributions to select optimal timing and highest-yield actions.',
    },
    {
      num: '05',
      title: 'Policy Safety & Guardrails',
      subtitle: 'Deterministic Safety Engine',
      tag: 'Governance',
      color: '#10b981',
      accentGradient: 'linear-gradient(90deg, #10b981, #06b6d4)',
      description:
        'Hard business rules enforce merchant velocity limits, customer communication opt-outs, and transaction ceilings.',
    },
    {
      num: '06',
      title: 'Authoritative Verification',
      subtitle: 'SHA-256 Cryptographic Audit Ledger',
      tag: 'Verification',
      color: '#a855f7',
      accentGradient: 'linear-gradient(90deg, #a855f7, #6366f1)',
      description:
        'Every recovered rupee is validated directly against gateway settlement webhooks and permanently recorded with cryptographic block hashing.',
    },
  ];

  // Horizontal Scroll Feature Cards
  const horizontalCards = [
    {
      id: 'h1',
      num: '01',
      title: '1-Command Razorpay Layer',
      desc: 'Plug AI payment recovery directly into your existing Razorpay setup with `npx reviveos init`. Normal payments never get proxied.',
      gradient: 'linear-gradient(135deg, #082436 0%, #03101c 100%)',
      accent: '#38bdf8',
      icon: <Zap size={18} color="#38bdf8" />,
    },
    {
      id: 'h2',
      num: '02',
      title: 'ML Fraud & Revenue Risk Guard',
      desc: 'Random Forest and Gradient Boosting models evaluate fraud probability and return risk to block bad actors before retrying.',
      gradient: 'linear-gradient(135deg, #361a08 0%, #1a0b02 100%)',
      accent: '#f59e0b',
      icon: <ShieldCheck size={18} color="#f59e0b" />,
    },
    {
      id: 'h3',
      num: '03',
      title: 'Zero-Touch Token Retries',
      desc: 'Recover failed subscriptions and recurring autopay automatically using stored RBI-compliant tokens on optimal bank switch timings.',
      gradient: 'linear-gradient(135deg, #1e0836 0%, #0c0418 100%)',
      accent: '#c084fc',
      icon: <RefreshCw size={18} color="#c084fc" />,
    },
    {
      id: 'h4',
      num: '04',
      title: 'Smart Fallback Links',
      desc: 'Generate direct 1-click payment links with custom branded checkout pages delivered straight via Resend email and WhatsApp.',
      gradient: 'linear-gradient(135deg, #36081e 0%, #1c030f 100%)',
      accent: '#ec4899',
      icon: <Mail size={18} color="#ec4899" />,
    },
    {
      id: 'h5',
      num: '05',
      title: 'Customer Recovery Memory',
      desc: 'Closed-loop calibration tracks customer-specific payment behavior, preferred methods, and historical recovery latency.',
      gradient: 'linear-gradient(135deg, #08361e 0%, #031a0e 100%)',
      accent: '#10b981',
      icon: <Cpu size={18} color="#10b981" />,
    },
    {
      id: 'h6',
      num: '06',
      title: 'SHA-256 Audit Ledger',
      desc: 'Immutable cryptographic audit trail for complete financial compliance. Every state transition is hash-chained.',
      gradient: 'linear-gradient(135deg, #1a0836 0%, #0d031c 100%)',
      accent: '#8b5cf6',
      icon: <Lock size={18} color="#8b5cf6" />,
    },
  ];

  return (
    <div style={{ position: 'relative', backgroundColor: 'var(--bg-primary)', minHeight: '100vh', overflowX: 'hidden' }}>
      {/* 1. Scroll Progress Bar at Top */}
      <ScrollProgressBar />

      {/* 2. Hero Ambient Spotlight */}
      <HeroSpotlight />

      {/* 3. Sticky Glass Navbar with Theme Toggle */}
      <header
        style={{
          position: 'fixed',
          top: 0,
          left: 0,
          right: 0,
          zIndex: 50,
          backdropFilter: 'blur(20px)',
          WebkitBackdropFilter: 'blur(20px)',
          backgroundColor: 'var(--navbar-bg)',
          borderBottom: '1px solid var(--border-subtle)',
          padding: '0 28px',
          height: '60px',
          transition: 'background-color 0.3s ease',
        }}
      >
        <div style={{ maxWidth: '1280px', margin: '0 auto', display: 'flex', alignItems: 'center', justifyContent: 'space-between', height: '100%' }}>
          <ReviveLogo size="md" href="/" />

          <nav style={{ display: 'flex', alignItems: 'center', gap: '28px' }}>
            {[
              { label: 'Pipeline', href: '#pipeline' },
              { label: 'Documentation', href: '/docs' },
              { label: 'Capabilities', href: '#features' },
              { label: 'Simulator', href: '#simulator' },
              { label: 'ROI Estimator', href: '#calculator' },
            ].map((item) => (
              <a
                key={item.href}
                href={item.href}
                style={{
                  color: 'var(--text-secondary)',
                  fontSize: '13.5px',
                  fontWeight: 500,
                  transition: 'color 0.2s ease',
                }}
                onMouseEnter={(e) => (e.currentTarget.style.color = 'var(--text-primary)')}
                onMouseLeave={(e) => (e.currentTarget.style.color = 'var(--text-secondary)')}
              >
                {item.label}
              </a>
            ))}
          </nav>

          <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
            {/* Dark / Light / System Mode Switcher */}
            <ThemeToggle />

            <Link href="/developer">
              <button className="btn-ghost" style={{ padding: '7px 14px', fontSize: '13px', display: 'flex', alignItems: 'center', gap: '6px' }}>
                <Code2 size={14} color="var(--color-accent-blue)" />
                <span>SDK / CLI</span>
              </button>
            </Link>
            <Link href="/merchant">
              <button className="btn-primary" style={{ padding: '7px 18px', fontSize: '13px' }}>
                Merchant Portal <ArrowRight size={13} />
              </button>
            </Link>
          </div>
        </div>
      </header>

      {/* 4. CINEMATIC NEON LIGHT STREAK & ENERGY TRAIL HERO ENTRY */}
      <LightStreakHero />

      {/* 5. WORD-BY-WORD TEXT SCRUB REVEAL */}
      <TextScrubReveal text="Every failed payment tells a story. The best payment platforms do not just log errors — they orchestrate recovery with mathematical precision. Timing creates assurance. AI creates conversion. And the scroll becomes the narrative." />

      {/* 6. TECHNICAL AI PIPELINE SCHEMATIC ARCHITECTURE */}
      <div id="pipeline">
        <TechnicalPipelineDiagram />
      </div>

      {/* 7. DEVELOPER PLATFORM & 1-COMMAND SDK INTEGRATION DOCS */}
      <DeveloperDocsSection />

      {/* 8. KINETIC MARQUEE RIBBONS */}
      <KineticMarquee />

      {/* 9. HORIZONTAL FEATURE CAROUSEL SLIDER (Smooth & No Black Void) */}
      <div id="features">
        <HorizontalScrollSection
          cards={horizontalCards}
          heading="Built for Zero Involuntary Churn"
          subheading="Explore ReviveOS autonomous payment recovery and ML fraud protection infrastructure."
        />
      </div>

      {/* 10. 3D PARTICLE MORPHING / GENERATIVE FLOW FIELD & CONSTELLATION NETWORK */}
      <ParticleMorphingSection />

      {/* 11. INTERACTIVE LIVE SIMULATOR (Medium/Compact Size) */}
      <section
        id="simulator"
        style={{
          position: 'relative',
          zIndex: 10,
          padding: '60px 24px',
          maxWidth: '880px',
          margin: '0 auto',
        }}
      >
        <div style={{ textAlign: 'center', marginBottom: '36px' }}>
          <div style={{ display: 'inline-flex', marginBottom: '10px' }}>
            <BadgePulse text="Live Interactive Demo" variant="info" />
          </div>
          <h2 style={{ fontSize: 'clamp(26px, 4vw, 40px)', fontWeight: 800, letterSpacing: '-1px', marginBottom: '8px', color: 'var(--text-primary)' }}>
            Adaptive Recovery Engine
          </h2>
          <p style={{ color: 'var(--text-secondary)', maxWidth: '520px', margin: '0 auto', fontSize: '14.5px' }}>
            Select a decline scenario and watch ReviveOS diagnose, predict, and execute recovery in real time.
          </p>
        </div>

        <div
          style={{
            background: 'var(--bg-card)',
            backdropFilter: 'blur(20px)',
            border: '1px solid var(--border-subtle)',
            borderRadius: '24px',
            padding: '28px',
            boxShadow: 'var(--card-shadow)',
          }}
        >
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: '24px' }}>
            {/* Left Controls */}
            <div>
              <h3 style={{ fontSize: '15px', fontWeight: 700, marginBottom: '16px', display: 'flex', alignItems: 'center', gap: '8px', color: 'var(--text-primary)' }}>
                <Cpu size={16} color="#c084fc" /> 1. Select Decline Scenario
              </h3>

              <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', marginBottom: '20px' }}>
                {Object.keys(failureProfiles).map((key) => {
                  const item = failureProfiles[key];
                  const isSelected = selectedError === key;
                  return (
                    <div
                      key={key}
                      onClick={() => {
                        setSelectedError(key);
                        setSimStage(0);
                      }}
                      style={{
                        padding: '12px 14px',
                        borderRadius: '12px',
                        background: isSelected ? 'rgba(139, 92, 246, 0.15)' : 'var(--bg-input)',
                        border: `1px solid ${isSelected ? '#8b5cf6' : 'var(--border-subtle)'}`,
                        cursor: 'pointer',
                        transition: 'all 0.2s ease',
                      }}
                    >
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '4px' }}>
                        <span style={{ fontWeight: 700, color: isSelected ? '#c084fc' : 'var(--text-primary)', fontSize: '13px' }}>
                          {item.title}
                        </span>
                        <span style={{ fontSize: '10.5px', color: isSelected ? '#c084fc' : 'var(--text-muted)' }} className="mono">
                          {key}
                        </span>
                      </div>
                      <p style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>{item.desc}</p>
                    </div>
                  );
                })}
              </div>

              {/* Amount Selector */}
              <div style={{ marginBottom: '20px' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '6px' }}>
                  <label style={{ fontSize: '13px', fontWeight: 600, color: 'var(--text-secondary)' }}>
                    Transaction Amount
                  </label>
                  <span style={{ color: '#ec4899', fontWeight: 800, fontSize: '14.5px' }} className="mono">
                    ₹{simAmount.toLocaleString()}
                  </span>
                </div>
                <input
                  type="range"
                  min="499"
                  max="19999"
                  step="500"
                  value={simAmount}
                  onChange={(e) => setSimAmount(Number(e.target.value))}
                  style={{ width: '100%' }}
                />
              </div>

              <button
                onClick={runSimulation}
                disabled={isSimulating}
                className="btn-primary"
                style={{ width: '100%', padding: '11px', fontSize: '14px' }}
              >
                <Sparkles size={15} />
                {isSimulating ? 'Analyzing Transaction...' : 'Run AI Recovery'}
              </button>
            </div>

            {/* Right Execution Trace */}
            <div
              style={{
                background: 'var(--bg-elevated)',
                borderRadius: '16px',
                border: '1px solid var(--border-subtle)',
                padding: '20px',
                display: 'flex',
                flexDirection: 'column',
                justifyContent: 'space-between',
              }}
            >
              <div>
                <div
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    marginBottom: '16px',
                    borderBottom: '1px solid var(--border-subtle)',
                    paddingBottom: '10px',
                  }}
                >
                  <span style={{ fontSize: '12px', fontWeight: 700, letterSpacing: '0.06em', color: 'var(--text-secondary)' }}>
                    ORCHESTRATION TIMELINE
                  </span>
                  <BadgePulse
                    text={simStage === 5 ? 'RECOVERED' : simStage > 0 ? 'PROCESSING' : 'IDLE'}
                    variant={simStage === 5 ? 'success' : simStage > 0 ? 'warning' : 'info'}
                  />
                </div>

                <div style={{ display: 'flex', flexDirection: 'column', gap: '14px' }}>
                  {/* Step 1 */}
                  <div style={{ display: 'flex', gap: '10px', opacity: simStage >= 1 ? 1 : 0.25, transition: 'opacity 0.3s' }}>
                    <div style={{ width: '22px', height: '22px', borderRadius: '50%', background: simStage >= 1 ? '#38bdf8' : 'var(--border-default)', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '11px', fontWeight: 800, flexShrink: 0, color: '#fff' }}>
                      1
                    </div>
                    <div>
                      <div style={{ fontWeight: 700, fontSize: '13px', color: 'var(--text-primary)' }}>Diagnose Failure Taxonomy</div>
                      <div style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>
                        Mapped to: <span className="mono" style={{ color: '#38bdf8' }}>{selectedError}</span>
                      </div>
                    </div>
                  </div>

                  {/* Step 2 */}
                  <div style={{ display: 'flex', gap: '10px', opacity: simStage >= 2 ? 1 : 0.25, transition: 'opacity 0.3s' }}>
                    <div style={{ width: '22px', height: '22px', borderRadius: '50%', background: simStage >= 2 ? '#c084fc' : 'var(--border-default)', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '11px', fontWeight: 800, flexShrink: 0, color: '#fff' }}>
                      2
                    </div>
                    <div>
                      <div style={{ fontWeight: 700, fontSize: '13px', color: 'var(--text-primary)' }}>Probability Prediction</div>
                      <div style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>
                        Chance: <span style={{ color: '#10b981', fontWeight: 700 }}>{failureProfiles[selectedError].prob}</span>
                      </div>
                    </div>
                  </div>

                  {/* Step 3 */}
                  <div style={{ display: 'flex', gap: '10px', opacity: simStage >= 3 ? 1 : 0.25, transition: 'opacity 0.3s' }}>
                    <div style={{ width: '22px', height: '22px', borderRadius: '50%', background: simStage >= 3 ? '#ec4899' : 'var(--border-default)', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '11px', fontWeight: 800, flexShrink: 0, color: '#fff' }}>
                      3
                    </div>
                    <div>
                      <div style={{ fontWeight: 700, fontSize: '13px', color: 'var(--text-primary)' }}>DeepSeek-R1 Decision</div>
                      <div style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>
                        Action: <span style={{ color: '#ec4899', fontWeight: 700 }}>{failureProfiles[selectedError].action}</span>
                      </div>
                    </div>
                  </div>

                  {/* Step 4 */}
                  <div style={{ display: 'flex', gap: '10px', opacity: simStage >= 4 ? 1 : 0.25, transition: 'opacity 0.3s' }}>
                    <div style={{ width: '22px', height: '22px', borderRadius: '50%', background: simStage >= 4 ? '#f59e0b' : 'var(--border-default)', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '11px', fontWeight: 800, flexShrink: 0, color: '#fff' }}>
                      4
                    </div>
                    <div>
                      <div style={{ fontWeight: 700, fontSize: '13px', color: 'var(--text-primary)' }}>Autonomous Execution</div>
                      <div style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>
                        {failureProfiles[selectedError].strategy}
                      </div>
                    </div>
                  </div>

                  {/* Step 5 */}
                  <div style={{ display: 'flex', gap: '10px', opacity: simStage >= 5 ? 1 : 0.25, transition: 'opacity 0.3s' }}>
                    <div style={{ width: '22px', height: '22px', borderRadius: '50%', background: simStage >= 5 ? '#10b981' : 'var(--border-default)', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '11px', fontWeight: 800, flexShrink: 0, color: '#fff' }}>
                      ✓
                    </div>
                    <div>
                      <div style={{ fontWeight: 700, fontSize: '13px', color: simStage >= 5 ? '#10b981' : 'var(--text-primary)' }}>
                        Provider Verification
                      </div>
                      <div style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>
                        Status: <span style={{ color: '#10b981', fontWeight: 700 }}>CAPTURED (₹{simAmount.toLocaleString()})</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              {simStage === 5 && (
                <div
                  style={{
                    marginTop: '16px',
                    padding: '10px 14px',
                    borderRadius: '10px',
                    background: 'var(--color-emerald-bg)',
                    border: '1px solid var(--color-emerald-border)',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '10px',
                  }}
                >
                  <CheckCircle2 size={16} color="#10b981" />
                  <span style={{ fontSize: '12px', color: '#10b981', fontWeight: 600 }}>
                    ₹{simAmount.toLocaleString()} recovered without customer friction.
                  </span>
                </div>
              )}
            </div>
          </div>
        </div>
      </section>

      {/* 11. FINANCIAL YIELD / ROI ESTIMATOR (Medium/Compact Size) */}
      <section
        id="calculator"
        style={{
          position: 'relative',
          zIndex: 10,
          padding: '60px 24px',
          maxWidth: '880px',
          margin: '0 auto',
        }}
      >
        <div
          style={{
            background: 'var(--bg-card)',
            border: '1px solid var(--border-subtle)',
            borderRadius: '24px',
            padding: '36px 32px',
            boxShadow: 'var(--card-shadow)',
          }}
        >
          <div style={{ textAlign: 'center', marginBottom: '32px' }}>
            <BadgePulse text="Financial Yield Estimator" variant="success" />
            <h2 style={{ fontSize: 'clamp(24px, 3.5vw, 36px)', fontWeight: 800, marginTop: '12px', marginBottom: '6px', color: 'var(--text-primary)' }}>
              How Much Revenue is Leaking?
            </h2>
            <p style={{ color: 'var(--text-secondary)', fontSize: '14.5px' }}>
              Calculate recurring revenue recapture with ReviveOS.
            </p>
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(260px, 1fr))', gap: '32px', alignItems: 'center' }}>
            {/* Slider */}
            <div>
              <label style={{ fontSize: '13px', fontWeight: 600, color: 'var(--text-secondary)', display: 'block', marginBottom: '8px' }}>
                Monthly Payment Volume (GMV):
              </label>
              <div style={{ fontSize: '28px', fontWeight: 900, color: '#ec4899', marginBottom: '14px' }} className="mono">
                ₹{monthlyGMV.toLocaleString()} <span style={{ fontSize: '13px', color: 'var(--text-muted)' }}>/ month</span>
              </div>
              <input
                type="range"
                min="500000"
                max="50000000"
                step="500000"
                value={monthlyGMV}
                onChange={(e) => setMonthlyGMV(Number(e.target.value))}
                style={{ width: '100%', marginBottom: '10px' }}
              />
              <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '11px', color: 'var(--text-muted)' }}>
                <span>₹5 Lakhs</span>
                <span>₹2.5 Crores</span>
                <span>₹5 Crores</span>
              </div>
            </div>

            {/* Output Cards */}
            <div style={{ display: 'flex', flexDirection: 'column', gap: '14px' }}>
              <div
                style={{
                  padding: '16px 18px',
                  borderRadius: '14px',
                  background: 'var(--color-red-bg)',
                  border: '1px solid var(--color-red-border)',
                }}
              >
                <div style={{ fontSize: '11px', color: '#ef4444', fontWeight: 700, letterSpacing: '0.04em' }}>
                  INVOLUNTARY CHURN (9%)
                </div>
                <div style={{ fontSize: '22px', fontWeight: 900, color: '#ef4444', marginTop: '4px' }} className="mono">
                  -₹{Math.round(lostRevenue).toLocaleString()} <span style={{ fontSize: '12px' }}>/mo</span>
                </div>
              </div>

              <div
                style={{
                  padding: '16px 18px',
                  borderRadius: '14px',
                  background: 'var(--color-emerald-bg)',
                  border: '1px solid var(--color-emerald-border)',
                }}
              >
                <div style={{ fontSize: '11px', color: '#10b981', fontWeight: 700, letterSpacing: '0.04em' }}>
                  RECOVERED BY REVIVEOS (68.4%)
                </div>
                <div style={{ fontSize: '26px', fontWeight: 900, color: '#10b981', marginTop: '4px' }} className="mono">
                  +₹{Math.round(recoveredMonthly).toLocaleString()} <span style={{ fontSize: '13px' }}>/month</span>
                </div>
                <div style={{ fontSize: '12px', color: 'var(--text-secondary)', marginTop: '4px' }}>
                  ≈ <strong style={{ color: 'var(--text-primary)' }}>+₹{Math.round(recoveredAnnual).toLocaleString()}</strong> in recovered annual revenue.
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* 12. CALL TO ACTION SECTION */}
      <section
        style={{
          position: 'relative',
          zIndex: 10,
          padding: '80px 24px',
          maxWidth: '1000px',
          margin: '0 auto',
          textAlign: 'center',
        }}
      >
        <div
          style={{
            background: 'linear-gradient(135deg, rgba(139, 92, 246, 0.15) 0%, rgba(236, 72, 153, 0.15) 100%)',
            border: '1px solid var(--border-default)',
            borderRadius: '28px',
            padding: '52px 28px',
            backdropFilter: 'blur(20px)',
            boxShadow: 'var(--card-shadow)',
          }}
        >
          <h2 style={{ fontSize: 'clamp(26px, 4vw, 42px)', fontWeight: 900, marginBottom: '12px', letterSpacing: '-1px', color: 'var(--text-primary)' }}>
            Stop Losing 9% of Your Revenue
          </h2>
          <p style={{ color: 'var(--text-secondary)', fontSize: '16px', maxWidth: '580px', margin: '0 auto 32px', lineHeight: 1.6 }}>
            Launch ReviveOS in less than 5 minutes. Connect your Razorpay or Stripe account and start recovering payments immediately.
          </p>
          <div style={{ display: 'flex', justifyContent: 'center', gap: '14px', flexWrap: 'wrap' }}>
            <Link href="/login">
              <button className="btn-primary" style={{ padding: '12px 28px', fontSize: '14px' }}>
                Launch Merchant Portal <ArrowRight size={14} />
              </button>
            </Link>
            <Link href="/login">
              <button className="btn-ghost" style={{ padding: '12px 24px', fontSize: '14px' }}>
                View Admin Hub
              </button>
            </Link>
          </div>
        </div>
      </section>

      {/* 13. FOOTER */}
      <footer
        style={{
          borderTop: '1px solid var(--border-subtle)',
          padding: '36px 28px',
          backgroundColor: 'var(--bg-surface)',
          position: 'relative',
          zIndex: 10,
        }}
      >
        <div
          style={{
            maxWidth: '1280px',
            margin: '0 auto',
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            flexWrap: 'wrap',
            gap: '20px',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
            <ReviveLogo size="sm" href="/" />
            <span style={{ fontSize: '12px', color: 'var(--text-muted)' }}>
              © 2026 ReviveOS Platform. All rights reserved.
            </span>
          </div>

          <div style={{ display: 'flex', gap: '20px', fontSize: '13px', color: 'var(--text-secondary)' }}>
            <Link href="/login" style={{ transition: 'color 0.2s' }}>Merchant Sign In</Link>
            <Link href="/login" style={{ transition: 'color 0.2s' }}>Admin Hub</Link>
            <a
              href="https://github.com/JATINCHULETT/ReviveOS"
              target="_blank"
              rel="noreferrer"
              style={{ display: 'flex', alignItems: 'center', gap: '6px' }}
            >
              GitHub <ExternalLink size={12} />
            </a>
          </div>
        </div>
      </footer>

      {/* Floating Theme Switcher Button */}
      <FloatingThemeToggle />
    </div>
  );
}

'use client';

import React, { useState, useEffect, useMemo } from 'react';
import Link from 'next/link';
import {
  Search,
  Sparkles,
  Terminal,
  Copy,
  Check,
  ChevronRight,
  BookOpen,
  Zap,
  ShieldCheck,
  ArrowRight,
  ArrowLeft,
  X,
  Send,
  HelpCircle,
  FileCode,
} from 'lucide-react';

// Documentation Section Types
interface DocArticle {
  id: string;
  title: string;
  category: string;
  description: string;
  readTime: string;
  badge?: string;
  content: {
    heading: string;
    subheading: string;
    sections: Array<{
      title: string;
      body: string;
      code?: string;
      language?: string;
      note?: { type: 'tip' | 'important' | 'warning'; text: string };
      table?: { headers: string[]; rows: string[][] };
    }>;
  };
}

export default function DocsPage() {
  const [activeArticleId, setActiveArticleId] = useState<string>('overview');
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [isSearchOpen, setIsSearchOpen] = useState(false);
  const [isAiOpen, setIsAiOpen] = useState(false);
  const [aiQuestion, setAiQuestion] = useState('');
  const [aiMessages, setAiMessages] = useState<Array<{ role: 'user' | 'assistant'; text: string }>>([
    {
      role: 'assistant',
      text: 'Hi! I am the ReviveOS AI Documentation Assistant. Ask me anything about integrating @reviveos/razorpay, CLI configuration, ML fraud protection, or webhook handlers.',
    },
  ]);
  const [aiLoading, setAiLoading] = useState(false);

  // Keyboard shortcut listener for ⌘K and ⌘I
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault();
        setIsSearchOpen((prev) => !prev);
      }
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'i') {
        e.preventDefault();
        setIsAiOpen((prev) => !prev);
      }
      if (e.key === 'Escape') {
        setIsSearchOpen(false);
        setIsAiOpen(false);
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  const copyToClipboard = (text: string, id: string) => {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  // Documentation Articles Catalog (API Reference removed per user requirement)
  const docArticles: Record<string, DocArticle> = {
    overview: {
      id: 'overview',
      title: 'Start building with ReviveOS',
      category: 'WELCOME',
      description: 'Autonomous payment recovery, ML fraud intelligence, and revenue protection — from first install to production.',
      readTime: '3 min read',
      content: {
        heading: 'Start building with ReviveOS',
        subheading: 'Autonomous payment recovery, ML fraud intelligence, and revenue protection — from first install to production.',
        sections: [
          {
            title: 'What is ReviveOS?',
            body: 'ReviveOS is an AI payment recovery and revenue protection layer that plugs into existing payment infrastructure (starting with Razorpay) via a single command without requiring merchants to replace their gateway or proxy normal successful transactions.\n\nNormal payment traffic continues directly through Razorpay. ReviveOS intercepts webhook failure events, executes deterministic classification, scores fraud and return risk, and orchestrates optimal recovery via stored token retries or smart interactive checkout links.',
            note: {
              type: 'important',
              text: 'ReviveOS does NOT replace Razorpay and does NOT proxy every transaction. Only payment failures and recovery events are processed by the ReviveOS engine.',
            },
          },
          {
            title: 'Core Product Architecture',
            body: 'The ReviveOS platform is structured around a 6-stage autonomous lifecycle:\n\n1. Diagnose: Deterministic gateway failure classification across 10+ bank error categories.\n2. Predict: Statistical machine learning model predicting recovery probability.\n3. Risk Guard: Multi-model ML fraud detection (Random Forest + Gradient Boosting).\n4. Decide: DeepSeek-R1 reasoning engine selecting the next-best action.\n5. Time: Dynamic scheduling aligning retries with salary credit and bank recovery windows.\n6. Verify: Authoritative gateway confirmation recorded on an immutable SHA-256 ledger.',
          },
        ],
      },
    },
    installation: {
      id: 'installation',
      title: 'Installation & SDK Setup',
      category: 'GETTING STARTED',
      description: 'Install the @reviveos/razorpay SDK and TypeScript definitions.',
      readTime: '2 min read',
      content: {
        heading: 'SDK Installation',
        subheading: 'Install the lightweight, zero-native-dependency SDK in any Node.js, Next.js, or Express project.',
        sections: [
          {
            title: '1. Install Package',
            body: 'Install the official ReviveOS Razorpay SDK via npm, pnpm, or yarn:',
            code: 'npm install @reviveos/razorpay\n# or\npnpm add @reviveos/razorpay\n# or\nyarn add @reviveos/razorpay',
            language: 'bash',
          },
          {
            title: 'Package Footprint',
            body: 'The @reviveos/razorpay package is designed to be ultra-lightweight with zero native C++ binaries or heavy ML packages. All machine learning models and LLM reasoning run on the ReviveOS backend API.',
            note: {
              type: 'tip',
              text: 'The SDK includes complete TypeScript definitions (.d.ts) and source maps out of the box.',
            },
          },
        ],
      },
    },
    'cli-config': {
      id: 'cli-config',
      title: 'API Keys & CLI Configuration',
      category: 'GETTING STARTED',
      description: 'Run npx reviveos init to configure your environment and webhook handlers.',
      readTime: '3 min read',
      badge: 'CLI',
      content: {
        heading: 'API Keys & CLI Configuration',
        subheading: 'Use the ReviveOS CLI to inspect your project, generate environment files, and scaffold webhook routes automatically.',
        sections: [
          {
            title: '1. Run the Auto-Configurator',
            body: 'In your project root directory, run:',
            code: 'npx reviveos init',
            language: 'bash',
          },
          {
            title: 'What the CLI Does',
            body: 'The CLI automatically:\n• Detects your Node.js project and package.json\n• Checks for existing Razorpay dependencies\n• Identifies your framework (Next.js App Router, Pages Router, Express, or Fastify)\n• Generates `.env.reviveos` with required variables\n• Creates a ready-to-run webhook handler tailored to your framework\n• Verifies API connectivity and credentials',
          },
          {
            title: 'Environment Variables Template (.env.reviveos)',
            body: 'Configure the following variables in your `.env` or `.env.local`:',
            code: `# Your ReviveOS Merchant API Key (obtained from ReviveOS Merchant Console)\nREVIVEOS_API_KEY=rvo_test_acme_secret_key_12345\n\n# ReviveOS Core API URL\nREVIVEOS_API_URL=https://api.revive-os.me\n\n# Razorpay Webhook Secret for HMAC-SHA256 verification\nRAZORPAY_WEBHOOK_SECRET=whsec_your_razorpay_webhook_secret\n\n# Mode: 'live', 'test', or 'mock' (for local sandbox testing)\nREVIVEOS_MODE=live`,
            language: 'env',
          },
          {
            title: 'Generating Merchant API Keys',
            body: 'To generate a new API key:\n1. Navigate to the Developer Platform in the Merchant Console.\n2. Click "API Keys & Secrets".\n3. Choose your mode (Test Sandbox or Production Live) and click "Create API Key".\n4. Copy the raw key secret immediately (keys are SHA-256 hashed and never displayed again).',
            note: {
              type: 'warning',
              text: 'Never commit your REVIVEOS_API_KEY or RAZORPAY_WEBHOOK_SECRET to public version control.',
            },
          },
        ],
      },
    },
    quickstart: {
      id: 'quickstart',
      title: '60-Second Quickstart',
      category: 'GETTING STARTED',
      description: 'Complete end-to-end integration tutorial in under 60 seconds.',
      readTime: '2 min read',
      badge: 'Popular',
      content: {
        heading: '60-Second Quickstart',
        subheading: 'Add autonomous recovery to your existing Razorpay setup with minimal code changes.',
        sections: [
          {
            title: 'Step 1: Install & Init',
            body: 'Run the setup commands in your repository:',
            code: 'npm install @reviveos/razorpay\nnpx reviveos init',
            language: 'bash',
          },
          {
            title: 'Step 2: Add Next.js App Router Webhook',
            body: 'Create `app/api/reviveos/webhook/route.ts`:',
            code: `import { NextResponse } from "next/server";\nimport { ReviveOS } from "@reviveos/razorpay";\n\nconst revive = new ReviveOS({\n  apiKey: process.env.REVIVEOS_API_KEY,\n  webhookSecret: process.env.RAZORPAY_WEBHOOK_SECRET,\n});\n\nexport async function POST(req: Request) {\n  const rawBody = await req.text();\n  const signature = req.headers.get("x-razorpay-signature") || "";\n\n  // 1. Verify cryptographic HMAC signature & normalize event\n  const event = revive.webhooks.verifyAndNormalize(rawBody, signature);\n\n  // 2. Ingest into ReviveOS AI recovery pipeline\n  const result = await revive.events.process(event);\n\n  return NextResponse.json(result);\n}`,
            language: 'typescript',
          },
          {
            title: 'Step 3: Point Razorpay Dashboard Webhook',
            body: '1. Go to Razorpay Dashboard > Settings > Webhooks > Add New Webhook.\n2. Set Webhook URL to: `https://your-domain.com/api/reviveos/webhook`\n3. Subscribe to events: `payment.failed`, `payment.captured`, `refund.created`\n4. Copy the Webhook Secret and set it as `RAZORPAY_WEBHOOK_SECRET` in your environment.',
          },
        ],
      },
    },
    webhooks: {
      id: 'webhooks',
      title: 'Webhook Verification & Idempotency',
      category: 'GETTING STARTED',
      description: 'Cryptographic signature verification, replay attack prevention, and duplicate event deduplication.',
      readTime: '4 min read',
      content: {
        heading: 'Webhook Verification & Idempotency',
        subheading: 'How ReviveOS secures incoming webhooks with constant-time HMAC-SHA256 checks and transactional idempotency.',
        sections: [
          {
            title: 'Cryptographic Signature Verification',
            body: 'The ReviveOS Webhook Manager uses Node.js `crypto.timingSafeEqual` to verify the `X-Razorpay-Signature` against your `RAZORPAY_WEBHOOK_SECRET` without opening timing attack vulnerabilities.',
            code: `import { RazorpayWebhookManager } from "@reviveos/razorpay";\n\nconst webhook = new RazorpayWebhookManager(process.env.RAZORPAY_WEBHOOK_SECRET);\n\n// Throws ReviveOSSignatureError if invalid\nconst event = webhook.verifyAndNormalize(rawBody, signatureHeader);`,
            language: 'typescript',
          },
          {
            title: 'Idempotent Event Deduplication',
            body: 'Webhooks can be delivered multiple times by payment gateways. ReviveOS implements strict database-backed deduplication using `idempotency_key` and `event_id`.\n\nIf the same event is received twice, ReviveOS acknowledges the request with status `DUPLICATE_IGNORED` and prevents redundant retry actions.',
            table: {
              headers: ['Field', 'Type', 'Description'],
              rows: [
                ['eventId', 'string', 'Unique event ID from gateway header (e.g. evt_12345)'],
                ['idempotencyKey', 'string', 'Derived from eventId to prevent duplicate execution'],
                ['paymentId', 'string', 'Gateway payment ID (e.g. pay_L8k3n2)'],
                ['status', 'string', 'INGESTED or DUPLICATE_IGNORED'],
              ],
            },
          },
        ],
      },
    },
    'ml-fraud': {
      id: 'ml-fraud',
      title: 'Revenue Risk & ML Fraud Guard',
      category: 'CORE CAPABILITIES',
      description: 'Dual-model AI engine evaluating transaction anomaly scores, return risk, and expected financial loss.',
      readTime: '4 min read',
      badge: 'AI Engine',
      content: {
        heading: 'Revenue Risk & ML Fraud Guard',
        subheading: 'Protect merchant revenue by predicting fraud probability and product return risk before executing retries.',
        sections: [
          {
            title: 'Dual-Model Risk Architecture',
            body: 'ReviveOS integrates two specialized machine learning models:\n\n1. Random Forest Fraud Classifier: Analyzes customer velocity, failure codes, and billing anomalies to compute a fraud probability score (0.0 to 1.0) and risk level (LOW, MEDIUM, HIGH).\n\n2. Gradient Boosting Return Risk Model: Predicts the likelihood of post-recovery customer disputes or return abuse.\n\n3. Expected Loss Calculation: Computes Expected Loss = (Fraud Probability × Transaction Amount) to evaluate financial risk vs reward.',
          },
          {
            title: 'Synchronous Risk Inspection',
            body: 'You can query the ML Risk Engine directly via the SDK:',
            code: `const diagnosis = await revive.payments.analyze({\n  paymentId: "pay_test_8921",\n  amount: 85000, // High-value transaction\n  currency: "INR",\n  failureCode: "GATEWAY_TIMEOUT",\n  customerId: "cust_alex_revive",\n  paymentMethod: "card"\n});\n\nconsole.log(diagnosis.fraudRisk);\n// Output:\n// {\n//   fraudProbability: 0.78,\n//   riskLevel: "HIGH",\n//   expectedLoss: 66300.0,\n//   overallRisk: "HIGH"\n// }\nconsole.log(diagnosis.decision); // "BLOCK"`,
            language: 'typescript',
          },
        ],
      },
    },
    'customer-memory': {
      id: 'customer-memory',
      title: 'Customer Recovery Memory',
      category: 'CORE CAPABILITIES',
      description: 'Closed-loop calibration tracking customer habits, preferred payment rails, and recovery latency.',
      readTime: '3 min read',
      content: {
        heading: 'Customer Recovery Memory',
        subheading: 'Make personalized recovery decisions using customer-specific historical payment performance.',
        sections: [
          {
            title: 'Customer Behavior Calibration',
            body: 'Instead of treating all failed payments identically, ReviveOS maintains a dynamic recovery memory profile for every customer. It tracks historical failure frequencies, successful recapture rates, preferred payment rails (e.g. UPI vs Card), and average recovery latency.',
          },
          {
            title: 'Querying Customer Recovery Profile',
            body: 'Retrieve a customer profile using the SDK:',
            code: `const profile = await revive.customers.getRecoveryProfile("cust_alex_revive");\n\nconsole.log(profile);\n// Output:\n// {\n//   customerId: "cust_alex_revive",\n//   previousFailures: 4,\n//   previousSuccessfulRecoveries: 3,\n//   preferredPaymentMethods: ["card", "upi"],\n//   averageRecoveryTime: 180, // in seconds\n//   recoveryProbability: 0.81,\n//   communicationOptOut: false\n// }`,
            language: 'typescript',
          },
        ],
      },
    },
  };

  // Group navigation by category
  const navCategories = useMemo(() => {
    const categories: Record<string, DocArticle[]> = {};
    Object.values(docArticles).forEach((doc) => {
      if (!categories[doc.category]) {
        categories[doc.category] = [];
      }
      categories[doc.category].push(doc);
    });
    return categories;
  }, []);

  // Search filtering
  const searchResults = useMemo(() => {
    if (!searchQuery.trim()) return [];
    const q = searchQuery.toLowerCase();
    return Object.values(docArticles).filter(
      (doc) =>
        doc.title.toLowerCase().includes(q) ||
        doc.description.toLowerCase().includes(q) ||
        doc.category.toLowerCase().includes(q)
    );
  }, [searchQuery]);

  const currentArticle = docArticles[activeArticleId] || docArticles['overview'];

  // AI Assistant Query Handler
  const handleAskAi = (e: React.FormEvent) => {
    e.preventDefault();
    if (!aiQuestion.trim() || aiLoading) return;

    const userQ = aiQuestion.trim();
    setAiMessages((prev) => [...prev, { role: 'user', text: userQ }]);
    setAiQuestion('');
    setAiLoading(true);

    setTimeout(() => {
      let answer = '';
      const qLower = userQ.toLowerCase();

      if (qLower.includes('install') || qLower.includes('npm')) {
        answer = 'To install the SDK, run `npm install @reviveos/razorpay` and then `npx reviveos init` to auto-configure your environment template and webhook handlers.';
      } else if (qLower.includes('api key') || qLower.includes('key')) {
        answer = 'You can generate API keys (`rvo_test_...` or `rvo_live_...`) in the Merchant Console under "API Keys & Secrets". Keys are SHA-256 hashed and displayed only once.';
      } else if (qLower.includes('fraud') || qLower.includes('risk')) {
        answer = 'ReviveOS uses a dual-model ML engine (Random Forest for fraud anomaly detection and Gradient Boosting for return risk) to score expected loss before retries are triggered.';
      } else if (qLower.includes('webhook') || qLower.includes('signature')) {
        answer = 'ReviveOS verifies incoming Razorpay webhooks using HMAC-SHA256 with constant-time buffer comparison (`crypto.timingSafeEqual`) to prevent timing and replay attacks.';
      } else {
        answer = `ReviveOS connects to Razorpay via webhook ingestion. You can run \`npx reviveos init\` in your project root to generate the appropriate route handler and start recovering failed transactions.`;
      }

      setAiMessages((prev) => [...prev, { role: 'assistant', text: answer }]);
      setAiLoading(false);
    }, 600);
  };

  return (
    <div className="page-container" style={{ paddingBottom: '60px' }}>
      {/* Top Banner / Search bar matching platform styling */}
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          flexWrap: 'wrap',
          gap: '16px',
          marginBottom: '28px',
          paddingBottom: '20px',
          borderBottom: '1px solid var(--border-subtle)',
        }}
      >
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
            <h1 style={{ fontSize: '24px', fontWeight: 800, color: 'var(--text-primary)', margin: 0 }}>
              Documentation & Guides
            </h1>
            <span
              style={{
                background: 'var(--color-accent-bg)',
                border: '1px solid var(--color-accent-border)',
                color: 'var(--color-accent-light)',
                fontSize: '11px',
                fontWeight: 700,
                padding: '2px 8px',
                borderRadius: 'var(--radius-full)',
              }}
            >
              v1.0 · Stable
            </span>
          </div>
          <p style={{ fontSize: '13.5px', color: 'var(--text-secondary)', margin: '4px 0 0' }}>
            Integration guides, SDK setup, ML fraud intelligence, and autonomous recovery architecture
          </p>
        </div>

        {/* Search & AI action buttons */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
          <button
            onClick={() => setIsSearchOpen(true)}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '10px',
              background: 'var(--bg-surface)',
              border: '1px solid var(--border-subtle)',
              borderRadius: 'var(--radius-full)',
              padding: '8px 16px',
              color: 'var(--text-muted)',
              fontSize: '13px',
              cursor: 'pointer',
              transition: 'all 0.2s ease',
            }}
          >
            <Search size={14} color="var(--text-muted)" />
            <span>Search docs...</span>
            <kbd
              style={{
                background: 'var(--bg-elevated)',
                border: '1px solid var(--border-default)',
                color: 'var(--text-secondary)',
                fontSize: '10px',
                padding: '1px 5px',
                borderRadius: '4px',
                fontFamily: 'var(--font-mono)',
              }}
            >
              ⌘K
            </kbd>
          </button>

          <button
            onClick={() => setIsAiOpen(true)}
            className="btn-primary"
            style={{ padding: '8px 16px', fontSize: '13px' }}
          >
            <Sparkles size={14} />
            <span>Ask AI Assistant</span>
            <kbd
              style={{
                background: 'rgba(255, 255, 255, 0.2)',
                color: '#ffffff',
                fontSize: '10px',
                padding: '1px 5px',
                borderRadius: '4px',
                fontFamily: 'var(--font-mono)',
              }}
            >
              ⌘I
            </kbd>
          </button>
        </div>
      </div>

      {/* 2-Column Documentation Layout */}
      <div style={{ display: 'grid', gridTemplateColumns: '260px 1fr', gap: '32px', alignItems: 'start' }}>
        {/* Left Topics Nav */}
        <div
          style={{
            background: 'var(--bg-surface)',
            border: '1px solid var(--border-default)',
            borderRadius: 'var(--radius-lg)',
            padding: '20px 14px',
            position: 'sticky',
            top: '80px',
            maxHeight: 'calc(100vh - 120px)',
            overflowY: 'auto',
          }}
        >
          {Object.entries(navCategories).map(([category, articles]) => (
            <div key={category} style={{ marginBottom: '20px' }}>
              <div
                style={{
                  fontSize: '11px',
                  fontWeight: 700,
                  letterSpacing: '0.06em',
                  color: 'var(--text-muted)',
                  marginBottom: '6px',
                  paddingLeft: '10px',
                }}
              >
                {category}
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '3px' }}>
                {articles.map((art) => {
                  const isActive = activeArticleId === art.id;
                  return (
                    <button
                      key={art.id}
                      onClick={() => setActiveArticleId(art.id)}
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'space-between',
                        padding: '8px 12px',
                        borderRadius: 'var(--radius-md)',
                        background: isActive ? 'var(--color-accent-bg)' : 'transparent',
                        border: isActive ? '1px solid var(--color-accent-border)' : '1px solid transparent',
                        color: isActive ? 'var(--color-accent-light)' : 'var(--text-secondary)',
                        fontSize: '13px',
                        fontWeight: isActive ? 700 : 500,
                        cursor: 'pointer',
                        textAlign: 'left',
                        transition: 'all 0.15s ease',
                      }}
                    >
                      <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                        {art.title}
                      </span>
                      {art.badge && (
                        <span
                          style={{
                            fontSize: '9.5px',
                            fontWeight: 700,
                            padding: '1px 5px',
                            borderRadius: '4px',
                            background: isActive ? 'var(--color-accent)' : 'var(--bg-elevated)',
                            color: isActive ? '#ffffff' : 'var(--text-muted)',
                          }}
                        >
                          {art.badge}
                        </span>
                      )}
                    </button>
                  );
                })}
              </div>
            </div>
          ))}
        </div>

        {/* Right Article Content */}
        <div
          style={{
            background: 'var(--bg-surface)',
            border: '1px solid var(--border-default)',
            borderRadius: 'var(--radius-lg)',
            padding: '36px 40px',
            boxShadow: 'var(--card-shadow)',
          }}
        >
          {/* Breadcrumb */}
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
              fontSize: '12px',
              color: 'var(--text-muted)',
              marginBottom: '16px',
            }}
          >
            <span>Docs</span>
            <ChevronRight size={12} />
            <span>{currentArticle.category}</span>
            <ChevronRight size={12} />
            <span style={{ color: 'var(--color-accent-light)', fontWeight: 600 }}>{currentArticle.title}</span>
          </div>

          {/* Heading */}
          <h1
            style={{
              fontSize: '32px',
              fontWeight: 800,
              letterSpacing: '-0.8px',
              color: 'var(--text-primary)',
              margin: '0 0 10px',
            }}
          >
            {currentArticle.content.heading}
          </h1>
          <p style={{ fontSize: '15px', color: 'var(--text-secondary)', lineHeight: 1.6, margin: '0 0 28px' }}>
            {currentArticle.content.subheading}
          </p>

          {/* If Overview Page: Show Popular Cards Grid */}
          {activeArticleId === 'overview' && (
            <div style={{ marginBottom: '36px' }}>
              <div style={{ fontSize: '13px', fontWeight: 700, color: 'var(--text-primary)', marginBottom: '14px' }}>
                POPULAR INTEGRATIONS
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: '14px', marginBottom: '28px' }}>
                {[
                  { id: 'quickstart', title: '60s Quickstart', desc: 'Run your first autonomous recovery workflow in minutes.', icon: Zap, color: '#8b5cf6' },
                  { id: 'cli-config', title: 'One-Command Init', desc: 'Configure SDK, env, and webhook routes with npx reviveos init.', icon: Terminal, color: '#10b981' },
                  { id: 'ml-fraud', title: 'ML Fraud Engine', desc: 'Predict fraud risk and expected loss before triggering retries.', icon: ShieldCheck, color: '#f59e0b' },
                ].map((card) => {
                  const Icon = card.icon;
                  return (
                    <div
                      key={card.id}
                      onClick={() => setActiveArticleId(card.id)}
                      style={{
                        background: 'var(--bg-card)',
                        border: '1px solid var(--border-subtle)',
                        borderRadius: 'var(--radius-md)',
                        padding: '18px',
                        cursor: 'pointer',
                        transition: 'all 0.2s ease',
                      }}
                      onMouseEnter={(e) => {
                        e.currentTarget.style.borderColor = 'var(--color-accent)';
                        e.currentTarget.style.transform = 'translateY(-2px)';
                      }}
                      onMouseLeave={(e) => {
                        e.currentTarget.style.borderColor = 'var(--border-subtle)';
                        e.currentTarget.style.transform = 'none';
                      }}
                    >
                      <div
                        style={{
                          width: '34px',
                          height: '34px',
                          borderRadius: '8px',
                          background: 'var(--color-accent-bg)',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          marginBottom: '12px',
                        }}
                      >
                        <Icon size={17} color={card.color} />
                      </div>
                      <h4 style={{ margin: '0 0 4px', fontSize: '14px', fontWeight: 700, color: 'var(--text-primary)' }}>
                        {card.title}
                      </h4>
                      <p style={{ margin: 0, fontSize: '12px', color: 'var(--text-muted)', lineHeight: 1.5 }}>
                        {card.desc}
                      </p>
                    </div>
                  );
                })}
              </div>
            </div>
          )}

          {/* Article Content Sections */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: '28px' }}>
            {currentArticle.content.sections.map((sec, idx) => (
              <div key={idx}>
                <h3 style={{ fontSize: '18px', fontWeight: 700, color: 'var(--text-primary)', margin: '0 0 10px' }}>
                  {sec.title}
                </h3>
                <p
                  style={{
                    fontSize: '14px',
                    color: 'var(--text-secondary)',
                    lineHeight: 1.7,
                    margin: '0 0 14px',
                    whiteSpace: 'pre-line',
                  }}
                >
                  {sec.body}
                </p>

                {/* Code Block */}
                {sec.code && (
                  <div
                    style={{
                      background: 'var(--bg-elevated)',
                      border: '1px solid var(--border-default)',
                      borderRadius: 'var(--radius-md)',
                      overflow: 'hidden',
                      marginBottom: '16px',
                    }}
                  >
                    <div
                      style={{
                        display: 'flex',
                        justifyContent: 'space-between',
                        alignItems: 'center',
                        background: 'var(--bg-surface)',
                        padding: '8px 14px',
                        borderBottom: '1px solid var(--border-subtle)',
                      }}
                    >
                      <span style={{ fontSize: '11px', color: 'var(--text-muted)', fontFamily: 'var(--font-mono)' }}>
                        {sec.language || 'code'}
                      </span>
                      <button
                        onClick={() => copyToClipboard(sec.code!, `code-${idx}`)}
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          gap: '5px',
                          background: 'none',
                          border: 'none',
                          color: 'var(--text-muted)',
                          fontSize: '11px',
                          cursor: 'pointer',
                        }}
                      >
                        {copiedId === `code-${idx}` ? <Check size={13} color="var(--color-emerald)" /> : <Copy size={13} />}
                        <span>{copiedId === `code-${idx}` ? 'Copied' : 'Copy'}</span>
                      </button>
                    </div>
                    <pre
                      style={{
                        margin: 0,
                        padding: '16px',
                        fontSize: '12.5px',
                        fontFamily: 'var(--font-mono)',
                        color: 'var(--text-primary)',
                        overflowX: 'auto',
                        lineHeight: 1.6,
                      }}
                    >
                      {sec.code}
                    </pre>
                  </div>
                )}

                {/* Alert Note */}
                {sec.note && (
                  <div
                    style={{
                      padding: '12px 16px',
                      borderRadius: 'var(--radius-md)',
                      background:
                        sec.note.type === 'warning'
                          ? 'var(--color-amber-bg)'
                          : sec.note.type === 'important'
                          ? 'var(--color-red-bg)'
                          : 'var(--color-accent-bg)',
                      border: `1px solid ${
                        sec.note.type === 'warning'
                          ? 'var(--color-amber-border)'
                          : sec.note.type === 'important'
                          ? 'var(--color-red-border)'
                          : 'var(--color-accent-border)'
                      }`,
                      color:
                        sec.note.type === 'warning'
                          ? 'var(--color-amber)'
                          : sec.note.type === 'important'
                          ? 'var(--color-red)'
                          : 'var(--color-accent-light)',
                      fontSize: '12.5px',
                      lineHeight: 1.5,
                      marginBottom: '16px',
                    }}
                  >
                    <strong>{sec.note.type.toUpperCase()}:</strong> {sec.note.text}
                  </div>
                )}

                {/* Table */}
                {sec.table && (
                  <div className="table-container" style={{ margin: '14px 0 16px' }}>
                    <table className="data-table">
                      <thead>
                        <tr>
                          {sec.table.headers.map((h, i) => (
                            <th key={i}>{h}</th>
                          ))}
                        </tr>
                      </thead>
                      <tbody>
                        {sec.table.rows.map((r, i) => (
                          <tr key={i}>
                            {r.map((c, j) => (
                              <td key={j} style={{ fontFamily: j === 0 || j === 1 ? 'var(--font-mono)' : 'inherit' }}>
                                {c}
                              </td>
                            ))}
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </div>
            ))}
          </div>

          {/* Bottom Next/Prev Pagination */}
          <div
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              borderTop: '1px solid var(--border-subtle)',
              marginTop: '40px',
              paddingTop: '20px',
            }}
          >
            <button
              onClick={() => setActiveArticleId('overview')}
              className="btn-ghost"
              style={{ fontSize: '13px' }}
            >
              <ArrowLeft size={14} /> Back to Overview
            </button>
            <Link href="/developer" className="btn-secondary" style={{ fontSize: '13px' }}>
              <span>Open Developer Simulator</span>
              <ArrowRight size={14} />
            </Link>
          </div>
        </div>
      </div>

      {/* INTERACTIVE SEARCH MODAL (⌘K) */}
      {isSearchOpen && (
        <div
          style={{
            position: 'fixed',
            inset: 0,
            zIndex: 100,
            background: 'rgba(0, 0, 0, 0.65)',
            backdropFilter: 'blur(10px)',
            display: 'flex',
            alignItems: 'flex-start',
            justifyContent: 'center',
            paddingTop: '100px',
          }}
          onClick={() => setIsSearchOpen(false)}
        >
          <div
            style={{
              background: 'var(--bg-surface)',
              border: '1px solid var(--border-default)',
              borderRadius: 'var(--radius-lg)',
              width: '100%',
              maxWidth: '540px',
              boxShadow: '0 25px 50px rgba(0,0,0,0.5)',
              overflow: 'hidden',
            }}
            onClick={(e) => e.stopPropagation()}
          >
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                padding: '14px 18px',
                borderBottom: '1px solid var(--border-subtle)',
                gap: '12px',
              }}
            >
              <Search size={17} color="var(--text-muted)" />
              <input
                type="text"
                placeholder="Search documentation, guides, webhook handlers..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                autoFocus
                style={{
                  flex: 1,
                  background: 'none',
                  border: 'none',
                  color: 'var(--text-primary)',
                  fontSize: '14px',
                  outline: 'none',
                }}
              />
              <kbd
                style={{
                  background: 'var(--bg-elevated)',
                  border: '1px solid var(--border-subtle)',
                  color: 'var(--text-muted)',
                  fontSize: '11px',
                  padding: '2px 6px',
                  borderRadius: '4px',
                  fontFamily: 'var(--font-mono)',
                }}
              >
                ESC
              </kbd>
            </div>

            <div style={{ maxHeight: '320px', overflowY: 'auto', padding: '8px' }}>
              {searchResults.length > 0 ? (
                searchResults.map((res) => (
                  <div
                    key={res.id}
                    onClick={() => {
                      setActiveArticleId(res.id);
                      setIsSearchOpen(false);
                    }}
                    style={{
                      padding: '10px 14px',
                      borderRadius: 'var(--radius-md)',
                      cursor: 'pointer',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'space-between',
                      transition: 'background 0.15s ease',
                    }}
                    onMouseEnter={(e) => (e.currentTarget.style.background = 'var(--bg-card-hover)')}
                    onMouseLeave={(e) => (e.currentTarget.style.background = 'transparent')}
                  >
                    <div>
                      <div style={{ fontSize: '13.5px', fontWeight: 600, color: 'var(--text-primary)' }}>
                        {res.title}
                      </div>
                      <div style={{ fontSize: '11.5px', color: 'var(--text-muted)' }}>{res.description}</div>
                    </div>
                    <span
                      style={{
                        fontSize: '10px',
                        color: 'var(--color-accent-light)',
                        background: 'var(--color-accent-bg)',
                        padding: '2px 6px',
                        borderRadius: '4px',
                      }}
                    >
                      {res.category}
                    </span>
                  </div>
                ))
              ) : searchQuery ? (
                <div style={{ padding: '24px', textAlign: 'center', color: 'var(--text-muted)', fontSize: '13px' }}>
                  No results found for "{searchQuery}"
                </div>
              ) : (
                <div style={{ padding: '16px', color: 'var(--text-muted)', fontSize: '12px' }}>
                  Suggested: 60s Quickstart, Webhook Verification, ML Fraud Guard, Customer Memory
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* INTERACTIVE ASK AI ASSISTANT MODAL (⌘I) */}
      {isAiOpen && (
        <div
          style={{
            position: 'fixed',
            inset: 0,
            zIndex: 100,
            background: 'rgba(0, 0, 0, 0.65)',
            backdropFilter: 'blur(10px)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            padding: '20px',
          }}
          onClick={() => setIsAiOpen(false)}
        >
          <div
            style={{
              background: 'var(--bg-surface)',
              border: '1px solid var(--border-default)',
              borderRadius: 'var(--radius-lg)',
              width: '100%',
              maxWidth: '600px',
              height: '500px',
              display: 'flex',
              flexDirection: 'column',
              boxShadow: '0 25px 50px rgba(0,0,0,0.5)',
              overflow: 'hidden',
            }}
            onClick={(e) => e.stopPropagation()}
          >
            {/* AI Header */}
            <div
              style={{
                padding: '14px 18px',
                borderBottom: '1px solid var(--border-subtle)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
              }}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                <Sparkles size={16} color="var(--color-accent-light)" />
                <span style={{ fontSize: '14px', fontWeight: 700, color: 'var(--text-primary)' }}>
                  ReviveOS AI Doc Assistant
                </span>
              </div>
              <button
                onClick={() => setIsAiOpen(false)}
                style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}
              >
                <X size={16} />
              </button>
            </div>

            {/* AI Messages Body */}
            <div style={{ flex: 1, padding: '18px', overflowY: 'auto', display: 'flex', flexDirection: 'column', gap: '12px' }}>
              {aiMessages.map((msg, idx) => (
                <div
                  key={idx}
                  style={{
                    alignSelf: msg.role === 'user' ? 'flex-end' : 'flex-start',
                    maxWidth: '85%',
                    padding: '10px 14px',
                    borderRadius: 'var(--radius-md)',
                    background: msg.role === 'user' ? 'linear-gradient(135deg, #8b5cf6, #ec4899)' : 'var(--bg-elevated)',
                    color: msg.role === 'user' ? '#ffffff' : 'var(--text-primary)',
                    fontSize: '13px',
                    lineHeight: 1.5,
                  }}
                >
                  {msg.text}
                </div>
              ))}
              {aiLoading && (
                <div
                  style={{
                    alignSelf: 'flex-start',
                    background: 'var(--bg-elevated)',
                    padding: '8px 12px',
                    borderRadius: 'var(--radius-md)',
                    fontSize: '12px',
                    color: 'var(--text-muted)',
                  }}
                >
                  AI is analyzing documentation...
                </div>
              )}
            </div>

            {/* AI Input Form */}
            <form
              onSubmit={handleAskAi}
              style={{
                padding: '12px 14px',
                borderTop: '1px solid var(--border-subtle)',
                display: 'flex',
                gap: '8px',
              }}
            >
              <input
                type="text"
                placeholder="Ask about SDK setup, Razorpay webhook verification, ML fraud..."
                value={aiQuestion}
                onChange={(e) => setAiQuestion(e.target.value)}
                autoFocus
                style={{
                  flex: 1,
                  background: 'var(--bg-input)',
                  border: '1px solid var(--border-subtle)',
                  borderRadius: 'var(--radius-md)',
                  padding: '9px 12px',
                  color: 'var(--text-primary)',
                  fontSize: '13px',
                  outline: 'none',
                }}
              />
              <button
                type="submit"
                disabled={aiLoading}
                className="btn-primary"
                style={{ padding: '8px 14px', fontSize: '13px' }}
              >
                <Send size={13} />
                <span>Ask</span>
              </button>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

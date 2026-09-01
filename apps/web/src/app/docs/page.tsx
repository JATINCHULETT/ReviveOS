'use client';

import React, { useState, useEffect, useMemo } from 'react';
import Link from 'next/link';
import {
  Search,
  Sparkles,
  ExternalLink,
  Terminal,
  Code2,
  Copy,
  Check,
  ChevronRight,
  ChevronDown,
  BookOpen,
  Zap,
  ShieldCheck,
  Cpu,
  Key,
  Layers,
  ArrowRight,
  ArrowLeft,
  FileCode,
  CheckCircle2,
  AlertTriangle,
  RefreshCw,
  Server,
  Lock,
  Boxes,
  HelpCircle,
  X,
  MessageSquare,
  Send,
} from 'lucide-react';
import ReviveLogo from '@/components/ui/ReviveLogo';
import { ThemeToggle } from '@/components/ui/ThemeToggle';

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

  // Documentation Articles Catalog
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
            code: `# Your ReviveOS Merchant API Key (obtained from ReviveOS Merchant Console)
REVIVEOS_API_KEY=rvo_test_acme_secret_key_12345

# ReviveOS Core API URL
REVIVEOS_API_URL=http://localhost:8080

# Razorpay Webhook Secret for HMAC-SHA256 verification
RAZORPAY_WEBHOOK_SECRET=whsec_your_razorpay_webhook_secret

# Mode: 'live', 'test', or 'mock' (for local sandbox testing)
REVIVEOS_MODE=live`,
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
            code: `import { NextResponse } from "next/server";
import { ReviveOS } from "@reviveos/razorpay";

const revive = new ReviveOS({
  apiKey: process.env.REVIVEOS_API_KEY,
  webhookSecret: process.env.RAZORPAY_WEBHOOK_SECRET,
});

export async function POST(req: Request) {
  const rawBody = await req.text();
  const signature = req.headers.get("x-razorpay-signature") || "";

  // 1. Verify cryptographic HMAC signature & normalize event
  const event = revive.webhooks.verifyAndNormalize(rawBody, signature);

  // 2. Ingest into ReviveOS AI recovery pipeline
  const result = await revive.events.process(event);

  return NextResponse.json(result);
}`,
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
            code: `import { RazorpayWebhookManager } from "@reviveos/razorpay";

const webhook = new RazorpayWebhookManager(process.env.RAZORPAY_WEBHOOK_SECRET);

// Throws ReviveOSSignatureError if invalid
const event = webhook.verifyAndNormalize(rawBody, signatureHeader);`,
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
            code: `const diagnosis = await revive.payments.analyze({
  paymentId: "pay_test_8921",
  amount: 85000, // High-value transaction
  currency: "INR",
  failureCode: "GATEWAY_TIMEOUT",
  customerId: "cust_99",
  paymentMethod: "card"
});

console.log(diagnosis.fraudRisk);
// Output:
// {
//   fraudProbability: 0.78,
//   riskLevel: "HIGH",
//   expectedLoss: 66300.0,
//   overallRisk: "HIGH"
// }
console.log(diagnosis.decision); // "BLOCK"`,
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
            code: `const profile = await revive.customers.getRecoveryProfile("cust_alex_123");

console.log(profile);
// Output:
// {
//   customerId: "cust_alex_123",
//   previousFailures: 4,
//   previousSuccessfulRecoveries: 3,
//   preferredPaymentMethods: ["card", "upi"],
//   averageRecoveryTime: 180, // in seconds
//   recoveryProbability: 0.81,
//   communicationOptOut: false
// }`,
            language: 'typescript',
          },
        ],
      },
    },
    'api-reference': {
      id: 'api-reference',
      title: 'REST API v1 Reference',
      category: 'SDK & CLI REFERENCE',
      description: 'Complete specification of ReviveOS REST API v1 endpoints and parameters.',
      readTime: '5 min read',
      badge: 'v1 Spec',
      content: {
        heading: 'REST API v1 Reference',
        subheading: 'Direct HTTP specification for all ReviveOS developer endpoints.',
        sections: [
          {
            title: 'Authentication',
            body: 'Authenticate all API requests by providing your API key in the `Authorization` header:',
            code: 'Authorization: Bearer rvo_test_your_secret_key\n# or\nX-API-Key: rvo_test_your_secret_key',
            language: 'http',
          },
          {
            title: 'Endpoints Overview',
            body: 'ReviveOS exposes the following core v1 endpoints:',
            table: {
              headers: ['Method', 'Endpoint', 'Description'],
              rows: [
                ['POST', '/v1/events', 'Ingest normalized payment events from webhooks or SDK'],
                ['POST', '/v1/payments/analyze', 'Synchronous AI failure classification & recovery prediction'],
                ['POST', '/v1/recovery/decision', 'Deterministic policy decision evaluation'],
                ['POST', '/v1/recovery/execute', 'Execute payment recovery retry or link generation'],
                ['GET', '/v1/payments/:id', 'Retrieve payment lifecycle and recovery actions'],
                ['GET', '/v1/customers/:id/recovery-profile', 'Retrieve customer recovery memory profile'],
                ['GET', '/v1/keys', 'List merchant API keys'],
                ['POST', '/v1/keys', 'Generate new API key (SHA-256 hashed)'],
              ],
            },
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
    <div style={{ backgroundColor: '#090a0f', color: '#e6edf3', minHeight: '100vh', display: 'flex', flexDirection: 'column', fontFamily: 'system-ui, -apple-system, sans-serif' }}>
      {/* 1. TOP NAVBAR */}
      <header
        style={{
          position: 'sticky',
          top: 0,
          zIndex: 50,
          height: '60px',
          borderBottom: '1px solid #1f242c',
          backgroundColor: 'rgba(9, 10, 15, 0.85)',
          backdropFilter: 'blur(16px)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '0 24px',
        }}
      >
        {/* Left: Brand & Docs Badge */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
          <ReviveLogo size="sm" href="/docs" />
          <span
            style={{
              background: '#161b22',
              border: '1px solid #30363d',
              color: '#58a6ff',
              fontSize: '12px',
              fontWeight: 700,
              padding: '2px 8px',
              borderRadius: '6px',
            }}
          >
            Docs
          </span>
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px', fontSize: '12px', color: '#8b949e', borderLeft: '1px solid #21262d', paddingLeft: '14px' }}>
            <span style={{ width: '6px', height: '6px', borderRadius: '50%', background: '#3fb950', display: 'inline-block' }} />
            <span>v1 · stable</span>
            <span style={{ color: '#484f58' }}>|</span>
            <span>2026.09</span>
          </div>
        </div>

        {/* Center: Search & Ask AI Trigger */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '10px', maxWidth: '480px', width: '100%' }}>
          <button
            onClick={() => setIsSearchOpen(true)}
            style={{
              flex: 1,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              background: '#161b22',
              border: '1px solid #30363d',
              borderRadius: '8px',
              padding: '7px 12px',
              color: '#8b949e',
              fontSize: '13px',
              cursor: 'pointer',
              transition: 'border-color 0.15s ease',
            }}
          >
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
              <Search size={14} color="#8b949e" />
              <span>Search documentation...</span>
            </div>
            <kbd style={{ background: '#21262d', border: '1px solid #30363d', color: '#c9d1d9', fontSize: '11px', padding: '2px 6px', borderRadius: '4px' }}>
              ⌘K
            </kbd>
          </button>

          <button
            onClick={() => setIsAiOpen(true)}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '6px',
              background: 'linear-gradient(135deg, rgba(56, 189, 248, 0.15), rgba(139, 92, 246, 0.15))',
              border: '1px solid rgba(56, 189, 248, 0.3)',
              borderRadius: '8px',
              padding: '7px 14px',
              color: '#38bdf8',
              fontSize: '13px',
              fontWeight: 600,
              cursor: 'pointer',
              whiteSpace: 'nowrap',
            }}
          >
            <Sparkles size={14} />
            <span>Ask AI</span>
            <kbd style={{ background: 'rgba(56, 189, 248, 0.2)', color: '#7dd3fc', fontSize: '10px', padding: '1px 5px', borderRadius: '4px' }}>
              ⌘I
            </kbd>
          </button>
        </div>

        {/* Right: Actions */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
          <ThemeToggle />
          <Link href="/merchant">
            <button
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '6px',
                background: '#f0f6fc',
                color: '#0d1117',
                border: 'none',
                borderRadius: '8px',
                padding: '7px 14px',
                fontSize: '13px',
                fontWeight: 700,
                cursor: 'pointer',
              }}
            >
              <span>Open dashboard</span>
              <ExternalLink size={13} />
            </button>
          </Link>
        </div>
      </header>

      {/* 2. BODY LAYOUT (Sidebar + Main Article Area) */}
      <div style={{ display: 'flex', flex: 1, maxWidth: '1440px', margin: '0 auto', width: '100%' }}>
        {/* Left Navigation Sidebar */}
        <aside
          style={{
            width: '280px',
            borderRight: '1px solid #1f242c',
            padding: '24px 16px',
            height: 'calc(100vh - 60px)',
            position: 'sticky',
            top: '60px',
            overflowY: 'auto',
            flexShrink: 0,
          }}
        >
          {Object.entries(navCategories).map(([category, articles]) => (
            <div key={category} style={{ marginBottom: '24px' }}>
              <div style={{ fontSize: '11px', fontWeight: 700, letterSpacing: '0.06em', color: '#8b949e', marginBottom: '8px', paddingLeft: '8px' }}>
                {category}
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '2px' }}>
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
                        borderRadius: '6px',
                        background: isActive ? '#1f293d' : 'transparent',
                        border: 'none',
                        color: isActive ? '#58a6ff' : '#c9d1d9',
                        fontSize: '13px',
                        fontWeight: isActive ? 600 : 400,
                        cursor: 'pointer',
                        textAlign: 'left',
                        transition: 'background 0.15s ease',
                      }}
                    >
                      <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                        {art.title}
                      </span>
                      {art.badge && (
                        <span
                          style={{
                            fontSize: '10px',
                            fontWeight: 700,
                            padding: '2px 6px',
                            borderRadius: '4px',
                            background: isActive ? 'rgba(88, 166, 255, 0.2)' : '#21262d',
                            color: isActive ? '#58a6ff' : '#8b949e',
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
        </aside>

        {/* Center: Main Documentation Article */}
        <main style={{ flex: 1, padding: '40px 48px', maxWidth: '880px', margin: '0 auto' }}>
          {/* Breadcrumb */}
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px', fontSize: '12px', color: '#8b949e', marginBottom: '20px' }}>
            <span>Docs</span>
            <ChevronRight size={12} />
            <span>{currentArticle.category}</span>
            <ChevronRight size={12} />
            <span style={{ color: '#58a6ff', fontWeight: 600 }}>{currentArticle.title}</span>
          </div>

          {/* Heading */}
          <h1 style={{ fontSize: '38px', fontWeight: 800, letterSpacing: '-1px', color: '#f0f6fc', margin: '0 0 12px' }}>
            {currentArticle.content.heading}
          </h1>
          <p style={{ fontSize: '16px', color: '#8b949e', lineHeight: 1.6, margin: '0 0 32px' }}>
            {currentArticle.content.subheading}
          </p>

          {/* If Overview Page: Show Popular Cards Grid (Matching reference screenshot) */}
          {activeArticleId === 'overview' && (
            <div>
              {/* Popular Section */}
              <div style={{ marginBottom: '40px' }}>
                <div style={{ fontSize: '14px', fontWeight: 700, color: '#f0f6fc', marginBottom: '16px' }}>Popular</div>
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: '16px' }}>
                  {[
                    { id: 'quickstart', title: '60s Quickstart', desc: 'Run your first autonomous recovery workflow in minutes.', icon: Zap, color: '#58a6ff' },
                    { id: 'cli-config', title: 'One-Command Init', desc: 'Configure SDK, env, and webhook routes with npx reviveos init.', icon: Terminal, color: '#3fb950' },
                    { id: 'ml-fraud', title: 'ML Fraud Engine', desc: 'Predict fraud risk and expected loss before triggering retries.', icon: ShieldCheck, color: '#d29922' },
                  ].map((card) => {
                    const Icon = card.icon;
                    return (
                      <div
                        key={card.id}
                        onClick={() => setActiveArticleId(card.id)}
                        style={{
                          background: '#161b22',
                          border: '1px solid #30363d',
                          borderRadius: '12px',
                          padding: '20px',
                          cursor: 'pointer',
                          transition: 'all 0.2s ease',
                        }}
                        onMouseEnter={(e) => {
                          e.currentTarget.style.borderColor = '#58a6ff';
                          e.currentTarget.style.transform = 'translateY(-2px)';
                        }}
                        onMouseLeave={(e) => {
                          e.currentTarget.style.borderColor = '#30363d';
                          e.currentTarget.style.transform = 'none';
                        }}
                      >
                        <div style={{ width: '36px', height: '36px', borderRadius: '8px', background: 'rgba(88, 166, 255, 0.1)', display: 'flex', alignItems: 'center', justifyContent: 'center', marginBottom: '14px' }}>
                          <Icon size={18} color={card.color} />
                        </div>
                        <h4 style={{ margin: '0 0 6px', fontSize: '15px', fontWeight: 700, color: '#f0f6fc' }}>{card.title}</h4>
                        <p style={{ margin: 0, fontSize: '12.5px', color: '#8b949e', lineHeight: 1.5 }}>{card.desc}</p>
                      </div>
                    );
                  })}
                </div>
              </div>

              {/* ReviveOS Basics Section */}
              <div style={{ marginBottom: '40px' }}>
                <div style={{ fontSize: '14px', fontWeight: 700, color: '#f0f6fc', marginBottom: '16px' }}>ReviveOS basics</div>
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '14px' }}>
                  {[
                    { id: 'installation', title: 'Installation', desc: 'npm install @reviveos/razorpay' },
                    { id: 'cli-config', title: 'API Keys', desc: 'SHA-256 hashed merchant credentials' },
                    { id: 'webhooks', title: 'Webhooks', desc: 'HMAC-SHA256 signature verification' },
                    { id: 'customer-memory', title: 'Customer Memory', desc: 'Historical behavior calibration' },
                  ].map((basic) => (
                    <div
                      key={basic.id}
                      onClick={() => setActiveArticleId(basic.id)}
                      style={{
                        background: '#0d1117',
                        border: '1px solid #21262d',
                        borderRadius: '10px',
                        padding: '16px',
                        cursor: 'pointer',
                        transition: 'border-color 0.15s ease',
                      }}
                      onMouseEnter={(e) => (e.currentTarget.style.borderColor = '#58a6ff')}
                      onMouseLeave={(e) => (e.currentTarget.style.borderColor = '#21262d')}
                    >
                      <div style={{ fontSize: '14px', fontWeight: 700, color: '#f0f6fc', marginBottom: '4px' }}>{basic.title}</div>
                      <div style={{ fontSize: '12px', color: '#8b949e' }}>{basic.desc}</div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}

          {/* Article Content Sections */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: '32px' }}>
            {currentArticle.content.sections.map((sec, idx) => (
              <div key={idx}>
                <h3 style={{ fontSize: '20px', fontWeight: 700, color: '#f0f6fc', margin: '0 0 12px' }}>
                  {sec.title}
                </h3>
                <p style={{ fontSize: '14.5px', color: '#c9d1d9', lineHeight: 1.7, margin: '0 0 16px', whiteSpace: 'pre-line' }}>
                  {sec.body}
                </p>

                {/* Code Block */}
                {sec.code && (
                  <div
                    style={{
                      background: '#0d1117',
                      border: '1px solid #30363d',
                      borderRadius: '10px',
                      overflow: 'hidden',
                      marginBottom: '16px',
                    }}
                  >
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', background: '#161b22', padding: '8px 14px', borderBottom: '1px solid #21262d' }}>
                      <span style={{ fontSize: '11px', color: '#8b949e', fontFamily: 'monospace' }}>{sec.language || 'code'}</span>
                      <button
                        onClick={() => copyToClipboard(sec.code!, `code-${idx}`)}
                        style={{ display: 'flex', alignItems: 'center', gap: '6px', background: 'none', border: 'none', color: '#8b949e', fontSize: '11px', cursor: 'pointer' }}
                      >
                        {copiedId === `code-${idx}` ? <Check size={13} color="#3fb950" /> : <Copy size={13} />}
                        <span>{copiedId === `code-${idx}` ? 'Copied' : 'Copy'}</span>
                      </button>
                    </div>
                    <pre style={{ margin: 0, padding: '16px', fontSize: '13px', fontFamily: '"Fira Code", monospace', color: '#e6edf3', overflowX: 'auto', lineHeight: 1.55 }}>
                      {sec.code}
                    </pre>
                  </div>
                )}

                {/* Alert Note */}
                {sec.note && (
                  <div
                    style={{
                      padding: '14px 16px',
                      borderRadius: '8px',
                      background: sec.note.type === 'warning' ? 'rgba(210, 153, 34, 0.1)' : 'rgba(56, 189, 248, 0.1)',
                      border: `1px solid ${sec.note.type === 'warning' ? 'rgba(210, 153, 34, 0.3)' : 'rgba(56, 189, 248, 0.3)'}`,
                      color: sec.note.type === 'warning' ? '#e3b341' : '#7dd3fc',
                      fontSize: '13px',
                      lineHeight: 1.5,
                      marginBottom: '16px',
                    }}
                  >
                    <strong>{sec.note.type.toUpperCase()}:</strong> {sec.note.text}
                  </div>
                )}

                {/* Table */}
                {sec.table && (
                  <div style={{ overflowX: 'auto', marginBottom: '16px' }}>
                    <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '13px', textAlign: 'left' }}>
                      <thead>
                        <tr style={{ borderBottom: '1px solid #30363d', color: '#8b949e' }}>
                          {sec.table.headers.map((h, i) => (
                            <th key={i} style={{ padding: '10px 12px' }}>{h}</th>
                          ))}
                        </tr>
                      </thead>
                      <tbody>
                        {sec.table.rows.map((r, i) => (
                          <tr key={i} style={{ borderBottom: '1px solid #21262d' }}>
                            {r.map((c, j) => (
                              <td key={j} style={{ padding: '10px 12px', color: j === 0 ? '#58a6ff' : '#c9d1d9', fontFamily: j === 0 || j === 1 ? 'monospace' : 'inherit' }}>
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
          <div style={{ display: 'flex', justifyContent: 'space-between', borderTop: '1px solid #1f242c', marginTop: '48px', paddingTop: '24px' }}>
            <Link href="/docs">
              <span style={{ fontSize: '13px', color: '#8b949e', display: 'flex', alignItems: 'center', gap: '6px' }}>
                <ArrowLeft size={14} /> Back to Overview
              </span>
            </Link>
            <Link href="/developer">
              <span style={{ fontSize: '13px', color: '#58a6ff', fontWeight: 600, display: 'flex', alignItems: 'center', gap: '6px' }}>
                Open Developer Simulator <ArrowRight size={14} />
              </span>
            </Link>
          </div>
        </main>
      </div>

      {/* 3. INTERACTIVE SEARCH MODAL (⌘K) */}
      {isSearchOpen && (
        <div
          style={{
            position: 'fixed',
            inset: 0,
            zIndex: 100,
            background: 'rgba(0,0,0,0.7)',
            backdropFilter: 'blur(8px)',
            display: 'flex',
            alignItems: 'flex-start',
            justifyContent: 'center',
            paddingTop: '120px',
          }}
          onClick={() => setIsSearchOpen(false)}
        >
          <div
            style={{
              background: '#161b22',
              border: '1px solid #30363d',
              borderRadius: '12px',
              width: '100%',
              maxWidth: '560px',
              boxShadow: '0 20px 40px rgba(0,0,0,0.6)',
              overflow: 'hidden',
            }}
            onClick={(e) => e.stopPropagation()}
          >
            <div style={{ display: 'flex', alignItems: 'center', padding: '14px 16px', borderBottom: '1px solid #21262d', gap: '10px' }}>
              <Search size={18} color="#8b949e" />
              <input
                type="text"
                placeholder="Search documentation, SDK methods, endpoints..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                autoFocus
                style={{
                  flex: 1,
                  background: 'none',
                  border: 'none',
                  color: '#f0f6fc',
                  fontSize: '14px',
                  outline: 'none',
                }}
              />
              <kbd style={{ background: '#21262d', border: '1px solid #30363d', color: '#8b949e', fontSize: '11px', padding: '2px 6px', borderRadius: '4px' }}>
                ESC
              </kbd>
            </div>

            <div style={{ maxHeight: '340px', overflowY: 'auto', padding: '8px' }}>
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
                      borderRadius: '8px',
                      cursor: 'pointer',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'space-between',
                      transition: 'background 0.15s ease',
                    }}
                    onMouseEnter={(e) => (e.currentTarget.style.background = '#21262d')}
                    onMouseLeave={(e) => (e.currentTarget.style.background = 'transparent')}
                  >
                    <div>
                      <div style={{ fontSize: '13.5px', fontWeight: 600, color: '#f0f6fc' }}>{res.title}</div>
                      <div style={{ fontSize: '11.5px', color: '#8b949e' }}>{res.description}</div>
                    </div>
                    <span style={{ fontSize: '10.5px', color: '#58a6ff', background: 'rgba(88, 166, 255, 0.1)', padding: '2px 6px', borderRadius: '4px' }}>
                      {res.category}
                    </span>
                  </div>
                ))
              ) : searchQuery ? (
                <div style={{ padding: '24px', textAlign: 'center', color: '#8b949e', fontSize: '13px' }}>
                  No results found for "{searchQuery}"
                </div>
              ) : (
                <div style={{ padding: '16px', color: '#8b949e', fontSize: '12px' }}>
                  Quick links: 60s Quickstart, Webhook Verification, ML Fraud Guard, REST API
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* 4. INTERACTIVE ASK AI ASSISTANT MODAL (⌘I) */}
      {isAiOpen && (
        <div
          style={{
            position: 'fixed',
            inset: 0,
            zIndex: 100,
            background: 'rgba(0,0,0,0.7)',
            backdropFilter: 'blur(8px)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            padding: '20px',
          }}
          onClick={() => setIsAiOpen(false)}
        >
          <div
            style={{
              background: '#161b22',
              border: '1px solid #30363d',
              borderRadius: '16px',
              width: '100%',
              maxWidth: '620px',
              height: '520px',
              display: 'flex',
              flexDirection: 'column',
              boxShadow: '0 25px 50px rgba(0,0,0,0.7)',
              overflow: 'hidden',
            }}
            onClick={(e) => e.stopPropagation()}
          >
            {/* AI Header */}
            <div style={{ padding: '14px 20px', borderBottom: '1px solid #21262d', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                <Sparkles size={16} color="#38bdf8" />
                <span style={{ fontSize: '14px', fontWeight: 700, color: '#f0f6fc' }}>ReviveOS AI Doc Assistant</span>
              </div>
              <button
                onClick={() => setIsAiOpen(false)}
                style={{ background: 'none', border: 'none', color: '#8b949e', cursor: 'pointer' }}
              >
                <X size={16} />
              </button>
            </div>

            {/* AI Messages Body */}
            <div style={{ flex: 1, padding: '20px', overflowY: 'auto', display: 'flex', flexDirection: 'column', gap: '14px' }}>
              {aiMessages.map((msg, idx) => (
                <div
                  key={idx}
                  style={{
                    alignSelf: msg.role === 'user' ? 'flex-end' : 'flex-start',
                    maxWidth: '85%',
                    padding: '12px 16px',
                    borderRadius: '12px',
                    background: msg.role === 'user' ? '#1f6feb' : '#21262d',
                    color: '#f0f6fc',
                    fontSize: '13.5px',
                    lineHeight: 1.5,
                  }}
                >
                  {msg.text}
                </div>
              ))}
              {aiLoading && (
                <div style={{ alignSelf: 'flex-start', background: '#21262d', padding: '10px 14px', borderRadius: '12px', fontSize: '12px', color: '#8b949e' }}>
                  AI is thinking...
                </div>
              )}
            </div>

            {/* AI Input Form */}
            <form onSubmit={handleAskAi} style={{ padding: '14px 16px', borderTop: '1px solid #21262d', display: 'flex', gap: '10px' }}>
              <input
                type="text"
                placeholder="Ask about installation, Razorpay webhook verification, ML fraud..."
                value={aiQuestion}
                onChange={(e) => setAiQuestion(e.target.value)}
                autoFocus
                style={{
                  flex: 1,
                  background: '#0d1117',
                  border: '1px solid #30363d',
                  borderRadius: '8px',
                  padding: '10px 14px',
                  color: '#f0f6fc',
                  fontSize: '13px',
                  outline: 'none',
                }}
              />
              <button
                type="submit"
                disabled={aiLoading}
                style={{
                  background: '#38bdf8',
                  color: '#090a0f',
                  border: 'none',
                  borderRadius: '8px',
                  padding: '0 16px',
                  fontWeight: 700,
                  fontSize: '13px',
                  cursor: 'pointer',
                  display: 'flex',
                  alignItems: 'center',
                  gap: '6px',
                }}
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

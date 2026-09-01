'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import {
  Terminal,
  Copy,
  Check,
  Code2,
  Key,
  ShieldAlert,
  Zap,
  ArrowRight,
  Layers,
  CheckCircle2,
  Cpu,
  Sparkles,
  ExternalLink,
} from 'lucide-react';

export default function DeveloperDocsSection() {
  const [activeCodeTab, setActiveCodeTab] = useState<'app-router' | 'pages-router' | 'express' | 'sync-analyze'>('app-router');
  const [copiedSnippet, setCopiedSnippet] = useState<string | null>(null);

  const copyText = (text: string, id: string) => {
    navigator.clipboard.writeText(text);
    setCopiedSnippet(id);
    setTimeout(() => setCopiedSnippet(null), 2000);
  };

  const codeSnippets = {
    'app-router': `// app/api/reviveos/webhook/route.ts
import { NextResponse } from "next/server";
import { ReviveOS } from "@reviveos/razorpay";

const revive = new ReviveOS({
  apiKey: process.env.REVIVEOS_API_KEY,
  webhookSecret: process.env.RAZORPAY_WEBHOOK_SECRET,
});

export async function POST(req: Request) {
  const rawBody = await req.text();
  const signature = req.headers.get("x-razorpay-signature") || "";

  // 1. Cryptographically verify HMAC-SHA256 signature
  const event = revive.webhooks.verifyAndNormalize(rawBody, signature);

  // 2. Autonomous Ingestion: Classify -> Fraud Check -> Recovery Prediction
  const result = await revive.events.process(event);

  return NextResponse.json(result);
}`,
    'pages-router': `// pages/api/reviveos/webhook.ts
import type { NextApiRequest, NextApiResponse } from "next";
import { ReviveOS } from "@reviveos/razorpay";

// Disable Next.js body parser to preserve raw bytes for HMAC verification
export const config = { api: { bodyParser: false } };

async function getRawBody(req: NextApiRequest): Promise<string> {
  const chunks: Buffer[] = [];
  for await (const chunk of req) chunks.push(Buffer.from(chunk));
  return Buffer.concat(chunks).toString("utf8");
}

const revive = new ReviveOS();

export default async function handler(req: NextApiRequest, res: NextApiResponse) {
  if (req.method !== "POST") return res.status(405).end();
  
  const rawBody = await getRawBody(req);
  const signature = req.headers["x-razorpay-signature"] as string;

  const event = revive.webhooks.verifyAndNormalize(rawBody, signature);
  const result = await revive.events.process(event);
  
  return res.status(200).json(result);
}`,
    'express': `// src/routes/reviveos-webhook.ts
import express, { Request, Response } from "express";
import { ReviveOS } from "@reviveos/razorpay";

const app = express();
const revive = new ReviveOS({
  apiKey: process.env.REVIVEOS_API_KEY,
  webhookSecret: process.env.RAZORPAY_WEBHOOK_SECRET,
});

// Use express.raw() to capture unparsed bytes for cryptographic signature verification
app.post(
  "/api/reviveos/webhook",
  express.raw({ type: "application/json" }),
  async (req: Request, res: Response) => {
    try {
      const signature = req.headers["x-razorpay-signature"] as string;
      const event = revive.webhooks.verifyAndNormalize(req.body, signature);
      const result = await revive.events.process(event);
      return res.status(200).json(result);
    } catch (err: any) {
      return res.status(400).json({ error: err.message });
    }
  }
);`,
    'sync-analyze': `// Synchronous Payment Failure Diagnosis & Recovery Prediction
import { ReviveOS } from "@reviveos/razorpay";

const revive = new ReviveOS({ apiKey: process.env.REVIVEOS_API_KEY });

// Execute synchronous AI diagnosis & optimal timing analysis
const diagnosis = await revive.payments.analyze({
  paymentId: "pay_L9k3x82n",
  amount: 4999.00,
  currency: "INR",
  failureCode: "BAD_REQUEST_ERROR",
  failureReason: "Payment failed due to temporary issuer bank downtime",
  customerId: "cust_alex_99",
  paymentMethod: "card"
});

console.log(diagnosis.failureCategory);      // "BANK_UNAVAILABLE"
console.log(diagnosis.recoveryProbability);  // 0.94 (94.0%)
console.log(diagnosis.fraudRisk);            // { fraudProbability: 0.04, riskLevel: 'LOW' }
console.log(diagnosis.nextBestAction);       // "RETRY_LATER"
console.log(diagnosis.delaySeconds);         // 120 (Scheduled in 2 mins)
console.log(diagnosis.decision);             // "RECOVER"`,
  };

  return (
    <section
      id="docs"
      style={{
        position: 'relative',
        zIndex: 10,
        padding: '80px 24px',
        maxWidth: '1240px',
        margin: '0 auto',
      }}
    >
      {/* Section Header */}
      <div style={{ textAlign: 'center', marginBottom: '48px' }}>
        <div
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: '8px',
            padding: '6px 14px',
            borderRadius: '9999px',
            background: 'rgba(14, 165, 233, 0.1)',
            border: '1px solid rgba(14, 165, 233, 0.25)',
            marginBottom: '14px',
          }}
        >
          <Terminal size={14} color="var(--color-accent-blue)" />
          <span style={{ fontSize: '12px', fontWeight: 700, color: 'var(--color-accent-blue)', letterSpacing: '0.04em' }}>
            DEVELOPER PLATFORM & SDK GUIDE
          </span>
        </div>
        <h2
          style={{
            fontSize: 'clamp(28px, 4.5vw, 46px)',
            fontWeight: 900,
            letterSpacing: '-1.5px',
            color: 'var(--text-primary)',
            marginBottom: '14px',
          }}
        >
          Integrate ReviveOS in 60 Seconds
        </h2>
        <p
          style={{
            fontSize: '16px',
            color: 'var(--text-secondary)',
            maxWidth: '680px',
            margin: '0 auto',
            lineHeight: 1.6,
          }}
        >
          Add autonomous AI payment recovery and ML fraud protection to your existing Razorpay integration without rebuilding your payment flow.
        </p>
      </div>

      {/* 3 Step Onboarding Flow Cards */}
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))',
          gap: '20px',
          marginBottom: '36px',
        }}
      >
        {/* Step 1 */}
        <div
          style={{
            background: 'var(--bg-card)',
            border: '1px solid var(--border-subtle)',
            borderRadius: '16px',
            padding: '24px',
            display: 'flex',
            flexDirection: 'column',
            justifyContent: 'space-between',
          }}
        >
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '14px' }}>
              <div
                style={{
                  width: '32px',
                  height: '32px',
                  borderRadius: '8px',
                  background: 'var(--color-accent-blue)',
                  color: '#fff',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontWeight: 800,
                  fontSize: '14px',
                }}
              >
                1
              </div>
              <h3 style={{ margin: 0, fontSize: '16px', fontWeight: 700, color: 'var(--text-primary)' }}>
                Install & Auto-Configure
              </h3>
            </div>
            <p style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '16px', lineHeight: 1.5 }}>
              Install the lightweight SDK and run the smart project scanner to generate your environment template and webhook handler.
            </p>
          </div>

          <div
            style={{
              background: '#0d1117',
              borderRadius: '8px',
              padding: '12px 14px',
              fontFamily: 'monospace',
              fontSize: '12.5px',
              color: '#7ee787',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              gap: '8px',
            }}
          >
            <div>
              <div style={{ color: '#8b949e', fontSize: '11px' }}>$ npm install @reviveos/razorpay</div>
              <div style={{ color: '#79c0ff', marginTop: '4px' }}>$ npx reviveos init</div>
            </div>
            <button
              onClick={() => copyText('npm install @reviveos/razorpay && npx reviveos init', 'step1')}
              style={{ background: 'none', border: 'none', color: '#8b949e', cursor: 'pointer' }}
              title="Copy commands"
            >
              {copiedSnippet === 'step1' ? <Check size={14} color="#7ee787" /> : <Copy size={14} />}
            </button>
          </div>
        </div>

        {/* Step 2 */}
        <div
          style={{
            background: 'var(--bg-card)',
            border: '1px solid var(--border-subtle)',
            borderRadius: '16px',
            padding: '24px',
            display: 'flex',
            flexDirection: 'column',
            justifyContent: 'space-between',
          }}
        >
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '14px' }}>
              <div
                style={{
                  width: '32px',
                  height: '32px',
                  borderRadius: '8px',
                  background: '#8b5cf6',
                  color: '#fff',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontWeight: 800,
                  fontSize: '14px',
                }}
              >
                2
              </div>
              <h3 style={{ margin: 0, fontSize: '16px', fontWeight: 700, color: 'var(--text-primary)' }}>
                Generate Merchant API Key
              </h3>
            </div>
            <p style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '16px', lineHeight: 1.5 }}>
              Open your ReviveOS Merchant Console to create a secure, SHA-256 hashed API key (`rvo_test_...` or `rvo_live_...`).
            </p>
          </div>

          <div
            style={{
              background: '#0d1117',
              borderRadius: '8px',
              padding: '12px 14px',
              fontFamily: 'monospace',
              fontSize: '12.5px',
              color: '#e6edf3',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              gap: '8px',
            }}
          >
            <div style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
              <span style={{ color: '#8b949e' }}>REVIVEOS_API_KEY=</span>
              <span style={{ color: '#ec4899' }}>rvo_test_••••••••••••</span>
            </div>
            <Link href="/developer">
              <span style={{ color: 'var(--color-accent-blue)', fontSize: '12px', fontWeight: 600, display: 'flex', alignItems: 'center', gap: '4px' }}>
                Keys <ArrowRight size={12} />
              </span>
            </Link>
          </div>
        </div>

        {/* Step 3 */}
        <div
          style={{
            background: 'var(--bg-card)',
            border: '1px solid var(--border-subtle)',
            borderRadius: '16px',
            padding: '24px',
            display: 'flex',
            flexDirection: 'column',
            justifyContent: 'space-between',
          }}
        >
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '14px' }}>
              <div
                style={{
                  width: '32px',
                  height: '32px',
                  borderRadius: '8px',
                  background: '#10b981',
                  color: '#fff',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontWeight: 800,
                  fontSize: '14px',
                }}
              >
                3
              </div>
              <h3 style={{ margin: 0, fontSize: '16px', fontWeight: 700, color: 'var(--text-primary)' }}>
                Point Razorpay Webhook
              </h3>
            </div>
            <p style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '16px', lineHeight: 1.5 }}>
              In Razorpay Dashboard &gt; Settings &gt; Webhooks, add your endpoint URL and subscribe to `payment.failed`, `payment.captured`, and `refund.created`.
            </p>
          </div>

          <div
            style={{
              background: '#0d1117',
              borderRadius: '8px',
              padding: '12px 14px',
              fontFamily: 'monospace',
              fontSize: '12px',
              color: '#38bdf8',
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
            }}
          >
            https://your-domain.com/api/reviveos/webhook
          </div>
        </div>
      </div>

      {/* Interactive Code Docs Hub */}
      <div
        style={{
          background: 'var(--bg-card)',
          border: '1px solid var(--border-subtle)',
          borderRadius: '20px',
          overflow: 'hidden',
          boxShadow: 'var(--card-shadow)',
        }}
      >
        {/* Code Tabs Header */}
        <div
          style={{
            background: 'var(--bg-secondary)',
            borderBottom: '1px solid var(--border-subtle)',
            padding: '12px 20px',
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            flexWrap: 'wrap',
            gap: '12px',
          }}
        >
          <div style={{ display: 'flex', gap: '8px', flexWrap: 'wrap' }}>
            {[
              { id: 'app-router', label: 'Next.js (App Router)' },
              { id: 'pages-router', label: 'Next.js (Pages Router)' },
              { id: 'express', label: 'Express.js / Node' },
              { id: 'sync-analyze', label: 'Synchronous API' },
            ].map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveCodeTab(tab.id as any)}
                style={{
                  padding: '6px 14px',
                  borderRadius: '8px',
                  border: 'none',
                  background: activeCodeTab === tab.id ? 'var(--color-accent-blue)' : 'transparent',
                  color: activeCodeTab === tab.id ? '#fff' : 'var(--text-secondary)',
                  fontSize: '12.5px',
                  fontWeight: 600,
                  cursor: 'pointer',
                  transition: 'all 0.15s ease',
                }}
              >
                {tab.label}
              </button>
            ))}
          </div>

          <button
            onClick={() => copyText(codeSnippets[activeCodeTab], 'snippet')}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '6px',
              background: 'var(--bg-card)',
              border: '1px solid var(--border-subtle)',
              color: 'var(--text-primary)',
              borderRadius: '6px',
              padding: '6px 12px',
              fontSize: '12px',
              fontWeight: 600,
              cursor: 'pointer',
            }}
          >
            {copiedSnippet === 'snippet' ? <Check size={14} color="#10b981" /> : <Copy size={14} />}
            <span>{copiedSnippet === 'snippet' ? 'Copied!' : 'Copy Code'}</span>
          </button>
        </div>

        {/* Code Body */}
        <div
          style={{
            background: '#0d1117',
            padding: '24px',
            overflowX: 'auto',
            maxHeight: '440px',
          }}
        >
          <pre
            style={{
              margin: 0,
              fontFamily: '"Fira Code", monospace',
              fontSize: '13px',
              lineHeight: 1.6,
              color: '#e6edf3',
            }}
          >
            {codeSnippets[activeCodeTab]}
          </pre>
        </div>

        {/* Footer info banner */}
        <div
          style={{
            padding: '16px 24px',
            background: 'var(--bg-surface)',
            borderTop: '1px solid var(--border-subtle)',
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            flexWrap: 'wrap',
            gap: '12px',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px', fontSize: '13px', color: 'var(--text-secondary)' }}>
            <CheckCircle2 size={16} color="#10b981" />
            <span>Zero Gateway Proxying • Cryptographic HMAC-SHA256 • Built-in Replay Protection</span>
          </div>

          <Link href="/developer">
            <button
              className="btn-primary"
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: '8px',
                padding: '8px 18px',
                fontSize: '13px',
              }}
            >
              Open Developer Portal <ExternalLink size={13} />
            </button>
          </Link>
        </div>
      </div>
    </section>
  );
}

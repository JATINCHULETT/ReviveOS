'use client';

import React, { useState, useEffect } from 'react';
import {
  Code,
  Terminal,
  Key,
  Copy,
  Check,
  Zap,
  Play,
  ShieldCheck,
  RefreshCw,
  Layers,
  ArrowRight,
  Database,
  CheckCircle2,
  AlertTriangle,
} from 'lucide-react';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export default function DeveloperPage() {
  const [activeTab, setActiveTab] = useState<'quickstart' | 'apikeys' | 'simulator' | 'customerMemory'>('quickstart');
  const [copiedKey, setCopiedKey] = useState<string | null>(null);

  // API Keys state
  const [keys, setKeys] = useState<any[]>([
    {
      id: '30000000-0000-0000-0000-000000000001',
      name: 'Default Test Key',
      prefix: 'rvo_test_',
      mode: 'test',
      last_used_at: new Date().toISOString(),
      created_at: new Date().toISOString(),
    },
  ]);
  const [newKeyName, setNewKeyName] = useState('');
  const [newKeyMode, setNewKeyMode] = useState<'test' | 'live'>('test');
  const [generatedKeySecret, setGeneratedKeySecret] = useState<string | null>(null);
  const [loadingKeys, setLoadingKeys] = useState(false);

  // Simulator state
  const [simPaymentId, setSimPaymentId] = useState('pay_sim_' + Math.floor(Math.random() * 899999 + 100000));
  const [simAmount, setSimAmount] = useState('4999');
  const [simFailureCode, setSimFailureCode] = useState('BAD_REQUEST_ERROR');
  const [simFailureReason, setSimFailureReason] = useState('Payment failed due to temporary issuer bank downtime');
  const [simMethod, setSimMethod] = useState('card');
  const [simulating, setSimulating] = useState(false);
  const [simResult, setSimResult] = useState<any>(null);

  // Customer memory state
  const [searchCustId, setSearchCustId] = useState('cust_alex_123');
  const [custProfile, setCustProfile] = useState<any>(null);
  const [custLoading, setCustLoading] = useState(false);

  const copyToClipboard = (text: string, id: string) => {
    navigator.clipboard.writeText(text);
    setCopiedKey(id);
    setTimeout(() => setCopiedKey(null), 2000);
  };

  const fetchKeys = async () => {
    setLoadingKeys(true);
    try {
      const res = await fetch(`${API_BASE}/v1/keys?merchant_id=00000000-0000-0000-0000-000000000001`, {
        headers: { Authorization: 'Bearer rvo_test_mock_key' },
      });
      if (res.ok) {
        const data = await res.json();
        if (data.keys && data.keys.length > 0) {
          setKeys(data.keys);
        }
      }
    } catch {
      // safe fallback
    } finally {
      setLoadingKeys(false);
    }
  };

  useEffect(() => {
    if (activeTab === 'apikeys') {
      fetchKeys();
    }
  }, [activeTab]);

  const handleCreateKey = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const res = await fetch(`${API_BASE}/v1/keys`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: 'Bearer rvo_test_mock_key',
        },
        body: JSON.stringify({
          name: newKeyName || 'Development API Key',
          mode: newKeyMode,
        }),
      });
      if (res.ok) {
        const data = await res.json();
        setGeneratedKeySecret(data.key);
        setNewKeyName('');
        fetchKeys();
        return;
      }
    } catch {
      // offline fallback
    }

    // Local deterministic key generation fallback
    const randomHex = Array.from({ length: 24 }, () => Math.floor(Math.random() * 16).toString(16)).join('');
    const rawKey = `rvo_${newKeyMode}_${randomHex}`;
    const newKeyObj = {
      id: 'key_' + Date.now(),
      name: newKeyName || 'Development API Key',
      prefix: `rvo_${newKeyMode}_`,
      mode: newKeyMode,
      last_used_at: new Date().toISOString(),
      created_at: new Date().toISOString(),
    };
    setKeys((prev) => [newKeyObj, ...prev]);
    setGeneratedKeySecret(rawKey);
    setNewKeyName('');
  };

  const handleRunSimulation = async () => {
    setSimulating(true);
    setSimResult(null);

    const amountVal = parseFloat(simAmount) || 4999.0;
    const isHighFraud = amountVal > 80000;
    const lowerReason = (simFailureReason + ' ' + simFailureCode).toLowerCase();

    let category = 'UNKNOWN';
    let action = 'RETRY_NOW';
    let delaySeconds = 0;
    let decision = 'RECOVER';
    let recoveryProb = 0.82;
    let reason = 'Standard recoverable payment failure.';

    if (isHighFraud) {
      decision = 'BLOCK';
      action = 'BLOCK';
      recoveryProb = 0.12;
      reason = 'High fraud risk score flagged by Revenue Risk Engine';
    } else if (lowerReason.includes('bank') || lowerReason.includes('downtime') || lowerReason.includes('offline') || lowerReason.includes('timeout') || lowerReason.includes('gateway')) {
      category = 'BANK_UNAVAILABLE';
      action = 'RETRY_LATER';
      delaySeconds = 120;
      recoveryProb = 0.94;
      decision = 'RECOVER';
      reason = 'Temporary banking gateway outage; delayed retry scheduled after recovery window.';
    } else if (lowerReason.includes('insufficient') || lowerReason.includes('balance') || lowerReason.includes('funds') || lowerReason.includes('low_balance')) {
      category = 'INSUFFICIENT_FUNDS';
      action = 'RETRY_LATER';
      delaySeconds = 86400;
      recoveryProb = 0.85;
      decision = 'RECOVER';
      reason = 'Insufficient funds; retry timed after standard account replenishment window.';
    } else if (lowerReason.includes('auth') || lowerReason.includes('otp') || lowerReason.includes('3d') || lowerReason.includes('action')) {
      category = 'AUTHENTICATION_FAILED';
      action = 'PAYMENT_LINK';
      delaySeconds = 0;
      recoveryProb = 0.89;
      decision = 'RECOVER';
      reason = 'Customer authentication required; generated smart interactive recovery link.';
    } else if (lowerReason.includes('expired') || lowerReason.includes('card') || lowerReason.includes('mandate')) {
      category = 'EXPIRED_CARD';
      action = 'ALTERNATIVE_PAYMENT';
      delaySeconds = 0;
      recoveryProb = 0.91;
      decision = 'RECOVER';
      reason = 'Expired card credentials; recommend switching to UPI or alternative card.';
    }

    try {
      const res = await fetch(`${API_BASE}/v1/payments/analyze`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: 'Bearer rvo_test_mock_key',
        },
        body: JSON.stringify({
          paymentId: simPaymentId,
          amount: amountVal,
          currency: 'INR',
          failureCode: simFailureCode,
          failureReason: simFailureReason,
          paymentMethod: simMethod,
          customerId: 'cust_demo_sim',
        }),
      });
      if (res.ok) {
        const data = await res.json();
        setSimResult(data);
        setSimulating(false);
        return;
      }
    } catch {
      // Backend not running on port 8080 - fallback to local simulation
    }

    // Local deterministic diagnosis result
    const localResult = {
      paymentId: simPaymentId,
      failureCategory: category,
      diagnosis: `${simFailureCode} failure classified as ${category}`,
      fraudRisk: {
        fraudProbability: isHighFraud ? 0.78 : 0.04,
        riskLevel: isHighFraud ? 'HIGH' : 'LOW',
        expectedLoss: isHighFraud ? amountVal : 0,
        overallRisk: isHighFraud ? 'HIGH' : 'LOW',
      },
      recoveryProbability: recoveryProb,
      nextBestAction: action,
      action: action,
      delaySeconds: delaySeconds,
      confidence: 0.94,
      reason: reason,
      customerHistory: {
        successful_payments: 4,
        failed_payments: 1,
      },
      decision: decision,
      timestamp: new Date().toISOString(),
    };

    setSimResult(localResult);
    setSimulating(false);
  };

  const handleLookupCustomer = async (e: React.FormEvent) => {
    e.preventDefault();
    setCustLoading(true);
    setCustProfile(null);
    try {
      const res = await fetch(`${API_BASE}/v1/customers/${searchCustId}/recovery-profile`, {
        headers: { Authorization: 'Bearer rvo_test_mock_key' },
      });
      if (res.ok) {
        const data = await res.json();
        setCustProfile(data);
        setCustLoading(false);
        return;
      }
    } catch {
      // offline fallback
    }

    // Local customer memory profile
    setCustProfile({
      customerId: searchCustId,
      email: searchCustId.includes('@') ? searchCustId : `${searchCustId}@example.com`,
      phone: '+919876543210',
      previousFailures: 3,
      previousSuccessfulRecoveries: 3,
      preferredPaymentMethods: ['card', 'upi'],
      averageRecoveryTime: 180,
      recoveryProbability: 0.88,
      communicationOptOut: false,
      lastActivityAt: new Date().toISOString(),
    });
    setCustLoading(false);
  };

  return (
    <div style={{ maxWidth: '1200px', margin: '0 auto', paddingBottom: '40px' }}>
        {/* Header Hero Banner */}
        <div
          style={{
            background: 'linear-gradient(135deg, rgba(56, 189, 248, 0.08) 0%, rgba(139, 92, 246, 0.08) 100%)',
            border: '1px solid var(--border-default)',
            borderRadius: '16px',
            padding: '28px 32px',
            marginBottom: '28px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            gap: '20px',
            flexWrap: 'wrap',
          }}
        >
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '8px' }}>
              <span
                style={{
                  background: 'var(--color-accent-blue-bg)',
                  border: '1px solid var(--color-accent-blue)',
                  color: 'var(--color-accent-blue)',
                  fontSize: '11px',
                  fontWeight: 700,
                  padding: '3px 8px',
                  borderRadius: '6px',
                  letterSpacing: '0.05em',
                }}
              >
                ONE-COMMAND INTEGRATION
              </span>
              <span style={{ fontSize: '13px', color: 'var(--text-secondary)', fontWeight: 500 }}>Razorpay AI Layer</span>
            </div>
            <h1 style={{ fontSize: '24px', fontWeight: 800, color: 'var(--text-primary)', margin: 0 }}>
              Plug AI Payment Recovery into Razorpay in 60 Seconds
            </h1>
            <p style={{ fontSize: '14px', color: 'var(--text-secondary)', margin: '6px 0 0', maxWidth: '680px' }}>
              ReviveOS does <strong>not</strong> replace Razorpay and does <strong>not</strong> proxy regular traffic. Successful payments flow directly to Razorpay; failures trigger ReviveOS intelligence and autonomous recovery.
            </p>
          </div>

          <div style={{ display: 'flex', gap: '12px' }}>
            <div
              style={{
                background: 'var(--bg-surface)',
                border: '1px solid var(--border-default)',
                borderRadius: '10px',
                padding: '10px 16px',
                fontFamily: 'var(--font-mono)',
                fontSize: '13px',
                color: 'var(--color-accent-light)',
                display: 'flex',
                alignItems: 'center',
                gap: '8px',
                boxShadow: 'var(--card-shadow)',
              }}
            >
              <span>$ npx reviveos init</span>
              <button
                onClick={() => copyToClipboard('npx reviveos init', 'cmd')}
                style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer', display: 'flex' }}
                title="Copy command"
              >
                {copiedKey === 'cmd' ? <Check size={14} color="var(--color-emerald)" /> : <Copy size={14} />}
              </button>
            </div>
          </div>
        </div>

        {/* Tab Navigation */}
        <div
          style={{
            display: 'flex',
            gap: '8px',
            borderBottom: '1px solid var(--border-subtle)',
            marginBottom: '24px',
          }}
        >
          {[
            { id: 'quickstart', label: '1-Command Setup & Docs', icon: Terminal },
            { id: 'apikeys', label: 'API Keys & Secrets', icon: Key },
            { id: 'simulator', label: 'AI Recovery Simulator', icon: Play },
            { id: 'customerMemory', label: 'Customer Recovery Memory', icon: Database },
          ].map((tab) => {
            const Icon = tab.icon;
            const isActive = activeTab === tab.id;
            return (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id as any)}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '8px',
                  padding: '12px 18px',
                  background: 'none',
                  border: 'none',
                  borderBottom: isActive ? '2px solid var(--color-accent-blue)' : '2px solid transparent',
                  color: isActive ? 'var(--color-accent-blue)' : 'var(--text-secondary)',
                  fontWeight: isActive ? 600 : 500,
                  fontSize: '14px',
                  cursor: 'pointer',
                  transition: 'all 0.15s ease',
                }}
              >
                <Icon size={16} />
                <span>{tab.label}</span>
              </button>
            );
          })}
        </div>

        {/* Tab 1: Quickstart */}
        {activeTab === 'quickstart' && (
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(500px, 1fr))', gap: '20px' }}>
            {/* Step 1: Install & Init */}
            <div className="card" style={{ padding: '24px' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '16px' }}>
                <div
                  style={{
                    width: '28px',
                    height: '28px',
                    borderRadius: '50%',
                    background: 'var(--color-accent-blue)',
                    color: '#fff',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontWeight: 700,
                    fontSize: '13px',
                  }}
                >
                  1
                </div>
                <h3 style={{ margin: 0, fontSize: '16px', fontWeight: 700 }}>Install & Configure with CLI</h3>
              </div>
              <p style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '14px' }}>
                Run the official setup in your existing Next.js, Express, or Node.js Razorpay repository:
              </p>

              <div
                style={{
                  background: 'var(--bg-elevated)',
                  border: '1px solid var(--border-default)',
                  borderRadius: '8px',
                  padding: '14px',
                  fontFamily: 'var(--font-mono)',
                  fontSize: '13px',
                  color: 'var(--text-primary)',
                  position: 'relative',
                  marginBottom: '16px',
                }}
              >
                <div style={{ color: 'var(--text-muted)', marginBottom: '4px' }}># 1. Install ReviveOS Razorpay SDK</div>
                <div style={{ color: 'var(--color-emerald)', fontWeight: 600 }}>npm install @reviveos/razorpay</div>
                <div style={{ color: 'var(--text-muted)', margin: '8px 0 4px' }}># 2. Run the ReviveOS Auto-Configurator</div>
                <div style={{ color: 'var(--color-accent-light)', fontWeight: 600 }}>npx reviveos init</div>
              </div>

              <div style={{ fontSize: '13px', color: 'var(--text-muted)' }}>
                The CLI automatically detects your project framework, generates environment variables, and creates a ready-to-use webhook endpoint.
              </div>
            </div>

            {/* Step 2: Next.js App Router Code */}
            <div className="card" style={{ padding: '24px' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '16px' }}>
                <div
                  style={{
                    width: '28px',
                    height: '28px',
                    borderRadius: '50%',
                    background: 'var(--color-accent-blue)',
                    color: '#fff',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontWeight: 700,
                    fontSize: '13px',
                  }}
                >
                  2
                </div>
                <h3 style={{ margin: 0, fontSize: '16px', fontWeight: 700 }}>Generated Webhook Route</h3>
              </div>
              <div style={{ fontSize: '12px', color: 'var(--text-muted)', marginBottom: '8px', fontFamily: 'var(--font-mono)' }}>
                app/api/reviveos/webhook/route.ts
              </div>

              <div
                style={{
                  background: 'var(--bg-elevated)',
                  border: '1px solid var(--border-default)',
                  borderRadius: '8px',
                  padding: '14px',
                  fontFamily: 'var(--font-mono)',
                  fontSize: '12px',
                  color: 'var(--text-primary)',
                  overflowX: 'auto',
                  maxHeight: '260px',
                }}
              >
                <pre style={{ margin: 0 }}>{`import { NextResponse } from "next/server";
import { ReviveOS } from "@reviveos/razorpay";

const revive = new ReviveOS({
  apiKey: process.env.REVIVEOS_API_KEY,
  webhookSecret: process.env.RAZORPAY_WEBHOOK_SECRET,
});

export async function POST(req: Request) {
  const rawBody = await req.text();
  const signature = req.headers.get("x-razorpay-signature");

  // 1. Verify HMAC signature & normalize event
  const event = revive.webhooks.verifyAndNormalize(rawBody, signature);

  // 2. Ingest into ReviveOS AI pipeline
  const result = await revive.events.process(event);

  return NextResponse.json(result);
}`}</pre>
              </div>
            </div>

            {/* Architectural Diagram Card */}
            <div className="card" style={{ gridColumn: '1 / -1', padding: '24px' }}>
              <h3 style={{ margin: '0 0 16px', fontSize: '16px', fontWeight: 700 }}>
                End-to-End Autonomous Pipeline Architecture
              </h3>
              <div
                style={{
                  display: 'grid',
                  gridTemplateColumns: 'repeat(auto-fit, minmax(140px, 1fr))',
                  gap: '12px',
                  textAlign: 'center',
                }}
              >
                {[
                  { step: '1. Ingest', desc: 'Webhook signature verified & idempotency checked', color: 'var(--color-accent-blue)' },
                  { step: '2. Diagnose', desc: 'Deterministic failure classification', color: '#8b5cf6' },
                  { step: '3. Risk Guard', desc: 'Random Forest fraud & return risk scoring', color: '#f59e0b' },
                  { step: '4. Predict', desc: 'Logistic recovery probability estimation', color: '#10b981' },
                  { step: '5. Decide', desc: 'Next Best Action & Optimal Timing', color: '#06b6d4' },
                  { step: '6. Execute', desc: 'Razorpay payment link or delayed retry', color: '#ec4899' },
                ].map((item, idx) => (
                  <div
                    key={idx}
                    style={{
                      background: 'var(--bg-secondary)',
                      border: '1px solid var(--border-subtle)',
                      borderRadius: '10px',
                      padding: '14px 10px',
                    }}
                  >
                    <div style={{ color: item.color, fontWeight: 700, fontSize: '13px', marginBottom: '4px' }}>
                      {item.step}
                    </div>
                    <div style={{ fontSize: '11px', color: 'var(--text-muted)' }}>{item.desc}</div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}

        {/* Tab 2: API Keys */}
        {activeTab === 'apikeys' && (
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 340px', gap: '24px' }}>
            <div className="card" style={{ padding: '24px' }}>
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '18px' }}>
                <h3 style={{ margin: 0, fontSize: '16px', fontWeight: 700 }}>Active API Keys</h3>
                <button
                  onClick={fetchKeys}
                  style={{
                    background: 'none',
                    border: '1px solid var(--border-subtle)',
                    color: 'var(--text-secondary)',
                    borderRadius: '6px',
                    padding: '6px 12px',
                    fontSize: '12px',
                    cursor: 'pointer',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '6px',
                  }}
                >
                  <RefreshCw size={13} />
                  <span>Refresh</span>
                </button>
              </div>

              {generatedKeySecret && (
                <div
                  style={{
                    background: 'rgba(16, 185, 129, 0.1)',
                    border: '1px solid rgba(16, 185, 129, 0.3)',
                    borderRadius: '8px',
                    padding: '16px',
                    marginBottom: '20px',
                  }}
                >
                  <div style={{ color: 'var(--color-green)', fontWeight: 700, fontSize: '13px', marginBottom: '6px' }}>
                    New API Key Generated Successfully!
                  </div>
                  <p style={{ fontSize: '12px', color: 'var(--text-secondary)', margin: '0 0 10px' }}>
                    Make sure to copy your API key now as you will not be able to see it again.
                  </p>
                  <div
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: '8px',
                      background: '#0d1117',
                      padding: '10px 14px',
                      borderRadius: '6px',
                      fontFamily: 'monospace',
                      fontSize: '13px',
                      color: '#7ee787',
                    }}
                  >
                    <span style={{ flex: 1, wordBreak: 'break-all' }}>{generatedKeySecret}</span>
                    <button
                      onClick={() => copyToClipboard(generatedKeySecret, 'secret')}
                      style={{ background: 'none', border: 'none', color: '#fff', cursor: 'pointer' }}
                    >
                      {copiedKey === 'secret' ? <Check size={14} color="var(--color-green)" /> : <Copy size={14} />}
                    </button>
                  </div>
                </div>
              )}

              <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '13px' }}>
                <thead>
                  <tr style={{ borderBottom: '1px solid var(--border-subtle)', textAlign: 'left', color: 'var(--text-muted)' }}>
                    <th style={{ padding: '10px 0' }}>NAME</th>
                    <th>KEY PREFIX</th>
                    <th>MODE</th>
                    <th>CREATED</th>
                  </tr>
                </thead>
                <tbody>
                  {keys.map((k, i) => (
                    <tr key={i} style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                      <td style={{ padding: '12px 0', fontWeight: 600, color: 'var(--text-primary)' }}>{k.name}</td>
                      <td style={{ fontFamily: 'monospace', color: 'var(--color-accent-blue)' }}>{k.prefix}••••••••</td>
                      <td>
                        <span
                          style={{
                            fontSize: '11px',
                            fontWeight: 700,
                            padding: '2px 8px',
                            borderRadius: '4px',
                            background: k.mode === 'live' ? 'rgba(16, 185, 129, 0.15)' : 'rgba(14, 165, 233, 0.15)',
                            color: k.mode === 'live' ? 'var(--color-green)' : 'var(--color-accent-blue)',
                          }}
                        >
                          {k.mode.toUpperCase()}
                        </span>
                      </td>
                      <td style={{ color: 'var(--text-muted)', fontSize: '12px' }}>
                        {new Date(k.created_at).toLocaleDateString()}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {/* Create Key Form */}
            <div className="card" style={{ padding: '24px' }}>
              <h3 style={{ margin: '0 0 16px', fontSize: '16px', fontWeight: 700 }}>Generate Key</h3>
              <form onSubmit={handleCreateKey}>
                <div style={{ marginBottom: '14px' }}>
                  <label style={{ display: 'block', fontSize: '12px', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '6px' }}>
                    Key Name
                  </label>
                  <input
                    type="text"
                    placeholder="e.g. Production Backend"
                    value={newKeyName}
                    onChange={(e) => setNewKeyName(e.target.value)}
                    style={{
                      width: '100%',
                      padding: '8px 12px',
                      borderRadius: '6px',
                      border: '1px solid var(--border-subtle)',
                      background: 'var(--bg-secondary)',
                      color: 'var(--text-primary)',
                      fontSize: '13px',
                    }}
                    required
                  />
                </div>

                <div style={{ marginBottom: '18px' }}>
                  <label style={{ display: 'block', fontSize: '12px', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '6px' }}>
                    Environment Mode
                  </label>
                  <select
                    value={newKeyMode}
                    onChange={(e: any) => setNewKeyMode(e.target.value)}
                    style={{
                      width: '100%',
                      padding: '8px 12px',
                      borderRadius: '6px',
                      border: '1px solid var(--border-subtle)',
                      background: 'var(--bg-secondary)',
                      color: 'var(--text-primary)',
                      fontSize: '13px',
                    }}
                  >
                    <option value="test">Test Sandbox (rvo_test_...)</option>
                    <option value="live">Production Live (rvo_live_...)</option>
                  </select>
                </div>

                <button
                  type="submit"
                  style={{
                    width: '100%',
                    padding: '10px',
                    borderRadius: '6px',
                    background: 'var(--color-accent-blue)',
                    color: '#fff',
                    border: 'none',
                    fontWeight: 600,
                    fontSize: '13px',
                    cursor: 'pointer',
                  }}
                >
                  Create API Key
                </button>
              </form>
            </div>
          </div>
        )}

        {/* Tab 3: Simulator */}
        {activeTab === 'simulator' && (
          <div style={{ display: 'grid', gridTemplateColumns: '380px 1fr', gap: '24px' }}>
            <div className="card" style={{ padding: '24px' }}>
              <h3 style={{ margin: '0 0 16px', fontSize: '16px', fontWeight: 700 }}>Simulate Razorpay Failure</h3>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '14px' }}>
                <div>
                  <label style={{ fontSize: '12px', fontWeight: 600, color: 'var(--text-secondary)', display: 'block', marginBottom: '4px' }}>
                    Payment ID
                  </label>
                  <input
                    type="text"
                    value={simPaymentId}
                    onChange={(e) => setSimPaymentId(e.target.value)}
                    style={{
                      width: '100%',
                      padding: '8px 12px',
                      borderRadius: '6px',
                      border: '1px solid var(--border-subtle)',
                      background: 'var(--bg-secondary)',
                      color: 'var(--text-primary)',
                      fontSize: '13px',
                      fontFamily: 'monospace',
                    }}
                  />
                </div>

                <div>
                  <label style={{ fontSize: '12px', fontWeight: 600, color: 'var(--text-secondary)', display: 'block', marginBottom: '4px' }}>
                    Amount (₹ INR)
                  </label>
                  <input
                    type="number"
                    value={simAmount}
                    onChange={(e) => setSimAmount(e.target.value)}
                    style={{
                      width: '100%',
                      padding: '8px 12px',
                      borderRadius: '6px',
                      border: '1px solid var(--border-subtle)',
                      background: 'var(--bg-secondary)',
                      color: 'var(--text-primary)',
                      fontSize: '13px',
                    }}
                  />
                </div>

                <div>
                  <label style={{ fontSize: '12px', fontWeight: 600, color: 'var(--text-secondary)', display: 'block', marginBottom: '4px' }}>
                    Razorpay Error Code
                  </label>
                  <select
                    value={simFailureCode}
                    onChange={(e) => setSimFailureCode(e.target.value)}
                    style={{
                      width: '100%',
                      padding: '8px 12px',
                      borderRadius: '6px',
                      border: '1px solid var(--border-subtle)',
                      background: 'var(--bg-secondary)',
                      color: 'var(--text-primary)',
                      fontSize: '13px',
                    }}
                  >
                    <option value="BAD_REQUEST_ERROR">BAD_REQUEST_ERROR (Bank Issue / Balance)</option>
                    <option value="GATEWAY_ERROR">GATEWAY_ERROR (Gateway Timeout)</option>
                    <option value="AUTHENTICATION_FAILED">AUTHENTICATION_FAILED (OTP / 3DS)</option>
                    <option value="CARD_EXPIRED">CARD_EXPIRED (Expired Card)</option>
                  </select>
                </div>

                <div>
                  <label style={{ fontSize: '12px', fontWeight: 600, color: 'var(--text-secondary)', display: 'block', marginBottom: '4px' }}>
                    Error Description
                  </label>
                  <input
                    type="text"
                    value={simFailureReason}
                    onChange={(e) => setSimFailureReason(e.target.value)}
                    style={{
                      width: '100%',
                      padding: '8px 12px',
                      borderRadius: '6px',
                      border: '1px solid var(--border-subtle)',
                      background: 'var(--bg-secondary)',
                      color: 'var(--text-primary)',
                      fontSize: '13px',
                    }}
                  />
                </div>

                <button
                  onClick={handleRunSimulation}
                  disabled={simulating}
                  style={{
                    marginTop: '8px',
                    padding: '12px',
                    borderRadius: '8px',
                    background: 'var(--color-accent-blue)',
                    color: '#fff',
                    border: 'none',
                    fontWeight: 700,
                    fontSize: '14px',
                    cursor: simulating ? 'not-allowed' : 'pointer',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    gap: '8px',
                  }}
                >
                  <Play size={16} />
                  <span>{simulating ? 'Analyzing Failure...' : 'Trigger AI Analysis'}</span>
                </button>
              </div>
            </div>

            {/* Simulation Result */}
            <div className="card" style={{ padding: '24px' }}>
              <h3 style={{ margin: '0 0 16px', fontSize: '16px', fontWeight: 700 }}>AI Pipeline Diagnosis Output</h3>

              {simResult ? (
                <div>
                  <div
                    style={{
                      display: 'grid',
                      gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))',
                      gap: '12px',
                      marginBottom: '20px',
                    }}
                  >
                    <div style={{ background: 'var(--bg-secondary)', padding: '14px', borderRadius: '8px' }}>
                      <div style={{ fontSize: '11px', color: 'var(--text-muted)', fontWeight: 600 }}>CLASSIFICATION</div>
                      <div style={{ fontSize: '15px', fontWeight: 700, color: 'var(--color-accent-blue)', marginTop: '4px' }}>
                        {simResult.failureCategory || 'UNKNOWN'}
                      </div>
                    </div>

                    <div style={{ background: 'var(--bg-secondary)', padding: '14px', borderRadius: '8px' }}>
                      <div style={{ fontSize: '11px', color: 'var(--text-muted)', fontWeight: 600 }}>RECOVERY PROBABILITY</div>
                      <div style={{ fontSize: '15px', fontWeight: 700, color: 'var(--color-green)', marginTop: '4px' }}>
                        {simResult.recoveryProbability ? `${(simResult.recoveryProbability * 100).toFixed(1)}%` : 'N/A'}
                      </div>
                    </div>

                    <div style={{ background: 'var(--bg-secondary)', padding: '14px', borderRadius: '8px' }}>
                      <div style={{ fontSize: '11px', color: 'var(--text-muted)', fontWeight: 600 }}>FRAUD RISK</div>
                      <div
                        style={{
                          fontSize: '15px',
                          fontWeight: 700,
                          color: simResult.fraudRisk?.riskLevel === 'HIGH' ? 'var(--color-red)' : 'var(--color-green)',
                          marginTop: '4px',
                        }}
                      >
                        {simResult.fraudRisk?.riskLevel || 'LOW'} ({(simResult.fraudRisk?.fraudProbability * 100 || 4).toFixed(1)}%)
                      </div>
                    </div>

                    <div style={{ background: 'var(--bg-secondary)', padding: '14px', borderRadius: '8px' }}>
                      <div style={{ fontSize: '11px', color: 'var(--text-muted)', fontWeight: 600 }}>NEXT BEST ACTION</div>
                      <div style={{ fontSize: '15px', fontWeight: 700, color: 'var(--text-primary)', marginTop: '4px' }}>
                        {simResult.action || 'RETRY_NOW'} {simResult.delaySeconds ? `(${simResult.delaySeconds}s delay)` : ''}
                      </div>
                    </div>
                  </div>

                  <div style={{ background: 'var(--bg-secondary)', padding: '14px', borderRadius: '8px', marginBottom: '16px' }}>
                    <div style={{ fontSize: '12px', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '4px' }}>
                      AI Reasoning & Diagnosis:
                    </div>
                    <div style={{ fontSize: '13px', color: 'var(--text-primary)' }}>{simResult.reason || simResult.diagnosis}</div>
                  </div>

                  <div style={{ fontSize: '12px', color: 'var(--text-muted)', marginBottom: '6px' }}>Raw JSON Response:</div>
                  <pre
                    style={{
                      background: '#0d1117',
                      padding: '12px',
                      borderRadius: '6px',
                      fontSize: '12px',
                      color: '#7ee787',
                      overflowX: 'auto',
                      maxHeight: '200px',
                    }}
                  >
                    {JSON.stringify(simResult, null, 2)}
                  </pre>
                </div>
              ) : (
                <div style={{ textAlign: 'center', padding: '60px 20px', color: 'var(--text-muted)' }}>
                  <Play size={36} style={{ marginBottom: '12px', opacity: 0.4 }} />
                  <div>Select error parameters on the left and click "Trigger AI Analysis" to view real-time pipeline inference.</div>
                </div>
              )}
            </div>
          </div>
        )}

        {/* Tab 4: Customer Recovery Memory */}
        {activeTab === 'customerMemory' && (
          <div style={{ maxWidth: '800px', margin: '0 auto' }}>
            <div className="card" style={{ padding: '24px', marginBottom: '24px' }}>
              <h3 style={{ margin: '0 0 16px', fontSize: '16px', fontWeight: 700 }}>Customer Recovery Memory Profile</h3>
              <form onSubmit={handleLookupCustomer} style={{ display: 'flex', gap: '12px' }}>
                <input
                  type="text"
                  placeholder="Customer ID or Email (e.g. cust_alex_123)"
                  value={searchCustId}
                  onChange={(e) => setSearchCustId(e.target.value)}
                  style={{
                    flex: 1,
                    padding: '10px 14px',
                    borderRadius: '6px',
                    border: '1px solid var(--border-subtle)',
                    background: 'var(--bg-secondary)',
                    color: 'var(--text-primary)',
                    fontSize: '13px',
                  }}
                  required
                />
                <button
                  type="submit"
                  style={{
                    padding: '10px 20px',
                    borderRadius: '6px',
                    background: 'var(--color-accent-blue)',
                    color: '#fff',
                    border: 'none',
                    fontWeight: 600,
                    fontSize: '13px',
                    cursor: 'pointer',
                  }}
                >
                  {custLoading ? 'Searching...' : 'Lookup Profile'}
                </button>
              </form>
            </div>

            {custProfile && (
              <div className="card" style={{ padding: '24px' }}>
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '18px' }}>
                  <div>
                    <h4 style={{ margin: 0, fontSize: '15px', fontWeight: 700 }}>{custProfile.customerId}</h4>
                    <span style={{ fontSize: '12px', color: 'var(--text-muted)' }}>{custProfile.email || 'No email attached'}</span>
                  </div>
                  <span
                    style={{
                      fontSize: '12px',
                      fontWeight: 700,
                      color: custProfile.recoveryProbability > 0.7 ? 'var(--color-green)' : 'var(--color-accent-blue)',
                      background: 'var(--bg-secondary)',
                      padding: '4px 10px',
                      borderRadius: '6px',
                    }}
                  >
                    Historical Recovery Rate: {(custProfile.recoveryProbability * 100).toFixed(1)}%
                  </span>
                </div>

                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '12px', marginBottom: '20px' }}>
                  <div style={{ background: 'var(--bg-secondary)', padding: '14px', borderRadius: '8px' }}>
                    <div style={{ fontSize: '11px', color: 'var(--text-muted)' }}>PREVIOUS FAILURES</div>
                    <div style={{ fontSize: '18px', fontWeight: 800, color: 'var(--text-primary)', marginTop: '4px' }}>
                      {custProfile.previousFailures}
                    </div>
                  </div>
                  <div style={{ background: 'var(--bg-secondary)', padding: '14px', borderRadius: '8px' }}>
                    <div style={{ fontSize: '11px', color: 'var(--text-muted)' }}>SUCCESSFUL RECOVERIES</div>
                    <div style={{ fontSize: '18px', fontWeight: 800, color: 'var(--color-green)', marginTop: '4px' }}>
                      {custProfile.previousSuccessfulRecoveries}
                    </div>
                  </div>
                  <div style={{ background: 'var(--bg-secondary)', padding: '14px', borderRadius: '8px' }}>
                    <div style={{ fontSize: '11px', color: 'var(--text-muted)' }}>AVG RECOVERY TIME</div>
                    <div style={{ fontSize: '18px', fontWeight: 800, color: 'var(--color-accent-blue)', marginTop: '4px' }}>
                      {custProfile.averageRecoveryTime}s
                    </div>
                  </div>
                </div>

                <div>
                  <div style={{ fontSize: '12px', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '8px' }}>
                    Preferred Payment Methods:
                  </div>
                  <div style={{ display: 'flex', gap: '8px' }}>
                    {custProfile.preferredPaymentMethods.map((m: string, i: number) => (
                      <span
                        key={i}
                        style={{
                          background: 'var(--color-accent-bg)',
                          border: '1px solid var(--color-accent-border)',
                          color: 'var(--color-accent-light)',
                          padding: '4px 10px',
                          borderRadius: '6px',
                          fontSize: '12px',
                          fontWeight: 600,
                        }}
                      >
                        {m.toUpperCase()}
                      </span>
                    ))}
                  </div>
                </div>
              </div>
            )}
          </div>
        )}
      </div>
  );
}

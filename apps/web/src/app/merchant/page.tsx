'use client';

import React, { useEffect, useState } from 'react';
import { 
  Building2, 
  CreditCard, 
  RefreshCw, 
  CheckCircle2, 
  XCircle, 
  Plus, 
  Send, 
  Zap, 
  ArrowUpRight,
  ExternalLink,
  ShieldCheck,
  Clock,
  Play
} from 'lucide-react';
import { getMerchantDashboard, createSubscription, createSandboxPaymentLink } from '@/lib/api';

export default function MerchantPortalPage() {
  const [data, setData] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<'subscriptions' | 'sandbox' | 'customers'>('subscriptions');

  // New Subscription Form State
  const [subEmail, setSubEmail] = useState('');
  const [subAmount, setSubAmount] = useState('1499.00');
  const [subPlan, setSubPlan] = useState('plan_pro_monthly');
  const [subInterval, setSubInterval] = useState('monthly');
  const [creatingSub, setCreatingSub] = useState(false);

  // Sandbox Form State
  const [sbEmail, setSbEmail] = useState('');
  const [sbAmount, setSbAmount] = useState('500.00');
  const [sbFailureCode, setSbFailureCode] = useState('INSUFFICIENT_FUNDS');
  const [sbTriggerFail, setSbTriggerFail] = useState(true);
  const [creatingLink, setCreatingLink] = useState(false);
  const [lastCreatedLink, setLastCreatedLink] = useState<any>(null);

  const fetchDashboard = async () => {
    try {
      setLoading(true);
      const res = await getMerchantDashboard();
      setData(res);
    } catch (err) {
      console.error('Failed to load merchant dashboard:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDashboard();
  }, []);

  const handleCreateSubscription = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!data?.merchant?.id) return;
    setCreatingSub(true);
    try {
      await createSubscription({
        merchant_id: data.merchant.id,
        customer_email: subEmail,
        amount: parseFloat(subAmount) || 1000,
        plan_id: subPlan,
        billing_interval: subInterval,
      });
      setSubEmail('');
      await fetchDashboard();
      alert('Recurring Subscription created successfully!');
    } catch (err: any) {
      alert(`Error: ${err.message}`);
    } finally {
      setCreatingSub(false);
    }
  };

  const handleCreateSandboxLink = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!data?.merchant?.id) return;
    setCreatingLink(true);
    try {
      const res = await createSandboxPaymentLink({
        merchant_id: data.merchant.id,
        customer_email: sbEmail.trim() || 'delivered@resend.dev',
        amount: parseFloat(sbAmount) || 500,
        trigger_failure_immediately: sbTriggerFail,
        failure_code: sbFailureCode,
      });
      setLastCreatedLink(res);
      await fetchDashboard();
    } catch (err: any) {
      alert(`Error: ${err.message}`);
    } finally {
      setCreatingLink(false);
    }
  };

  if (loading && !data) {
    return (
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', minHeight: '60vh', color: 'var(--text-muted)' }}>
        <RefreshCw size={24} className="spin" />
        <span style={{ marginLeft: '12px' }}>Loading Merchant Portal...</span>
      </div>
    );
  }

  const metrics = data?.metrics || {};

  return (
    <div style={{ padding: '24px 32px' }}>
      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '28px' }}>
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '4px' }}>
            <Building2 size={24} color="#38BDF8" />
            <h1 style={{ fontSize: '24px', fontWeight: 700, margin: 0, color: 'var(--text-primary)' }}>
              {data?.merchant?.name || 'Acme Cloud Services'}
            </h1>
            <span style={{ background: 'rgba(56, 189, 248, 0.15)', color: '#38BDF8', fontSize: '11px', fontWeight: 600, padding: '2px 8px', borderRadius: '4px' }}>
              MERCHANT PORTAL
            </span>
          </div>
          <p style={{ fontSize: '13px', color: 'var(--text-muted)', margin: 0 }}>
            Manage recurring subscriptions, test checkout sandbox links, and monitor AI payment recovery in real-time.
          </p>
        </div>

        <button 
          onClick={fetchDashboard}
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
            background: 'var(--bg-card)',
            border: '1px solid var(--border-subtle)',
            color: 'var(--text-secondary)',
            padding: '8px 16px',
            borderRadius: '8px',
            cursor: 'pointer',
            fontSize: '13px',
            fontWeight: 500
          }}
        >
          <RefreshCw size={14} className={loading ? 'spin' : ''} />
          <span>Refresh</span>
        </button>
      </div>

      {/* Metrics Row */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '16px', marginBottom: '28px' }}>
        <div className="metric-card" style={{ background: 'var(--bg-card)', border: '1px solid var(--border-subtle)', padding: '20px', borderRadius: '12px' }}>
          <div style={{ fontSize: '12px', fontWeight: 600, color: 'var(--text-muted)', textTransform: 'uppercase' }}>Active Subscriptions</div>
          <div style={{ fontSize: '26px', fontWeight: 700, color: 'var(--text-primary)', marginTop: '8px' }}>
            {metrics.active_subscriptions ?? metrics.ActiveSubscriptions ?? (data?.subscriptions?.filter((s: any) => s.status === 'ACTIVE').length || 0)}
          </div>
          <div style={{ fontSize: '12px', color: '#10B981', marginTop: '4px' }}>Recurring revenue protected</div>
        </div>

        <div className="metric-card" style={{ background: 'var(--bg-card)', border: '1px solid var(--border-subtle)', padding: '20px', borderRadius: '12px' }}>
          <div style={{ fontSize: '12px', fontWeight: 600, color: 'var(--text-muted)', textTransform: 'uppercase' }}>At-Risk Volume</div>
          <div style={{ fontSize: '26px', fontWeight: 700, color: '#F87171', marginTop: '8px' }}>
            ₹{(metrics.total_at_risk_revenue ?? metrics.TotalAtRiskRevenue ?? 0).toLocaleString('en-IN', { minimumFractionDigits: 2 })}
          </div>
          <div style={{ fontSize: '12px', color: 'var(--text-muted)', marginTop: '4px' }}>{metrics.pending_recoveries ?? metrics.PendingRecoveries ?? 0} pending workflows</div>
        </div>

        <div className="metric-card" style={{ background: 'var(--bg-card)', border: '1px solid var(--border-subtle)', padding: '20px', borderRadius: '12px' }}>
          <div style={{ fontSize: '12px', fontWeight: 600, color: 'var(--text-muted)', textTransform: 'uppercase' }}>Recovered Revenue</div>
          <div style={{ fontSize: '26px', fontWeight: 700, color: '#34D399', marginTop: '8px' }}>
            ₹{(metrics.total_recovered_revenue ?? metrics.TotalRecovered ?? 0).toLocaleString('en-IN', { minimumFractionDigits: 2 })}
          </div>
          <div style={{ fontSize: '12px', color: '#34D399', marginTop: '4px' }}>Verified by payment provider</div>
        </div>

        <div className="metric-card" style={{ background: 'var(--bg-card)', border: '1px solid var(--border-subtle)', padding: '20px', borderRadius: '12px' }}>
          <div style={{ fontSize: '12px', fontWeight: 600, color: 'var(--text-muted)', textTransform: 'uppercase' }}>Overall Recovery Rate</div>
          <div style={{ fontSize: '26px', fontWeight: 700, color: '#38BDF8', marginTop: '8px' }}>
            {(metrics.recovery_rate ?? metrics.RecoveryRate ?? 0).toFixed(1)}%
          </div>
          <div style={{ fontSize: '12px', color: '#38BDF8', marginTop: '4px' }}>AI Adaptive Orchestration</div>
        </div>
      </div>

      {/* Tabs */}
      <div style={{ display: 'flex', gap: '8px', borderBottom: '1px solid var(--border-subtle)', marginBottom: '24px' }}>
        <button
          onClick={() => setActiveTab('subscriptions')}
          style={{
            padding: '10px 18px',
            background: 'transparent',
            border: 'none',
            borderBottom: activeTab === 'subscriptions' ? '2px solid #38BDF8' : '2px solid transparent',
            color: activeTab === 'subscriptions' ? '#38BDF8' : 'var(--text-secondary)',
            fontWeight: 600,
            fontSize: '14px',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            gap: '8px'
          }}
        >
          <CreditCard size={16} />
          <span>Recurring Subscriptions ({data?.subscriptions?.length || 0})</span>
        </button>

        <button
          onClick={() => setActiveTab('sandbox')}
          style={{
            padding: '10px 18px',
            background: 'transparent',
            border: 'none',
            borderBottom: activeTab === 'sandbox' ? '2px solid #38BDF8' : '2px solid transparent',
            color: activeTab === 'sandbox' ? '#38BDF8' : 'var(--text-secondary)',
            fontWeight: 600,
            fontSize: '14px',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            gap: '8px'
          }}
        >
          <Zap size={16} />
          <span>Payment Link Sandbox Simulator</span>
        </button>

        <button
          onClick={() => setActiveTab('customers')}
          style={{
            padding: '10px 18px',
            background: 'transparent',
            border: 'none',
            borderBottom: activeTab === 'customers' ? '2px solid #38BDF8' : '2px solid transparent',
            color: activeTab === 'customers' ? '#38BDF8' : 'var(--text-secondary)',
            fontWeight: 600,
            fontSize: '14px',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            gap: '8px'
          }}
        >
          <ShieldCheck size={16} />
          <span>Customer Recovery Status ({data?.customers?.length || 0})</span>
        </button>
      </div>

      {/* Tab Content: Subscriptions */}
      {activeTab === 'subscriptions' && (
        <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: '24px' }}>
          {/* Subscriptions Table */}
          <div style={{ background: 'var(--bg-card)', border: '1px solid var(--border-subtle)', borderRadius: '12px', padding: '20px' }}>
            <h3 style={{ fontSize: '16px', fontWeight: 600, margin: '0 0 16px', color: 'var(--text-primary)' }}>
              Active & Recurring Subscriptions
            </h3>
            {(!data?.subscriptions || data.subscriptions.length === 0) ? (
              <div style={{ padding: '32px', textAlign: 'center', color: 'var(--text-muted)' }}>
                No subscriptions registered yet. Create one using the form on the right.
              </div>
            ) : (
              <div style={{ overflowX: 'auto' }}>
                <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '13px' }}>
                  <thead>
                    <tr style={{ borderBottom: '1px solid var(--border-subtle)', color: 'var(--text-muted)', textAlign: 'left' }}>
                      <th style={{ padding: '10px 12px' }}>Customer</th>
                      <th style={{ padding: '10px 12px' }}>Plan / Interval</th>
                      <th style={{ padding: '10px 12px' }}>Amount</th>
                      <th style={{ padding: '10px 12px' }}>Status</th>
                      <th style={{ padding: '10px 12px' }}>Next Billing</th>
                    </tr>
                  </thead>
                  <tbody>
                    {data.subscriptions.map((sub: any) => (
                      <tr key={sub.id} style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                        <td style={{ padding: '12px', fontWeight: 500, color: 'var(--text-primary)' }}>{sub.customer_email}</td>
                        <td style={{ padding: '12px', color: 'var(--text-secondary)' }}>
                          <div>{sub.plan_id}</div>
                          <span style={{ fontSize: '11px', color: 'var(--text-muted)', textTransform: 'uppercase' }}>{sub.billing_interval}</span>
                        </td>
                        <td style={{ padding: '12px', fontWeight: 600, color: 'var(--text-primary)' }}>₹{sub.amount.toFixed(2)}</td>
                        <td style={{ padding: '12px' }}>
                          <span style={{
                            padding: '3px 8px',
                            borderRadius: '4px',
                            fontSize: '11px',
                            fontWeight: 600,
                            background: sub.status === 'ACTIVE' ? 'rgba(52, 211, 153, 0.15)' : 'rgba(248, 113, 113, 0.15)',
                            color: sub.status === 'ACTIVE' ? '#34D399' : '#F87171'
                          }}>
                            {sub.status}
                          </span>
                        </td>
                        <td style={{ padding: '12px', color: 'var(--text-muted)', fontSize: '12px' }}>
                          {sub.next_billing_at ? new Date(sub.next_billing_at).toLocaleDateString() : 'N/A'}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>

          {/* Create Subscription Form */}
          <div style={{ background: 'var(--bg-card)', border: '1px solid var(--border-subtle)', borderRadius: '12px', padding: '20px' }}>
            <h3 style={{ fontSize: '16px', fontWeight: 600, margin: '0 0 16px', color: 'var(--text-primary)', display: 'flex', alignItems: 'center', gap: '8px' }}>
              <Plus size={16} color="#38BDF8" />
              <span>New Recurring Subscription</span>
            </h3>

            <form onSubmit={handleCreateSubscription}>
              <div style={{ marginBottom: '14px' }}>
                <label style={{ display: 'block', fontSize: '12px', color: 'var(--text-secondary)', marginBottom: '4px' }}>Customer Email</label>
                <input
                  type="email"
                  required
                  placeholder="subscriber@example.com"
                  value={subEmail}
                  onChange={(e) => setSubEmail(e.target.value)}
                  style={{ width: '100%', padding: '8px 12px', borderRadius: '6px', background: 'var(--bg-input)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                />
              </div>

              <div style={{ marginBottom: '14px' }}>
                <label style={{ display: 'block', fontSize: '12px', color: 'var(--text-secondary)', marginBottom: '4px' }}>Monthly Amount (INR)</label>
                <input
                  type="number"
                  required
                  value={subAmount}
                  onChange={(e) => setSubAmount(e.target.value)}
                  style={{ width: '100%', padding: '8px 12px', borderRadius: '6px', background: 'var(--bg-input)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                />
              </div>

              <div style={{ marginBottom: '14px' }}>
                <label style={{ display: 'block', fontSize: '12px', color: 'var(--text-secondary)', marginBottom: '4px' }}>Plan Tier</label>
                <select
                  value={subPlan}
                  onChange={(e) => setSubPlan(e.target.value)}
                  style={{ width: '100%', padding: '8px 12px', borderRadius: '6px', background: 'var(--bg-input)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                >
                  <option value="plan_starter_monthly">Starter Tier (₹499/mo)</option>
                  <option value="plan_pro_monthly">Pro Business Tier (₹1,499/mo)</option>
                  <option value="plan_enterprise_monthly">Enterprise Tier (₹4,999/mo)</option>
                </select>
              </div>

              <button
                type="submit"
                disabled={creatingSub}
                style={{
                  width: '100%',
                  padding: '10px',
                  borderRadius: '6px',
                  background: 'linear-gradient(135deg, #2563EB, #1D4ED8)',
                  color: '#FFF',
                  border: 'none',
                  fontWeight: 600,
                  fontSize: '13px',
                  cursor: creatingSub ? 'not-allowed' : 'pointer'
                }}
              >
                {creatingSub ? 'Creating...' : 'Enroll Subscription'}
              </button>
            </form>
          </div>
        </div>
      )}

      {/* Tab Content: Sandbox */}
      {activeTab === 'sandbox' && (
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '24px' }}>
          {/* Sandbox Generator */}
          <div style={{ background: 'var(--bg-card)', border: '1px solid var(--border-subtle)', borderRadius: '12px', padding: '24px' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '8px' }}>
              <Zap size={20} color="#F59E0B" />
              <h3 style={{ fontSize: '18px', fontWeight: 600, margin: 0, color: 'var(--text-primary)' }}>
                Payment Link Sandbox
              </h3>
            </div>
            <p style={{ fontSize: '13px', color: 'var(--text-muted)', marginBottom: '20px' }}>
              Simulate checkout transactions with specific card failure triggers to test the AI adaptive recovery workflow end-to-end.
            </p>

            <form onSubmit={handleCreateSandboxLink}>
              <div style={{ marginBottom: '16px' }}>
                <label style={{ display: 'block', fontSize: '12px', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '6px' }}>
                  Customer Email
                </label>
                <input
                  type="email"
                  placeholder="test_user@gmail.com"
                  value={sbEmail}
                  onChange={(e) => setSbEmail(e.target.value)}
                  style={{ width: '100%', padding: '10px 12px', borderRadius: '8px', background: 'var(--bg-input)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                />
              </div>

              <div style={{ marginBottom: '16px' }}>
                <label style={{ display: 'block', fontSize: '12px', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '6px' }}>
                  Amount (INR)
                </label>
                <input
                  type="number"
                  value={sbAmount}
                  onChange={(e) => setSbAmount(e.target.value)}
                  style={{ width: '100%', padding: '10px 12px', borderRadius: '8px', background: 'var(--bg-input)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                />
              </div>

              <div style={{ marginBottom: '16px' }}>
                <label style={{ display: 'block', fontSize: '12px', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '6px' }}>
                  Failure Reason / Gateway Error Code
                </label>
                <select
                  value={sbFailureCode}
                  onChange={(e) => setSbFailureCode(e.target.value)}
                  style={{ width: '100%', padding: '10px 12px', borderRadius: '8px', background: 'var(--bg-input)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                >
                  <option value="INSUFFICIENT_FUNDS">INSUFFICIENT_FUNDS (Smart Delayed Retry)</option>
                  <option value="EXPIRED_CARD">EXPIRED_CARD (Customer Payment Link Trigger)</option>
                  <option value="BANK_UNAVAILABLE">BANK_UNAVAILABLE (Immediate / Jitter Retry)</option>
                  <option value="AUTHENTICATION_FAILED">AUTHENTICATION_FAILED (3D Secure Required)</option>
                  <option value="LIMIT_EXCEEDED">LIMIT_EXCEEDED (Escalation / Card Update)</option>
                </select>
              </div>

              <div style={{ marginBottom: '20px', display: 'flex', alignItems: 'center', gap: '8px' }}>
                <input
                  type="checkbox"
                  id="triggerFail"
                  checked={sbTriggerFail}
                  onChange={(e) => setSbTriggerFail(e.target.checked)}
                  style={{ cursor: 'pointer' }}
                />
                <label htmlFor="triggerFail" style={{ fontSize: '13px', color: 'var(--text-primary)', cursor: 'pointer' }}>
                  Trigger immediate failure & launch AI Recovery Pipeline
                </label>
              </div>

              <button
                type="submit"
                disabled={creatingLink}
                style={{
                  width: '100%',
                  padding: '12px',
                  borderRadius: '8px',
                  background: 'linear-gradient(135deg, #F59E0B, #D97706)',
                  color: '#FFF',
                  border: 'none',
                  fontWeight: 600,
                  fontSize: '14px',
                  cursor: creatingLink ? 'not-allowed' : 'pointer',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: '8px'
                }}
              >
                <Play size={16} />
                <span>{creatingLink ? 'Generating...' : 'Generate & Simulate Recovery'}</span>
              </button>
            </form>
          </div>

          {/* Last Created Link Preview */}
          <div style={{ background: 'var(--bg-card)', border: '1px solid var(--border-subtle)', borderRadius: '12px', padding: '24px' }}>
            <h3 style={{ fontSize: '18px', fontWeight: 600, margin: '0 0 16px', color: 'var(--text-primary)' }}>
              Sandbox Output & Link
            </h3>

            {lastCreatedLink ? (
              <div style={{ background: 'var(--bg-input)', border: '1px solid var(--border-subtle)', borderRadius: '8px', padding: '16px' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px', color: '#10B981', fontWeight: 600, fontSize: '14px', marginBottom: '12px' }}>
                  <CheckCircle2 size={18} />
                  <span>Sandbox Transaction Dispatched</span>
                </div>

                <div style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '8px' }}>
                  <strong>Payment ID:</strong> <span style={{ fontFamily: 'monospace', color: 'var(--text-primary)' }}>{lastCreatedLink.payment_id}</span>
                </div>
                <div style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '8px' }}>
                  <strong>Amount:</strong> ₹{lastCreatedLink.amount}
                </div>
                <div style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '16px' }}>
                  <strong>Recovery Queued:</strong> <span style={{ color: '#38BDF8' }}>{lastCreatedLink.recovery_queued ? 'YES (AI Orchestrator Active)' : 'NO'}</span>
                </div>

                <div style={{ borderTop: '1px solid var(--border-subtle)', paddingTop: '12px' }}>
                  <div style={{ fontSize: '12px', color: 'var(--text-muted)', marginBottom: '6px' }}>Payment Link URL:</div>
                  <a
                    href={lastCreatedLink.payment_link_url}
                    target="_blank"
                    rel="noreferrer"
                    style={{ color: '#38BDF8', fontSize: '13px', display: 'flex', alignItems: 'center', gap: '4px', wordBreak: 'break-all' }}
                  >
                    <span>{lastCreatedLink.payment_link_url}</span>
                    <ExternalLink size={14} />
                  </a>
                </div>
              </div>
            ) : (
              <div style={{ padding: '32px', textAlign: 'center', color: 'var(--text-muted)', fontSize: '13px' }}>
                Generate a sandbox payment link using the simulator to view its live URL and tracking status here.
              </div>
            )}
          </div>
        </div>
      )}

      {/* Tab Content: Customers Recovery Status */}
      {activeTab === 'customers' && (
        <div style={{ background: 'var(--bg-card)', border: '1px solid var(--border-subtle)', borderRadius: '12px', padding: '20px' }}>
          <h3 style={{ fontSize: '16px', fontWeight: 600, margin: '0 0 16px', color: 'var(--text-primary)' }}>
            Customer Recovery Status & History
          </h3>

          {(!data?.customers || data.customers.length === 0) ? (
            <div style={{ padding: '32px', textAlign: 'center', color: 'var(--text-muted)' }}>
              No customer recovery records found for this merchant.
            </div>
          ) : (
            <div style={{ overflowX: 'auto' }}>
              <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '13px' }}>
                <thead>
                  <tr style={{ borderBottom: '1px solid var(--border-subtle)', color: 'var(--text-muted)', textAlign: 'left' }}>
                    <th style={{ padding: '10px 12px' }}>Customer Email</th>
                    <th style={{ padding: '10px 12px' }}>Communications</th>
                    <th style={{ padding: '10px 12px' }}>Failed Payments</th>
                    <th style={{ padding: '10px 12px' }}>Recovered</th>
                    <th style={{ padding: '10px 12px' }}>Latest Action</th>
                    <th style={{ padding: '10px 12px' }}>Status</th>
                  </tr>
                </thead>
                <tbody>
                  {data.customers.map((c: any) => {
                    const status = (c.last_status || 'NEW').toUpperCase();
                    let badgeBg = 'rgba(148, 163, 184, 0.15)';
                    let badgeColor = 'var(--text-secondary)';

                    if (status === 'CAPTURED' || status === 'RECOVERED' || status === 'SUCCESS') {
                      badgeBg = 'rgba(52, 211, 153, 0.15)';
                      badgeColor = '#34D399';
                    } else if (status === 'ESCALATED' || status === 'REQUIRES_HUMAN_REVIEW') {
                      badgeBg = 'rgba(245, 158, 11, 0.15)';
                      badgeColor = '#F59E0B';
                    } else if (status === 'ANALYZING' || status === 'SCHEDULED' || status === 'ACTIVE') {
                      badgeBg = 'rgba(56, 189, 248, 0.15)';
                      badgeColor = '#38BDF8';
                    } else if (status === 'FAILED' || status === 'HALTED') {
                      badgeBg = 'rgba(248, 113, 113, 0.15)';
                      badgeColor = '#F87171';
                    }

                    return (
                      <tr key={c.id} style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                        <td style={{ padding: '12px', fontWeight: 500, color: 'var(--text-primary)' }}>{c.email}</td>
                        <td style={{ padding: '12px' }}>
                          {c.communication_opt_out ? (
                            <span style={{ color: '#F87171', fontSize: '11px', fontWeight: 600 }}>OPTED OUT</span>
                          ) : (
                            <span style={{ color: '#34D399', fontSize: '11px', fontWeight: 600 }}>ALLOWED</span>
                          )}
                        </td>
                        <td style={{ padding: '12px', color: '#F87171', fontWeight: 600 }}>{c.failed_count ?? 0}</td>
                        <td style={{ padding: '12px', color: '#34D399', fontWeight: 600 }}>{c.recovered_count ?? 0}</td>
                        <td style={{ padding: '12px', fontFamily: 'monospace', color: 'var(--text-secondary)' }}>{c.last_action || 'N/A'}</td>
                        <td style={{ padding: '12px' }}>
                          <span style={{
                            padding: '3px 8px',
                            borderRadius: '4px',
                            fontSize: '11px',
                            fontWeight: 600,
                            background: badgeBg,
                            color: badgeColor
                          }}>
                            {status}
                          </span>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

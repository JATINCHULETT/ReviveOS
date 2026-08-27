'use client';

import React, { useEffect, useState } from 'react';
import { 
  Shield, 
  Building2, 
  Plus, 
  RefreshCw, 
  Send, 
  ExternalLink, 
  Users, 
  CheckCircle2, 
  Sliders,
  CreditCard,
  Layers
} from 'lucide-react';
import { getMerchants, createMerchant, createSandboxPaymentLink, createSubscription } from '@/lib/api';

export default function AdminHubPage() {
  const [merchants, setMerchants] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  // New Merchant Form
  const [newMerchantName, setNewMerchantName] = useState('');
  const [newMaxRetries, setNewMaxRetries] = useState('3');
  const [newConfThresh, setNewConfThresh] = useState('0.75');
  const [newAmountThresh, setNewAmountThresh] = useState('50000');
  const [creatingMerchant, setCreatingMerchant] = useState(false);

  // Push Link / Subscription Form
  const [selectedMerchantId, setSelectedMerchantId] = useState('');
  const [pushType, setPushType] = useState<'link' | 'subscription'>('link');
  const [customerEmail, setCustomerEmail] = useState('');
  const [amount, setAmount] = useState('1000.00');
  const [planId, setPlanId] = useState('plan_pro_tier');
  const [pushing, setPushing] = useState(false);
  const [pushResult, setPushResult] = useState<any>(null);

  const fetchMerchantsList = async () => {
    try {
      setLoading(true);
      const res = await getMerchants();
      setMerchants(res || []);
      if (res && res.length > 0 && !selectedMerchantId) {
        setSelectedMerchantId(res[0].id);
      }
    } catch (err) {
      console.error('Failed to fetch merchants:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchMerchantsList();
  }, []);

  const handleCreateMerchant = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newMerchantName.trim()) return;
    setCreatingMerchant(true);
    try {
      await createMerchant({
        name: newMerchantName.trim(),
        max_retries: parseInt(newMaxRetries, 10) || 3,
        confidence_threshold: parseFloat(newConfThresh) || 0.75,
        amount_threshold: parseFloat(newAmountThresh) || 50000,
      });
      setNewMerchantName('');
      await fetchMerchantsList();
      alert('Merchant created successfully!');
    } catch (err: any) {
      alert(`Failed to create merchant: ${err.message}`);
    } finally {
      setCreatingMerchant(false);
    }
  };

  const handlePushAction = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedMerchantId) {
      alert('Please select a merchant first');
      return;
    }
    setPushing(true);
    setPushResult(null);

    try {
      if (pushType === 'link') {
        const res = await createSandboxPaymentLink({
          merchant_id: selectedMerchantId,
          customer_email: customerEmail || `customer_${Date.now()}@example.com`,
          amount: parseFloat(amount) || 1000,
          trigger_failure_immediately: true,
          failure_code: 'INSUFFICIENT_FUNDS',
        });
        setPushResult(res);
      } else {
        const res = await createSubscription({
          merchant_id: selectedMerchantId,
          customer_email: customerEmail || `subscriber_${Date.now()}@example.com`,
          amount: parseFloat(amount) || 1499,
          plan_id: planId,
          billing_interval: 'monthly',
        });
        setPushResult(res);
      }
      await fetchMerchantsList();
    } catch (err: any) {
      alert(`Error pushing action: ${err.message}`);
    } finally {
      setPushing(false);
    }
  };

  return (
    <div style={{ padding: '24px 32px' }}>
      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '28px' }}>
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
            <Shield size={24} color="#3B82F6" />
            <h1 style={{ fontSize: '24px', fontWeight: 700, margin: 0, color: 'var(--text-primary)' }}>
              Admin Management Hub
            </h1>
          </div>
          <p style={{ fontSize: '13px', color: 'var(--text-muted)', marginTop: '4px' }}>
            Manage registered merchants, configure recovery policy thresholds, and manually push payment links and recurring subscriptions.
          </p>
        </div>

        <button 
          onClick={fetchMerchantsList}
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

      {/* Grid: Merchants Table & Add Merchant */}
      <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: '24px', marginBottom: '32px' }}>
        {/* Merchants Overview Table */}
        <div style={{ background: 'var(--bg-card)', border: '1px solid var(--border-subtle)', borderRadius: '12px', padding: '20px' }}>
          <h3 style={{ fontSize: '16px', fontWeight: 600, margin: '0 0 16px', color: 'var(--text-primary)', display: 'flex', alignItems: 'center', gap: '8px' }}>
            <Building2 size={18} color="#3B82F6" />
            <span>Registered Merchants ({merchants.length})</span>
          </h3>

          {merchants.length === 0 ? (
            <div style={{ padding: '32px', textAlign: 'center', color: 'var(--text-muted)' }}>
              No merchants registered yet. Register one on the right.
            </div>
          ) : (
            <div style={{ overflowX: 'auto' }}>
              <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '13px' }}>
                <thead>
                  <tr style={{ borderBottom: '1px solid var(--border-subtle)', color: 'var(--text-muted)', textAlign: 'left' }}>
                    <th style={{ padding: '10px 12px' }}>Merchant Name</th>
                    <th style={{ padding: '10px 12px' }}>Customers</th>
                    <th style={{ padding: '10px 12px' }}>Subscriptions</th>
                    <th style={{ padding: '10px 12px' }}>Failed</th>
                    <th style={{ padding: '10px 12px' }}>Recovered</th>
                    <th style={{ padding: '10px 12px' }}>Recovery %</th>
                  </tr>
                </thead>
                <tbody>
                  {merchants.map((m) => (
                    <tr key={m.id} style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                      <td style={{ padding: '12px', fontWeight: 600, color: 'var(--text-primary)' }}>
                        <div>{m.name}</div>
                        <div style={{ fontSize: '10px', color: 'var(--text-muted)', fontFamily: 'monospace' }}>{m.id}</div>
                      </td>
                      <td style={{ padding: '12px', color: 'var(--text-secondary)' }}>{m.total_customers}</td>
                      <td style={{ padding: '12px', color: '#10B981', fontWeight: 600 }}>{m.active_subscriptions}</td>
                      <td style={{ padding: '12px', color: '#F87171', fontWeight: 600 }}>{m.failed_payments}</td>
                      <td style={{ padding: '12px', color: '#34D399', fontWeight: 600 }}>{m.recovered_payments}</td>
                      <td style={{ padding: '12px', fontWeight: 700, color: '#38BDF8' }}>
                        {m.recovery_rate ? m.recovery_rate.toFixed(1) : '0.0'}%
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>

        {/* Add Merchant Form */}
        <div style={{ background: 'var(--bg-card)', border: '1px solid var(--border-subtle)', borderRadius: '12px', padding: '20px' }}>
          <h3 style={{ fontSize: '16px', fontWeight: 600, margin: '0 0 16px', color: 'var(--text-primary)', display: 'flex', alignItems: 'center', gap: '8px' }}>
            <Plus size={18} color="#10B981" />
            <span>Register New Merchant</span>
          </h3>

          <form onSubmit={handleCreateMerchant}>
            <div style={{ marginBottom: '14px' }}>
              <label style={{ display: 'block', fontSize: '12px', color: 'var(--text-secondary)', marginBottom: '4px' }}>Merchant Business Name</label>
              <input
                type="text"
                required
                placeholder="e.g. Apex Gaming Inc."
                value={newMerchantName}
                onChange={(e) => setNewMerchantName(e.target.value)}
                style={{ width: '100%', padding: '8px 12px', borderRadius: '6px', background: 'var(--bg-input)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
              />
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px', marginBottom: '14px' }}>
              <div>
                <label style={{ display: 'block', fontSize: '11px', color: 'var(--text-secondary)', marginBottom: '4px' }}>Max Retries</label>
                <input
                  type="number"
                  value={newMaxRetries}
                  onChange={(e) => setNewMaxRetries(e.target.value)}
                  style={{ width: '100%', padding: '8px 10px', borderRadius: '6px', background: 'var(--bg-input)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                />
              </div>
              <div>
                <label style={{ display: 'block', fontSize: '11px', color: 'var(--text-secondary)', marginBottom: '4px' }}>AI Conf. Threshold</label>
                <input
                  type="number"
                  step="0.05"
                  value={newConfThresh}
                  onChange={(e) => setNewConfThresh(e.target.value)}
                  style={{ width: '100%', padding: '8px 10px', borderRadius: '6px', background: 'var(--bg-input)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                />
              </div>
            </div>

            <div style={{ marginBottom: '16px' }}>
              <label style={{ display: 'block', fontSize: '11px', color: 'var(--text-secondary)', marginBottom: '4px' }}>High-Value Threshold (INR)</label>
              <input
                type="number"
                value={newAmountThresh}
                onChange={(e) => setNewAmountThresh(e.target.value)}
                style={{ width: '100%', padding: '8px 12px', borderRadius: '6px', background: 'var(--bg-input)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
              />
            </div>

            <button
              type="submit"
              disabled={creatingMerchant}
              style={{
                width: '100%',
                padding: '10px',
                borderRadius: '6px',
                background: 'linear-gradient(135deg, #10B981, #059669)',
                color: '#FFF',
                border: 'none',
                fontWeight: 600,
                fontSize: '13px',
                cursor: creatingMerchant ? 'not-allowed' : 'pointer'
              }}
            >
              {creatingMerchant ? 'Registering...' : 'Register Merchant'}
            </button>
          </form>
        </div>
      </div>

      {/* Push Payment Links & Subscriptions Tool */}
      <div style={{ background: 'var(--bg-card)', border: '1px solid var(--border-subtle)', borderRadius: '12px', padding: '24px' }}>
        <h3 style={{ fontSize: '18px', fontWeight: 600, margin: '0 0 6px', color: 'var(--text-primary)', display: 'flex', alignItems: 'center', gap: '8px' }}>
          <Send size={18} color="#38BDF8" />
          <span>Manual Push: Payment Links & Subscriptions</span>
        </h3>
        <p style={{ fontSize: '13px', color: 'var(--text-muted)', marginBottom: '20px' }}>
          Admin tool to push Razorpay payment links or recurring subscription authorizations directly to a merchant's customer profile.
        </p>

        <form onSubmit={handlePushAction} style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr) auto', gap: '14px', alignItems: 'flex-end' }}>
          <div>
            <label style={{ display: 'block', fontSize: '12px', color: 'var(--text-secondary)', marginBottom: '4px' }}>Target Merchant</label>
            <select
              value={selectedMerchantId}
              onChange={(e) => setSelectedMerchantId(e.target.value)}
              style={{ width: '100%', padding: '9px 10px', borderRadius: '6px', background: 'var(--bg-input)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
            >
              {merchants.map((m) => (
                <option key={m.id} value={m.id}>{m.name}</option>
              ))}
            </select>
          </div>

          <div>
            <label style={{ display: 'block', fontSize: '12px', color: 'var(--text-secondary)', marginBottom: '4px' }}>Action Type</label>
            <select
              value={pushType}
              onChange={(e) => setPushType(e.target.value as any)}
              style={{ width: '100%', padding: '9px 10px', borderRadius: '6px', background: 'var(--bg-input)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
            >
              <option value="link">Payment Link (One-time Recovery)</option>
              <option value="subscription">Recurring Subscription</option>
            </select>
          </div>

          <div>
            <label style={{ display: 'block', fontSize: '12px', color: 'var(--text-secondary)', marginBottom: '4px' }}>Customer Email</label>
            <input
              type="email"
              placeholder="customer@example.com"
              value={customerEmail}
              onChange={(e) => setCustomerEmail(e.target.value)}
              style={{ width: '100%', padding: '9px 10px', borderRadius: '6px', background: 'var(--bg-input)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
            />
          </div>

          <div>
            <label style={{ display: 'block', fontSize: '12px', color: 'var(--text-secondary)', marginBottom: '4px' }}>Amount (INR)</label>
            <input
              type="number"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              style={{ width: '100%', padding: '9px 10px', borderRadius: '6px', background: 'var(--bg-input)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
            />
          </div>

          <button
            type="submit"
            disabled={pushing}
            style={{
              padding: '10px 20px',
              borderRadius: '6px',
              background: 'linear-gradient(135deg, #2563EB, #1D4ED8)',
              color: '#FFF',
              border: 'none',
              fontWeight: 600,
              fontSize: '13px',
              cursor: pushing ? 'not-allowed' : 'pointer',
              whiteSpace: 'nowrap'
            }}
          >
            {pushing ? 'Pushing...' : 'Dispatch'}
          </button>
        </form>

        {pushResult && (
          <div style={{ marginTop: '20px', background: 'var(--bg-input)', border: '1px solid var(--border-subtle)', borderRadius: '8px', padding: '14px', fontSize: '13px' }}>
            <div style={{ color: '#10B981', fontWeight: 600, display: 'flex', alignItems: 'center', gap: '6px', marginBottom: '8px' }}>
              <CheckCircle2 size={16} />
              <span>Dispatched Successfully</span>
            </div>
            <div style={{ color: 'var(--text-secondary)' }}>
              <strong>Generated Link:</strong>{' '}
              <a href={pushResult.payment_link_url} target="_blank" rel="noreferrer" style={{ color: '#38BDF8' }}>
                {pushResult.payment_link_url}
              </a>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

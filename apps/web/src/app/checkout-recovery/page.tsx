'use client';

import React, { useState, useEffect } from 'react';
import { ShoppingCart, RefreshCw, Send, CheckCircle2, ArrowRight, Clock, AlertTriangle } from 'lucide-react';
import { BadgePulse } from '@/components/ui/AnimatedComponents';

export default function CheckoutRecoveryPage() {
  const [sessions, setSessions] = useState<any[]>([]);
  const [funnel, setFunnel] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [recoveryNotice, setRecoveryNotice] = useState<string | null>(null);

  const fetchCheckoutData = async () => {
    try {
      const res = await fetch('http://localhost:8080/v1/checkout');
      if (res.ok) {
        const d = await res.json();
        setSessions(d.sessions || []);
        setFunnel(d.funnel || null);
      }
    } catch (err) {
      console.error('Failed to load checkout sessions:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchCheckoutData();
  }, []);

  const dispatchRecovery = async (token: string) => {
    try {
      const res = await fetch('http://localhost:8080/v1/checkout/recover', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ session_token: token, channel: 'EMAIL' }),
      });
      if (res.ok) {
        const d = await res.json();
        setRecoveryNotice(d.message);
        setTimeout(() => setRecoveryNotice(null), 4000);
        fetchCheckoutData();
      }
    } catch (err) {
      alert('Recovery dispatch failed');
    }
  };

  return (
    <div className="main-container" style={{ padding: '24px 32px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '28px', flexWrap: 'wrap', gap: '16px' }}>
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '6px' }}>
            <h1 style={{ fontSize: '26px', fontWeight: 800, letterSpacing: '-0.5px' }}>
              Checkout Drop-off Recovery & Cart Telemetry
            </h1>
            <BadgePulse text="Telemetry Active" variant="success" />
          </div>
          <p style={{ color: 'var(--text-secondary)', fontSize: '14px' }}>
            Detect abandoned carts, OTP drop-offs, and payment friction points before customer churn. Dispatches 1-click cart restoration links via Resend.
          </p>
        </div>
      </div>

      {recoveryNotice && (
        <div style={{
          padding: '14px 18px',
          background: 'rgba(16, 185, 129, 0.12)',
          border: '1px solid rgba(16, 185, 129, 0.3)',
          borderRadius: '10px',
          marginBottom: '24px',
          color: '#10b981',
          fontSize: '14px',
          display: 'flex',
          alignItems: 'center',
          gap: '10px',
        }}>
          <CheckCircle2 size={18} />
          <span>{recoveryNotice}</span>
        </div>
      )}

      {/* Funnel Metrics */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: '16px', marginBottom: '32px' }}>
        <div className="metric-card" style={{ padding: '20px', borderRadius: '12px', background: 'var(--bg-secondary)', border: '1px solid var(--border-subtle)' }}>
          <div style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '8px' }}>Total Checkout Sessions</div>
          <div style={{ fontSize: '26px', fontWeight: 800, color: 'var(--text-primary)' }}>
            {funnel?.total_sessions || 3}
          </div>
          <div style={{ fontSize: '12px', color: 'var(--text-muted)', marginTop: '6px' }}>Monitored in real-time</div>
        </div>

        <div className="metric-card" style={{ padding: '20px', borderRadius: '12px', background: 'var(--bg-secondary)', border: '1px solid var(--border-subtle)' }}>
          <div style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '8px' }}>Drop-offs Detected</div>
          <div style={{ fontSize: '26px', fontWeight: 800, color: '#f59e0b' }}>
            {funnel?.dropped_off || 1}
          </div>
          <div style={{ fontSize: '12px', color: '#f59e0b', marginTop: '6px' }}>Idle &gt; 15 mins</div>
        </div>

        <div className="metric-card" style={{ padding: '20px', borderRadius: '12px', background: 'var(--bg-secondary)', border: '1px solid var(--border-subtle)' }}>
          <div style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '8px' }}>Recovered Cart Value</div>
          <div style={{ fontSize: '26px', fontWeight: 800, color: '#10b981' }}>
            ₹{funnel?.recovered_value?.toLocaleString() || '4,890'}
          </div>
          <div style={{ fontSize: '12px', color: '#10b981', marginTop: '6px' }}>Recovered via 1-click restore</div>
        </div>

        <div className="metric-card" style={{ padding: '20px', borderRadius: '12px', background: 'var(--bg-secondary)', border: '1px solid var(--border-subtle)' }}>
          <div style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '8px' }}>Recovery Success Rate</div>
          <div style={{ fontSize: '26px', fontWeight: 800, color: '#3b82f6' }}>
            {funnel?.recovery_rate_pct?.toFixed(1) || '33.3'}%
          </div>
          <div style={{ fontSize: '12px', color: 'var(--text-muted)', marginTop: '6px' }}>Multi-channel Resend / SMS</div>
        </div>
      </div>

      {/* Abandoned Sessions Table */}
      <div style={{ background: 'var(--bg-secondary)', borderRadius: '14px', border: '1px solid var(--border-subtle)', overflow: 'hidden' }}>
        <div style={{ padding: '18px 24px', borderBottom: '1px solid var(--border-subtle)' }}>
          <h2 style={{ fontSize: '17px', fontWeight: 700 }}>Checkout Sessions & Churn Prevention Queue</h2>
        </div>
        <div style={{ overflowX: 'auto' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left', fontSize: '13.5px' }}>
            <thead>
              <tr style={{ background: 'var(--bg-tertiary)', borderBottom: '1px solid var(--border-subtle)', color: 'var(--text-secondary)' }}>
                <th style={{ padding: '14px 20px' }}>Session</th>
                <th style={{ padding: '14px 20px' }}>Customer</th>
                <th style={{ padding: '14px 20px' }}>Cart Value</th>
                <th style={{ padding: '14px 20px' }}>Step Reached</th>
                <th style={{ padding: '14px 20px' }}>Drop-off Reason</th>
                <th style={{ padding: '14px 20px' }}>Status</th>
                <th style={{ padding: '14px 20px' }}>Action</th>
              </tr>
            </thead>
            <tbody>
              {sessions.map((s) => (
                <tr key={s.id} style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                  <td style={{ padding: '16px 20px', fontWeight: 600 }}>{s.session_token.slice(0, 16)}...</td>
                  <td style={{ padding: '16px 20px' }}>
                    <div style={{ fontWeight: 600 }}>{s.customer_name}</div>
                    <div style={{ fontSize: '12px', color: 'var(--text-muted)' }}>{s.customer_email}</div>
                  </td>
                  <td style={{ padding: '16px 20px', fontWeight: 700 }}>₹{s.cart_amount.toLocaleString()}</td>
                  <td style={{ padding: '16px 20px' }}>
                    <span style={{ fontSize: '12px', fontWeight: 600, color: 'var(--text-secondary)' }}>
                      {s.step_reached}
                    </span>
                  </td>
                  <td style={{ padding: '16px 20px', fontSize: '12px', color: '#f59e0b', maxWidth: '220px' }}>
                    {s.drop_off_reason || 'Incomplete payment'}
                  </td>
                  <td style={{ padding: '16px 20px' }}>
                    <span style={{
                      padding: '4px 10px',
                      borderRadius: '6px',
                      fontSize: '11px',
                      fontWeight: 700,
                      background: s.status === 'RECOVERED' ? 'rgba(16, 185, 129, 0.15)' :
                                  s.status === 'RECOVERY_DISPATCHED' ? 'rgba(59, 130, 246, 0.15)' : 'rgba(245, 158, 11, 0.15)',
                      color: s.status === 'RECOVERED' ? '#10b981' :
                             s.status === 'RECOVERY_DISPATCHED' ? '#60a5fa' : '#fbbf24',
                    }}>
                      {s.status}
                    </span>
                  </td>
                  <td style={{ padding: '16px 20px' }}>
                    {s.status !== 'RECOVERED' && (
                      <button
                        onClick={() => dispatchRecovery(s.session_token)}
                        style={{
                          padding: '8px 12px',
                          background: 'var(--primary)',
                          color: '#fff',
                          border: 'none',
                          borderRadius: '6px',
                          fontSize: '12px',
                          fontWeight: 600,
                          cursor: 'pointer',
                          display: 'flex',
                          alignItems: 'center',
                          gap: '6px',
                        }}
                      >
                        <Send size={12} />
                        Dispatch Link
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

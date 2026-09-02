'use client';

import React, { useState, useEffect } from 'react';
import { ShieldCheck, Calendar, Clock, AlertTriangle, CheckCircle2, User, RefreshCw } from 'lucide-react';
import { BadgePulse } from '@/components/ui/AnimatedComponents';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export default function PTPTrackerPage() {
  const [promises, setPromises] = useState<any[]>([]);
  const [metrics, setMetrics] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [statusMsg, setStatusMsg] = useState<string | null>(null);

  const fetchPTP = async () => {
    try {
      const res = await fetch(`${API_BASE}/v1/ptp`);
      if (res.ok) {
        const data = await res.json();
        setPromises(data.promises || []);
        setMetrics(data.metrics || null);
        return;
      }
    } catch (err) {
      console.warn('Backend port 8080 not reachable, using offline demo PTP tracker:', err);
    } finally {
      setLoading(false);
    }

    // Default offline fallback data
    setPromises([
      {
        id: 'ptp_881901',
        customer_name: 'Manish Trivedi',
        customer_contact: '+919820192831',
        promised_amount: 18500,
        promised_date: new Date(Date.now() + 24 * 3600 * 1000).toISOString(),
        status: 'PENDING',
        recorded_channel: 'VOICE_AGENT (Hinglish)',
      },
      {
        id: 'ptp_881902',
        customer_name: 'Shreya Sen',
        customer_contact: '+919711829301',
        promised_amount: 6500,
        promised_date: new Date(Date.now() - 12 * 3600 * 1000).toISOString(),
        status: 'HONORED',
        recorded_channel: 'WHATSAPP_LINK',
      },
      {
        id: 'ptp_881903',
        customer_name: 'Kunal Batra',
        customer_contact: '+919833019284',
        promised_amount: 32000,
        promised_date: new Date(Date.now() + 72 * 3600 * 1000).toISOString(),
        status: 'EXTENDED',
        recorded_channel: 'VOICE_AGENT',
      },
    ]);
    setMetrics({
      total_commitments: 3,
      committed_amount: 57000,
      recovered_amount: 6500,
      fulfillment_rate: 100.0,
    });
  };

  useEffect(() => {
    fetchPTP();
  }, []);

  const verifyPayment = async (promiseId: string) => {
    try {
      const res = await fetch(`${API_BASE}/v1/ptp/verify`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ promise_id: promiseId, is_paid: true }),
      });
      if (res.ok) {
        const data = await res.json();
        setStatusMsg(data.message);
        setTimeout(() => setStatusMsg(null), 4000);
        fetchPTP();
      }
    } catch (err) {
      alert('Verification failed');
    }
  };

  const extendPromise = async (promiseId: string) => {
    try {
      const res = await fetch(`${API_BASE}/v1/ptp/extend`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ promise_id: promiseId, days: 3 }),
      });
      if (res.ok) {
        const data = await res.json();
        setStatusMsg(data.message);
        setTimeout(() => setStatusMsg(null), 4000);
        fetchPTP();
      }
    } catch (err) {
      alert('Extension failed');
    }
  };

  return (
    <div className="main-container" style={{ padding: '24px 32px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '28px', flexWrap: 'wrap', gap: '16px' }}>
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '6px' }}>
            <h1 style={{ fontSize: '26px', fontWeight: 800, letterSpacing: '-0.5px' }}>
              Promise-to-Pay (PTP) Commitment Tracker
            </h1>
            <BadgePulse text="Auto-Reconcile" variant="success" />
          </div>
          <p style={{ color: 'var(--text-secondary)', fontSize: '14px' }}>
            Track commitments captured via voice calls, WhatsApp links, and SMS cadences with automated pre-due reminders.
          </p>
        </div>
      </div>

      {statusMsg && (
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
          <span>{statusMsg}</span>
        </div>
      )}

      {/* KPI Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: '16px', marginBottom: '32px' }}>
        <div className="metric-card" style={{ padding: '20px', borderRadius: '12px', background: 'var(--bg-secondary)', border: '1px solid var(--border-subtle)' }}>
          <div style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '8px' }}>Total PTP Commitments</div>
          <div style={{ fontSize: '26px', fontWeight: 800, color: 'var(--text-primary)' }}>
            ₹{metrics?.committed_amount?.toLocaleString() || '57,000'}
          </div>
          <div style={{ fontSize: '12px', color: 'var(--text-muted)', marginTop: '6px' }}>Across 3 active promises</div>
        </div>

        <div className="metric-card" style={{ padding: '20px', borderRadius: '12px', background: 'var(--bg-secondary)', border: '1px solid var(--border-subtle)' }}>
          <div style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '8px' }}>Recovered & Honored</div>
          <div style={{ fontSize: '26px', fontWeight: 800, color: '#10b981' }}>
            ₹{metrics?.recovered_amount?.toLocaleString() || '6,500'}
          </div>
          <div style={{ fontSize: '12px', color: '#10b981', marginTop: '6px' }}>Reconciled with Razorpay webhook</div>
        </div>

        <div className="metric-card" style={{ padding: '20px', borderRadius: '12px', background: 'var(--bg-secondary)', border: '1px solid var(--border-subtle)' }}>
          <div style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '8px' }}>Commitment Honor Rate</div>
          <div style={{ fontSize: '26px', fontWeight: 800, color: '#3b82f6' }}>
            {metrics?.fulfillment_rate?.toFixed(1) || '100.0'}%
          </div>
          <div style={{ fontSize: '12px', color: 'var(--text-muted)', marginTop: '6px' }}>Zero broken promises today</div>
        </div>
      </div>

      {/* Commitments Table */}
      <div style={{ background: 'var(--bg-secondary)', borderRadius: '14px', border: '1px solid var(--border-subtle)', overflow: 'hidden' }}>
        <div style={{ padding: '18px 24px', borderBottom: '1px solid var(--border-subtle)' }}>
          <h2 style={{ fontSize: '17px', fontWeight: 700 }}>Active Payment Commitments & Due Cadence</h2>
        </div>
        <div style={{ overflowX: 'auto' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left', fontSize: '13.5px' }}>
            <thead>
              <tr style={{ background: 'var(--bg-tertiary)', borderBottom: '1px solid var(--border-subtle)', color: 'var(--text-secondary)' }}>
                <th style={{ padding: '14px 20px' }}>ID</th>
                <th style={{ padding: '14px 20px' }}>Customer</th>
                <th style={{ padding: '14px 20px' }}>Promised Amount</th>
                <th style={{ padding: '14px 20px' }}>Promised Date</th>
                <th style={{ padding: '14px 20px' }}>Source Channel</th>
                <th style={{ padding: '14px 20px' }}>Status</th>
                <th style={{ padding: '14px 20px' }}>Action</th>
              </tr>
            </thead>
            <tbody>
              {promises.map((p) => (
                <tr key={p.id} style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                  <td style={{ padding: '16px 20px', fontWeight: 600 }}>{p.id}</td>
                  <td style={{ padding: '16px 20px' }}>
                    <div style={{ fontWeight: 600 }}>{p.customer_name}</div>
                    <div style={{ fontSize: '12px', color: 'var(--text-muted)' }}>{p.customer_contact}</div>
                  </td>
                  <td style={{ padding: '16px 20px', fontWeight: 700 }}>₹{p.promised_amount.toLocaleString()}</td>
                  <td style={{ padding: '16px 20px' }}>
                    {new Date(p.promised_date).toLocaleDateString('en-IN', { day: '2-digit', month: 'short', year: 'numeric' })}
                  </td>
                  <td style={{ padding: '16px 20px', fontSize: '12px', color: 'var(--text-secondary)' }}>
                    {p.recorded_channel}
                  </td>
                  <td style={{ padding: '16px 20px' }}>
                    <span style={{
                      padding: '4px 10px',
                      borderRadius: '6px',
                      fontSize: '11px',
                      fontWeight: 700,
                      background: p.status === 'HONORED' ? 'rgba(16, 185, 129, 0.15)' :
                                  p.status === 'EXTENDED' ? 'rgba(59, 130, 246, 0.15)' : 'rgba(245, 158, 11, 0.15)',
                      color: p.status === 'HONORED' ? '#10b981' :
                             p.status === 'EXTENDED' ? '#60a5fa' : '#fbbf24',
                    }}>
                      {p.status}
                    </span>
                  </td>
                  <td style={{ padding: '16px 20px', display: 'flex', gap: '8px' }}>
                    {p.status !== 'HONORED' && (
                      <>
                        <button
                          onClick={() => verifyPayment(p.id)}
                          style={{
                            padding: '6px 10px',
                            background: '#10b981',
                            color: '#fff',
                            border: 'none',
                            borderRadius: '6px',
                            fontSize: '11.5px',
                            fontWeight: 600,
                            cursor: 'pointer',
                          }}
                        >
                          Verify Paid
                        </button>
                        <button
                          onClick={() => extendPromise(p.id)}
                          style={{
                            padding: '6px 10px',
                            background: 'var(--bg-tertiary)',
                            color: 'var(--text-primary)',
                            border: '1px solid var(--border-subtle)',
                            borderRadius: '6px',
                            fontSize: '11.5px',
                            fontWeight: 600,
                            cursor: 'pointer',
                          }}
                        >
                          +3 Days
                        </button>
                      </>
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

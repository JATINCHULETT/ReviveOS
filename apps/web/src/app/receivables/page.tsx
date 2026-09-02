'use client';

import React, { useState, useEffect } from 'react';
import { Building2, AlertCircle, Send, CheckCircle2, Clock, Calendar, ArrowUpRight } from 'lucide-react';
import { BadgePulse } from '@/components/ui/AnimatedComponents';

export default function ReceivablesPage() {
  const [invoices, setInvoices] = useState<any[]>([]);
  const [summary, setSummary] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [dunningNotice, setDunningNotice] = useState<string | null>(null);

  const fetchReceivables = async () => {
    try {
      const res = await fetch('http://localhost:8080/v1/receivables');
      if (res.ok) {
        const data = await res.json();
        setInvoices(data.invoices || []);
        setSummary(data.summary || null);
      }
    } catch (err) {
      console.error('Failed to load receivables:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchReceivables();
  }, []);

  const triggerDunning = async (invId: string) => {
    try {
      const res = await fetch('http://localhost:8080/v1/receivables/dunning/trigger', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ invoice_id: invId }),
      });
      if (res.ok) {
        const data = await res.json();
        setDunningNotice(data.message);
        setTimeout(() => setDunningNotice(null), 5000);
      }
    } catch (err) {
      alert('Failed to trigger dunning sequence');
    }
  };

  return (
    <div className="main-container" style={{ padding: '24px 32px' }}>
      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '28px', flexWrap: 'wrap', gap: '16px' }}>
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '6px' }}>
            <h1 style={{ fontSize: '26px', fontWeight: 800, letterSpacing: '-0.5px' }}>
              B2B Receivables Chaser & Dunning
            </h1>
            <BadgePulse text="Net Terms Monitor" variant="success" />
          </div>
          <p style={{ color: 'var(--text-secondary)', fontSize: '14px' }}>
            Automate AR dunning cadences, aging bucket tracking, and self-serve payment link generation via Resend.
          </p>
        </div>
      </div>

      {dunningNotice && (
        <div style={{
          padding: '14px 18px',
          background: 'rgba(16, 185, 129, 0.12)',
          border: '1px solid rgba(16, 185, 129, 0.3)',
          borderRadius: '10px',
          marginBottom: '24px',
          display: 'flex',
          alignItems: 'center',
          gap: '10px',
          color: '#10b981',
          fontSize: '14px',
        }}>
          <CheckCircle2 size={18} />
          <span>{dunningNotice}</span>
        </div>
      )}

      {/* Aging Metric Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: '16px', marginBottom: '32px' }}>
        <div className="metric-card" style={{ padding: '20px', borderRadius: '12px', background: 'var(--bg-secondary)', border: '1px solid var(--border-subtle)' }}>
          <div style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '8px' }}>Total Outstanding AR</div>
          <div style={{ fontSize: '26px', fontWeight: 800, color: 'var(--text-primary)' }}>
            ₹{summary?.total_outstanding?.toLocaleString() || '7,24,000'}
          </div>
          <div style={{ fontSize: '12px', color: '#10b981', marginTop: '6px' }}>Across 3 Enterprise Buyers</div>
        </div>

        <div className="metric-card" style={{ padding: '20px', borderRadius: '12px', background: 'var(--bg-secondary)', border: '1px solid var(--border-subtle)' }}>
          <div style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '8px' }}>Current (Due &lt;30d)</div>
          <div style={{ fontSize: '26px', fontWeight: 800, color: '#3b82f6' }}>
            ₹{summary?.current_due?.toLocaleString() || '89,000'}
          </div>
          <div style={{ fontSize: '12px', color: 'var(--text-muted)', marginTop: '6px' }}>Courteous reminders</div>
        </div>

        <div className="metric-card" style={{ padding: '20px', borderRadius: '12px', background: 'var(--bg-secondary)', border: '1px solid var(--border-subtle)' }}>
          <div style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '8px' }}>1 - 30 Days Past Due</div>
          <div style={{ fontSize: '26px', fontWeight: 800, color: '#f59e0b' }}>
            ₹{summary?.bucket_1_30?.toLocaleString() || '4,50,000'}
          </div>
          <div style={{ fontSize: '12px', color: '#f59e0b', marginTop: '6px' }}>Automated firm notice</div>
        </div>

        <div className="metric-card" style={{ padding: '20px', borderRadius: '12px', background: 'var(--bg-secondary)', border: '1px solid var(--border-subtle)' }}>
          <div style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '8px' }}>31 - 60 Days Past Due</div>
          <div style={{ fontSize: '26px', fontWeight: 800, color: '#ef4444' }}>
            ₹{summary?.bucket_31_60?.toLocaleString() || '1,85,000'}
          </div>
          <div style={{ fontSize: '12px', color: '#ef4444', marginTop: '6px' }}>Urgent / Voice call escalated</div>
        </div>
      </div>

      {/* Invoice Ledger Table */}
      <div style={{ background: 'var(--bg-secondary)', borderRadius: '14px', border: '1px solid var(--border-subtle)', overflow: 'hidden' }}>
        <div style={{ padding: '18px 24px', borderBottom: '1px solid var(--border-subtle)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <h2 style={{ fontSize: '17px', fontWeight: 700 }}>Overdue Invoices & Automated Dunning Queue</h2>
          <span style={{ fontSize: '13px', color: 'var(--text-muted)' }}>Auto-sync with Razorpay Invoices</span>
        </div>

        <div style={{ overflowX: 'auto' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left', fontSize: '13.5px' }}>
            <thead>
              <tr style={{ background: 'var(--bg-tertiary)', borderBottom: '1px solid var(--border-subtle)', color: 'var(--text-secondary)' }}>
                <th style={{ padding: '14px 20px' }}>Invoice #</th>
                <th style={{ padding: '14px 20px' }}>Buyer Entity</th>
                <th style={{ padding: '14px 20px' }}>Amount</th>
                <th style={{ padding: '14px 20px' }}>Aging Bucket</th>
                <th style={{ padding: '14px 20px' }}>Days Past Due</th>
                <th style={{ padding: '14px 20px' }}>Dunning Action</th>
              </tr>
            </thead>
            <tbody>
              {invoices.map((inv) => (
                <tr key={inv.id} style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                  <td style={{ padding: '16px 20px', fontWeight: 600 }}>{inv.invoice_number}</td>
                  <td style={{ padding: '16px 20px' }}>
                    <div style={{ fontWeight: 600, color: 'var(--text-primary)' }}>{inv.buyer_company}</div>
                    <div style={{ fontSize: '12px', color: 'var(--text-muted)' }}>{inv.buyer_name} ({inv.buyer_email})</div>
                  </td>
                  <td style={{ padding: '16px 20px', fontWeight: 700 }}>₹{inv.amount.toLocaleString()}</td>
                  <td style={{ padding: '16px 20px' }}>
                    <span style={{
                      padding: '4px 10px',
                      borderRadius: '6px',
                      fontSize: '11px',
                      fontWeight: 700,
                      background: inv.current_bucket === 'CURRENT' ? 'rgba(59, 130, 246, 0.15)' :
                                  inv.current_bucket === '1_30' ? 'rgba(245, 158, 11, 0.15)' : 'rgba(239, 68, 68, 0.15)',
                      color: inv.current_bucket === 'CURRENT' ? '#60a5fa' :
                             inv.current_bucket === '1_30' ? '#fbbf24' : '#f87171',
                    }}>
                      {inv.current_bucket}
                    </span>
                  </td>
                  <td style={{ padding: '16px 20px' }}>
                    {inv.days_past_due > 0 ? (
                      <span style={{ color: '#ef4444', fontWeight: 600 }}>{inv.days_past_due} days</span>
                    ) : (
                      <span style={{ color: '#10b981', fontWeight: 600 }}>On Schedule</span>
                    )}
                  </td>
                  <td style={{ padding: '16px 20px' }}>
                    <button
                      onClick={() => triggerDunning(inv.id)}
                      style={{
                        padding: '8px 14px',
                        background: 'var(--primary)',
                        color: '#fff',
                        border: 'none',
                        borderRadius: '6px',
                        fontSize: '12.5px',
                        fontWeight: 600,
                        cursor: 'pointer',
                        display: 'flex',
                        alignItems: 'center',
                        gap: '6px',
                      }}
                    >
                      <Send size={13} />
                      Chase Now
                    </button>
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

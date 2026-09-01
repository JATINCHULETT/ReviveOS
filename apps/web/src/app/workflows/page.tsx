'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { getWorkflows, getInterventions, approveWorkflow, rejectWorkflow } from '@/lib/api';
import { getSyntheticWorkflows } from '@/lib/syntheticDataset';
import { WorkflowSummary } from '@/lib/types';
import { StatusBadge } from '@/components/ui/StatusBadge';
import { LoadingView, EmptyView, ErrorView } from '@/components/ui/StateViews';
import { RefreshCw, Search, Filter, ShieldAlert, CheckCircle, XCircle, UserCheck, AlertTriangle } from 'lucide-react';

export default function WorkflowsPage() {
  const [workflows, setWorkflows] = useState<WorkflowSummary[]>(() => getSyntheticWorkflows().slice(0, 50));
  const [interventions, setInterventions] = useState<any[]>([]);
  const [activeTab, setActiveTab] = useState<'ALL' | 'INTERVENTIONS'>('ALL');
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [statusFilter, setStatusFilter] = useState<string>('ALL');
  const [searchQuery, setSearchQuery] = useState<string>('');
  const [actionLoadingId, setActionLoadingId] = useState<string | null>(null);

  const loadData = async () => {
    try {
      setLoading(true);
      setError(null);
      const [wfRes, intRes] = await Promise.all([
        getWorkflows({
          status: statusFilter !== 'ALL' ? statusFilter : undefined,
          limit: 50,
        }),
        getInterventions(),
      ]);
      setWorkflows(wfRes.data || wfRes.workflows || []);
      setInterventions(intRes.data || []);
    } catch (err: any) {
      setError(err.message || 'Failed to fetch workflows');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [statusFilter]);

  const handleInlineApprove = async (wfId: string) => {
    try {
      setActionLoadingId(wfId);
      await approveWorkflow(wfId, { notes: 'Approved via Human Intervention Queue' });
      await loadData();
    } catch (err: any) {
      alert(`Approval error: ${err.message}`);
    } finally {
      setActionLoadingId(null);
    }
  };

  const handleInlineReject = async (wfId: string) => {
    const reason = prompt('Enter reason for rejecting this recovery attempt:') || 'Rejected by operator';
    try {
      setActionLoadingId(wfId);
      await rejectWorkflow(wfId, { reason, notes: 'Rejected via Human Intervention Queue' });
      await loadData();
    } catch (err: any) {
      alert(`Rejection error: ${err.message}`);
    } finally {
      setActionLoadingId(null);
    }
  };

  const filteredWorkflows = workflows.filter((wf) => {
    if (!searchQuery) return true;
    const q = searchQuery.toLowerCase();
    return (
      wf.payment_id.toLowerCase().includes(q) ||
      (wf.customer_email && wf.customer_email.toLowerCase().includes(q)) ||
      wf.failure_code.toLowerCase().includes(q) ||
      (wf.selected_action && wf.selected_action.toLowerCase().includes(q))
    );
  });

  const filteredInterventions = interventions.filter((item) => {
    if (!searchQuery) return true;
    const q = searchQuery.toLowerCase();
    return (
      item.payment_id.toLowerCase().includes(q) ||
      (item.customer_email && item.customer_email.toLowerCase().includes(q)) ||
      item.failure_code.toLowerCase().includes(q) ||
      (item.escalation_reason && item.escalation_reason.toLowerCase().includes(q))
    );
  });

  return (
    <div className="page-container">
      {/* Header & Controls */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '20px' }}>
        <div>
          <h1 style={{ fontSize: '22px', fontWeight: 800, color: 'var(--text-primary)' }}>Recovery Orchestration</h1>
          <p style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
            Monitor automated lifecycle executions and resolve high-risk escalated workflows
          </p>
        </div>
        <button onClick={loadData} className="btn-secondary" style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
          <RefreshCw size={14} className={loading ? 'spinning' : ''} />
          <span>Refresh</span>
        </button>
      </div>

      {/* Tabs Header */}
      <div style={{ display: 'flex', gap: '12px', marginBottom: '20px', borderBottom: '1px solid var(--border-subtle)', paddingBottom: '12px' }}>
        <button
          onClick={() => setActiveTab('ALL')}
          style={{
            background: activeTab === 'ALL' ? 'var(--bg-elevated)' : 'transparent',
            border: activeTab === 'ALL' ? '1px solid var(--color-accent)' : '1px solid transparent',
            color: activeTab === 'ALL' ? 'var(--text-primary)' : 'var(--text-secondary)',
            padding: '8px 16px',
            borderRadius: '8px',
            fontSize: '13px',
            fontWeight: 700,
            cursor: 'pointer',
          }}
        >
          All Workflows ({workflows.length})
        </button>
        <button
          onClick={() => setActiveTab('INTERVENTIONS')}
          style={{
            background: activeTab === 'INTERVENTIONS' ? '#78350f' : 'rgba(245, 158, 11, 0.1)',
            border: activeTab === 'INTERVENTIONS' ? '1px solid #f59e0b' : '1px solid rgba(245, 158, 11, 0.3)',
            color: activeTab === 'INTERVENTIONS' ? '#fef3c7' : '#f59e0b',
            padding: '8px 16px',
            borderRadius: '8px',
            fontSize: '13px',
            fontWeight: 700,
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
          }}
        >
          <ShieldAlert size={16} />
          <span>🚨 Needs Human Review</span>
          <span style={{ background: '#f59e0b', color: '#000', borderRadius: '12px', padding: '2px 8px', fontSize: '11px', fontWeight: 800 }}>
            {interventions.length}
          </span>
        </button>
      </div>

      {/* Filter / Search Bar */}
      <div style={{ display: 'flex', gap: '16px', marginBottom: '20px', alignItems: 'center' }}>
        {/* Search */}
        <div style={{ position: 'relative', flex: 1 }}>
          <Search size={16} color="#64748b" style={{ position: 'absolute', left: '12px', top: '10px' }} />
          <input
            type="text"
            placeholder="Search by Payment ID, customer email, failure code, or reason..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            style={{
              width: '100%',
              background: 'var(--bg-surface)',
              border: '1px solid var(--border-subtle)',
              padding: '8px 12px 8px 36px',
              borderRadius: '6px',
              color: 'var(--text-primary)',
              fontSize: '13px',
              outline: 'none',
            }}
          />
        </div>

        {/* Status Filter (Active only in ALL tab) */}
        {activeTab === 'ALL' && (
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <Filter size={16} color="#64748b" />
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              style={{
                background: 'var(--bg-surface)',
                border: '1px solid var(--border-subtle)',
                padding: '8px 12px',
                borderRadius: '6px',
                color: 'var(--text-primary)',
                fontSize: '13px',
                outline: 'none',
                cursor: 'pointer',
              }}
            >
              <option value="ALL">All Statuses</option>
              <option value="ANALYZING">ANALYZING</option>
              <option value="SCHEDULED">SCHEDULED</option>
              <option value="ESCALATED">ESCALATED</option>
              <option value="EXECUTING">EXECUTING</option>
              <option value="VERIFYING">VERIFYING</option>
              <option value="RECOVERED">RECOVERED</option>
              <option value="FAILED">FAILED</option>
              <option value="HALTED">HALTED</option>
            </select>
          </div>
        )}
      </div>

      {/* Workflows / Interventions Table */}
      <div className="table-container">
        {loading && (activeTab === 'ALL' ? workflows.length === 0 : interventions.length === 0) ? (
          <div style={{ padding: '40px' }}>
            <LoadingView message="Loading workflows from database..." />
          </div>
        ) : error ? (
          <div style={{ padding: '40px' }}>
            <ErrorView message={error} onRetry={loadData} />
          </div>
        ) : activeTab === 'INTERVENTIONS' ? (
          filteredInterventions.length > 0 ? (
            <table className="data-table">
              <thead>
                <tr>
                  <th>Payment / Link</th>
                  <th>Customer Profile & History</th>
                  <th>Amount</th>
                  <th>Failure Reason</th>
                  <th>Escalation Reason</th>
                  <th>AI Recommended Action</th>
                  <th>Human Decision Action</th>
                </tr>
              </thead>
              <tbody>
                {filteredInterventions.map((item) => (
                  <tr key={item.id}>
                    <td>
                      <Link href={`/workflows/${item.id}`} style={{ display: 'block' }}>
                        <div className="mono" style={{ color: '#60a5fa', fontWeight: 700, fontSize: '13px' }}>
                          {item.payment_id ? `${item.payment_id.substring(0, 18)}...` : item.id.substring(0, 8)}
                        </div>
                        <div style={{ fontSize: '11px', color: 'var(--text-muted)', marginTop: '2px' }}>
                          Created: {new Date(item.created_at).toLocaleString()}
                        </div>
                      </Link>
                    </td>
                    <td>
                      <div style={{ fontWeight: 600, color: 'var(--text-primary)', fontSize: '13px' }}>
                        {item.customer_email || '—'}
                      </div>
                      <div style={{ display: 'flex', gap: '8px', fontSize: '11px', marginTop: '4px' }}>
                        <span style={{ color: '#10b981' }}>✓ {item.customer_success_count || 0} paid</span>
                        <span style={{ color: '#ef4444' }}>✗ {item.customer_failed_count || 0} failed</span>
                        {item.communication_opt_out && (
                          <span style={{ color: '#f59e0b', fontWeight: 700 }}>⚠️ Opted Out</span>
                        )}
                      </div>
                    </td>
                    <td>
                      <span className="mono" style={{ fontWeight: 800, fontSize: '14px' }}>
                        ₹{item.amount.toLocaleString('en-IN', { minimumFractionDigits: 2 })}
                      </span>
                      <div style={{ fontSize: '11px', color: 'var(--text-muted)' }}>{item.currency}</div>
                      <div style={{ marginTop: '4px' }}>
                        <span className={`badge ${(item.fraud_probability ?? 0) >= 0.7 ? 'badge-danger' : (item.fraud_probability ?? 0) >= 0.35 ? 'badge-warning' : 'badge-success'}`} style={{ fontSize: '10px' }}>
                          🛡️ {((item.fraud_probability ?? 0.08) * 100).toFixed(0)}% Fraud
                        </span>
                      </div>
                    </td>
                    <td>
                      <span className="badge badge-danger" style={{ fontSize: '11px' }}>
                        {item.failure_code || 'UNKNOWN'}
                      </span>
                    </td>
                    <td>
                      <div style={{ background: 'rgba(245, 158, 11, 0.1)', border: '1px solid rgba(245, 158, 11, 0.3)', padding: '6px 10px', borderRadius: '6px', fontSize: '12px', color: '#fef3c7' }}>
                        <AlertTriangle size={12} style={{ display: 'inline', marginRight: '4px', verticalAlign: 'text-top', color: '#f59e0b' }} />
                        {item.escalation_reason || 'Policy threshold exceeded'}
                      </div>
                    </td>
                    <td>
                      <div style={{ fontWeight: 700, color: 'var(--color-accent)', fontSize: '13px' }}>
                        {item.selected_action || 'PAYMENT_LINK'}
                      </div>
                      <div style={{ fontSize: '11px', color: 'var(--text-muted)', marginTop: '2px' }}>
                        Confidence: {((item.latest_confidence || item.recovery_probability || 0.75) * 100).toFixed(0)}%
                      </div>
                    </td>
                    <td>
                      <div style={{ display: 'flex', gap: '6px', alignItems: 'center' }}>
                        <button
                          onClick={() => handleInlineApprove(item.id)}
                          disabled={actionLoadingId === item.id}
                          className="btn-primary"
                          style={{ padding: '6px 12px', fontSize: '12px', display: 'flex', alignItems: 'center', gap: '4px' }}
                        >
                          <CheckCircle size={14} />
                          <span>{actionLoadingId === item.id ? 'Executing...' : 'Approve'}</span>
                        </button>
                        <button
                          onClick={() => handleInlineReject(item.id)}
                          disabled={actionLoadingId === item.id}
                          className="btn-secondary"
                          style={{ padding: '6px 10px', fontSize: '12px', color: '#ef4444', borderColor: 'rgba(239, 68, 68, 0.3)' }}
                        >
                          <XCircle size={14} />
                          <span>Reject</span>
                        </button>
                        <Link href={`/workflows/${item.id}`} className="btn-secondary" style={{ padding: '6px 8px', fontSize: '12px' }}>
                          Review
                        </Link>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          ) : (
            <div style={{ padding: '40px', textAlign: 'center' }}>
              <CheckCircle size={40} color="#10b981" style={{ margin: '0 auto 12px' }} />
              <h3 style={{ fontSize: '16px', fontWeight: 700, color: 'var(--text-primary)', margin: '0 0 6px' }}>
                All Clear! No Pending Human Interventions
              </h3>
              <p style={{ fontSize: '13px', color: 'var(--text-secondary)', margin: 0 }}>
                High-value, elevated-fraud, and edge-case transactions requiring human review will appear here automatically.
              </p>
            </div>
          )
        ) : filteredWorkflows.length > 0 ? (
          <table className="data-table">
            <thead>
              <tr>
                <th>Payment</th>
                <th>Customer</th>
                <th>Amount</th>
                <th>Risk Level</th>
                <th>Failure Reason</th>
                <th>Probability</th>
                <th>Recommended Action</th>
                <th>Status</th>
                <th>Attempts</th>
                <th>Updated</th>
              </tr>
            </thead>
            <tbody>
              {filteredWorkflows.map((wf) => {
                const fraudScore = wf.fraud_probability ?? 0.08;
                const riskLevel = wf.overall_risk ?? (fraudScore >= 0.7 ? 'HIGH' : fraudScore >= 0.35 ? 'MEDIUM' : 'LOW');

                return (
                  <tr key={wf.id} className="clickable">
                    <td>
                      <Link href={`/workflows/${wf.id}`} style={{ display: 'block', width: '100%' }}>
                        <span className="mono" style={{ color: '#60a5fa', fontWeight: 600 }}>
                          {wf.payment_id ? `${wf.payment_id.substring(0, 18)}...` : wf.id.substring(0, 8)}
                        </span>
                      </Link>
                    </td>
                    <td>
                      <div style={{ fontSize: '13px', color: 'var(--text-primary)' }}>
                        {wf.customer_email || '—'}
                      </div>
                    </td>
                    <td>
                      <span className="mono">
                        ₹{wf.amount.toLocaleString('en-IN', { minimumFractionDigits: 2 })}
                      </span>
                    </td>
                    <td>
                      <span className={`badge ${riskLevel === 'HIGH' ? 'badge-danger' : riskLevel === 'MEDIUM' ? 'badge-warning' : 'badge-success'}`} style={{ fontSize: '11px' }}>
                        🛡️ {(fraudScore * 100).toFixed(0)}% ({riskLevel})
                      </span>
                    </td>
                    <td>
                      <span className="mono" style={{ fontSize: '12px' }}>
                        {wf.failure_code || 'UNKNOWN'}
                      </span>
                    </td>
                    <td>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                        <div style={{ width: '40px', height: '4px', background: 'var(--bg-elevated)', borderRadius: '2px', overflow: 'hidden' }}>
                          <div style={{
                            height: '100%',
                            width: `${(wf.recovery_probability * 100).toFixed(0)}%`,
                            background: wf.recovery_probability > 0.6 ? '#10b981' : wf.recovery_probability > 0.3 ? '#f59e0b' : '#ef4444',
                          }} />
                        </div>
                        <span className="mono" style={{ fontSize: '12px' }}>
                          {(wf.recovery_probability * 100).toFixed(0)}%
                        </span>
                      </div>
                    </td>
                    <td>
                      <span className="mono" style={{ fontSize: '12px' }}>
                        {wf.selected_action || 'PENDING'}
                      </span>
                    </td>
                    <td>
                      <StatusBadge status={wf.status} />
                    </td>
                    <td>
                      <span className="mono">{wf.attempt_count}</span>
                    </td>
                    <td style={{ fontSize: '12px', color: 'var(--text-muted)' }}>
                      {new Date(wf.updated_at || wf.created_at).toLocaleString()}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        ) : (
          <div style={{ padding: '40px' }}>
            <EmptyView
              title="No workflows match query"
              description="Adjust search filters or intake a new payment failure webhook."
            />
          </div>
        )}
      </div>
    </div>
  );
}

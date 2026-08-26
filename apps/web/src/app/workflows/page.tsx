'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { getWorkflows } from '@/lib/api';
import { WorkflowSummary } from '@/lib/types';
import { StatusBadge } from '@/components/ui/StatusBadge';
import { LoadingView, EmptyView, ErrorView } from '@/components/ui/StateViews';
import { RefreshCw, Search, Filter } from 'lucide-react';

export default function WorkflowsPage() {
  const [workflows, setWorkflows] = useState<WorkflowSummary[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [statusFilter, setStatusFilter] = useState<string>('ALL');
  const [searchQuery, setSearchQuery] = useState<string>('');

  const loadWorkflows = async () => {
    try {
      setLoading(true);
      setError(null);
      const res = await getWorkflows({
        status: statusFilter !== 'ALL' ? statusFilter : undefined,
        limit: 50,
      });
      setWorkflows(res.data || res.workflows || []);
    } catch (err: any) {
      setError(err.message || 'Failed to fetch workflows');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadWorkflows();
  }, [statusFilter]);

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

  return (
    <div className="page-container">
      {/* Header & Controls */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '24px' }}>
        <div>
          <h1 style={{ fontSize: '20px', fontWeight: 700, color: 'var(--text-primary)' }}>Recovery Workflows</h1>
          <p style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
            Active and resolved lifecycle executions orchestrated by ReviveOS
          </p>
        </div>
        <button onClick={loadWorkflows} className="btn-secondary" style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
          <RefreshCw size={14} className={loading ? 'spinning' : ''} />
          <span>Refresh</span>
        </button>
      </div>

      {/* Filter / Search Bar */}
      <div style={{ display: 'flex', gap: '16px', marginBottom: '20px', alignItems: 'center' }}>
        {/* Search */}
        <div style={{ position: 'relative', flex: 1 }}>
          <Search size={16} color="#64748b" style={{ position: 'absolute', left: '12px', top: '10px' }} />
          <input
            type="text"
            placeholder="Search by Payment ID, customer email, or failure code..."
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

        {/* Status Filter */}
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
            <option value="EXECUTING">EXECUTING</option>
            <option value="VERIFYING">VERIFYING</option>
            <option value="RECOVERED">RECOVERED</option>
            <option value="FAILED">FAILED</option>
            <option value="BLOCKED">BLOCKED</option>
            <option value="HALTED">HALTED</option>
          </select>
        </div>
      </div>

      {/* Workflows Table */}
      <div className="table-container">
        {loading && workflows.length === 0 ? (
          <div style={{ padding: '40px' }}>
            <LoadingView message="Loading workflows from database..." />
          </div>
        ) : error ? (
          <div style={{ padding: '40px' }}>
            <ErrorView message={error} onRetry={loadWorkflows} />
          </div>
        ) : filteredWorkflows.length > 0 ? (
          <table className="data-table">
            <thead>
              <tr>
                <th>Payment</th>
                <th>Customer</th>
                <th>Amount</th>
                <th>Failure Reason</th>
                <th>Probability</th>
                <th>Recommended Action</th>
                <th>Status</th>
                <th>Attempts</th>
                <th>Updated</th>
              </tr>
            </thead>
            <tbody>
              {filteredWorkflows.map((wf) => (
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
              ))}
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

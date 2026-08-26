'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { getAnalyticsOverview, getSystemQueues, getWorkflows } from '@/lib/api';
import { AnalyticsOverview, SystemQueuesResponse, WorkflowSummary } from '@/lib/types';
import { MetricCard } from '@/components/ui/MetricCard';
import { StatusBadge } from '@/components/ui/StatusBadge';
import { LoadingView, EmptyView, ErrorView } from '@/components/ui/StateViews';
import { ArrowRight, RefreshCw, Layers } from 'lucide-react';

export default function OverviewPage() {
  const [analytics, setAnalytics] = useState<AnalyticsOverview | null>(null);
  const [queues, setQueues] = useState<SystemQueuesResponse | null>(null);
  const [recentWorkflows, setRecentWorkflows] = useState<WorkflowSummary[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const loadData = async () => {
    try {
      setLoading(true);
      setError(null);
      const [analyticsData, queuesData, workflowsData] = await Promise.all([
        getAnalyticsOverview(),
        getSystemQueues().catch(() => null),
        getWorkflows({ limit: 5 }),
      ]);
      setAnalytics(analyticsData);
      setQueues(queuesData);
      setRecentWorkflows(workflowsData?.data || workflowsData?.workflows || []);
    } catch (err: any) {
      setError(err.message || 'Failed to load overview data');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
    const interval = setInterval(loadData, 15000);
    return () => clearInterval(interval);
  }, []);

  if (loading && !analytics) {
    return (
      <div className="page-container">
        <LoadingView message="Loading real recovery metrics from API..." />
      </div>
    );
  }

  if (error && !analytics) {
    return (
      <div className="page-container">
        <ErrorView message={error} onRetry={loadData} />
      </div>
    );
  }

  const recoveryRatePercent = analytics ? (analytics.recovery_rate * 100).toFixed(1) : '0.0';
  const totalQueued = queues?.queues?.reduce((acc, q) => acc + q.pending + q.active + q.scheduled, 0) ?? 0;
  const activeWorkers = queues?.servers?.length ?? 0;

  return (
    <div className="page-container">
      {/* Page Header */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '24px' }}>
        <div>
          <h1 style={{ fontSize: '20px', fontWeight: 700, color: 'var(--text-primary)' }}>Recovery Overview</h1>
          <p style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
            Real-time telemetry across failure intake, AI decisioning, and verified recoveries
          </p>
        </div>
        <button onClick={loadData} className="btn-secondary" style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
          <RefreshCw size={14} className={loading ? 'spinning' : ''} />
          <span>Refresh</span>
        </button>
      </div>

      {/* Primary Metrics Grid */}
      <div className="metrics-grid">
        <MetricCard
          label="Failed Volume"
          value={`₹${(analytics?.total_at_risk_revenue || 0).toLocaleString('en-IN', { minimumFractionDigits: 2 })}`}
          sub={`${analytics?.total_payments || 0} failed payments intercepted`}
        />
        <MetricCard
          label="Recovered Revenue"
          value={`₹${(analytics?.recovered_revenue || 0).toLocaleString('en-IN', { minimumFractionDigits: 2 })}`}
          sub={`${analytics?.recovered_workflows || 0} verified recovered`}
        />
        <MetricCard
          label="Recovery Rate"
          value={`${recoveryRatePercent}%`}
          sub={`${analytics?.recovered_workflows || 0} of ${analytics?.total_workflows || analytics?.total_payments || 0} captured`}
        />
        <MetricCard
          label="Active Recoveries"
          value={analytics?.active_recoveries || 0}
          sub={`${analytics?.pending_actions || 0} actions pending`}
        />
        <MetricCard
          label="Queue State"
          value={queues?.redis_status === 'connected' ? 'CONNECTED' : 'DISCONNECTED'}
          sub={`${totalQueued} tasks | ${activeWorkers} active workers`}
        />
      </div>

      {/* Category Breakdown & Queue Activity */}
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '20px', marginBottom: '28px' }}>
        {/* Failure Category Distribution */}
        <div className="table-container" style={{ padding: '20px' }}>
          <h3 style={{ fontSize: '14px', fontWeight: 600, marginBottom: '16px', display: 'flex', alignItems: 'center', gap: '8px' }}>
            <Layers size={16} color="#3b82f6" />
            <span>Failure Intake by Category</span>
          </h3>
          {analytics?.category_breakdown && Object.keys(analytics.category_breakdown).length > 0 ? (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
              {Object.entries(analytics.category_breakdown).map(([category, count]) => {
                const total = analytics.total_payments || 1;
                const pct = ((count / total) * 100).toFixed(0);
                return (
                  <div key={category}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '12px', marginBottom: '4px' }}>
                      <span className="mono" style={{ color: 'var(--text-primary)' }}>{category}</span>
                      <span style={{ color: 'var(--text-muted)' }}>{count} ({pct}%)</span>
                    </div>
                    <div style={{ height: '6px', background: 'var(--bg-elevated)', borderRadius: '3px', overflow: 'hidden' }}>
                      <div style={{ height: '100%', width: `${pct}%`, background: '#3b82f6' }} />
                    </div>
                  </div>
                );
              })}
            </div>
          ) : (
            <EmptyView title="No failure categories recorded" description="Intake webhook events to see breakdown." />
          )}
        </div>

        {/* Asynq / Redis Queue Activity */}
        <div className="table-container" style={{ padding: '20px' }}>
          <h3 style={{ fontSize: '14px', fontWeight: 600, marginBottom: '16px' }}>
            Queue & Infrastructure Load
          </h3>
          {queues?.queues && queues.queues.length > 0 ? (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '14px' }}>
              {queues.queues.map((q) => (
                <div key={q.queue} style={{ background: 'var(--bg-elevated)', padding: '12px 16px', borderRadius: '6px' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '6px' }}>
                    <span className="mono" style={{ fontWeight: 600 }}>Queue: {q.queue}</span>
                    <span style={{ fontSize: '11px', color: 'var(--text-muted)' }}>
                      {(q.memory_usage_bytes / 1024).toFixed(1)} KB RAM
                    </span>
                  </div>
                  <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '8px', fontSize: '12px' }}>
                    <div><span style={{ color: 'var(--text-muted)' }}>Active:</span> <b>{q.active}</b></div>
                    <div><span style={{ color: 'var(--text-muted)' }}>Pending:</span> <b>{q.pending}</b></div>
                    <div><span style={{ color: 'var(--text-muted)' }}>Scheduled:</span> <b>{q.scheduled}</b></div>
                    <div><span style={{ color: 'var(--text-muted)' }}>Retry:</span> <b>{q.retry}</b></div>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <EmptyView title="No queue activity" description="Asynq queues are currently idle." />
          )}
        </div>
      </div>

      {/* Recent Workflows Table */}
      <div className="table-container">
        <div className="table-header-bar">
          <h3 style={{ fontSize: '14px', fontWeight: 600 }}>Recent Recovery Workflows</h3>
          <Link href="/workflows" style={{ display: 'flex', alignItems: 'center', gap: '4px', fontSize: '12px', color: '#60a5fa' }}>
            <span>View All</span>
            <ArrowRight size={14} />
          </Link>
        </div>

        {recentWorkflows.length > 0 ? (
          <table className="data-table">
            <thead>
              <tr>
                <th>Payment ID</th>
                <th>Amount</th>
                <th>Failure</th>
                <th>AI Probability</th>
                <th>Action</th>
                <th>Status</th>
                <th>Created</th>
              </tr>
            </thead>
            <tbody>
              {recentWorkflows.map((wf) => (
                <tr key={wf.id} className="clickable">
                  <td>
                    <Link href={`/workflows/${wf.id}`} style={{ display: 'block', width: '100%' }}>
                      <span className="mono" style={{ color: '#60a5fa' }}>
                        {wf.payment_id ? `${wf.payment_id.substring(0, 16)}...` : wf.id.substring(0, 8)}
                      </span>
                    </Link>
                  </td>
                  <td>
                    <span className="mono">
                      ₹{wf.amount.toLocaleString('en-IN', { minimumFractionDigits: 2 })}
                    </span>
                  </td>
                  <td>
                    <span className="mono" style={{ fontSize: '12px' }}>{wf.failure_code || 'UNKNOWN'}</span>
                  </td>
                  <td>
                    <span className="mono">
                      {(wf.recovery_probability * 100).toFixed(0)}%
                    </span>
                  </td>
                  <td>
                    <span className="mono" style={{ fontSize: '12px' }}>{wf.selected_action || 'PENDING'}</span>
                  </td>
                  <td>
                    <StatusBadge status={wf.status} />
                  </td>
                  <td style={{ fontSize: '12px', color: 'var(--text-muted)' }}>
                    {new Date(wf.created_at).toLocaleTimeString()}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        ) : (
          <div style={{ padding: '32px' }}>
            <EmptyView
              title="No recovery workflows yet"
              description="Incoming payment failure webhooks will automatically spawn recovery workflows."
            />
          </div>
        )}
      </div>
    </div>
  );
}

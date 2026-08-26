'use client';

import { useEffect, useState } from 'react';
import { getSystemHealth, getSystemQueues } from '@/lib/api';
import { SystemHealthResponse, SystemQueuesResponse } from '@/lib/types';
import { StatusBadge } from '@/components/ui/StatusBadge';
import { LoadingView, ErrorView } from '@/components/ui/StateViews';
import { RefreshCw, Server, Activity, Database, Cpu, HardDrive } from 'lucide-react';

export default function SystemPage() {
  const [health, setHealth] = useState<SystemHealthResponse | null>(null);
  const [queues, setQueues] = useState<SystemQueuesResponse | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const loadSystemInfo = async () => {
    try {
      setLoading(true);
      setError(null);
      const [h, q] = await Promise.all([
        getSystemHealth(),
        getSystemQueues().catch(() => null),
      ]);
      setHealth(h);
      setQueues(q);
    } catch (err: any) {
      setError(err.message || 'Failed to load system diagnostics');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadSystemInfo();
    const interval = setInterval(loadSystemInfo, 10000);
    return () => clearInterval(interval);
  }, []);

  if (loading && !health) {
    return (
      <div className="page-container">
        <LoadingView message="Querying live infrastructure dependencies..." />
      </div>
    );
  }

  if (error && !health) {
    return (
      <div className="page-container">
        <ErrorView message={error} onRetry={loadSystemInfo} />
      </div>
    );
  }

  return (
    <div className="page-container">
      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '24px' }}>
        <div>
          <h1 style={{ fontSize: '20px', fontWeight: 700, color: 'var(--text-primary)' }}>System Infrastructure Health</h1>
          <p style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
            Real-time telemetry across database, queuing, worker pool, and local AI runtime
          </p>
        </div>
        <button onClick={loadSystemInfo} className="btn-secondary" style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
          <RefreshCw size={14} className={loading ? 'spinning' : ''} />
          <span>Refresh</span>
        </button>
      </div>

      {/* Component Health Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))', gap: '20px', marginBottom: '32px' }}>
        {health?.components.map((comp) => {
          let Icon = Server;
          if (comp.name.includes('PostgreSQL')) Icon = Database;
          if (comp.name.includes('Redis') || comp.name.includes('Queue')) Icon = HardDrive;
          if (comp.name.includes('Ollama') || comp.name.includes('DeepSeek')) Icon = Cpu;
          if (comp.name.includes('API')) Icon = Activity;

          return (
            <div key={comp.name} className="metric-card" style={{ gap: '12px' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                  <Icon size={18} color="#3b82f6" />
                  <span style={{ fontWeight: 600, color: 'var(--text-primary)' }}>{comp.name}</span>
                </div>
                <StatusBadge status={comp.status} />
              </div>

              <div style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>
                {comp.message || 'Operational'}
              </div>

              <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '11px', color: 'var(--text-muted)', borderTop: '1px solid var(--border-subtle)', paddingTop: '8px' }}>
                <span>Latency: {comp.latency_ms}ms</span>
                {comp.details && (
                  <span className="mono">{JSON.stringify(comp.details)}</span>
                )}
              </div>
            </div>
          );
        })}
      </div>

      {/* Queues & Workers Breakdown */}
      <div className="table-container" style={{ padding: '24px', marginBottom: '28px' }}>
        <h3 style={{ fontSize: '14px', fontWeight: 600, marginBottom: '16px' }}>
          Asynq Background Queue Infrastructure
        </h3>

        {queues?.queues && queues.queues.length > 0 ? (
          <table className="data-table">
            <thead>
              <tr>
                <th>Queue</th>
                <th>Active</th>
                <th>Pending</th>
                <th>Scheduled</th>
                <th>Retry</th>
                <th>Archived</th>
                <th>Completed</th>
                <th>Memory (RAM)</th>
              </tr>
            </thead>
            <tbody>
              {queues.queues.map((q) => (
                <tr key={q.queue}>
                  <td><span className="mono" style={{ fontWeight: 600 }}>{q.queue}</span></td>
                  <td><span className="mono">{q.active}</span></td>
                  <td><span className="mono">{q.pending}</span></td>
                  <td><span className="mono">{q.scheduled}</span></td>
                  <td><span className="mono">{q.retry}</span></td>
                  <td><span className="mono">{q.archived}</span></td>
                  <td><span className="mono">{q.completed}</span></td>
                  <td><span className="mono">{(q.memory_usage_bytes / 1024).toFixed(1)} KB</span></td>
                </tr>
              ))}
            </tbody>
          </table>
        ) : (
          <div style={{ color: 'var(--text-muted)', fontSize: '13px' }}>
            No active queue inspectors found.
          </div>
        )}
      </div>

      {/* Worker Server Pool */}
      <div className="table-container" style={{ padding: '24px' }}>
        <h3 style={{ fontSize: '14px', fontWeight: 600, marginBottom: '16px' }}>
          Registered Worker Processes
        </h3>

        {queues?.servers && queues.servers.length > 0 ? (
          <table className="data-table">
            <thead>
              <tr>
                <th>Server Host / PID</th>
                <th>Concurrency</th>
                <th>Queues Handled</th>
                <th>Status</th>
                <th>Started At</th>
              </tr>
            </thead>
            <tbody>
              {queues.servers.map((srv) => (
                <tr key={srv.id}>
                  <td>
                    <span className="mono">{srv.host} (PID: {srv.pid})</span>
                  </td>
                  <td><span className="mono">{srv.concurrency} threads</span></td>
                  <td>
                    <span className="mono" style={{ fontSize: '12px' }}>{srv.queues.join(', ')}</span>
                  </td>
                  <td><StatusBadge status={srv.status} /></td>
                  <td style={{ fontSize: '12px', color: 'var(--text-muted)' }}>
                    {new Date(srv.started).toLocaleTimeString()}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        ) : (
          <div style={{ color: 'var(--text-muted)', fontSize: '13px' }}>
            No worker processes currently connected.
          </div>
        )}
      </div>
    </div>
  );
}

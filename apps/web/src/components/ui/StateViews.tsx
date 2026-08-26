import React from 'react';
import { AlertCircle, Inbox, Loader2 } from 'lucide-react';

export function LoadingView({ message = 'Loading live data from ReviveOS API...' }: { message?: string }) {
  return (
    <div className="state-box">
      <Loader2 size={32} color="#3b82f6" className="spinning" />
      <div className="state-title">{message}</div>
      <div className="state-desc">Querying backend PostgreSQL and Asynq queues</div>
    </div>
  );
}

export function EmptyView({
  title = 'No records found',
  description = 'No data currently exists for this view.',
  action,
}: {
  title?: string;
  description?: string;
  action?: React.ReactNode;
}) {
  return (
    <div className="state-box">
      <Inbox size={36} color="#64748b" />
      <div className="state-title">{title}</div>
      <div className="state-desc">{description}</div>
      {action && <div style={{ marginTop: '8px' }}>{action}</div>}
    </div>
  );
}

export function ErrorView({
  title = 'Unable to connect to ReviveOS API',
  message,
  onRetry,
}: {
  title?: string;
  message: string;
  onRetry?: () => void;
}) {
  return (
    <div className="state-box" style={{ borderColor: 'var(--color-danger-border)' }}>
      <AlertCircle size={36} color="#ef4444" />
      <div className="state-title" style={{ color: '#ef4444' }}>{title}</div>
      <div className="state-desc">{message}</div>
      {onRetry && (
        <button onClick={onRetry} className="btn-secondary" style={{ marginTop: '8px' }}>
          Retry Connection
        </button>
      )}
    </div>
  );
}

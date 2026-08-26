import React from 'react';

interface StatusBadgeProps {
  status: string;
  type?: 'workflow' | 'policy' | 'decision' | 'health';
}

export function StatusBadge({ status, type = 'workflow' }: StatusBadgeProps) {
  const norm = (status || 'UNKNOWN').toUpperCase();

  let badgeClass = 'badge-neutral';

  switch (norm) {
    case 'RECOVERED':
    case 'CAPTURED':
    case 'ALLOW':
    case 'HEALTHY':
    case 'AVAILABLE':
    case 'EXECUTED':
    case 'SUCCESS':
      badgeClass = 'badge-success';
      break;

    case 'SCHEDULED':
    case 'ANALYZING':
    case 'EXECUTING':
    case 'VERIFYING':
    case 'PENDING':
    case 'DEGRADED':
    case 'MODIFY':
      badgeClass = 'badge-warning';
      break;

    case 'FAILED':
    case 'BLOCKED':
    case 'HALTED':
    case 'UNHEALTHY':
    case 'UNAVAILABLE':
      badgeClass = 'badge-danger';
      break;

    case 'ESCALATE':
    case 'ESCALATED':
      badgeClass = 'badge-accent';
      break;

    default:
      badgeClass = 'badge-neutral';
  }

  return (
    <span className={`badge ${badgeClass}`}>
      {norm}
    </span>
  );
}

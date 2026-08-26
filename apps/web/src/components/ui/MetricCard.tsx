import React from 'react';

interface MetricCardProps {
  label: string;
  value: string | number;
  sub?: string;
  badge?: React.ReactNode;
}

export function MetricCard({ label, value, sub, badge }: MetricCardProps) {
  return (
    <div className="metric-card">
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <span className="metric-label">{label}</span>
        {badge}
      </div>
      <div className="metric-value">{value}</div>
      {sub && <div className="metric-sub">{sub}</div>}
    </div>
  );
}

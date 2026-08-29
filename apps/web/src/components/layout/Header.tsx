'use client';

import { useEffect, useState } from 'react';
import { getSystemHealth } from '@/lib/api';
import { Activity, RefreshCw } from 'lucide-react';
import { ThemeToggle } from '@/components/ui/ThemeToggle';

export default function Header({ title }: { title?: string }) {
  const [status, setStatus] = useState<string>('CHECKING');
  const [loading, setLoading] = useState<boolean>(false);

  const checkHealth = async () => {
    try {
      setLoading(true);
      const res = await getSystemHealth();
      setStatus(res.overall_status || 'HEALTHY');
    } catch {
      setStatus('UNHEALTHY');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    checkHealth();
    const timer = setInterval(checkHealth, 15000);
    return () => clearInterval(timer);
  }, []);

  const statusColor = status === 'HEALTHY' ? 'var(--color-emerald)' : status === 'DEGRADED' ? 'var(--color-amber)' : 'var(--color-red)';

  return (
    <header className="top-header">
      <div className="page-title">{title || 'ReviveOS Operations'}</div>

      <div style={{ display: 'flex', alignItems: 'center', gap: '14px' }}>
        <ThemeToggle />

        <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
          <Activity size={14} color={statusColor} />
          <span style={{ fontSize: '12px', fontWeight: 600, color: statusColor }}>
            {status}
          </span>
        </div>

        <button
          onClick={checkHealth}
          className="btn-ghost"
          style={{ padding: '5px 12px', fontSize: '12px', gap: '5px' }}
          title="Refresh system status"
        >
          <RefreshCw size={12} className={loading ? 'spinning' : ''} />
          Sync
        </button>
      </div>
    </header>
  );
}

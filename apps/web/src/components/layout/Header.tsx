'use client';

import { useEffect, useState } from 'react';
import { getSystemHealth } from '@/lib/api';
import { Activity, RefreshCw } from 'lucide-react';

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

  return (
    <header className="top-header">
      <div className="page-title">{title || 'ReviveOS Operations'}</div>

      <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <Activity size={16} color={status === 'HEALTHY' ? '#10b981' : status === 'DEGRADED' ? '#f59e0b' : '#ef4444'} />
          <span style={{ 
            fontSize: '12px', 
            fontWeight: 600, 
            color: status === 'HEALTHY' ? '#10b981' : status === 'DEGRADED' ? '#f59e0b' : '#ef4444' 
          }}>
            SYSTEM {status}
          </span>
        </div>

        <button 
          onClick={checkHealth}
          className="btn-secondary" 
          style={{ padding: '6px 10px', display: 'flex', alignItems: 'center', gap: '6px', fontSize: '12px' }}
          title="Refresh system status"
        >
          <RefreshCw size={12} className={loading ? 'spinning' : ''} />
          <span>Sync</span>
        </button>
      </div>
    </header>
  );
}

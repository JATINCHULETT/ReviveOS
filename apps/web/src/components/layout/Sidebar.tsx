'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { 
  LayoutDashboard, 
  GitBranch, 
  BarChart3, 
  Server, 
  ShieldCheck,
  Building2,
  Shield,
  LogIn
} from 'lucide-react';

export default function Sidebar() {
  const pathname = usePathname();

  const navItems = [
    { label: 'Overview', href: '/', icon: LayoutDashboard },
    { label: 'Merchant Portal', href: '/merchant', icon: Building2 },
    { label: 'Admin Hub', href: '/admin', icon: Shield },
    { label: 'Workflows', href: '/workflows', icon: GitBranch },
    { label: 'Analytics', href: '/analytics', icon: BarChart3 },
    { label: 'System Health', href: '/system', icon: Server },
    { label: 'Sign In / Roles', href: '/login', icon: LogIn },
  ];

  return (
    <aside className="sidebar">
      <div className="sidebar-header">
        <div className="logo-badge">REVIVE</div>
        <div>
          <div className="logo-text">ReviveOS</div>
          <div className="logo-sub">Recovery Engine</div>
        </div>
      </div>

      <nav className="sidebar-nav">
        {navItems.map((item) => {
          const Icon = item.icon;
          const isActive = pathname === item.href || (item.href !== '/' && pathname.startsWith(item.href));
          return (
            <Link
              key={item.href}
              href={item.href}
              className={`nav-link ${isActive ? 'active' : ''}`}
            >
              <Icon size={18} />
              <span>{item.label}</span>
            </Link>
          );
        })}
      </nav>

      <div style={{ padding: '16px 20px', borderTop: '1px solid var(--border-subtle)', fontSize: '11px', color: 'var(--text-muted)' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '6px', marginBottom: '4px' }}>
          <ShieldCheck size={14} color="#10b981" />
          <span style={{ color: '#10b981', fontWeight: 600 }}>Ledger Validated</span>
        </div>
        <div>Hash-chained audit log active</div>
      </div>
    </aside>
  );
}

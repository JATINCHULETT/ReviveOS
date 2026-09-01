'use client';

import React, { useEffect, useState } from 'react';
import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import {
  LayoutDashboard,
  GitBranch,
  BarChart3,
  Server,
  ShieldCheck,
  Building2,
  Shield,
  LogIn,
  Zap,
  Globe,
  LogOut,
  User,
  Users,
  Terminal,
  BookOpen,
} from 'lucide-react';
import ReviveLogo from '@/components/ui/ReviveLogo';
import { ThemeToggle } from '@/components/ui/ThemeToggle';

export default function Sidebar() {
  const pathname = usePathname();
  const router = useRouter();
  const [currentUser, setCurrentUser] = useState<any>(null);

  useEffect(() => {
    try {
      const stored = localStorage.getItem('revive_user');
      if (stored) {
        setCurrentUser(JSON.parse(stored));
      }
    } catch {
      // ignore
    }
  }, [pathname]);

  const navItems = [
    { label: 'Analytics', href: '/analytics', icon: BarChart3 },
    { label: 'Workflows', href: '/workflows', icon: GitBranch },
    { label: 'Customers & Retries', href: '/merchant', icon: Users },
    { label: 'Developer & SDK', href: '/developer', icon: Terminal },
    { label: 'Documentation', href: '/docs', icon: BookOpen },
    { label: 'System Health', href: '/system', icon: Server },
  ];

  const handleLogout = () => {
    localStorage.removeItem('revive_token');
    localStorage.removeItem('revive_user');
    setCurrentUser(null);
    router.push('/login');
  };

  return (
    <aside className="sidebar">
      {/* Brand */}
      <div className="sidebar-header" style={{ padding: '16px 20px' }}>
        <ReviveLogo size="md" href="/analytics" />
      </div>

      {/* Navigation */}
      <nav className="sidebar-nav">
        <div style={{ fontSize: '11px', fontWeight: 600, letterSpacing: '0.05em', color: 'var(--text-muted)', padding: '0 12px 8px', marginTop: '4px' }}>
          MAIN CONSOLE
        </div>
        {navItems.map((item) => {
          const Icon = item.icon;
          const isActive = pathname === item.href || (item.href !== '/analytics' && pathname.startsWith(item.href));
          return (
            <Link key={item.href} href={item.href} className={`nav-item ${isActive ? 'active' : ''}`}>
              <Icon size={17} />
              <span>{item.label}</span>
            </Link>
          );
        })}
      </nav>

      {/* Theme Switcher in Sidebar */}
      <div style={{ padding: '10px 14px', display: 'flex', alignItems: 'center', justifyContent: 'space-between', borderTop: '1px solid var(--border-subtle)' }}>
        <span style={{ fontSize: '12px', color: 'var(--text-muted)', fontWeight: 600 }}>THEME</span>
        <ThemeToggle />
      </div>

      {/* User Session */}
      <div style={{ padding: '14px 12px', borderTop: '1px solid var(--border-subtle)' }}>
        {currentUser ? (
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px', minWidth: 0 }}>
              <div style={{
                width: '28px', height: '28px', borderRadius: 'var(--radius-full)',
                background: 'var(--color-accent-bg)',
                border: '1px solid var(--color-accent-border)',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                flexShrink: 0,
              }}>
                <User size={13} color="var(--color-accent-light)" />
              </div>
              <div style={{ minWidth: 0 }}>
                <div style={{ fontSize: '12px', fontWeight: 600, color: 'var(--text-primary)', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                  {currentUser.email}
                </div>
                <div style={{ fontSize: '10px', color: 'var(--color-accent-light)', fontWeight: 500 }}>
                  {currentUser.role}
                </div>
              </div>
            </div>
            <button
              onClick={handleLogout}
              title="Sign Out"
              style={{
                background: 'transparent',
                border: 'none',
                color: 'var(--text-muted)',
                cursor: 'pointer',
                padding: '6px',
                borderRadius: 'var(--radius-sm)',
                transition: 'color 0.15s ease',
                display: 'flex',
              }}
              onMouseEnter={(e) => (e.currentTarget.style.color = 'var(--color-red)')}
              onMouseLeave={(e) => (e.currentTarget.style.color = 'var(--text-muted)')}
            >
              <LogOut size={15} />
            </button>
          </div>
        ) : (
          <Link
            href="/login"
            style={{
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              gap: '8px', padding: '8px',
              borderRadius: 'var(--radius-sm)',
              background: 'var(--bg-card)',
              border: '1px solid var(--border-subtle)',
              fontSize: '12px', fontWeight: 500,
              color: 'var(--color-accent-light)',
              transition: 'background 0.15s ease',
            }}
          >
            <LogIn size={13} />
            <span>Sign In / Switch Role</span>
          </Link>
        )}
      </div>
    </aside>
  );
}

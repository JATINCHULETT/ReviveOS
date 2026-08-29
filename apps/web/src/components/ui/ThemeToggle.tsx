'use client';

import React, { useEffect, useState } from 'react';
import { Sun, Moon, Monitor } from 'lucide-react';

export type ThemeMode = 'dark' | 'light' | 'system';

export function ThemeToggle({ showLabels = false }: { showLabels?: boolean }) {
  const [theme, setTheme] = useState<ThemeMode>('dark');
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
    try {
      const saved = (localStorage.getItem('revive_theme') as ThemeMode) || 'dark';
      setTheme(saved);
      applyTheme(saved);
    } catch {
      // fallback
    }
  }, []);

  const applyTheme = (mode: ThemeMode) => {
    if (typeof document === 'undefined') return;
    const root = document.documentElement;
    if (mode === 'system') {
      const isSystemDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
      if (isSystemDark) {
        root.classList.remove('light');
        root.classList.add('dark');
      } else {
        root.classList.remove('dark');
        root.classList.add('light');
      }
    } else if (mode === 'light') {
      root.classList.remove('dark');
      root.classList.add('light');
    } else {
      root.classList.remove('light');
      root.classList.add('dark');
    }
  };

  const handleSelect = (mode: ThemeMode) => {
    setTheme(mode);
    try {
      localStorage.setItem('revive_theme', mode);
    } catch {}
    applyTheme(mode);
  };

  const currentTheme = mounted ? theme : 'dark';

  return (
    <div
      suppressHydrationWarning
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        background: 'rgba(255, 255, 255, 0.08)',
        border: '1px solid var(--border-default)',
        borderRadius: '9999px',
        padding: '3px 4px',
        gap: '3px',
        boxShadow: '0 2px 8px rgba(0, 0, 0, 0.2)',
      }}
    >
      {/* Light Button */}
      <button
        type="button"
        onClick={() => handleSelect('light')}
        title="Light Mode"
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: '4px',
          padding: showLabels ? '4px 10px' : '5px',
          borderRadius: '9999px',
          border: 'none',
          cursor: 'pointer',
          background: currentTheme === 'light' ? 'linear-gradient(135deg, #8b5cf6, #ec4899)' : 'transparent',
          color: currentTheme === 'light' ? '#ffffff' : 'var(--text-muted)',
          fontSize: '12px',
          fontWeight: 600,
          transition: 'all 0.2s ease',
        }}
      >
        <Sun size={14} />
        {showLabels && <span>Light</span>}
      </button>

      {/* Dark Button */}
      <button
        type="button"
        onClick={() => handleSelect('dark')}
        title="Dark Mode"
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: '4px',
          padding: showLabels ? '4px 10px' : '5px',
          borderRadius: '9999px',
          border: 'none',
          cursor: 'pointer',
          background: currentTheme === 'dark' ? 'linear-gradient(135deg, #8b5cf6, #ec4899)' : 'transparent',
          color: currentTheme === 'dark' ? '#ffffff' : 'var(--text-muted)',
          fontSize: '12px',
          fontWeight: 600,
          transition: 'all 0.2s ease',
        }}
      >
        <Moon size={14} />
        {showLabels && <span>Dark</span>}
      </button>

      {/* System Button */}
      <button
        type="button"
        onClick={() => handleSelect('system')}
        title="System Auto Mode"
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: '4px',
          padding: showLabels ? '4px 10px' : '5px',
          borderRadius: '9999px',
          border: 'none',
          cursor: 'pointer',
          background: currentTheme === 'system' ? 'linear-gradient(135deg, #8b5cf6, #ec4899)' : 'transparent',
          color: currentTheme === 'system' ? '#ffffff' : 'var(--text-muted)',
          fontSize: '12px',
          fontWeight: 600,
          transition: 'all 0.2s ease',
        }}
      >
        <Monitor size={14} />
        {showLabels && <span>System</span>}
      </button>
    </div>
  );
}

/* Floating Quick Toggle in bottom right */
export function FloatingThemeToggle() {
  const [theme, setTheme] = useState<ThemeMode>('dark');
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
    try {
      const saved = (localStorage.getItem('revive_theme') as ThemeMode) || 'dark';
      setTheme(saved);
    } catch {}
  }, []);

  const toggle = () => {
    const next: ThemeMode = theme === 'dark' ? 'light' : theme === 'light' ? 'system' : 'dark';
    setTheme(next);
    if (typeof document === 'undefined') return;
    const root = document.documentElement;
    if (next === 'system') {
      const isSystemDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
      root.classList.toggle('dark', isSystemDark);
      root.classList.toggle('light', !isSystemDark);
    } else if (next === 'light') {
      root.classList.remove('dark');
      root.classList.add('light');
    } else {
      root.classList.remove('light');
      root.classList.add('dark');
    }
    try {
      localStorage.setItem('revive_theme', next);
    } catch {}
  };

  const currentTheme = mounted ? theme : 'dark';

  return (
    <button
      type="button"
      onClick={toggle}
      suppressHydrationWarning
      title={`Theme: ${currentTheme.toUpperCase()} (Click to toggle)`}
      style={{
        position: 'fixed',
        bottom: '24px',
        right: '24px',
        zIndex: 999,
        display: 'flex',
        alignItems: 'center',
        gap: '8px',
        padding: '10px 18px',
        borderRadius: '9999px',
        background: 'var(--bg-surface)',
        border: '1px solid var(--border-default)',
        color: 'var(--text-primary)',
        fontSize: '13px',
        fontWeight: 600,
        boxShadow: '0 8px 24px rgba(0, 0, 0, 0.4)',
        cursor: 'pointer',
        backdropFilter: 'blur(16px)',
        transition: 'all 0.25s ease',
      }}
    >
      {currentTheme === 'light' && <Sun size={16} color="#d97706" />}
      {currentTheme === 'dark' && <Moon size={16} color="#c084fc" />}
      {currentTheme === 'system' && <Monitor size={16} color="#38bdf8" />}
      <span style={{ textTransform: 'capitalize' }}>{currentTheme} Mode</span>
    </button>
  );
}

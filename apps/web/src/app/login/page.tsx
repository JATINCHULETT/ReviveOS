'use client';

import React, { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { motion, AnimatePresence } from 'framer-motion';
import {
  Zap,
  Lock,
  Mail,
  ArrowRight,
  ShieldCheck,
  Building,
  Key,
  CheckCircle2,
  AlertCircle,
  Sparkles,
} from 'lucide-react';
import { loginUser } from '@/lib/api';
import GradientMesh from '@/components/3d/GradientMesh';
import { BadgePulse } from '@/components/ui/AnimatedComponents';
import ReviveLogo from '@/components/ui/ReviveLogo';

export default function LoginPage() {
  const router = useRouter();
  const [roleTab, setRoleTab] = useState<'MERCHANT' | 'ADMIN' | 'DEMO'>('MERCHANT');
  const [email, setEmail] = useState('merchant@acme.com');
  const [password, setPassword] = useState('merchant');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const handleRoleSelect = (role: 'MERCHANT' | 'ADMIN' | 'DEMO') => {
    setRoleTab(role);
    setError(null);
    if (role === 'MERCHANT') {
      setEmail('merchant@acme.com');
      setPassword('merchant');
    } else if (role === 'ADMIN') {
      setEmail('admin@reviveos.io');
      setPassword('admin');
    } else {
      setEmail('demo@reviveos.io');
      setPassword('demo123');
    }
  };

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    if (roleTab === 'DEMO') {
      localStorage.setItem('revive_token', 'demo_token');
      localStorage.setItem('revive_user', JSON.stringify({ email: 'demo@reviveos.io', role: 'MERCHANT', name: 'Demo Merchant' }));
      setSuccess(true);
      setTimeout(() => router.push('/merchant'), 600);
      return;
    }

    try {
      const res = await loginUser({ email, password });
      localStorage.setItem('revive_token', res.token);
      localStorage.setItem('revive_user', JSON.stringify(res.user));

      setSuccess(true);
      setTimeout(() => {
        if (res.user.role === 'ADMIN') {
          router.push('/admin');
        } else {
          router.push('/merchant');
        }
      }, 500);
    } catch (err: any) {
      setError(err?.message || 'Invalid email or password. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const roles = [
    { id: 'MERCHANT' as const, label: 'Merchant', icon: <Building size={14} /> },
    { id: 'ADMIN' as const, label: 'Admin', icon: <ShieldCheck size={14} /> },
    { id: 'DEMO' as const, label: 'Demo', icon: <Sparkles size={14} /> },
  ];

  return (
    <div
      style={{
        position: 'relative',
        minHeight: '100vh',
        backgroundColor: 'var(--bg-primary)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '24px',
        overflow: 'hidden',
      }}
    >
      <GradientMesh />

      {/* Back Link */}
      <Link
        href="/"
        style={{
          position: 'absolute',
          top: '20px',
          left: '20px',
          zIndex: 30,
          display: 'flex',
          alignItems: 'center',
          gap: '8px',
          padding: '8px 14px',
          borderRadius: 'var(--radius-full)',
          background: 'var(--bg-card)',
          border: '1px solid var(--border-subtle)',
          color: 'var(--text-muted)',
          fontSize: '13px',
          fontWeight: 500,
          transition: 'all 0.2s ease',
        }}
      >
        <Zap size={14} color="var(--color-accent-light)" />
        ReviveOS
      </Link>

      {/* Login Card */}
      <motion.div
        initial={{ opacity: 0, y: 16 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.4, ease: [0.25, 0.4, 0.25, 1] }}
        style={{
          position: 'relative',
          zIndex: 20,
          width: '100%',
          maxWidth: '420px',
          background: 'rgba(9, 9, 11, 0.85)',
          backdropFilter: 'blur(20px)',
          WebkitBackdropFilter: 'blur(20px)',
          border: '1px solid var(--border-subtle)',
          borderRadius: 'var(--radius-xl)',
          padding: '36px',
        }}
      >
        {/* Header */}
        <div style={{ textAlign: 'center', marginBottom: '28px' }}>
          <div style={{ display: 'flex', justifyContent: 'center', marginBottom: '14px' }}>
            <ReviveLogo size="lg" showTagline={true} href="/" />
          </div>
          <p style={{ fontSize: '13px', color: 'var(--text-secondary)', marginTop: '8px' }}>
            Sign in to your autonomous AI recovery console
          </p>
        </div>

        {/* Role Tabs */}
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(3, 1fr)',
            gap: '4px',
            padding: '4px',
            borderRadius: 'var(--radius-md)',
            background: 'rgba(255, 255, 255, 0.03)',
            border: '1px solid var(--border-subtle)',
            marginBottom: '24px',
          }}
        >
          {roles.map((tab) => {
            const isSelected = roleTab === tab.id;
            return (
              <button
                key={tab.id}
                type="button"
                onClick={() => handleRoleSelect(tab.id)}
                style={{
                  position: 'relative',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: '6px',
                  padding: '9px 8px',
                  borderRadius: 'var(--radius-sm)',
                  border: 'none',
                  fontSize: '13px',
                  fontWeight: isSelected ? 600 : 500,
                  fontFamily: 'var(--font-sans)',
                  cursor: 'pointer',
                  background: isSelected ? 'var(--color-accent-bg)' : 'transparent',
                  color: isSelected ? 'var(--color-accent-light)' : 'var(--text-muted)',
                  transition: 'all 0.2s ease',
                }}
              >
                {tab.icon}
                {tab.label}
              </button>
            );
          })}
        </div>

        {/* Error Alert */}
        <AnimatePresence>
          {error && (
            <motion.div
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: 'auto' }}
              exit={{ opacity: 0, height: 0 }}
              style={{
                marginBottom: '16px',
                padding: '12px',
                borderRadius: 'var(--radius-sm)',
                background: 'var(--color-red-bg)',
                border: '1px solid var(--color-red-border)',
                display: 'flex',
                alignItems: 'center',
                gap: '8px',
                fontSize: '13px',
                color: '#ef4444',
              }}
            >
              <AlertCircle size={15} />
              <span>{error}</span>
            </motion.div>
          )}
        </AnimatePresence>

        {/* Success Alert */}
        {success && (
          <div
            style={{
              marginBottom: '16px',
              padding: '12px',
              borderRadius: 'var(--radius-sm)',
              background: 'var(--color-emerald-bg)',
              border: '1px solid var(--color-emerald-border)',
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
              fontSize: '13px',
              color: '#10b981',
            }}
          >
            <CheckCircle2 size={15} />
            <span>Authenticated — launching console...</span>
          </div>
        )}

        {/* Login Form */}
        <form onSubmit={handleLogin} style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
          <div>
            <label style={{ fontSize: '13px', fontWeight: 500, color: 'var(--text-muted)', display: 'block', marginBottom: '6px' }}>
              Email
            </label>
            <div style={{ position: 'relative' }}>
              <Mail size={15} color="var(--text-muted)" style={{ position: 'absolute', left: '12px', top: '13px' }} />
              <input
                type="email"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="input"
                style={{ paddingLeft: '38px' }}
              />
            </div>
          </div>

          <div>
            <label style={{ fontSize: '13px', fontWeight: 500, color: 'var(--text-muted)', display: 'block', marginBottom: '6px' }}>
              Password
            </label>
            <div style={{ position: 'relative' }}>
              <Lock size={15} color="var(--text-muted)" style={{ position: 'absolute', left: '12px', top: '13px' }} />
              <input
                type="password"
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="input"
                style={{ paddingLeft: '38px' }}
              />
            </div>
          </div>

          <motion.button
            whileHover={{ scale: 1.01 }}
            whileTap={{ scale: 0.99 }}
            type="submit"
            disabled={loading}
            className="btn-primary"
            style={{ width: '100%', marginTop: '6px', padding: '12px', fontSize: '14px' }}
          >
            {loading ? 'Authenticating...' : roleTab === 'DEMO' ? 'Enter Sandbox' : 'Sign In'}
            <ArrowRight size={15} />
          </motion.button>
        </form>

        {/* Security Footer */}
        <div style={{ marginTop: '22px', textAlign: 'center', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '6px', fontSize: '11px', color: 'var(--text-muted)' }}>
          <Key size={11} color="var(--color-emerald)" />
          <span>HMAC-SHA256 Scoped Session • PCI-DSS Encrypted</span>
        </div>
      </motion.div>
    </div>
  );
}

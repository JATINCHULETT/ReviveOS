'use client';

import React from 'react';
import { usePathname } from 'next/navigation';
import Sidebar from '@/components/layout/Sidebar';
import Header from '@/components/layout/Header';

export default function AppLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();

  // Standalone pages that do NOT show dashboard sidebar/header
  const isStandalone = pathname === '/' || pathname === '/login';

  if (isStandalone) {
    return <main style={{ minHeight: '100vh' }}>{children}</main>;
  }

  return (
    <div className="dashboard-layout">
      <Sidebar />
      <div className="main-content">
        <Header />
        <main>{children}</main>
      </div>
    </div>
  );
}

'use client';

import React, { useEffect } from 'react';
import { useRouter } from 'next/navigation';

export default function AdminHubPage() {
  const router = useRouter();

  useEffect(() => {
    router.replace('/merchant');
  }, [router]);

  return null;
}

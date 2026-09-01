import { NextResponse } from 'next/server';
import { getSyntheticAnalyticsOverview } from '@/lib/syntheticDataset';

export async function GET() {
  const data = getSyntheticAnalyticsOverview();
  return NextResponse.json(data);
}

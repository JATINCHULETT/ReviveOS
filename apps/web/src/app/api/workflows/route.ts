import { NextRequest, NextResponse } from 'next/server';
import { getSyntheticWorkflows } from '@/lib/syntheticDataset';

export async function GET(request: NextRequest) {
  const searchParams = request.nextUrl.searchParams;
  const status = searchParams.get('status');
  const limit = parseInt(searchParams.get('limit') || '50', 10);
  const offset = parseInt(searchParams.get('offset') || '0', 10);

  const all = getSyntheticWorkflows();
  let filtered = all;
  if (status && status !== 'ALL') {
    filtered = all.filter((w) => w.status === status);
  }

  const slice = filtered.slice(offset, offset + limit);

  return NextResponse.json({
    data: slice,
    workflows: slice,
    total: filtered.length,
    limit,
    offset,
  });
}

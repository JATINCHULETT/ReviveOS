import { NextRequest, NextResponse } from 'next/server';
import { getSyntheticWorkflowDetail } from '@/lib/syntheticDataset';

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  const detail = getSyntheticWorkflowDetail(id);
  return NextResponse.json(detail);
}

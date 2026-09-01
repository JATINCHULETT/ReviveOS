import { NextRequest, NextResponse } from 'next/server';
import { getSyntheticInterventions } from '@/lib/syntheticDataset';

export async function GET(request: NextRequest) {
  const interventions = getSyntheticInterventions();
  return NextResponse.json({
    data: interventions,
    total: interventions.length,
  });
}

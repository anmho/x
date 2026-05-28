import { NextResponse } from 'next/server';
import { pauseMockRun } from '@/lib/trading/mock-store';
import { isXapiLiveEnabled, pauseRunInXapi } from '@/lib/trading/xapi';

export async function POST(
  _request: Request,
  context: { params: Promise<{ runId: string }> },
) {
  try {
    const { runId } = await context.params;

    if (isXapiLiveEnabled()) {
      const live = await pauseRunInXapi(runId);
      return NextResponse.json(live);
    }

    const run = pauseMockRun(runId);
    if (!run) {
      return NextResponse.json({ error: 'Run not found' }, { status: 404 });
    }

    return NextResponse.json({ run, source: 'mock' as const });
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Failed to pause run';
    return NextResponse.json({ error: message }, { status: 500 });
  }
}

import { NextRequest, NextResponse } from 'next/server';
import { createMockRun, listMockRuns } from '@/lib/trading/mock-store';
import { createRunInXapi, isXapiLiveEnabled, listRunsFromXapi } from '@/lib/trading/xapi';
import { CreateBotRunRequest } from '@/lib/trading/types';

export async function GET() {
  try {
    if (isXapiLiveEnabled()) {
      const live = await listRunsFromXapi();
      return NextResponse.json(live);
    }

    return NextResponse.json({ runs: listMockRuns(), source: 'mock' as const });
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Failed to load bot runs';
    return NextResponse.json({ error: message }, { status: 500 });
  }
}

export async function POST(request: NextRequest) {
  try {
    const body = (await request.json()) as CreateBotRunRequest;

    if (!body.name || !body.strategy || !body.market) {
      return NextResponse.json(
        { error: 'name, strategy, and market are required' },
        { status: 400 },
      );
    }

    if (isXapiLiveEnabled()) {
      const live = await createRunInXapi(body);
      return NextResponse.json(live, { status: 201 });
    }

    const run = createMockRun(body);
    return NextResponse.json({ run, source: 'mock' as const }, { status: 201 });
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Failed to create bot run';
    return NextResponse.json({ error: message }, { status: 500 });
  }
}

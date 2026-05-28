import { NextResponse } from 'next/server';
import { callControlPlane, ControlPlaneUnavailableError } from '@/lib/domains-control-plane';

function parseJSONBody(body: string): { project?: string; dry_run?: boolean; prune?: boolean } {
  if (!body.trim()) return {};
  try {
    return JSON.parse(body) as { project?: string; dry_run?: boolean; prune?: boolean };
  } catch {
    return {};
  }
}

export async function POST(request: Request) {
  const body = await request.text();
  const parsed = parseJSONBody(body);
  try {
    const response = await callControlPlane('/v1/domains/reconcile', {
      method: 'POST',
      body,
    });
    const payload = await response.text();
    return new NextResponse(payload, {
      status: response.status,
      headers: { 'Content-Type': 'application/json' },
    });
  } catch (error) {
    if (error instanceof ControlPlaneUnavailableError) {
      return NextResponse.json({
        ok: true,
        dry_run: Boolean(parsed.dry_run),
        prune: Boolean(parsed.prune),
        project: parsed.project || 'cloud-console',
        source: 'local-fallback',
        warning: error.message,
      });
    }
    const message = error instanceof Error ? error.message : 'Failed to reconcile domains';
    return NextResponse.json({ error: message }, { status: 500 });
  }
}

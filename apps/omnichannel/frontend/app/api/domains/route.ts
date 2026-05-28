import { NextResponse } from 'next/server';
import { callControlPlane, ControlPlaneUnavailableError } from '@/lib/domains-control-plane';
import { listLocalDomains } from '@/lib/domains-local-store';

export async function GET(request: Request) {
  const url = new URL(request.url);
  const project = url.searchParams.get('project') || undefined;
  try {
    const suffix = project ? `?project=${encodeURIComponent(project)}` : '';
    const response = await callControlPlane(`/v1/domains${suffix}`, { method: 'GET' });
    const payload = await response.text();
    return new NextResponse(payload, {
      status: response.status,
      headers: { 'Content-Type': 'application/json' },
    });
  } catch (error) {
    if (error instanceof ControlPlaneUnavailableError) {
      return NextResponse.json({
        domains: listLocalDomains(project),
        source: 'local-fallback',
        warning: error.message,
      });
    }
    const message = error instanceof Error ? error.message : 'Failed to list domains';
    return NextResponse.json({ error: message }, { status: 500 });
  }
}

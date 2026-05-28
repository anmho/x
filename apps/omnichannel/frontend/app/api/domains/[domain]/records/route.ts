import { NextResponse } from 'next/server';
import { callControlPlane, ControlPlaneUnavailableError } from '@/lib/domains-control-plane';
import { createLocalDomainRecord, listLocalDomainRecords } from '@/lib/domains-local-store';

export async function GET(request: Request, context: { params: Promise<{ domain: string }> }) {
  const { domain } = await context.params;
  const url = new URL(request.url);
  const provider = url.searchParams.get('provider');
  if (!provider) {
    return NextResponse.json({ error: 'provider query parameter is required' }, { status: 400 });
  }

  try {
    const response = await callControlPlane(`/v1/domains/${encodeURIComponent(domain)}/records?provider=${encodeURIComponent(provider)}`, {
      method: 'GET',
    });
    const payload = await response.text();
    return new NextResponse(payload, {
      status: response.status,
      headers: { 'Content-Type': 'application/json' },
    });
  } catch (error) {
    if (error instanceof ControlPlaneUnavailableError) {
      try {
        return NextResponse.json({
          records: listLocalDomainRecords(domain, provider),
          source: 'local-fallback',
          warning: error.message,
        });
      } catch (fallbackError) {
        const message = fallbackError instanceof Error ? fallbackError.message : 'Failed to list domain records';
        return NextResponse.json({ error: message }, { status: 404 });
      }
    }
    const message = error instanceof Error ? error.message : 'Failed to list domain records';
    return NextResponse.json({ error: message }, { status: 500 });
  }
}

export async function POST(request: Request, context: { params: Promise<{ domain: string }> }) {
  const { domain } = await context.params;
  const url = new URL(request.url);
  const provider = url.searchParams.get('provider');
  if (!provider) {
    return NextResponse.json({ error: 'provider query parameter is required' }, { status: 400 });
  }
  const body = await request.text();

  try {
    const response = await callControlPlane(`/v1/domains/${encodeURIComponent(domain)}/records?provider=${encodeURIComponent(provider)}`, {
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
      try {
        const input = (body ? JSON.parse(body) : {}) as {
          type?: string;
          name?: string;
          content?: string;
          ttl?: number;
          proxied?: boolean;
        };
        const record = createLocalDomainRecord(domain, provider, input);
        return NextResponse.json({ record, source: 'local-fallback', warning: error.message }, { status: 201 });
      } catch (fallbackError) {
        const message = fallbackError instanceof Error ? fallbackError.message : 'Failed to create domain record';
        return NextResponse.json({ error: message }, { status: 400 });
      }
    }
    const message = error instanceof Error ? error.message : 'Failed to create domain record';
    return NextResponse.json({ error: message }, { status: 500 });
  }
}

import { NextResponse } from 'next/server';
import { callControlPlane, ControlPlaneUnavailableError } from '@/lib/domains-control-plane';
import { deleteLocalDomainRecord, updateLocalDomainRecord } from '@/lib/domains-local-store';

export async function PATCH(
  request: Request,
  context: { params: Promise<{ domain: string; recordId: string }> },
) {
  const { domain, recordId } = await context.params;
  const url = new URL(request.url);
  const provider = url.searchParams.get('provider');
  if (!provider) {
    return NextResponse.json({ error: 'provider query parameter is required' }, { status: 400 });
  }
  const body = await request.text();

  try {
    const response = await callControlPlane(
      `/v1/domains/${encodeURIComponent(domain)}/records/${encodeURIComponent(recordId)}?provider=${encodeURIComponent(provider)}`,
      { method: 'PATCH', body },
    );
    const payload = await response.text();
    return new NextResponse(payload, {
      status: response.status,
      headers: { 'Content-Type': 'application/json' },
    });
  } catch (error) {
    if (error instanceof ControlPlaneUnavailableError) {
      try {
        const parsed = body ? JSON.parse(body) : {};
        const record = updateLocalDomainRecord(domain, provider, recordId, parsed);
        return NextResponse.json({ record, source: 'local-fallback', warning: error.message });
      } catch (fallbackError) {
        const message = fallbackError instanceof Error ? fallbackError.message : 'Failed to update domain record';
        return NextResponse.json({ error: message }, { status: 400 });
      }
    }
    const message = error instanceof Error ? error.message : 'Failed to update domain record';
    return NextResponse.json({ error: message }, { status: 500 });
  }
}

export async function DELETE(
  request: Request,
  context: { params: Promise<{ domain: string; recordId: string }> },
) {
  const { domain, recordId } = await context.params;
  const url = new URL(request.url);
  const provider = url.searchParams.get('provider');
  if (!provider) {
    return NextResponse.json({ error: 'provider query parameter is required' }, { status: 400 });
  }

  try {
    const response = await callControlPlane(
      `/v1/domains/${encodeURIComponent(domain)}/records/${encodeURIComponent(recordId)}?provider=${encodeURIComponent(provider)}`,
      { method: 'DELETE' },
    );
    const payload = await response.text();
    return new NextResponse(payload, {
      status: response.status,
      headers: { 'Content-Type': 'application/json' },
    });
  } catch (error) {
    if (error instanceof ControlPlaneUnavailableError) {
      const deleted = deleteLocalDomainRecord(domain, provider, recordId);
      if (!deleted) {
        return NextResponse.json({ error: `record "${recordId}" not found` }, { status: 404 });
      }
      return NextResponse.json({ deleted: true, source: 'local-fallback', warning: error.message });
    }
    const message = error instanceof Error ? error.message : 'Failed to delete domain record';
    return NextResponse.json({ error: message }, { status: 500 });
  }
}

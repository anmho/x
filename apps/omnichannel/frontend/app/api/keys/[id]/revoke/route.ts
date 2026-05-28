import { NextResponse } from 'next/server';

const ACCESS_API_BASE_URL = (process.env.ACCESS_API_URL || 'http://127.0.0.1:8090').replace(/\/$/, '');
const ACCESS_API_ADMIN_KEY = process.env.ACCESS_API_ADMIN_KEY || 'x-admin-dev-key';

async function accessAPI(path: string, init?: RequestInit) {
  return fetch(`${ACCESS_API_BASE_URL}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': ACCESS_API_ADMIN_KEY,
      ...(init?.headers || {}),
    },
    cache: 'no-store',
  });
}

export async function POST(_request: Request, context: { params: Promise<{ id: string }> }) {
  try {
    const { id } = await context.params;
    if (!id) {
      return NextResponse.json({ error: 'key id is required' }, { status: 400 });
    }

    const response = await accessAPI(`/v1/keys/${encodeURIComponent(id)}/revoke`, {
      method: 'POST',
    });
    const payload = await response.text();
    return new NextResponse(payload, {
      status: response.status,
      headers: { 'Content-Type': 'application/json' },
    });
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Failed to revoke key';
    return NextResponse.json({ error: message }, { status: 500 });
  }
}


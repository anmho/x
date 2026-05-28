import { NextResponse } from 'next/server';
import { MintAccessKeyRequest } from '@/lib/access-keys';

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

export async function GET() {
  try {
    const response = await accessAPI('/v1/keys', { method: 'GET' });
    const payload = await response.text();
    return new NextResponse(payload, {
      status: response.status,
      headers: { 'Content-Type': 'application/json' },
    });
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Failed to list keys';
    return NextResponse.json({ error: message }, { status: 500 });
  }
}

export async function POST(request: Request) {
  try {
    const body = (await request.json()) as MintAccessKeyRequest;
    if (!body.application || !body.owner || !body.environment || !Array.isArray(body.service_scopes) || body.service_scopes.length === 0) {
      return NextResponse.json(
        { error: 'application, owner, environment, and service_scopes are required' },
        { status: 400 },
      );
    }

    const response = await accessAPI('/v1/keys', {
      method: 'POST',
      body: JSON.stringify(body),
    });
    const payload = await response.text();
    return new NextResponse(payload, {
      status: response.status,
      headers: { 'Content-Type': 'application/json' },
    });
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Failed to mint key';
    return NextResponse.json({ error: message }, { status: 500 });
  }
}


import { NextResponse } from 'next/server';
import { MintAccessKeyRequest } from '@/lib/access-keys';
import { cookies } from 'next/headers';
import { isConsoleAuthDisabled, AUTH_COOKIE_NAME } from '@/lib/auth';
import { accessAPI } from '@/lib/access-api';

export async function GET() {
  if (!isConsoleAuthDisabled()) {
    const cookieStore = await cookies();
    const authCookie = cookieStore.get(AUTH_COOKIE_NAME);
    if (authCookie?.value !== '1') {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
  }
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
  if (!isConsoleAuthDisabled()) {
    const cookieStore = await cookies();
    const authCookie = cookieStore.get(AUTH_COOKIE_NAME);
    if (authCookie?.value !== '1') {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
  }
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


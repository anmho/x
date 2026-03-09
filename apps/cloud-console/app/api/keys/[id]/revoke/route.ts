import { NextResponse } from 'next/server';
import { cookies } from 'next/headers';
import { isConsoleAuthDisabled, AUTH_COOKIE_NAME } from '@/lib/auth';
import { accessAPI } from '@/lib/access-api';

export async function POST(_request: Request, context: { params: Promise<{ id: string }> }) {
  if (!isConsoleAuthDisabled()) {
    const cookieStore = await cookies();
    const authCookie = cookieStore.get(AUTH_COOKIE_NAME);
    if (authCookie?.value !== '1') {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
  }
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


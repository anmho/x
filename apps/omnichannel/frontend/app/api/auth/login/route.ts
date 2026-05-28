import { NextResponse } from 'next/server';
import { AUTH_COOKIE_NAME, getConsolePassword } from '@/lib/auth';

export async function POST(request: Request) {
  try {
    const body = (await request.json()) as { password?: string };
    const password = (body.password || '').trim();

    if (!password || password !== getConsolePassword()) {
      return NextResponse.json({ error: 'Invalid password' }, { status: 401 });
    }

    const response = NextResponse.json({ ok: true });
    response.cookies.set({
      name: AUTH_COOKIE_NAME,
      value: '1',
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'lax',
      path: '/',
      maxAge: 60 * 60 * 8,
    });

    return response;
  } catch {
    return NextResponse.json({ error: 'Invalid request' }, { status: 400 });
  }
}

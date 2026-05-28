import { NextResponse } from 'next/server';
import { AUTH_COOKIE_NAME } from '@/lib/auth';

function clearAuthCookie(response: NextResponse) {
  response.cookies.set({
    name: AUTH_COOKIE_NAME,
    value: '',
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    path: '/',
    maxAge: 0,
  });
  return response;
}

export async function POST() {
  return clearAuthCookie(NextResponse.json({ ok: true }));
}

export async function GET(request: Request) {
  return clearAuthCookie(NextResponse.redirect(new URL('/auth/login', request.url)));
}

import type { NextRequest } from 'next/server';
import { NextResponse } from 'next/server';
import { AUTH_COOKIE_NAME, isConsoleAuthDisabled } from '@/lib/auth';

const PUBLIC_PATHS = ['/auth/login', '/api/auth/login', '/api/auth/logout'];

function isPublicPath(pathname: string) {
  return PUBLIC_PATHS.some((path) => pathname === path || pathname.startsWith(`${path}/`));
}

function isProtectedPath(pathname: string) {
  return (
    pathname === '/' ||
    pathname.startsWith('/deployments') ||
    pathname.startsWith('/workflows') ||
    pathname.startsWith('/projects') ||
    pathname.startsWith('/applications') ||
    pathname.startsWith('/domains') ||
    pathname.startsWith('/c/') ||
    pathname.startsWith('/api-keys') ||
    pathname.startsWith('/console') ||
    pathname.startsWith('/services') ||
    pathname.startsWith('/templates') ||
    pathname.startsWith('/notifications') ||
    (pathname.startsWith('/api/') && !pathname.startsWith('/api/auth/'))
  );
}

export function proxy(request: NextRequest) {
  const { pathname, search } = request.nextUrl;

  if (pathname.startsWith('/_next') || pathname.startsWith('/favicon.ico')) {
    return NextResponse.next();
  }

  if (isConsoleAuthDisabled()) {
    if (pathname === '/auth/login') {
      return NextResponse.redirect(new URL('/', request.url));
    }
    return NextResponse.next();
  }

  const isAuthenticated = request.cookies.get(AUTH_COOKIE_NAME)?.value === '1';

  if (isPublicPath(pathname)) {
    if (pathname === '/auth/login' && isAuthenticated) {
      return NextResponse.redirect(new URL('/', request.url));
    }
    return NextResponse.next();
  }

  if (isProtectedPath(pathname) && !isAuthenticated) {
    const loginUrl = new URL('/auth/login', request.url);
    loginUrl.searchParams.set('next', `${pathname}${search}`);
    return NextResponse.redirect(loginUrl);
  }

  return NextResponse.next();
}

export const config = {
  matcher: ['/((?!_next/static|_next/image|favicon.ico).*)'],
};

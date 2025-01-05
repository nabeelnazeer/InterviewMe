import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

// Export as default function to fix the middleware error
export default function middleware(request: NextRequest) {
  if (request.nextUrl.pathname === '/') {
    return NextResponse.redirect(new URL('/cvScoring', request.url));
  }

  return NextResponse.next();
}

// Update matcher configuration
export const config = {
  matcher: ['/((?!api|_next/static|_next/image|favicon.ico).*)']
};

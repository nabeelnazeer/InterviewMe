import { type NextRequest, NextResponse } from 'next/server';

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // Guard analysis page
  if (pathname === '/analysis') {
    // Check if we have required data in the request (you might need to adjust this)
    const hasData = request.cookies.has('hasAnalysisData') || 
                   request.headers.get('x-has-analysis-data');
    
    if (!hasData) {
      console.log('Redirecting from analysis to cvScoring - no data found');
      return NextResponse.redirect(new URL('/cvScoring', request.url));
    }
  }

  // Add basic path protection
  if (pathname === '/') {
    return NextResponse.redirect(new URL('/cvScoring', request.url));
  }

  // Add response headers for better security
  const response = NextResponse.next();
  response.headers.set('x-middleware-cache', 'no-cache');
  response.headers.set('x-frame-options', 'DENY');
  response.headers.set('x-content-type-options', 'nosniff');

  return response;
}

// Configure matcher
export const config = {
  matcher: [
    /*
     * Match all paths except static files and api routes
     */
    '/((?!api|_next/static|_next/image|favicon.ico).*)',
  ],
};

import type { NextRequest } from 'next/server';
import { NextResponse } from 'next/server';

import { getCanonicalLocalhostUrl } from '@/lib/local-origin';

export function middleware(request: NextRequest) {
  const canonicalLocalhostUrl = getCanonicalLocalhostUrl(request.url);

  if (canonicalLocalhostUrl) {
    return NextResponse.redirect(canonicalLocalhostUrl, 307);
  }

  return NextResponse.next();
}

export const config = {
  matcher: ['/((?!_next/static|_next/image|favicon.ico).*)'],
};

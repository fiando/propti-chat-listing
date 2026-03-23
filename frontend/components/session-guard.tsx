'use client';

import { useEffect } from 'react';
import { signIn, useSession } from 'next-auth/react';

/**
 * Watches the current NextAuth session for token-refresh errors propagated
 * from the server-side JWT callback (e.g. expired Google refresh token) and
 * forces a fresh Google sign-in when recovery is no longer possible.
 *
 * Render this component once inside the SessionProvider tree (see providers.tsx).
 * It does not render any visible UI.
 */
export function SessionGuard() {
  const { data: session } = useSession();
  const error = (session as { error?: string } | null)?.error;

  useEffect(() => {
    if (error === 'RefreshAccessTokenError' || error === 'MissingRefreshToken') {
      signIn('google');
    }
  }, [error]);

  return null;
}

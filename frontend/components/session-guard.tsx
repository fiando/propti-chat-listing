'use client';

import { useEffect } from 'react';
import { signOut, useSession } from 'next-auth/react';

/**
 * Watches the current NextAuth session for token-refresh errors propagated
 * from the server-side JWT callback (e.g. expired Google refresh token) and
 * clears the invalid session when recovery is no longer possible so the user
 * can continue browsing public pages and choose to sign in manually.
 *
 * Render this component once inside the SessionProvider tree (see providers.tsx).
 * It does not render any visible UI.
 */
export function SessionGuard() {
  const { data: session } = useSession();
  const error = (session as { error?: string } | null)?.error;

  useEffect(() => {
    if (error === 'RefreshAccessTokenError' || error === 'MissingRefreshToken') {
      void signOut({ redirect: false });
    }
  }, [error]);

  return null;
}

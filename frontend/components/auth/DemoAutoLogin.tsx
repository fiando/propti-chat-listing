'use client';

import { signIn, useSession } from 'next-auth/react';
import { useEffect, useRef } from 'react';
import { buildPostLoginCallbackUrl } from '@/lib/auth-callback';

/**
 * DemoAutoLogin — invisible component that silently signs in via the demo
 * credentials provider when NEXT_PUBLIC_DEMO_MODE=true.
 *
 * Renders nothing. Fires once on mount if the user is unauthenticated.
 * Remove this component (or set NEXT_PUBLIC_DEMO_MODE=false) to disable.
 */
export function DemoAutoLogin({ callbackUrl = '/' }: { callbackUrl?: string }) {
  const { status } = useSession();
  const triggered = useRef(false);

  useEffect(() => {
    if (
      process.env.NEXT_PUBLIC_DEMO_MODE !== 'true' ||
      status === 'authenticated' ||
      status === 'loading' ||
      triggered.current
    ) {
      return;
    }

    triggered.current = true;
    signIn('credentials', {
      email: 'demo@propti.app',
      password: 'demo',
      callbackUrl: buildPostLoginCallbackUrl(callbackUrl),
      redirect: true,
    });
  }, [status, callbackUrl]);

  return null;
}

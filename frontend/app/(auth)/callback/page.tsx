'use client';

import { useEffect, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useSession } from 'next-auth/react';
import { Loader2 } from 'lucide-react';
import { getPostLoginDecision, getSafeAuthCallbackUrl } from '@/lib/auth-callback';

const MAX_SESSION_SYNC_ATTEMPTS = 4;
const SESSION_SYNC_RETRY_MS = 750;

export default function CallbackPage() {
  const { status, update } = useSession();
  const router = useRouter();
  const searchParams = useSearchParams();
  const [syncAttempt, setSyncAttempt] = useState(0);
  const callbackUrl = getSafeAuthCallbackUrl(searchParams.get('callbackUrl'));

  useEffect(() => {
    let cancelled = false;
    let retryTimer: number | undefined;

    const navigateTo = (targetUrl: string) => {
      router.refresh();
      window.location.replace(targetUrl);
    };

    const syncSessionAndRedirect = async () => {
      const decision = getPostLoginDecision({
        status,
        syncAttempt,
        maxSyncAttempts: MAX_SESSION_SYNC_ATTEMPTS,
      });

      if (decision === 'redirect-target') {
        navigateTo(callbackUrl);
        return;
      }

      if (decision === 'redirect-login') {
        navigateTo(`/login?callbackUrl=${encodeURIComponent(callbackUrl)}`);
        return;
      }

      if (status === 'loading') {
        return;
      }

      const refreshedSession = await update();
      if (cancelled) {
        return;
      }

      if (refreshedSession) {
        navigateTo(callbackUrl);
        return;
      }

      retryTimer = window.setTimeout(() => {
        setSyncAttempt((currentAttempt) => currentAttempt + 1);
      }, SESSION_SYNC_RETRY_MS);
    };

    void syncSessionAndRedirect();

    return () => {
      cancelled = true;
      if (retryTimer) {
        window.clearTimeout(retryTimer);
      }
    };
  }, [callbackUrl, router, status, syncAttempt, update]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-[#F8F9FA]">
      <div className="text-center">
        <Loader2 className="w-12 h-12 text-brand-primary animate-spin mx-auto mb-4" />
        <p className="text-gray-600 font-medium">Sedang masuk ke Propti...</p>
      </div>
    </div>
  );
}

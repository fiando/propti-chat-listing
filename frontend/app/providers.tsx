'use client';

import { SessionProvider } from 'next-auth/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useState } from 'react';
import { Toaster } from './toaster';
import { SessionGuard } from '@/components/session-guard';

// Refetch the session every 4 minutes so the JWT callback can refresh
// the backend token before it expires.
const SESSION_REFETCH_INTERVAL_SECONDS = 4 * 60;

export function Providers({ children }: { children: React.ReactNode }) {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 60 * 1000,
            retry: 1,
          },
        },
      })
  );

  return (
    <SessionProvider
      refetchInterval={SESSION_REFETCH_INTERVAL_SECONDS}
      refetchOnWindowFocus
      refetchWhenOffline={false}
    >
      <QueryClientProvider client={queryClient}>
        <SessionGuard />
        <Toaster>{children}</Toaster>
      </QueryClientProvider>
    </SessionProvider>
  );
}

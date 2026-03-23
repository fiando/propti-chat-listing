export const AUTH_PATHS = ['/login', '/callback', '/api/auth'];

type PostLoginDecisionInput = {
  status: 'loading' | 'authenticated' | 'unauthenticated';
  syncAttempt: number;
  maxSyncAttempts: number;
};

export function getSafeAuthCallbackUrl(rawCallbackUrl?: string | null): string {
  const callbackUrl = rawCallbackUrl || '/';

  return callbackUrl.startsWith('/') && !AUTH_PATHS.some((path) => callbackUrl.startsWith(path))
    ? callbackUrl
    : '/';
}

export function buildPostLoginCallbackUrl(rawCallbackUrl?: string | null): string {
  const callbackUrl = getSafeAuthCallbackUrl(rawCallbackUrl);

  return `/callback?callbackUrl=${encodeURIComponent(callbackUrl)}`;
}

export function getPostLoginDecision({
  status,
  syncAttempt,
  maxSyncAttempts,
}: PostLoginDecisionInput): 'wait' | 'redirect-target' | 'redirect-login' {
  if (status === 'authenticated') {
    return 'redirect-target';
  }

  if (status === 'loading') {
    return 'wait';
  }

  return syncAttempt < maxSyncAttempts ? 'wait' : 'redirect-login';
}

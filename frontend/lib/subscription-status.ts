import type { User } from '../types';

export type AuthStatus = 'loading' | 'authenticated' | 'unauthenticated';
export type SubscriptionStatus = 'loading' | 'free' | 'premium';

interface GetSubscriptionStatusInput {
  authStatus: AuthStatus;
  profile?: User;
  now?: number;
}

export function getSubscriptionStatus({
  authStatus,
  profile,
  now = Date.now(),
}: GetSubscriptionStatusInput): SubscriptionStatus {
  if (authStatus === 'loading') {
    return 'loading';
  }

  if (authStatus === 'authenticated' && !profile) {
    return 'loading';
  }

  const renewDate = profile?.subscription?.renewDate ? new Date(profile.subscription.renewDate) : null;
  const hasActiveRenewDate = renewDate
    ? Number.isFinite(renewDate.getTime()) && renewDate.getTime() > now
    : true;

  return profile?.subscription?.tier === 'premium' && hasActiveRenewDate ? 'premium' : 'free';
}

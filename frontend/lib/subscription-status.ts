import type { User, SubscriptionStatus } from '../types';

export type AuthStatus = 'loading' | 'authenticated' | 'unauthenticated';

export type { SubscriptionStatus };

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

  if (profile?.subscription?.tier !== 'premium') {
    return 'free';
  }

  // If the API returned a pre-computed status, trust it
  if (profile.subscriptionStatus && profile.subscriptionStatus !== 'loading') {
    return profile.subscriptionStatus;
  }

  return getDerivedStatus(profile.subscription.renewDate, now);
}

function getDerivedStatus(renewDateStr: string | undefined, now: number): SubscriptionStatus {
  if (!renewDateStr) {
    return 'expired';
  }
  const renewDate = new Date(renewDateStr);
  if (!Number.isFinite(renewDate.getTime())) {
    return 'expired';
  }
  const msRemaining = renewDate.getTime() - now;
  if (msRemaining < 0) {
    return 'expired';
  }
  if (msRemaining <= 7 * 24 * 60 * 60 * 1000) {
    return 'expiring_soon';
  }
  return 'active';
}

export function getDaysUntilExpiry(renewDate: string | undefined, now = Date.now()): number | null {
  if (!renewDate) return null;
  const d = new Date(renewDate);
  if (!Number.isFinite(d.getTime())) return null;
  const ms = d.getTime() - now;
  return Math.ceil(ms / (24 * 60 * 60 * 1000));
}

export function getExpiryMessage(status: SubscriptionStatus, renewDate: string | undefined): string {
  switch (status) {
    case 'active': {
      const days = getDaysUntilExpiry(renewDate);
      if (days === null) return 'Premium aktif';
      const date = renewDate ? new Date(renewDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' }) : '';
      return `Premium aktif sampai ${date}`;
    }
    case 'expiring_soon': {
      const days = getDaysUntilExpiry(renewDate);
      if (days === null || days <= 0) return 'Premium berakhir hari ini';
      return `Premium berakhir dalam ${days} hari`;
    }
    case 'expired':
      return 'Premium telah berakhir';
    case 'free':
      return 'Paket gratis';
    default:
      return '';
  }
}

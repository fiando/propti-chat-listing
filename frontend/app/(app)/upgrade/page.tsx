import type { Metadata } from 'next';
import { redirect } from 'next/navigation';
import { UpgradePageClient } from '@/components/upgrade/UpgradePageClient';
import { getServerAuthProfile } from '@/lib/server-profile';

export const metadata: Metadata = {
  title: 'Upgrade Paket — Propti',
};

export default async function UpgradePage({
  searchParams,
}: {
  searchParams: Promise<{ tier?: string }>;
}) {
  const { isAuthenticated, profile } = await getServerAuthProfile();

  const params = await searchParams;
  const rawTier = params?.tier ?? '';
  const validTiers = ['basic', 'premium', 'pro'] as const;
  type PaidTier = (typeof validTiers)[number];
  const initialTier: PaidTier = validTiers.includes(rawTier as PaidTier)
    ? (rawTier as PaidTier)
    : 'premium';

  if (!isAuthenticated) {
    const callbackUrl = `/upgrade?tier=${initialTier}`;
    redirect(`/login?callbackUrl=${encodeURIComponent(callbackUrl)}`);
  }

  return (
    <UpgradePageClient
      isAuthenticated={isAuthenticated}
      currentTier={profile?.subscription?.tier ?? 'free'}
      initialTier={initialTier}
    />
  );
}

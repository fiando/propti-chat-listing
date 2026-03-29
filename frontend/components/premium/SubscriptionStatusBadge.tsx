'use client';

import type { SubscriptionStatus } from '@/types';
import { getExpiryMessage } from '@/lib/subscription-status';

interface SubscriptionStatusBadgeProps {
  status: SubscriptionStatus;
  renewDate?: string;
  tier?: 'free' | 'basic' | 'premium' | 'pro';
  className?: string;
}

const statusStyles: Record<string, string> = {
  active: 'bg-emerald-50 text-emerald-700 border-emerald-200',
  expiring_soon: 'bg-amber-50 text-amber-700 border-amber-200',
  expired: 'bg-red-50 text-red-700 border-red-200',
  free: 'bg-gray-50 text-gray-600 border-gray-200',
  loading: 'bg-gray-50 text-gray-400 border-gray-200',
};

export function SubscriptionStatusBadge({ status, renewDate, tier = 'free', className = '' }: SubscriptionStatusBadgeProps) {
  const message = getExpiryMessage(status, renewDate);
  const tierLabel = tier === 'free' ? 'Gratis' : tier === 'basic' ? 'Basic' : tier === 'premium' ? 'Premium' : 'Pro';

  return (
    <span
      className={`inline-flex items-center rounded-full border px-3 py-1 text-sm font-medium ${statusStyles[status] ?? statusStyles.free} ${className}`}
    >
      {tierLabel}: {message}
    </span>
  );
}

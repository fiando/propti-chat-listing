import type { SubscriptionStatus } from '../types';

export function shouldShowRenewalCTA(status: SubscriptionStatus): boolean {
  return status === 'expiring_soon' || status === 'expired';
}

interface RenewalUXCopy {
  heading: string;
  body: string;
  ctaText: string;
}

export function getRenewalUXCopy(status: SubscriptionStatus, renewDate: string | undefined): RenewalUXCopy {
  if (status === 'expired') {
    return {
      heading: 'Premium Berakhir',
      body: 'Akses premiummu telah berakhir. Perpanjang sekarang untuk melanjutkan fitur premium.',
      ctaText: 'Perpanjang Premium',
    };
  }

  if (status === 'expiring_soon') {
    const dateStr = renewDate
      ? new Date(renewDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' })
      : '';
    return {
      heading: 'Premium Segera Berakhir',
      body: `Premiummu berakhir pada ${dateStr}. Perpanjang sekarang agar tidak ada hari yang terbuang.`,
      ctaText: 'Perpanjang Premium',
    };
  }

  return {
    heading: 'Upgrade ke Premium',
    body: 'Dapatkan akses fitur premium untuk memaksimalkan listing propertimu.',
    ctaText: 'Upgrade ke Premium',
  };
}

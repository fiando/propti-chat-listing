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
      heading: 'Paket Berbayar Berakhir',
      body: 'Akses paket berbayarmu telah berakhir. Perpanjang sekarang untuk melanjutkan semua fitur berbayar.',
      ctaText: 'Perpanjang Paket',
    };
  }

  if (status === 'expiring_soon') {
    const dateStr = renewDate
      ? new Date(renewDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' })
      : '';
    return {
      heading: 'Paket Segera Berakhir',
      body: `Paketmu berakhir pada ${dateStr}. Perpanjang sekarang agar tidak ada hari yang terbuang.`,
      ctaText: 'Perpanjang Paket',
    };
  }

  return {
    heading: 'Upgrade Paket',
    body: 'Dapatkan akses fitur berbayar untuk memaksimalkan listing propertimu.',
    ctaText: 'Upgrade Paket',
  };
}

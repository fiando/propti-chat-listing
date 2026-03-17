import type { Metadata } from 'next';
import { InfoPageLayout } from '@/components/common/InfoPageLayout';
import { contactSections } from './content';

export const metadata: Metadata = {
  title: 'Hubungi Kami',
  description: 'Panduan singkat untuk menghubungi tim Propti sesuai kebutuhanmu.',
};

export default function ContactPage() {
  return (
    <InfoPageLayout
      eyebrow="Hubungi Propti"
      title="Kami siap membantu kebutuhanmu"
      description="Gunakan halaman ini sebagai panduan singkat ketika kamu ingin menghubungi tim Propti untuk pertanyaan produk, masukan, atau peluang kerja sama."
      sections={contactSections}
      primaryCta={{ label: 'Masuk ke Akun', href: '/login' }}
      secondaryCta={{ label: 'Lihat Listing', href: '/search' }}
    />
  );
}

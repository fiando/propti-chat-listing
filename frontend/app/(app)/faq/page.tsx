import type { Metadata } from 'next';
import { InfoPageLayout } from '@/components/common/InfoPageLayout';

export const metadata: Metadata = {
  title: 'FAQ',
  description: 'Pertanyaan yang paling sering ditanyakan tentang penggunaan Propti.',
};

const sections = [
  {
    title: 'Bagaimana cara pasang iklan di Propti?',
    paragraphs: [
      'Masuk ke akun Propti, buka halaman pasang iklan, lalu isi data properti yang kamu miliki. Setelah dikirim, iklan akan melalui proses peninjauan sebelum tampil ke publik.',
    ],
  },
  {
    title: 'Apakah Propti gratis digunakan?',
    paragraphs: [
      'Ya, Propti bisa digunakan secara gratis untuk mulai memasarkan properti. Beberapa fitur tambahan tersedia di paket berbayar sesuai kebutuhan.',
    ],
  },
  {
    title: 'Kenapa iklan saya belum tampil?',
    paragraphs: [
      'Setelah dikirim atau diubah, iklan bisa masuk proses peninjauan terlebih dahulu. Ini dilakukan agar konten yang tampil ke publik tetap relevan, aman, dan memenuhi standar kualitas Propti.',
    ],
  },
  {
    title: 'Bisakah saya mengubah iklan setelah dipublikasikan?',
    paragraphs: [
      'Bisa. Saat iklan diperbarui, sistem dapat meninjau ulang konten tersebut sebelum perubahan ditampilkan ke publik.',
    ],
  },
  {
    title: 'Bagaimana AI membantu di Propti?',
    paragraphs: [
      'AI digunakan untuk membantu merapikan informasi iklan agar lebih terstruktur dan mudah dibaca. Sistem juga membantu proses peninjauan konten untuk menjaga kualitas platform.',
    ],
  },
];

export default function FaqPage() {
  return (
    <InfoPageLayout
      eyebrow="Pusat Bantuan"
      title="Pertanyaan yang sering ditanyakan"
      description="Kalau kamu baru mulai menggunakan Propti, halaman ini menjawab pertanyaan paling umum seputar pemasangan iklan, proses peninjauan, dan penggunaan platform."
      sections={sections}
      primaryCta={{ label: 'Pasang Iklan', href: '/listings/create' }}
      secondaryCta={{ label: 'Cari Properti', href: '/search' }}
    />
  );
}

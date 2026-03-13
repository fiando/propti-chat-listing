import type { Metadata } from 'next';
import { InfoPageLayout } from '@/components/common/InfoPageLayout';

export const metadata: Metadata = {
  title: 'Kebijakan Privasi',
  description: 'Ringkasan cara Propti mengelola data pengguna dan informasi listing.',
};

const sections = [
  {
    title: 'Data yang kami gunakan',
    paragraphs: [
      'Kami menggunakan data yang kamu berikan saat membuat akun, mengelola profil, dan memasang listing agar layanan Propti dapat berjalan dengan baik.',
      'Data ini dapat mencakup identitas akun, isi listing, media yang diunggah, serta interaksi dasar di dalam platform.',
    ],
  },
  {
    title: 'Tujuan penggunaan data',
    paragraphs: [
      'Data dipakai untuk mengoperasikan layanan, meningkatkan kualitas pengalaman pengguna, menjaga keamanan, dan membantu proses peninjauan konten.',
    ],
    bullets: [
      'Menampilkan dan mengelola listing properti.',
      'Memproses autentikasi dan aktivitas akun.',
      'Mendukung moderasi, pencegahan penyalahgunaan, dan peningkatan produk.',
    ],
  },
  {
    title: 'Perlindungan data',
    paragraphs: [
      'Kami berupaya menjaga keamanan data dengan kontrol teknis dan operasional yang wajar. Meski demikian, tidak ada sistem yang dapat menjamin keamanan absolut.',
    ],
  },
  {
    title: 'Perubahan kebijakan',
    paragraphs: [
      'Kebijakan privasi ini dapat diperbarui sewaktu-waktu untuk menyesuaikan perubahan produk, praktik operasional, atau kebutuhan hukum yang berlaku.',
    ],
  },
];

export default function PrivacyPage() {
  return (
    <InfoPageLayout
      eyebrow="Privasi"
      title="Kebijakan Privasi Propti"
      description="Kami berkomitmen menggunakan data secara bertanggung jawab untuk menjalankan layanan, menjaga keamanan platform, dan meningkatkan pengalaman pengguna."
      sections={sections}
      primaryCta={{ label: 'Syarat & Ketentuan', href: '/terms' }}
      secondaryCta={{ label: 'Hubungi Kami', href: '/contact' }}
    />
  );
}

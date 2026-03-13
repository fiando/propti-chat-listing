import type { Metadata } from 'next';
import { InfoPageLayout } from '@/components/common/InfoPageLayout';

export const metadata: Metadata = {
  title: 'Hubungi Kami',
  description: 'Panduan singkat untuk menghubungi tim Propti sesuai kebutuhanmu.',
};

const sections = [
  {
    title: 'Bantuan penggunaan platform',
    paragraphs: [
      'Jika kamu membutuhkan bantuan terkait akun, pemasangan listing, atau proses peninjauan, siapkan detail kendala secara singkat agar tim kami bisa memahami konteks dengan cepat.',
    ],
    bullets: [
      'Sertakan nama akun atau email yang digunakan.',
      'Tambahkan tautan atau ID listing jika pertanyaanmu terkait iklan tertentu.',
      'Jelaskan kapan masalah terjadi dan apa yang sudah kamu coba.',
    ],
  },
  {
    title: 'Kerja sama dan pertanyaan bisnis',
    paragraphs: [
      'Untuk partnership, kolaborasi komunitas, atau pertanyaan bisnis lainnya, kirimkan ringkasan kebutuhan, target kerja sama, dan cakupan wilayah agar diskusi bisa lebih efektif.',
    ],
  },
  {
    title: 'Waktu respons',
    paragraphs: [
      'Kami meninjau pesan yang masuk secara berkala pada hari kerja. Semakin lengkap detail yang kamu berikan, semakin cepat tim kami dapat menindaklanjuti.',
    ],
  },
];

export default function ContactPage() {
  return (
    <InfoPageLayout
      eyebrow="Hubungi Propti"
      title="Kami siap membantu kebutuhanmu"
      description="Gunakan halaman ini sebagai panduan singkat ketika kamu ingin menghubungi tim Propti untuk pertanyaan produk, masukan, atau peluang kerja sama."
      sections={sections}
      primaryCta={{ label: 'Masuk ke Akun', href: '/login' }}
      secondaryCta={{ label: 'Lihat Listing', href: '/search' }}
    />
  );
}

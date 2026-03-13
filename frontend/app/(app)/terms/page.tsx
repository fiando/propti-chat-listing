import type { Metadata } from 'next';
import { InfoPageLayout } from '@/components/common/InfoPageLayout';

export const metadata: Metadata = {
  title: 'Syarat & Ketentuan',
  description: 'Ringkasan syarat penggunaan layanan Propti.',
};

const sections = [
  {
    title: 'Penggunaan layanan',
    paragraphs: [
      'Dengan menggunakan Propti, kamu setuju untuk memakai layanan ini secara sah, wajar, dan sesuai tujuan platform sebagai sarana pencarian serta pemasaran properti.',
      'Kamu bertanggung jawab atas informasi, foto, dan materi lain yang kamu unggah ke dalam platform.',
    ],
  },
  {
    title: 'Akurasi informasi listing',
    paragraphs: [
      'Kamu wajib memastikan bahwa detail properti yang dipublikasikan akurat, tidak menyesatkan, dan memiliki hak untuk dipasarkan.',
    ],
    bullets: [
      'Tidak boleh memuat informasi palsu, menipu, atau merugikan pihak lain.',
      'Tidak boleh menggunakan konten yang melanggar hukum atau hak pihak ketiga.',
      'Tidak boleh menyalahgunakan platform untuk spam atau aktivitas yang tidak relevan dengan properti.',
    ],
  },
  {
    title: 'Peninjauan dan moderasi',
    paragraphs: [
      'Propti dapat meninjau, menahan, menyunting secara terbatas, atau menolak listing tertentu demi menjaga kualitas, keamanan, dan kepatuhan platform.',
      'Status peninjauan dapat berubah ketika listing baru dikirim atau ketika konten listing diperbarui.',
    ],
  },
  {
    title: 'Perubahan layanan',
    paragraphs: [
      'Kami dapat memperbarui fitur, batas penggunaan, atau kebijakan layanan dari waktu ke waktu. Perubahan penting akan berlaku setelah dipublikasikan di platform.',
    ],
  },
];

export default function TermsPage() {
  return (
    <InfoPageLayout
      eyebrow="Dokumen Legal"
      title="Syarat & Ketentuan penggunaan Propti"
      description="Halaman ini merangkum ketentuan dasar yang berlaku saat kamu menggunakan Propti untuk mencari atau memasarkan properti."
      sections={sections}
      primaryCta={{ label: 'Kebijakan Privasi', href: '/privacy' }}
      secondaryCta={{ label: 'FAQ', href: '/faq' }}
    />
  );
}

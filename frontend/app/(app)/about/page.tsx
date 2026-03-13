import type { Metadata } from 'next';
import { InfoPageLayout } from '@/components/common/InfoPageLayout';

export const metadata: Metadata = {
  title: 'Tentang Kami',
  description: 'Kenali Propti, misi kami, dan cara kami membantu pasar properti Indonesia bergerak lebih cepat.',
};

const sections = [
  {
    title: 'Apa itu Propti',
    paragraphs: [
      'Propti adalah platform properti Indonesia yang membantu pemilik, agen, dan pencari properti menemukan kecocokan lebih cepat dengan bantuan AI.',
      'Kami membangun pengalaman yang sederhana: dari iklan yang biasanya tersebar di chat atau pesan singkat menjadi listing yang lebih rapi, mudah dibaca, dan siap dipasarkan.',
    ],
  },
  {
    title: 'Misi kami',
    paragraphs: [
      'Kami ingin membuat proses jual, beli, dan sewa properti terasa sesederhana mengirim pesan. Informasi yang jelas, alur yang cepat, dan pengalaman yang nyaman adalah fokus utama kami.',
    ],
    bullets: [
      'Membantu pengguna mempublikasikan listing lebih cepat.',
      'Membuat informasi properti lebih terstruktur dan mudah dibandingkan.',
      'Menjaga kualitas listing agar pengalaman pencarian tetap aman dan relevan.',
    ],
  },
  {
    title: 'Untuk siapa Propti dibuat',
    paragraphs: [
      'Propti dirancang untuk pemilik rumah, agen properti, developer, pemilik kos, dan siapa pun yang ingin memasarkan properti dengan lebih efisien.',
      'Di sisi lain, pencari properti mendapatkan tampilan listing yang lebih jelas sehingga keputusan bisa dibuat dengan lebih percaya diri.',
    ],
  },
];

export default function AboutPage() {
  return (
    <InfoPageLayout
      eyebrow="Tentang Propti"
      title="Platform properti yang dibuat untuk pasar Indonesia"
      description="Kami percaya proses properti seharusnya tidak rumit. Propti hadir untuk merapikan informasi, mempercepat publikasi, dan membantu pengguna mengambil keputusan dengan lebih mudah."
      sections={sections}
      primaryCta={{ label: 'Cari Properti', href: '/search' }}
      secondaryCta={{ label: 'Pasang Iklan', href: '/listings/create' }}
    />
  );
}

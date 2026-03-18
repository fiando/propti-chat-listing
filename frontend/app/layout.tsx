import type { Metadata } from 'next';
import '../styles/globals.css';
import { Providers } from './providers';

export const metadata: Metadata = {
  title: {
    default: 'Propti — Jual Beli Properti Semudah Chat WhatsApp',
    template: '%s | Propti',
  },
  description:
    'Platform properti Indonesia berbasis AI. Paste teks iklan dari WhatsApp, AI kami rapikan jadi listing rapi, lengkap, dan siap tayang — dalam hitungan detik.',
  keywords: [
    'properti Indonesia',
    'jual rumah',
    'sewa rumah',
    'pasang iklan properti',
    'AI listing properti',
    'iklan rumah WhatsApp',
    'real estate Indonesia',
  ],
  authors: [{ name: 'Propti', url: 'https://propti.id' }],
  creator: 'Propti',
  metadataBase: new URL('https://propti.id'),
  openGraph: {
    type: 'website',
    locale: 'id_ID',
    url: 'https://propti.id',
    siteName: 'Propti',
    title: 'Propti — Jual Beli Properti Semudah Chat WhatsApp',
    description:
      'Paste iklan dari WhatsApp, AI kami ubah jadi listing properti rapi dan siap tayang. Gratis untuk 3 iklan pertama.',
    images: [
      {
        url: '/og-image.jpg',
        width: 1200,
        height: 630,
        alt: 'Propti — Platform Properti AI Indonesia',
      },
    ],
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Propti — Jual Beli Properti Semudah Chat WhatsApp',
    description:
      'Paste iklan dari WhatsApp, AI kami ubah jadi listing properti rapi dan siap tayang.',
    images: ['/og-image.jpg'],
  },
  icons: {
    icon: [
      { url: '/favicon.ico', sizes: 'any' },
      { url: '/favicon.svg', type: 'image/svg+xml' },
      { url: '/favicon-16x16.png', sizes: '16x16', type: 'image/png' },
      { url: '/favicon-32x32.png', sizes: '32x32', type: 'image/png' },
    ],
    apple: [{ url: '/apple-touch-icon.png', sizes: '180x180', type: 'image/png' }],
    other: [
      { rel: 'mask-icon', url: '/favicon.svg', color: '#1B4332' },
    ],
  },
  manifest: '/site.webmanifest',
  themeColor: '#1B4332',
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="id">
      <body className="bg-[#F8F9FA] text-[#1A1A2E] antialiased">
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}

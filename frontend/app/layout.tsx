import type { Metadata, Viewport } from 'next';
import { Analytics } from '@vercel/analytics/next';
import '../styles/globals.css';
import { Providers } from './providers';

export const metadata: Metadata = {
  applicationName: 'Propti',
  title: {
    default: 'Propti — Pasang Listing Properti Langsung dari WhatsApp',
    template: '%s | Propti',
  },
  description:
    'Kirim pesan WhatsApp ke Propti — teks atau voice note — dan AI kami ubah jadi listing properti rapi dan siap tayang dalam hitungan detik.',
  keywords: [
    'properti Indonesia',
    'jual rumah',
    'sewa rumah',
    'pasang iklan properti',
    'AI listing properti',
    'listing properti WhatsApp',
    'buat listing via WhatsApp',
    'pasang properti via WA',
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
    title: 'Propti — Pasang Listing Properti Langsung dari WhatsApp',
    description:
      'Kirim pesan ke WhatsApp Propti, AI parse jadi listing rapi dan siap tayang. Dukung teks dan pesan suara.',
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
    title: 'Propti — Pasang Listing Properti Langsung dari WhatsApp',
    description:
      'Kirim pesan ke WhatsApp Propti, AI parse jadi listing rapi dan siap tayang. Dukung teks dan pesan suara.',
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
  appleWebApp: {
    capable: true,
    title: 'Propti',
    statusBarStyle: 'default',
  },
};

export const viewport: Viewport = {
  width: 'device-width',
  initialScale: 1,
  viewportFit: 'cover',
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
        <Analytics />
      </body>
    </html>
  );
}

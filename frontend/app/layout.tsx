import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import '../styles/globals.css';
import { Providers } from './providers';

const inter = Inter({
  subsets: ['latin'],
  variable: '--font-inter',
  display: 'swap',
});

export const metadata: Metadata = {
  title: {
    default: 'Propti - Jual Beli Properti Semudah Chat WhatsApp',
    template: '%s | Propti',
  },
  description:
    'Platform properti Indonesia yang menggunakan AI untuk merapikan iklan properti dari WhatsApp. Jual, beli, dan sewa properti dengan mudah.',
  keywords: ['properti', 'rumah dijual', 'rumah disewa', 'real estate', 'Indonesia', 'AI'],
  authors: [{ name: 'Propti' }],
  creator: 'Propti',
  openGraph: {
    type: 'website',
    locale: 'id_ID',
    url: 'https://propti.id',
    siteName: 'Propti',
    title: 'Propti - Jual Beli Properti Semudah Chat WhatsApp',
    description:
      'Platform properti Indonesia dengan AI parsing. Paste iklan WhatsApp-mu, kami rapikan otomatis.',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Propti - Jual Beli Properti Semudah Chat WhatsApp',
    description:
      'Platform properti Indonesia dengan AI parsing. Paste iklan WhatsApp-mu, kami rapikan otomatis.',
  },
  icons: {
    icon: '/favicon.ico',
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="id" className={inter.variable}>
      <body className="bg-[#F8F9FA] text-[#1A1A2E] antialiased">
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}

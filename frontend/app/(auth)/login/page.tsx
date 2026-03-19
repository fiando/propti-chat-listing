import type { Metadata } from 'next';
import { GoogleLoginButton } from '@/components/auth/GoogleLoginButton';
import { ProptiLogo } from '@/components/common/ProptiLogo';
import { Home, Shield, Zap } from 'lucide-react';
import Link from 'next/link';

export const metadata: Metadata = {
  title: 'Masuk ke Propti',
  description: 'Login untuk mulai jual beli properti dengan mudah',
};

export default async function LoginPage({
  searchParams,
}: {
  searchParams?: Promise<{ callbackUrl?: string }>;
}) {
  const resolvedSearchParams = searchParams ? await searchParams : undefined;
  const callbackUrl =
    resolvedSearchParams?.callbackUrl && resolvedSearchParams.callbackUrl.startsWith('/')
      ? resolvedSearchParams.callbackUrl
      : '/';

  return (
      <div className="min-h-screen bg-gradient-hero flex flex-col">
      {/* Main content */}
      <main className="flex-1 flex items-center justify-center px-4 py-12">
        <div className="w-full max-w-md">
          {/* Card */}
          <div className="bg-white rounded-3xl shadow-2xl p-8">
            {/* Logo area */}
            <div className="text-center mb-8">
              <div className="flex items-center justify-center mx-auto mb-4">
                <ProptiLogo size={48} showWordmark />
              </div>
              <h1 className="text-2xl font-bold text-gray-900">Selamat Datang di Propti</h1>
              <p className="text-gray-500 mt-2 text-sm">
                Platform properti Indonesia paling mudah
              </p>
            </div>

            {/* Benefits */}
            <div className="space-y-3 mb-8">
              {[
                {
                  icon: <Zap className="w-4 h-4 text-brand-secondary" />,
                  text: 'Pasang iklan properti dalam 60 detik',
                },
                {
                  icon: <Shield className="w-4 h-4 text-brand-secondary" />,
                  text: 'AI otomatis rapikan iklan WhatsApp-mu',
                },
                {
                  icon: <Home className="w-4 h-4 text-brand-secondary" />,
                  text: 'Gratis untuk 3 listing pertama',
                },
              ].map((item, i) => (
                <div key={i} className="flex items-center gap-3 bg-brand-light/30 rounded-xl p-3">
                  <div className="w-8 h-8 bg-white rounded-lg flex items-center justify-center shadow-sm flex-shrink-0">
                    {item.icon}
                  </div>
                  <span className="text-sm text-gray-700">{item.text}</span>
                </div>
              ))}
            </div>

            {/* Login button */}
            <GoogleLoginButton callbackUrl={callbackUrl} />

            {/* Privacy note */}
            <p className="text-center text-xs text-gray-400 mt-6">
              Dengan masuk, kamu menyetujui{' '}
              <Link href="/terms" className="text-brand-secondary hover:underline">
                Syarat &amp; Ketentuan
              </Link>{' '}
              dan{' '}
              <Link href="/privacy" className="text-brand-secondary hover:underline">
                Kebijakan Privasi
              </Link>{' '}
              Propti.
            </p>
          </div>

          {/* Back to home */}
          <p className="text-center mt-6 text-white/80 text-sm">
            <Link href="/" className="hover:text-white transition-colors underline underline-offset-2">
              ← Kembali ke Beranda
            </Link>
          </p>
        </div>
      </main>
    </div>
  );
}

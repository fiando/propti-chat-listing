import Link from 'next/link';
import { redirect } from 'next/navigation';
import { ProfilePageClient } from '@/components/profile/ProfilePageClient';
import { getServerAuthProfile } from '@/lib/server-profile';

export default async function ProfilePage() {
  const { isAuthenticated, profile, profileError, session } = await getServerAuthProfile();

  if (!isAuthenticated) {
    redirect(`/login?callbackUrl=${encodeURIComponent('/profile')}`);
  }

  if (!profile) {
    return (
      <div className="max-w-2xl mx-auto px-4 py-12">
        <div className="rounded-3xl border border-red-200 bg-red-50 p-6">
          <h1 className="text-lg font-bold text-gray-900">Profil belum bisa dimuat</h1>
          <p className="mt-2 text-sm text-gray-600">
            {profileError ?? 'Coba muat ulang halaman dalam beberapa saat.'}
          </p>
          <div className="mt-4 flex gap-3">
            <Link href="/profile" className="btn-primary">
              Muat ulang
            </Link>
            <Link href="/" className="btn-secondary">
              Kembali ke beranda
            </Link>
          </div>
        </div>
      </div>
    );
  }

  return <ProfilePageClient profile={profile} sessionUser={session?.user} />;
}

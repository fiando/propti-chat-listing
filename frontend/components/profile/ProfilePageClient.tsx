'use client';

import { signOut } from 'next-auth/react';
import { useState } from 'react';
import {
  User,
  Mail,
  Phone,
  Crown,
  Home,
  TrendingUp,
  Settings,
  LogOut,
  Edit2,
  Check,
  Star,
} from 'lucide-react';
import { PremiumUpgradeModal } from '@/components/premium/PremiumUpgradeModal';
import Link from 'next/link';
import { formatDate } from '@/lib/utils';
import type { User as UserProfile } from '@/types';

type ProfilePageClientProps = {
  profile: UserProfile;
  sessionUser?: {
    name?: string | null;
    email?: string | null;
    image?: string | null;
  } | null;
};

export function ProfilePageClient({ profile, sessionUser }: ProfilePageClientProps) {
  const [showPremiumModal, setShowPremiumModal] = useState(false);
  const isPremium = profile.subscription.tier === 'premium';
  const activeListingsCount = profile.subscription.activeListingsCount ?? 0;

  const displayName = profile.name || sessionUser?.name || '';
  const displayEmail = profile.email || sessionUser?.email || '';
  const displayAvatar = profile.profilePicture || sessionUser?.image || null;
  const initials = displayName.split(' ').map((n) => n[0]).join('').slice(0, 2).toUpperCase();

  return (
    <div className="max-w-3xl mx-auto px-4 py-8">
      <h1 className="text-2xl font-black text-brand-primary mb-8">Profil Saya</h1>

      <div className="card p-6 mb-4">
        <div className="flex items-center gap-5">
          {displayAvatar ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img
              src={displayAvatar}
              alt={displayName}
              className="w-20 h-20 rounded-2xl object-cover shadow-md"
            />
          ) : (
            <div className="w-20 h-20 bg-gradient-hero rounded-2xl flex items-center justify-center text-white text-2xl font-black shadow-md">
              {initials}
            </div>
          )}
          <div className="flex-1">
            <div className="flex items-center gap-2 mb-1">
              <h2 className="text-xl font-bold text-gray-900">{displayName}</h2>
              {isPremium && (
                <span className="flex items-center gap-1 bg-amber-50 text-amber-600 text-xs font-semibold px-2 py-0.5 rounded-full border border-amber-200">
                  <Crown className="w-3 h-3" />
                  Premium
                </span>
              )}
            </div>
            <div className="flex items-center gap-1.5 text-gray-500 text-sm">
              <Mail className="w-3.5 h-3.5" />
              {displayEmail}
            </div>
            {profile.phone && (
              <div className="flex items-center gap-1.5 text-gray-500 text-sm mt-1">
                <Phone className="w-3.5 h-3.5" />
                {profile.phone}
              </div>
            )}
          </div>
        </div>

        <div className="grid grid-cols-3 gap-4 mt-6 pt-6 border-t">
          {[
            {
              icon: Home,
              label: 'Iklan Aktif',
              value: activeListingsCount,
            },
            {
              icon: User,
              label: 'Tipe Akun',
              value: isPremium ? 'Premium' : 'Gratis',
            },
            {
              icon: TrendingUp,
              label: 'Bergabung',
              value: formatDate(profile.createdAt).split(' ').slice(-1)[0],
            },
          ].map((stat) => (
            <div key={stat.label} className="text-center">
              <div className="w-10 h-10 bg-brand-light rounded-xl flex items-center justify-center mx-auto mb-2">
                <stat.icon className="w-5 h-5 text-brand-primary" />
              </div>
              <div className="font-bold text-gray-900">{stat.value}</div>
              <div className="text-xs text-gray-500">{stat.label}</div>
            </div>
          ))}
        </div>
      </div>

      <div className="card p-6 mb-4" id="premium">
        <div className="flex items-center justify-between mb-4">
          <h3 className="font-bold text-gray-900">Paket Berlangganan</h3>
        </div>

        {isPremium ? (
          <div className="flex items-center gap-4 bg-amber-50 border border-amber-200 rounded-2xl p-4">
            <div className="w-12 h-12 bg-brand-gold rounded-xl flex items-center justify-center">
              <Crown className="w-6 h-6 text-white" />
            </div>
            <div>
              <p className="font-bold text-amber-700">Propti Premium Aktif</p>
              {profile.subscription.renewDate && (
                <p className="text-xs text-amber-600">
                  Berlaku hingga {formatDate(profile.subscription.renewDate)}
                </p>
              )}
            </div>
            <Check className="w-5 h-5 text-amber-500 ml-auto" />
          </div>
        ) : (
          <div>
            <div className="bg-gray-50 rounded-2xl p-4 mb-4">
              <div className="flex items-center justify-between mb-3">
                <span className="font-semibold text-gray-700">Paket Gratis</span>
                <span className="badge bg-gray-100 text-gray-500">Aktif</span>
              </div>
              <ul className="space-y-2 text-sm text-gray-600">
                {[
                  '3 iklan pertama gratis',
                  'Maksimal 3 foto per iklan',
                  'Iklan tayang 30 hari',
                ].map((f) => (
                  <li key={f} className="flex items-center gap-2">
                    <Check className="w-3.5 h-3.5 text-gray-400" />
                    {f}
                  </li>
                ))}
              </ul>
            </div>
            <button
              type="button"
              onClick={() => setShowPremiumModal(true)}
              className="w-full btn-gold flex items-center justify-center gap-2"
            >
              <Crown className="w-4 h-4" />
              Upgrade ke Premium - Rp 49rb/bulan
            </button>
          </div>
        )}
      </div>

      <div className="card p-2 mb-4">
        {[
          { icon: Home, label: 'Iklan Saya', href: '/listings' },
          { icon: Star, label: 'Properti Tersimpan', href: '/saved' },
          { icon: Settings, label: 'Pengaturan', href: '/settings' },
        ].map((item) => (
          <Link
            key={item.href}
            href={item.href}
            className="flex items-center gap-3 px-4 py-3.5 rounded-xl hover:bg-gray-50 transition-colors group"
          >
            <div className="w-8 h-8 bg-brand-light rounded-lg flex items-center justify-center">
              <item.icon className="w-4 h-4 text-brand-primary" />
            </div>
            <span className="font-medium text-gray-700 group-hover:text-gray-900 flex-1">
              {item.label}
            </span>
            <Edit2 className="w-4 h-4 text-gray-300 group-hover:text-gray-400" />
          </Link>
        ))}
      </div>

      <button
        type="button"
        onClick={() => signOut({ callbackUrl: '/' })}
        className="w-full flex items-center justify-center gap-2 border-2 border-red-100 text-red-500 font-semibold py-3 px-4 rounded-2xl hover:bg-red-50 transition-colors"
      >
        <LogOut className="w-4 h-4" />
        Keluar dari Propti
      </button>

      <PremiumUpgradeModal
        isOpen={showPremiumModal}
        onClose={() => setShowPremiumModal(false)}
      />
    </div>
  );
}

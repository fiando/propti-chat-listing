'use client';

import Link from 'next/link';
import { signOut } from 'next-auth/react';
import { useState } from 'react';
import {
  Home,
  Search,
  Plus,
  Heart,
  User,
  LogOut,
  ChevronDown,
  Crown,
  Menu,
  X,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { useAuth } from '@/hooks/useAuth';
import { ProptiLogo } from './ProptiLogo';

export function Header() {
  const { session, isAuthenticated, isLoading, isPremium, isSubscriptionLoading } = useAuth();
  const [menuOpen, setMenuOpen] = useState(false);
  const [profileOpen, setProfileOpen] = useState(false);

  return (
    <header className="sticky top-0 z-40 bg-white border-b border-gray-100 shadow-sm">
      <div className="max-w-6xl mx-auto px-4 h-16 flex items-center justify-between">
        {/* Logo */}
        <Link href="/" className="flex items-center flex-shrink-0" aria-label="Propti - Beranda">
          <ProptiLogo size={36} showWordmark />
        </Link>

        {/* Desktop nav */}
        <nav className="hidden md:flex items-center gap-6">
          <Link
            href="/search"
            className="text-gray-600 hover:text-brand-primary font-medium transition-colors text-sm flex items-center gap-1.5"
          >
            <Search className="w-4 h-4" />
            Cari Properti
          </Link>
          {isAuthenticated && (
            <>
              <Link
                href="/listings"
                className="text-gray-600 hover:text-brand-primary font-medium transition-colors text-sm"
              >
                Iklan Saya
              </Link>
              <Link
                href="/saved"
                className="text-gray-600 hover:text-brand-primary font-medium transition-colors text-sm flex items-center gap-1.5"
              >
                <Heart className="w-4 h-4" />
                Tersimpan
              </Link>
            </>
          )}
        </nav>

        {/* Auth area */}
        <div className="flex items-center gap-3">
          <Link
            href="/listings/create"
            className="hidden md:flex items-center gap-2 btn-primary text-sm py-2 px-4"
          >
            <Plus className="w-4 h-4" />
            Pasang Iklan
          </Link>

          {isLoading ? (
            <div className="w-9 h-9 bg-gray-100 rounded-full animate-pulse" />
          ) : isAuthenticated ? (
            <div className="relative">
              <button
                onClick={() => setProfileOpen(!profileOpen)}
                className="flex items-center gap-2 hover:bg-gray-50 rounded-xl px-2 py-1.5 transition-colors"
              >
                {session?.user?.image ? (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img
                    src={session.user.image}
                    alt={session.user.name || ''}
                    className="w-8 h-8 rounded-full object-cover"
                  />
                ) : (
                  <div className="w-8 h-8 bg-brand-primary rounded-full flex items-center justify-center text-white text-sm font-bold">
                    {session?.user?.name?.[0] || 'U'}
                  </div>
                )}
                <span className="hidden md:block text-sm font-medium text-gray-700 max-w-[120px] truncate">
                  {session?.user?.name?.split(' ')[0]}
                </span>
                <ChevronDown className={cn('w-4 h-4 text-gray-400 transition-transform', profileOpen && 'rotate-180')} />
              </button>

              {profileOpen && (
                <>
                  <div className="fixed inset-0 z-10" onClick={() => setProfileOpen(false)} />
                  <div className="absolute right-0 top-full mt-2 w-56 bg-white rounded-2xl shadow-xl border border-gray-100 py-2 z-20">
                    <div className="px-4 py-3 border-b border-gray-100">
                      <p className="font-semibold text-gray-900 text-sm">{session?.user?.name}</p>
                      <p className="text-xs text-gray-500 truncate">{session?.user?.email}</p>
                    </div>
                    <Link
                      href="/profile"
                      className="flex items-center gap-3 px-4 py-2.5 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
                      onClick={() => setProfileOpen(false)}
                    >
                      <User className="w-4 h-4" />
                      Profil Saya
                    </Link>
                    <Link
                      href="/listings"
                      className="flex items-center gap-3 px-4 py-2.5 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
                      onClick={() => setProfileOpen(false)}
                    >
                      <Home className="w-4 h-4" />
                      Iklan Saya
                    </Link>
                    {isSubscriptionLoading ? (
                      <div className="flex items-center gap-3 px-4 py-2.5 text-sm text-gray-400">
                        <Crown className="w-4 h-4" />
                        Memuat status paket...
                      </div>
                    ) : isPremium ? (
                      <div className="flex items-center gap-3 px-4 py-2.5 text-sm text-brand-gold bg-amber-50/60">
                        <Crown className="w-4 h-4" />
                        Paket Premium Aktif
                      </div>
                    ) : (
                      <Link
                        href="/profile#premium"
                        className="flex items-center gap-3 px-4 py-2.5 text-sm text-brand-gold hover:bg-amber-50 transition-colors"
                        onClick={() => setProfileOpen(false)}
                      >
                        <Crown className="w-4 h-4" />
                        Upgrade Premium
                      </Link>
                    )}
                    <div className="border-t border-gray-100 mt-1 pt-1">
                      <button
                        onClick={() => {
                          setProfileOpen(false);
                          signOut({ callbackUrl: '/' });
                        }}
                        className="w-full flex items-center gap-3 px-4 py-2.5 text-sm text-red-500 hover:bg-red-50 transition-colors"
                      >
                        <LogOut className="w-4 h-4" />
                        Keluar
                      </button>
                    </div>
                  </div>
                </>
              )}
            </div>
          ) : (
            <Link href="/login" className="btn-primary text-sm py-2 px-4">
              Masuk
            </Link>
          )}

          {/* Mobile menu button */}
          <button
            onClick={() => setMenuOpen(!menuOpen)}
            className="md:hidden w-9 h-9 flex items-center justify-center rounded-xl hover:bg-gray-100 transition-colors"
          >
            {menuOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
          </button>
        </div>
      </div>

      {/* Mobile menu */}
      {menuOpen && (
        <div className="md:hidden border-t border-gray-100 bg-white px-4 py-4 space-y-1">
          <Link
            href="/search"
            className="flex items-center gap-3 px-3 py-2.5 rounded-xl text-gray-700 hover:bg-gray-50 font-medium text-sm"
            onClick={() => setMenuOpen(false)}
          >
            <Search className="w-4 h-4" />
            Cari Properti
          </Link>
          <Link
            href="/listings/create"
            className="flex items-center gap-3 px-3 py-2.5 rounded-xl bg-brand-primary text-white font-semibold text-sm"
            onClick={() => setMenuOpen(false)}
          >
            <Plus className="w-4 h-4" />
            Pasang Iklan Gratis
          </Link>
          {isAuthenticated && (
            <>
              <Link
                href="/listings"
                className="flex items-center gap-3 px-3 py-2.5 rounded-xl text-gray-700 hover:bg-gray-50 font-medium text-sm"
                onClick={() => setMenuOpen(false)}
              >
                <Home className="w-4 h-4" />
                Iklan Saya
              </Link>
              <Link
                href="/saved"
                className="flex items-center gap-3 px-3 py-2.5 rounded-xl text-gray-700 hover:bg-gray-50 font-medium text-sm"
                onClick={() => setMenuOpen(false)}
              >
                <Heart className="w-4 h-4" />
                Tersimpan
              </Link>
            </>
          )}
        </div>
      )}
    </header>
  );
}

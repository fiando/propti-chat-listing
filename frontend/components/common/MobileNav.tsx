'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { Home, Search, Plus, Heart, User } from 'lucide-react';
import { cn } from '@/lib/utils';
import { useSession } from 'next-auth/react';

const NAV_ITEMS = [
  { href: '/', icon: Home, label: 'Beranda' },
  { href: '/search', icon: Search, label: 'Cari' },
  { href: '/listings/create', icon: Plus, label: 'Pasang', isPrimary: true },
  { href: '/saved', icon: Heart, label: 'Simpan', requiresAuth: true },
  { href: '/profile', icon: User, label: 'Profil', requiresAuth: true },
];

export function MobileNav() {
  const pathname = usePathname();
  const { data: session } = useSession();

  return (
    <nav className="md:hidden fixed bottom-0 left-0 right-0 z-40 bg-white border-t border-gray-100 shadow-[0_-4px_16px_rgba(0,0,0,0.06)]">
      <div className="flex items-center justify-around px-2 h-16">
        {NAV_ITEMS.map((item) => {
          if (item.requiresAuth && !session) return null;

          const isActive = pathname === item.href || (item.href !== '/' && pathname.startsWith(item.href));

          if (item.isPrimary) {
            return (
              <Link
                key={item.href}
                href={item.href}
                className="flex flex-col items-center gap-1 -mt-6"
              >
                <div className="w-14 h-14 bg-gradient-hero rounded-2xl flex items-center justify-center shadow-lg shadow-brand-primary/30">
                  <item.icon className="w-7 h-7 text-white" />
                </div>
                <span className="text-xs font-semibold text-brand-primary">{item.label}</span>
              </Link>
            );
          }

          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                'flex flex-col items-center gap-1 px-3 py-1 rounded-xl transition-colors',
                isActive ? 'text-brand-primary' : 'text-gray-400 hover:text-gray-600'
              )}
            >
              <item.icon className={cn('w-5 h-5', isActive && 'fill-brand-light')} />
              <span className={cn('text-xs font-medium', isActive && 'font-semibold')}>
                {item.label}
              </span>
            </Link>
          );
        })}
        {!session && (
          <Link
            href="/login"
            className="flex flex-col items-center gap-1 px-3 py-1 rounded-xl text-gray-400 hover:text-gray-600 transition-colors"
          >
            <User className="w-5 h-5" />
            <span className="text-xs font-medium">Masuk</span>
          </Link>
        )}
      </div>
    </nav>
  );
}

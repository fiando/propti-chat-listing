'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { Home, Search, Plus, Heart, User, BriefcaseBusiness } from 'lucide-react';
import { cn } from '@/lib/utils';
import { useSession } from 'next-auth/react';

const NAV_ITEMS = [
  { href: '/', icon: Home, label: 'Beranda' },
  { href: '/search', icon: Search, label: 'Cari' },
  { href: '/agent', icon: BriefcaseBusiness, label: 'Agent', requiresAuth: true },
  { href: '/saved', icon: Heart, label: 'Simpan', requiresAuth: true },
  { href: '/profile', icon: User, label: 'Profil', requiresAuth: true },
];

export function MobileNav() {
  const pathname = usePathname();
  const { data: session } = useSession();

  const getHref = (href: string, requiresAuth?: boolean) => {
    if (!requiresAuth || session) {
      return href;
    }

    return `/login?callbackUrl=${encodeURIComponent(href)}`;
  };

  return (
    <>
      <nav className="md:hidden fixed bottom-0 left-0 right-0 z-40 bg-white border-t border-gray-100 shadow-[0_-4px_16px_rgba(0,0,0,0.06)]">
        <div className="flex items-center justify-around px-2 h-16">
        {NAV_ITEMS.map((item) => {
          const isActive = pathname === item.href || (item.href !== '/' && pathname.startsWith(item.href));

          return (
            <Link
              key={item.href}
              href={getHref(item.href, item.requiresAuth)}
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
        </div>
      </nav>
      <Link
        href={getHref('/listings/create')}
        className="md:hidden fixed bottom-20 right-4 z-50 flex h-14 w-14 items-center justify-center rounded-2xl bg-gradient-hero shadow-lg shadow-brand-primary/30 transition-transform active:scale-95"
        aria-label="Pasang iklan"
      >
        <Plus className="h-7 w-7 text-white" />
      </Link>
    </>
  );
}

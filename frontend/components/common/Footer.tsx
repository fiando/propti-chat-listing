import Link from 'next/link';
import { Facebook, Instagram, Twitter } from 'lucide-react';
import { ProptiLogo } from './ProptiLogo';

const FOOTER_LINKS = {
  Produk: [
    { label: 'Cari Properti', href: '/search' },
    { label: 'Pasang Iklan', href: '/listings/create' },
    { label: 'Premium', href: '/profile#premium' },
    { label: 'Harga', href: '/pricing' },
  ],
  Perusahaan: [
    { label: 'Tentang Kami', href: '/about' },
  ],
  Bantuan: [
    { label: 'FAQ', href: '/faq' },
    { label: 'Hubungi Kami', href: '/contact' },
    { label: 'Syarat & Ketentuan', href: '/terms' },
    { label: 'Kebijakan Privasi', href: '/privacy' },
  ],
};

export function Footer() {
  return (
    <footer className="bg-brand-primary text-white mt-auto hidden md:block">
      <div className="max-w-6xl mx-auto px-4 py-12">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-8">
          {/* Brand */}
          <div className="col-span-2 md:col-span-1">
            <div className="mb-4 flex items-center gap-3">
              <ProptiLogo size={36} />
              <span className="text-2xl font-black tracking-tight text-white">Propti</span>
            </div>
            <p className="text-white/60 text-sm leading-relaxed mb-6">
              Jual beli properti semudah chat WhatsApp. AI kami otomatis rapikan iklanmu.
            </p>
            <div className="flex items-center gap-3">
              {[Facebook, Instagram, Twitter].map((Icon, i) => (
                <a
                  key={i}
                  href="#"
                  className="w-9 h-9 bg-white/10 rounded-lg flex items-center justify-center hover:bg-white/20 transition-colors"
                >
                  <Icon className="w-4 h-4" />
                </a>
              ))}
            </div>
          </div>

          {/* Links */}
          {Object.entries(FOOTER_LINKS).map(([category, links]) => (
            <div key={category}>
              <h3 className="font-semibold text-white mb-4">{category}</h3>
              <ul className="space-y-2.5">
                {links.map((link) => (
                  <li key={link.href}>
                    <Link
                      href={link.href}
                      className="text-white/60 hover:text-white text-sm transition-colors"
                    >
                      {link.label}
                    </Link>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        <div className="border-t border-white/10 mt-10 pt-6 flex flex-col md:flex-row items-center justify-between gap-4">
          <p className="text-white/40 text-xs">
            © {new Date().getFullYear()} Propti. Hak cipta dilindungi.
          </p>
          <p className="text-white/40 text-xs">
            Dibuat dengan ❤️ untuk pasar properti Indonesia
          </p>
        </div>
      </div>
    </footer>
  );
}

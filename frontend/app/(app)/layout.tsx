import { Header } from '@/components/common/Header';
import { MobileNav } from '@/components/common/MobileNav';
import { Footer } from '@/components/common/Footer';

export default function AppLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex flex-col min-h-screen">
      <Header />
      <main className="flex-1 pb-20 md:pb-0">{children}</main>
      <Footer />
      <MobileNav />
    </div>
  );
}

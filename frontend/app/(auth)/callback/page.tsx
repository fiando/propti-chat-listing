'use client';

import { useEffect } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useSession } from 'next-auth/react';
import { Loader2 } from 'lucide-react';

export default function CallbackPage() {
  const { status } = useSession();
  const router = useRouter();
  const searchParams = useSearchParams();
  const callbackUrl = searchParams.get('callbackUrl') || '/';

  useEffect(() => {
    if (status === 'authenticated') {
      router.replace(callbackUrl);
    } else if (status === 'unauthenticated') {
      router.replace('/login');
    }
  }, [status, router, callbackUrl]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-[#F8F9FA]">
      <div className="text-center">
        <Loader2 className="w-12 h-12 text-brand-primary animate-spin mx-auto mb-4" />
        <p className="text-gray-600 font-medium">Sedang masuk ke Propti...</p>
      </div>
    </div>
  );
}

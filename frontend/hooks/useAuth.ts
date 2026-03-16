'use client';

import { useSession, signIn, signOut } from 'next-auth/react';
import { useQuery } from '@tanstack/react-query';
import { getProfile } from '@/lib/api';
import type { User } from '@/types';

export function useAuth() {
  const { data: session, status } = useSession();

  const { data: profile } = useQuery<User>({
    queryKey: ['profile'],
    queryFn: getProfile,
    enabled: status === 'authenticated',
    retry: 1,
  });

  const isAuthenticated = status === 'authenticated';
  const isLoading = status === 'loading';
  const renewDate = profile?.subscription?.renewDate ? new Date(profile.subscription.renewDate) : null;
  const hasActiveRenewDate = renewDate ? Number.isFinite(renewDate.getTime()) && renewDate.getTime() > Date.now() : true;
  const isPremium = profile?.subscription?.tier === 'premium' && hasActiveRenewDate;

  const login = () => signIn('google', { callbackUrl: '/' });
  const logout = () => signOut({ callbackUrl: '/' });

  return {
    session,
    profile,
    isAuthenticated,
    isLoading,
    isPremium,
    login,
    logout,
    user: session?.user,
  };
}

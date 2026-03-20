'use client';

import { useSession, signIn, signOut } from 'next-auth/react';
import { useQuery } from '@tanstack/react-query';
import { getProfile } from '@/lib/api';
import type { User } from '@/types';
import { getSubscriptionStatus } from '@/lib/subscription-status';

export function useAuth() {
  const { data: session, status } = useSession();

  const {
    data: profile,
    isError: isProfileError,
    isLoading: isProfileLoading,
    isFetching: isProfileFetching,
    isFetchedAfterMount: isProfileFetchedAfterMount,
  } = useQuery<User>({
    queryKey: ['profile'],
    queryFn: getProfile,
    enabled: status === 'authenticated',
    retry: 1,
    refetchOnMount: 'always',
  });

  const isAuthenticated = status === 'authenticated';
  const subscriptionStatus = getSubscriptionStatus({ authStatus: status, profile });
  const isSubscriptionLoading = subscriptionStatus === 'loading';
  const isLoading = status === 'loading' || (status === 'authenticated' && isProfileLoading);
  const isPremium = subscriptionStatus === 'active' || subscriptionStatus === 'expiring_soon';

  const login = () => signIn('google', { callbackUrl: '/' });
  const logout = () => signOut({ callbackUrl: '/' });

  return {
    session,
    profile,
    isAuthenticated,
    isLoading,
    isProfileError,
    isProfileFetchedAfterMount,
    isProfileFetching,
    isSubscriptionLoading,
    subscriptionStatus,
    isPremium,
    login,
    logout,
    user: session?.user,
  };
}

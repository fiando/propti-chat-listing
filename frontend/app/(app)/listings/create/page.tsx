import { CreateListingClient } from '@/components/listings/CreateListingClient';
import { getCreateListingAccessState } from '@/lib/create-listing-errors';
import { getServerAuthProfile } from '@/lib/server-profile';

export default async function CreateListingPage() {
  const { isAuthenticated, profile } = await getServerAuthProfile();
  const isPremium = profile?.subscription?.tier === 'premium';
  const initialCreateAccessState = getCreateListingAccessState({
    isAuthenticated,
    isPremium,
    hasFreshAccessResult: true,
    hasListingsError: isAuthenticated && !profile,
    activeListingsCount: profile?.subscription?.activeListingsCount,
  });

  return (
    <CreateListingClient
      initialIsAuthenticated={isAuthenticated}
      initialIsPremium={Boolean(isPremium)}
      initialPhone={profile?.phone ?? ''}
      initialCreateAccessState={initialCreateAccessState}
    />
  );
}

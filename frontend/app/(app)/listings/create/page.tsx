import { CreateListingClient } from '@/components/listings/CreateListingClient';
import { getCreateListingAccessState } from '@/lib/create-listing-errors';
import { getServerAuthProfile } from '@/lib/server-profile';

export default async function CreateListingPage() {
  const { isAuthenticated, profile } = await getServerAuthProfile();
  const tier = profile?.subscription?.tier ?? 'free';
  const initialCreateAccessState = getCreateListingAccessState({
    isAuthenticated,
    tier,
    hasFreshAccessResult: true,
    hasListingsError: isAuthenticated && !profile,
    activeListingsCount: profile?.subscription?.activeListingsCount,
  });

  return (
    <CreateListingClient
      initialIsAuthenticated={isAuthenticated}
      initialTier={tier}
      initialProfilePhone={profile?.phone ?? ''}
      initialCreateAccessState={initialCreateAccessState}
    />
  );
}

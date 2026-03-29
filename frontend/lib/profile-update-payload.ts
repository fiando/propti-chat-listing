import type { UpdateProfileRequest, UserPreferences } from '@/types';

type AccountRole = UpdateProfileRequest['role'] | '';

interface BuildProfileUpdatePayloadInput {
  role: AccountRole;
  notifications: boolean;
  preferences?: UserPreferences;
}

export function buildProfileUpdatePayload({
  role,
  notifications,
  preferences,
}: BuildProfileUpdatePayloadInput): UpdateProfileRequest {
  const payload: UpdateProfileRequest = {
    preferences: {
      favoriteLocations: preferences?.favoriteLocations ?? [],
      searchHistory: preferences?.searchHistory ?? [],
      notifications,
    },
  };

  if (role) {
    payload.role = role;
  }

  return payload;
}

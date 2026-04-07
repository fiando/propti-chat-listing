import type { UpdateProfileRequest, UserPreferences } from '@/types';

type AccountRole = UpdateProfileRequest['role'] | '';

interface BuildProfileUpdatePayloadInput {
  phone: string;
  role: AccountRole;
  notifications: boolean;
  preferences?: UserPreferences;
}

export function buildProfileUpdatePayload({
  phone,
  role,
  notifications,
  preferences,
}: BuildProfileUpdatePayloadInput): UpdateProfileRequest {
  const payload: UpdateProfileRequest = {
    phone: phone.trim(),
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

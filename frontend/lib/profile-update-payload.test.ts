import test from 'node:test';
import assert from 'node:assert/strict';
import type { UserPreferences } from '@/types';
import { buildProfileUpdatePayload } from './profile-update-payload.ts';

const basePreferences: UserPreferences = {
  favoriteLocations: ['Jakarta'],
  searchHistory: ['Menteng'],
  notifications: true,
};

test('buildProfileUpdatePayload omits role when no account role is selected', () => {
  assert.deepEqual(
    buildProfileUpdatePayload({
      role: '',
      notifications: false,
      preferences: basePreferences,
    }),
    {
      preferences: {
        ...basePreferences,
        notifications: false,
      },
    },
  );
});

test('buildProfileUpdatePayload includes role when account role is selected', () => {
  assert.deepEqual(
    buildProfileUpdatePayload({
      role: 'buyer',
      notifications: true,
      preferences: basePreferences,
    }),
    {
      role: 'buyer',
      preferences: basePreferences,
    },
  );
});

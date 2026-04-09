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
      phone: '081234567890',
      role: '',
      notifications: false,
      preferences: basePreferences,
    }),
    {
      preferences: {
        ...basePreferences,
        notifications: false,
      },
      phone: '081234567890',
    },
  );
});

test('buildProfileUpdatePayload includes role when account role is selected', () => {
  assert.deepEqual(
    buildProfileUpdatePayload({
      phone: '081234567890',
      role: 'buyer',
      notifications: true,
      preferences: basePreferences,
    }),
    {
      role: 'buyer',
      phone: '081234567890',
      preferences: basePreferences,
    },
  );
});

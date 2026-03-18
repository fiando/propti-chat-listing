import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import type { User } from '@/types';

const apiBaseUrl = process.env.NEXT_PUBLIC_API_URL || 'https://api.propti.id';

type SessionWithAccessToken = {
  accessToken?: string;
  user?: {
    name?: string | null;
    email?: string | null;
    image?: string | null;
  };
};

export async function getServerAuthProfile() {
  const session = (await getServerSession(authOptions)) as SessionWithAccessToken | null;
  const accessToken = session?.accessToken;

  if (!session?.user || !accessToken) {
    return {
      isAuthenticated: false,
      session,
      profile: null as User | null,
      profileError: null as string | null,
    };
  }

  try {
    const response = await fetch(`${apiBaseUrl}/auth/user`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${accessToken}`,
      },
      cache: 'no-store',
    });

    if (!response.ok) {
      throw new Error('Gagal memuat profil akun.');
    }

    const profile = (await response.json()) as User;
    return {
      isAuthenticated: true,
      session,
      profile,
      profileError: null as string | null,
    };
  } catch (error) {
    return {
      isAuthenticated: true,
      session,
      profile: null as User | null,
      profileError: error instanceof Error ? error.message : 'Gagal memuat profil akun.',
    };
  }
}

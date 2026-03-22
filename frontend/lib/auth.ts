import type { NextAuthOptions } from 'next-auth';
import GoogleProvider from 'next-auth/providers/google';

import {
  exchangeGoogleIdTokenForBackendSession,
  getJwtExpiryTimestamp,
  refreshBackendAuthTokenIfNeeded,
} from '@/lib/backend-auth';

const apiBaseUrl = process.env.NEXT_PUBLIC_API_URL || 'https://api.propti.id';

type BackendUser = {
  userId: string;
  email: string;
  name?: string;
  profilePicture?: string;
};

type TokenWithBackendAuth = {
  sub?: string;
  accessToken?: string;
  googleAccessToken?: string;
  googleAccessTokenExpiresAt?: number;
  googleId?: string;
  refreshToken?: string;
  backendAccessToken?: string;
  backendAccessTokenExpiresAt?: number;
  backendUser?: BackendUser;
  error?: 'MissingRefreshToken' | 'RefreshAccessTokenError';
};

export const authOptions: NextAuthOptions = {
  providers: [
    GoogleProvider({
      clientId: process.env.GOOGLE_CLIENT_ID!,
      clientSecret: process.env.GOOGLE_CLIENT_SECRET!,
      authorization: {
        params: {
          prompt: 'consent',
          access_type: 'offline',
          response_type: 'code',
        },
      },
    }),
  ],
  callbacks: {
    async jwt({ token, account, profile }) {
      const nextToken = token as typeof token & TokenWithBackendAuth;

      if (account && profile) {
        nextToken.googleAccessToken = account.access_token;
        nextToken.googleAccessTokenExpiresAt = account.expires_at ? account.expires_at * 1000 : undefined;
        nextToken.googleId = profile.sub;
        nextToken.refreshToken = account.refresh_token ?? nextToken.refreshToken;

        const googleIdToken = (account as typeof account & { id_token?: string }).id_token;
        if (!googleIdToken) {
          throw new Error('Missing Google ID token');
        }

        const backendSession = await exchangeGoogleIdTokenForBackendSession({
          apiBaseUrl,
          idToken: googleIdToken,
        });

        nextToken.backendAccessToken = backendSession.backendAccessToken;
        nextToken.backendAccessTokenExpiresAt = getJwtExpiryTimestamp(backendSession.backendAccessToken);
        nextToken.backendUser = backendSession.backendUser;
        nextToken.accessToken = backendSession.backendAccessToken;
        nextToken.sub = backendSession.backendUser.userId;
        nextToken.error = undefined;

        return nextToken;
      }

      return refreshBackendAuthTokenIfNeeded({
        token: nextToken,
        apiBaseUrl,
        googleClientId: process.env.GOOGLE_CLIENT_ID!,
        googleClientSecret: process.env.GOOGLE_CLIENT_SECRET!,
      });
    },
    async session({ session, token }) {
      const nextToken = token as typeof token & TokenWithBackendAuth;
      return {
        ...session,
        accessToken: nextToken.backendAccessToken,
        error: nextToken.error,
        user: {
          ...session.user,
          id: nextToken.backendUser?.userId ?? token.sub,
          email: nextToken.backendUser?.email ?? session.user?.email,
          name: nextToken.backendUser?.name ?? session.user?.name,
          image: nextToken.backendUser?.profilePicture ?? session.user?.image,
        },
      };
    },
  },
  pages: {
    signIn: '/login',
    error: '/login',
  },
  session: {
    strategy: 'jwt',
    maxAge: 30 * 24 * 60 * 60, // 30 days
  },
  secret: process.env.NEXTAUTH_SECRET,
};

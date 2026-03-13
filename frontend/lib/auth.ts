import type { NextAuthOptions } from 'next-auth';
import GoogleProvider from 'next-auth/providers/google';

import { exchangeGoogleIdTokenForBackendSession } from '@/lib/backend-auth';

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
  googleId?: string;
  backendAccessToken?: string;
  backendUser?: BackendUser;
};

export const authOptions: NextAuthOptions = {
  providers: [
    GoogleProvider({
      clientId: process.env.GOOGLE_CLIENT_ID!,
      clientSecret: process.env.GOOGLE_CLIENT_SECRET!,
    }),
  ],
  callbacks: {
    async jwt({ token, account, profile }) {
      const nextToken = token as typeof token & TokenWithBackendAuth;

      if (account && profile) {
        nextToken.googleAccessToken = account.access_token;
        nextToken.googleId = profile.sub;

        const googleIdToken = (account as typeof account & { id_token?: string }).id_token;
        if (!googleIdToken) {
          throw new Error('Missing Google ID token');
        }

        const backendSession = await exchangeGoogleIdTokenForBackendSession({
          apiBaseUrl,
          idToken: googleIdToken,
        });

        nextToken.backendAccessToken = backendSession.backendAccessToken;
        nextToken.backendUser = backendSession.backendUser;
        nextToken.accessToken = backendSession.backendAccessToken;
        nextToken.sub = backendSession.backendUser.userId;
      }

      return nextToken;
    },
    async session({ session, token }) {
      const nextToken = token as typeof token & TokenWithBackendAuth;
      return {
        ...session,
        accessToken: nextToken.backendAccessToken,
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

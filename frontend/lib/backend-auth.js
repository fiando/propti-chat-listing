export function buildBackendAuthPayload({ idToken }) {
  return { idToken };
}

const GOOGLE_TOKEN_ENDPOINT = 'https://oauth2.googleapis.com/token';
const BACKEND_TOKEN_REFRESH_SKEW_MS = 60 * 1000;

export function getJwtExpiryTimestamp(token) {
  if (!token) {
    return undefined;
  }

  const [, payload] = token.split('.');
  if (!payload) {
    return undefined;
  }

  try {
    const claims = JSON.parse(Buffer.from(payload, 'base64url').toString('utf8'));
    return typeof claims.exp === 'number' ? claims.exp * 1000 : undefined;
  } catch {
    return undefined;
  }
}

export async function exchangeGoogleIdTokenForBackendSession({ apiBaseUrl, idToken, fetchImpl = fetch }) {
  const response = await fetchImpl(`${apiBaseUrl}/auth/google`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(buildBackendAuthPayload({ idToken })),
  });

  if (!response.ok) {
    throw new Error('Backend authentication failed');
  }

  const data = await response.json();
  return {
    backendAccessToken: data.accessToken,
    backendUser: data.user,
  };
}

export async function refreshGoogleTokens({
  refreshToken,
  clientId,
  clientSecret,
  fetchImpl = fetch,
  now = () => Date.now(),
}) {
  const response = await fetchImpl(GOOGLE_TOKEN_ENDPOINT, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded',
    },
    body: new URLSearchParams({
      client_id: clientId,
      client_secret: clientSecret,
      grant_type: 'refresh_token',
      refresh_token: refreshToken,
    }),
  });

  if (!response.ok) {
    throw new Error('Google token refresh failed');
  }

  const data = await response.json();
  if (!data.id_token) {
    throw new Error('Google token refresh did not return an ID token');
  }

  return {
    accessToken: data.access_token,
    accessTokenExpiresAt:
      typeof data.expires_in === 'number' ? now() + data.expires_in * 1000 : undefined,
    refreshToken: data.refresh_token ?? refreshToken,
    idToken: data.id_token,
  };
}

export async function refreshBackendAuthTokenIfNeeded({
  token,
  apiBaseUrl,
  googleClientId,
  googleClientSecret,
  now = () => Date.now(),
  refreshGoogleTokens: refreshGoogleTokensImpl = refreshGoogleTokens,
  exchangeGoogleIdTokenForBackendSession: exchangeGoogleIdTokenForBackendSessionImpl = ({ idToken }) =>
    exchangeGoogleIdTokenForBackendSession({ apiBaseUrl, idToken }),
}) {
  const backendAccessTokenExpiresAt =
    token.backendAccessTokenExpiresAt ?? getJwtExpiryTimestamp(token.backendAccessToken);

  if (
    token.backendAccessToken &&
    typeof backendAccessTokenExpiresAt === 'number' &&
    backendAccessTokenExpiresAt - BACKEND_TOKEN_REFRESH_SKEW_MS > now()
  ) {
    return {
      ...token,
      backendAccessTokenExpiresAt,
    };
  }

  if (!token.refreshToken) {
    return {
      ...token,
      backendAccessTokenExpiresAt,
      error: 'MissingRefreshToken',
    };
  }

  try {
    const googleTokens = await refreshGoogleTokensImpl({
      refreshToken: token.refreshToken,
      clientId: googleClientId,
      clientSecret: googleClientSecret,
      now,
    });

    const backendSession = await exchangeGoogleIdTokenForBackendSessionImpl({
      idToken: googleTokens.idToken,
    });

    return {
      ...token,
      accessToken: backendSession.backendAccessToken,
      backendAccessToken: backendSession.backendAccessToken,
      backendAccessTokenExpiresAt: getJwtExpiryTimestamp(backendSession.backendAccessToken),
      backendUser: backendSession.backendUser,
      googleAccessToken: googleTokens.accessToken,
      googleAccessTokenExpiresAt: googleTokens.accessTokenExpiresAt,
      refreshToken: googleTokens.refreshToken,
      sub: backendSession.backendUser?.userId ?? token.sub,
      error: undefined,
    };
  } catch {
    return {
      ...token,
      backendAccessTokenExpiresAt,
      error: 'RefreshAccessTokenError',
    };
  }
}

export function getBackendAuthHeader({ backendAccessToken }) {
  return backendAccessToken ? `Bearer ${backendAccessToken}` : undefined;
}

export function getBackendProfilePath() {
  return '/auth/user';
}

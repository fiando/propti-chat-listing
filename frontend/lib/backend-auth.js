export function buildBackendAuthPayload({ idToken }) {
  return { idToken };
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

export function getBackendAuthHeader({ backendAccessToken }) {
  return backendAccessToken ? `Bearer ${backendAccessToken}` : undefined;
}

export function getBackendProfilePath() {
  return '/auth/user';
}

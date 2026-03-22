type HomeRedirectDecision = {
  isAuthenticated: boolean;
  hasProfile: boolean;
};

export function getAuthenticatedHomeRedirectPath({
  isAuthenticated,
  hasProfile,
}: HomeRedirectDecision): string | null {
  if (!isAuthenticated || !hasProfile) {
    return null;
  }

  return '/listings';
}


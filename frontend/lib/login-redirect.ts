type LoginRedirectDecision = {
  isAuthenticated: boolean;
  hasProfile: boolean;
};

export function shouldRedirectAuthenticatedLoginVisitor({
  isAuthenticated,
  hasProfile,
}: LoginRedirectDecision): boolean {
  return isAuthenticated && hasProfile;
}

export function shouldCanonicalizeLocalhostUrl(input) {
  const url = new URL(input);

  return !url.pathname.startsWith('/api/auth');
}

export function getCanonicalLocalhostUrl(input) {
  const url = new URL(input);

  if (url.hostname !== '127.0.0.1') {
    return null;
  }
  if (!shouldCanonicalizeLocalhostUrl(input)) {
    return null;
  }

  url.hostname = 'localhost';
  return url.toString();
}

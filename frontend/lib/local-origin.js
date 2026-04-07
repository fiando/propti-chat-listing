export function getCanonicalLocalhostUrl(input) {
  const url = new URL(input);

  if (url.hostname !== '127.0.0.1') {
    return null;
  }

  url.hostname = 'localhost';
  return url.toString();
}

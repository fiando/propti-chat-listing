export function getApiErrorMessage(payload: unknown, fallbackMessage?: string): string {
  if (payload && typeof payload === 'object') {
    const data = payload as { message?: unknown; error?: unknown };

    if (typeof data.message === 'string' && data.message.trim() !== '') {
      return data.message;
    }

    if (typeof data.error === 'string' && data.error.trim() !== '') {
      return data.error;
    }
  }

  return fallbackMessage || 'Terjadi kesalahan. Silakan coba lagi.';
}

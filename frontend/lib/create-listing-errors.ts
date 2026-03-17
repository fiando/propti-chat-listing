export function getCreateListingErrorMessage(error: unknown): string {
  const message = error instanceof Error ? error.message : '';
  const normalizedMessage = message.trim().toLowerCase();

  if (normalizedMessage.includes('free tier allows at most 3 listing')) {
    return 'Paket gratis hanya bisa memasang 3 listing. Upgrade ke Premium untuk memasang lebih banyak.';
  }

  return message || 'Terjadi kesalahan saat memasang iklan. Silakan coba lagi.';
}

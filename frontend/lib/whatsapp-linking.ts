export type WhatsAppWriteEligibilityResponse = {
  eligible: boolean;
  isLinked: boolean;
  linkedPhone?: string;
  verifiedAt?: string;
  reason?: string;
};

export type WhatsAppLinkChallengeResponse = {
  challengeId: string;
  phone: string;
  expiresAt: string;
  retryCount: number;
  messageText?: string;
  messageLink?: string;
};

function digitsOnly(value: string) {
  return value.replace(/\D+/g, '');
}

export function normalizeWhatsAppLinkPhone(phone: string) {
  const digits = digitsOnly(phone);
  if (!digits) {
    return '';
  }

  if (digits.startsWith('62')) {
    return digits;
  }

  if (digits.startsWith('0')) {
    return `62${digits.slice(1)}`;
  }

  return digits;
}

type WhatsAppEligibilityCopyTone = 'success' | 'warning';

export type WhatsAppEligibilityCopy = {
  tone: WhatsAppEligibilityCopyTone;
  title: string;
  description: string;
};

export function getWhatsAppWriteEligibilityCopy(
  status: Pick<WhatsAppWriteEligibilityResponse, 'eligible' | 'isLinked' | 'reason'>
): WhatsAppEligibilityCopy {
  if (status.eligible) {
    return {
      tone: 'success',
      title: 'WhatsApp terverifikasi',
      description: 'Akun kamu sudah eligible untuk fitur WhatsApp write.',
    };
  }

  if (status.isLinked) {
    return {
      tone: 'warning',
      title: 'Nomor sudah terhubung, tinggal kirim pesan verifikasi WhatsApp',
      description: 'Kirim pesan challenge dari WhatsApp yang sama agar akun kamu bisa pakai fitur WhatsApp write.',
    };
  }

  return {
    tone: 'warning',
    title: 'WhatsApp belum terhubung',
    description: 'Hubungkan nomor WhatsApp kamu lalu kirim pesan challenge untuk verifikasi.',
  };
}

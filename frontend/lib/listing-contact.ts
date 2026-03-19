function digitsOnly(value: string) {
  return value.replace(/\D+/g, '');
}

export function normalizeContactPhone(phone: string) {
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

export function buildWhatsAppMessage(title?: string, listingUrl?: string): string {
  let message = 'Halo, saya tertarik dengan properti';
  if (title) {
    message += ` *${title}*`;
  }
  message += ' di Propti. Apakah masih tersedia?';
  if (listingUrl) {
    message += `\n${listingUrl}`;
  }
  return message;
}

export function buildListingContactLinks(phone: string, title?: string, listingUrl?: string) {
  const normalizedPhone = normalizeContactPhone(phone);
  if (!normalizedPhone) {
    return {
      whatsappUrl: null,
      phoneUrl: null,
    };
  }

  const message = buildWhatsAppMessage(title, listingUrl);
  return {
    whatsappUrl: `https://wa.me/${normalizedPhone}?text=${encodeURIComponent(message)}`,
    phoneUrl: `tel:+${normalizedPhone}`,
  };
}

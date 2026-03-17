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

export function buildListingContactLinks(phone: string) {
  const normalizedPhone = normalizeContactPhone(phone);
  if (!normalizedPhone) {
    return {
      whatsappUrl: null,
      phoneUrl: null,
    };
  }

  return {
    whatsappUrl: `https://wa.me/${normalizedPhone}`,
    phoneUrl: `tel:+${normalizedPhone}`,
  };
}

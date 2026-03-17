const AMENITY_ALIASES: Record<string, string> = {
  'ruang tamu': 'ruang_tamu',
  'ruang keluarga': 'ruang_keluarga',
  'kolam renang': 'kolam_renang',
  'keamanan 24 jam': 'keamanan_24jam',
  'one gate system': 'one_gate_system',
  'listrik 3 phase': 'listrik_3_phase',
  'loading dock': 'loading_dock',
  'akses container': 'akses_container',
  'akses truk': 'akses_truk',
  wifi: 'internet_wifi',
  internet: 'internet_wifi',
};

export function formatListingPriceInput(value: number): string {
  if (!value) {
    return '';
  }

  return new Intl.NumberFormat('id-ID').format(value);
}

export function parseListingPriceInput(value: string): number {
  const digitsOnly = value.replace(/\D/g, '');
  return digitsOnly ? Number(digitsOnly) : 0;
}

function normalizeAmenityId(value: string): string {
  const normalized = value.trim().toLowerCase().replace(/[\s-]+/g, ' ');
  return AMENITY_ALIASES[normalized] || normalized.replace(/\s+/g, '_');
}

export function normalizeAmenityIds(values: string[] | undefined): string[] {
  if (!values?.length) {
    return [];
  }

  const uniqueIds = new Set<string>();

  for (const value of values) {
    const normalized = normalizeAmenityId(value);
    if (normalized) {
      uniqueIds.add(normalized);
    }
  }

  return Array.from(uniqueIds);
}

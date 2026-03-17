export function shouldRequirePhone(input: {
  profilePhone?: string | null;
  phoneOverride?: string | null;
}): boolean {
  return !(input.phoneOverride?.trim() || input.profilePhone?.trim());
}

export function getPhoneModalSubmitLabel(input: {
  isSavingPhone: boolean;
  isSubmittingListing: boolean;
}): string {
  if (input.isSavingPhone) {
    return 'Menyimpan nomor telepon...';
  }

  if (input.isSubmittingListing) {
    return 'Memasang iklan...';
  }

  return 'Simpan & Pasang Iklan';
}

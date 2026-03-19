'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import {
  MapPin,
  Bed,
  Bath,
  Maximize2,
  ZoomIn,
  Home,
  Eye,
  Heart,
  Share2,
  Edit2,
  Trash2,
  CheckCircle,
  Clock,
  XCircle,
  Star,
  Phone,
  MessageCircle,
  ChevronLeft,
  ChevronRight,
  Shield,
} from 'lucide-react';
import Link from 'next/link';
import Lightbox from 'yet-another-react-lightbox';
import Zoom from 'yet-another-react-lightbox/plugins/zoom';
import { formatAmenityLabel, formatPrice, formatDate, LISTING_TYPE_LABELS, PRICE_UNIT_LABELS } from '@/lib/utils';
import type { Listing, ContactRevealChannel, RevealListingContactResponse } from '@/types';
import { buildListingContactLinks } from '@/lib/listing-contact';
import { getCachedRevealedContact, getOrLoadRevealedContact } from '@/lib/revealed-contact-cache';
import { useToast } from '@/app/toaster';
import { useRevealListingContact } from '@/hooks/useListings';
import { getListingGalleryImages } from '@/lib/listing-images';

interface ListingDetailProps {
  listing: Listing;
  isOwner?: boolean;
  isAuthenticated?: boolean;
  isSaved?: boolean;
  isSaving?: boolean;
  onSave?: () => void;
  onDelete?: () => void;
}

const OWNER_MODERATION_NOTICES: Partial<
  Record<'approved' | 'pending' | 'rejected', { title: string; message: string; tone: string }>
> = {
  pending: {
    title: 'Iklan sedang diproses',
    message: 'Iklan kamu sudah tersimpan dan sedang kami proses. Sebentar lagi akan tampil ke publik.',
    tone: 'border-blue-200 bg-blue-50 text-blue-700',
  },
  rejected: {
    title: 'Iklan Ditolak',
    message: 'Iklan ini melanggar kebijakan Propti. Hapus dan buat iklan baru.',
    tone: 'border-red-200 bg-red-50 text-red-700',
  },
};

const STATUS_CONFIG = {
  approved: { icon: CheckCircle, label: 'Aktif', color: 'text-green-600 bg-green-50' },
  pending: { icon: Clock, label: 'Menunggu Review', color: 'text-amber-600 bg-amber-50' },
  rejected: { icon: XCircle, label: 'Ditolak', color: 'text-red-600 bg-red-50' },
};

export function ListingDetail({
  listing,
  isOwner = false,
  isAuthenticated = false,
  isSaved = false,
  isSaving = false,
  onSave,
  onDelete,
}: ListingDetailProps) {
  const [activeImage, setActiveImage] = useState(0);
  const [isLightboxOpen, setIsLightboxOpen] = useState(false);
  const [revealedContact, setRevealedContact] = useState<RevealListingContactResponse | null>(() =>
    getCachedRevealedContact(listing.listingId)
  );
  const router = useRouter();
  const { toast } = useToast();
  const { mutateAsync: revealListingContact } = useRevealListingContact();

  const status = STATUS_CONFIG[listing.moderationStatus] || STATUS_CONFIG.pending;
  const priceLabel = formatPrice(listing.price);
  const priceUnit = PRICE_UNIT_LABELS[listing.priceUnit];
  const typeLabel = LISTING_TYPE_LABELS[listing.listingType];

  const images = getListingGalleryImages(listing);
  const lightboxSlides = images.map((image, index) => ({
    src: image.url,
    alt: `${listing.title} - Foto ${index + 1}`,
  }));
  const sellerName = listing.sellerName?.trim() || 'Penjual Propti';
  const hasSellerContact = Boolean(listing.hasSellerPhone);
  const ownerModerationNotice = isOwner ? OWNER_MODERATION_NOTICES[listing.moderationStatus] : undefined;
  const shouldHideOwnerContent = isOwner && listing.moderationStatus === 'rejected';

  const handleRevealContact = async (channel: ContactRevealChannel) => {
    if (!isAuthenticated) {
      router.push(`/login?callbackUrl=${encodeURIComponent(`/listings/${listing.listingId}`)}`);
      return;
    }

    try {
      const contact = await getOrLoadRevealedContact(listing.listingId, () =>
        revealListingContact({ id: listing.listingId, channel })
      );
      setRevealedContact(contact);

      const listingUrl =
        typeof window !== 'undefined' ? `${window.location.origin}/listings/${listing.listingId}` : undefined;
      const nextLinks = buildListingContactLinks(contact.sellerPhone, listing.title, listingUrl);
      const nextUrl = channel === 'whatsapp' ? nextLinks.whatsappUrl : nextLinks.phoneUrl;
      if (!nextUrl) {
        toast('Nomor penjual tidak valid untuk dibuka.', 'error');
        return;
      }

      if (channel === 'whatsapp') {
        window.open(nextUrl, '_blank', 'noopener,noreferrer');
        return;
      }

      window.location.href = nextUrl;
    } catch (error) {
      toast(error instanceof Error ? error.message : 'Gagal membuka kontak penjual.', 'error');
    }
  };

  const handleShare = async () => {
    const shareUrl = typeof window !== 'undefined' ? window.location.href : '';

    try {
      if (navigator.share) {
        await navigator.share({
          title: listing.title,
          text: `${listing.title} — ${priceLabel}`,
          url: shareUrl,
        });
        return;
      }

      await navigator.clipboard.writeText(shareUrl);
      toast('Link iklan berhasil disalin.', 'success');
    } catch {
      toast('Gagal membagikan link iklan.', 'error');
    }
  };

  if (shouldHideOwnerContent) {
    return (
      <div className="max-w-5xl mx-auto px-4 py-8">
        <div className="grid lg:grid-cols-3 gap-8">
          <div className="lg:col-span-2 space-y-6">
            {ownerModerationNotice && (
              <div className={`rounded-2xl border px-4 py-4 text-sm ${ownerModerationNotice.tone}`}>
                <p className="font-semibold">{ownerModerationNotice.title}</p>
                <p className="mt-1 leading-6">{ownerModerationNotice.message}</p>
              </div>
            )}

            <div className="card overflow-hidden">
              <div className="flex h-72 items-center justify-center bg-gradient-to-br from-brand-primary/20 to-brand-secondary/30 md:h-96">
                <Home className="w-24 h-24 text-brand-secondary/30" />
              </div>
            </div>

            <div className="card p-6">
              <div className="flex flex-wrap items-start justify-between gap-4">
                <div>
                  <h1 className="text-2xl font-bold text-gray-900">
                    {listing.moderationStatus === 'rejected' ? 'Iklan Ditolak' : 'Sedang diproses'}
                  </h1>
                  <p className="mt-2 max-w-2xl text-sm leading-6 text-gray-600">
                    {listing.moderationStatus === 'rejected'
                      ? 'Iklan ini melanggar kebijakan Propti dan tidak tampil di pencarian.'
                      : 'Detail iklan belum tampil ke publik sampai review selesai.'}
                  </p>
                </div>

                <div className={`flex items-center gap-1.5 text-xs font-semibold px-3 py-1.5 rounded-full ${status.color}`}>
                  <status.icon className="w-3.5 h-3.5" />
                  {status.label}
                </div>
              </div>

              <div className="mt-6 grid gap-4 sm:grid-cols-2">
                <div className="rounded-2xl border border-gray-100 bg-gray-50 px-4 py-4">
                  <p className="text-xs font-semibold uppercase tracking-wide text-gray-500">ID iklan</p>
                  <p className="mt-1 text-sm font-semibold text-gray-900">{listing.listingId}</p>
                </div>
                <div className="rounded-2xl border border-gray-100 bg-gray-50 px-4 py-4">
                  <p className="text-xs font-semibold uppercase tracking-wide text-gray-500">Dikirim pada</p>
                  <p className="mt-1 text-sm font-semibold text-gray-900">{formatDate(listing.createdAt)}</p>
                </div>
              </div>
            </div>
          </div>

          <div className="space-y-4">
            <div className="card p-6 sticky top-20">
              <h2 className="text-lg font-bold text-gray-900">
                {listing.moderationStatus === 'rejected' ? 'Buat iklan baru' : 'Sedang dalam proses'}
              </h2>
              {listing.moderationStatus === 'rejected' ? (
                <>
                  <p className="mt-2 text-sm leading-6 text-gray-600">
                    Tips agar iklan lolos:
                  </p>
                  <ul className="mt-2 space-y-1.5 text-sm text-gray-600">
                    <li className="flex gap-2"><span>📸</span> Gunakan foto asli yang terang dan jelas</li>
                    <li className="flex gap-2"><span>✍️</span> Deskripsi jujur sesuai kondisi properti</li>
                    <li className="flex gap-2"><span>💰</span> Harga realistis sesuai pasaran</li>
                    <li className="flex gap-2"><span>🚫</span> Hindari kata berlebihan atau menyesatkan</li>
                  </ul>
                </>
              ) : (
                <p className="mt-2 text-sm leading-6 text-gray-600">
                  Biasanya selesai dalam beberapa jam. Kamu akan bisa melihat iklan setelah disetujui.
                </p>
              )}

              <button
                onClick={onDelete}
                disabled={!onDelete}
                className="mt-6 w-full flex items-center justify-center gap-2 border border-red-200 text-red-500 rounded-xl py-2.5 text-sm font-medium hover:bg-red-50 transition-colors disabled:cursor-not-allowed disabled:opacity-60"
              >
                <Trash2 className="w-4 h-4" />
                Hapus Iklan
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-5xl mx-auto px-4 py-8">
      <div className="grid lg:grid-cols-3 gap-8">
        {/* Left: Images + Details */}
        <div className="lg:col-span-2 space-y-6">
          {/* Image gallery */}
          <div className="card overflow-hidden">
            <div className="relative bg-gradient-to-br from-brand-primary/20 to-brand-secondary/30 h-72 md:h-96">
              {images[activeImage]?.url ? (
                <button
                  type="button"
                  onClick={() => setIsLightboxOpen(true)}
                  className="block h-full w-full cursor-zoom-in"
                  aria-label="Buka foto dalam tampilan penuh"
                >
                  {/* eslint-disable-next-line @next/next/no-img-element */}
                  <img
                    src={images[activeImage].url}
                    alt={listing.title}
                    className="w-full h-full object-cover"
                  />
                </button>
              ) : (
                <div className="w-full h-full flex items-center justify-center">
                  <Home className="w-24 h-24 text-brand-secondary/30" />
                </div>
              )}

              {images[activeImage]?.url && (
                <button
                  type="button"
                  onClick={() => setIsLightboxOpen(true)}
                  className="absolute right-4 top-4 z-10 inline-flex items-center gap-2 rounded-full bg-white/90 px-3 py-2 text-sm font-medium text-gray-700 shadow-lg transition-colors hover:bg-white"
                >
                  <ZoomIn className="w-4 h-4" />
                  Perbesar
                </button>
              )}

              {images.length > 1 && (
                <>
                  <button
                    type="button"
                    onClick={() => setActiveImage((p) => (p - 1 + images.length) % images.length)}
                    className="absolute left-3 top-1/2 z-10 -translate-y-1/2 w-10 h-10 bg-white/90 rounded-full flex items-center justify-center shadow-lg hover:bg-white transition-colors"
                  >
                    <ChevronLeft className="w-5 h-5" />
                  </button>
                  <button
                    type="button"
                    onClick={() => setActiveImage((p) => (p + 1) % images.length)}
                    className="absolute right-3 top-1/2 z-10 -translate-y-1/2 w-10 h-10 bg-white/90 rounded-full flex items-center justify-center shadow-lg hover:bg-white transition-colors"
                  >
                    <ChevronRight className="w-5 h-5" />
                  </button>
                  <div className="absolute bottom-3 left-1/2 z-10 -translate-x-1/2 flex gap-1.5">
                    {images.map((_, i) => (
                      <button
                        key={i}
                        type="button"
                        onClick={() => setActiveImage(i)}
                        className={`w-2 h-2 rounded-full transition-all ${
                          i === activeImage ? 'bg-white w-5' : 'bg-white/50'
                        }`}
                      />
                    ))}
                  </div>
                </>
              )}

              {/* Badges */}
              <div className="absolute top-4 left-4 z-10 flex gap-2">
                <span
                  className={`text-white text-xs font-bold px-3 py-1.5 rounded-full ${
                    listing.listingType === 'sell' ? 'bg-brand-primary' : 'bg-blue-600'
                  }`}
                >
                  {typeLabel}
                </span>
                {listing.premiumFeatures?.isFeatured && (
                  <span className="bg-brand-gold text-white text-xs font-bold px-3 py-1.5 rounded-full flex items-center gap-1">
                    <Star className="w-3 h-3" /> Unggulan
                  </span>
                )}
              </div>
            </div>

            {/* Thumbnail strip */}
            {images.length > 1 && (
              <div className="flex gap-2 p-4 overflow-x-auto">
                 {images.map((img, i) => (
                   <button
                     key={img.id}
                     onClick={() => setActiveImage(i)}
                     className={`flex-shrink-0 w-16 h-16 rounded-lg overflow-hidden border-2 transition-all ${
                       i === activeImage ? 'border-brand-primary' : 'border-transparent'
                     }`}
                   >
                     {/* eslint-disable-next-line @next/next/no-img-element */}
                     <img src={img.thumbnailUrl || img.url} alt="" className="w-full h-full object-cover" />
                   </button>
                 ))}
              </div>
            )}
          </div>

          <Lightbox
            open={isLightboxOpen}
            close={() => setIsLightboxOpen(false)}
            index={activeImage}
            slides={lightboxSlides}
            plugins={[Zoom]}
            on={{ view: ({ index }) => setActiveImage(index) }}
          />

          {ownerModerationNotice && (
            <div className={`rounded-2xl border px-4 py-4 text-sm ${ownerModerationNotice.tone}`}>
              <p className="font-semibold">{ownerModerationNotice.title}</p>
              <p className="mt-1 leading-6">{ownerModerationNotice.message}</p>
            </div>
          )}

          {/* Title & Price */}
          <div className="card p-6">
            <div className="flex items-start justify-between gap-4">
              <div className="flex-1">
                <h1 className="text-2xl font-bold text-gray-900 mb-2">{listing.title}</h1>
                <div className="flex items-center gap-1.5 text-gray-500 text-sm mb-4">
                  <MapPin className="w-4 h-4" />
                  <span>
                    {[listing.location?.district, listing.location?.city, listing.location?.province]
                      .filter(Boolean)
                      .join(', ')}
                  </span>
                </div>
                <div className="flex items-baseline gap-2">
                  <span className="text-3xl font-black text-brand-primary">{priceLabel}</span>
                  {listing.priceUnit !== 'total' && (
                    <span className="text-gray-500 text-sm">{priceUnit}</span>
                  )}
                </div>
              </div>

              {/* Status badge for owner */}
              {isOwner && (
                <div className={`flex items-center gap-1.5 text-xs font-semibold px-3 py-1.5 rounded-full ${status.color}`}>
                  <status.icon className="w-3.5 h-3.5" />
                  {status.label}
                </div>
              )}
            </div>

            {/* Stats row */}
            <div className="flex items-center gap-4 mt-4 pt-4 border-t text-sm text-gray-500">
              <span className="flex items-center gap-1">
                <Eye className="w-4 h-4" /> {listing.views} dilihat
              </span>
              <span className="flex items-center gap-1">
                <Heart className="w-4 h-4" /> {listing.saves} disimpan
              </span>
              <span>Dipasang {formatDate(listing.createdAt)}</span>
            </div>
          </div>

          {/* Property specs */}
          <div className="card p-6">
            <h2 className="font-bold text-gray-900 mb-4 text-lg">Detail Properti</h2>
            <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
              {[
                {
                  icon: Maximize2,
                  label: 'Luas Tanah',
                  value: listing.propertyDetails?.landArea
                    ? `${listing.propertyDetails.landArea} m²`
                    : '-',
                },
                {
                  icon: Home,
                  label: 'Luas Bangunan',
                  value: listing.propertyDetails?.buildingArea
                    ? `${listing.propertyDetails.buildingArea} m²`
                    : '-',
                },
                {
                  icon: Bed,
                  label: 'Kamar Tidur',
                  value: listing.propertyDetails?.bedrooms
                    ? `${listing.propertyDetails.bedrooms} KT`
                    : '-',
                },
                {
                  icon: Bath,
                  label: 'Kamar Mandi',
                  value: listing.propertyDetails?.bathrooms
                    ? `${listing.propertyDetails.bathrooms} KM`
                    : '-',
                },
                {
                  icon: Shield,
                  label: 'Sertifikat',
                  value: listing.propertyDetails?.legalStatus || '-',
                },
                {
                  icon: Star,
                  label: 'Orientasi',
                  value: listing.propertyDetails?.orientation || '-',
                },
              ].map((spec, i) => (
                <div key={i} className="flex items-center gap-3 bg-gray-50 rounded-xl p-3">
                  <div className="w-8 h-8 bg-brand-light rounded-lg flex items-center justify-center flex-shrink-0">
                    <spec.icon className="w-4 h-4 text-brand-primary" />
                  </div>
                  <div>
                    <div className="text-xs text-gray-500">{spec.label}</div>
                    <div className="font-semibold text-gray-900 text-sm">{spec.value}</div>
                  </div>
                </div>
              ))}
            </div>

            {/* Amenities */}
            {listing.propertyDetails?.amenities?.length > 0 && (
              <div className="mt-6">
                <h3 className="font-semibold text-gray-700 mb-3 text-sm">Fasilitas</h3>
                <div className="flex flex-wrap gap-2">
                  {listing.propertyDetails.amenities.map((a) => (
                    <span key={a} className="flex items-center gap-1.5 bg-brand-light text-brand-primary text-xs font-medium px-3 py-1.5 rounded-full">
                      <CheckCircle className="w-3 h-3" />
                      {formatAmenityLabel(a)}
                    </span>
                  ))}
                </div>
              </div>
            )}
          </div>

          {/* Description */}
          <div className="card p-6">
            <h2 className="font-bold text-gray-900 mb-3 text-lg">Deskripsi</h2>
            <p className="text-gray-600 leading-relaxed whitespace-pre-line">
              {listing.description}
            </p>
          </div>

          {/* Location */}
          {listing.location?.address && (
            <div className="card p-6">
              <h2 className="font-bold text-gray-900 mb-3 text-lg">Lokasi</h2>
              <div className="flex items-start gap-2 text-gray-600">
                <MapPin className="w-5 h-5 text-brand-primary mt-0.5 flex-shrink-0" />
                <span>{listing.location.address}</span>
              </div>
            </div>
          )}
        </div>

        {/* Right: Contact + Actions */}
        <div className="space-y-4">
          {/* Contact card */}
          <div className="card p-6 sticky top-20">
            <h3 className="font-bold text-gray-900 mb-4">Hubungi Penjual</h3>
            <div className="mb-4 rounded-2xl border border-gray-100 bg-gray-50 px-4 py-3">
              <p className="text-xs font-semibold uppercase tracking-wide text-gray-500">Nama penjual</p>
              <p className="mt-1 font-semibold text-gray-900">{revealedContact?.sellerName || sellerName}</p>
              {revealedContact?.sellerPhone ? (
                <p className="mt-1 text-sm text-gray-600">{revealedContact.sellerPhone}</p>
              ) : isAuthenticated ? (
                <div></div>
              ) : (
                <p className="mt-1 text-sm text-gray-500">Masuk untuk melihat nomor dan menghubungi penjual.</p>
              )}
            </div>
            <div className="space-y-3">
              {hasSellerContact ? (
                <>
                  <button
                    type="button"
                    onClick={() => void handleRevealContact('whatsapp')}
                    className="w-full flex items-center justify-center gap-2 bg-[#25D366] text-white font-semibold py-3 px-4 rounded-xl hover:opacity-90 transition-opacity"
                  >
                    <MessageCircle className="w-5 h-5" />
                    Chat WhatsApp
                  </button>
                  <button
                    type="button"
                    onClick={() => void handleRevealContact('phone')}
                    className="w-full flex items-center justify-center gap-2 btn-secondary"
                  >
                    <Phone className="w-5 h-5" />
                    Telepon
                  </button>
                  {!isAuthenticated && (
                    <button
                      type="button"
                      onClick={() => router.push(`/login?callbackUrl=${encodeURIComponent(`/listings/${listing.listingId}`)}`)}
                      className="w-full rounded-xl border border-gray-200 px-4 py-3 text-sm font-semibold text-gray-700 hover:bg-gray-50 transition-colors"
                    >
                      Masuk untuk melihat nomor
                    </button>
                  )}
                </>
              ) : (
                <div className="rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-700">
                  Nomor penjual belum tersedia.
                </div>
              )}
            </div>

            <div className="flex gap-2 mt-4">
              <button
                onClick={onSave}
                disabled={!onSave || isSaving}
                className={`flex-1 flex items-center justify-center gap-2 border rounded-xl py-2.5 text-sm font-medium transition-all ${
                  isSaved
                    ? 'bg-red-50 border-red-200 text-red-500'
                    : 'border-gray-200 text-gray-600 hover:bg-gray-50'
                }`}
              >
                <Heart className={`w-4 h-4 ${isSaved ? 'fill-red-500' : ''}`} />
                {isSaving ? 'Memproses...' : isSaved ? 'Tersimpan' : 'Simpan'}
              </button>
              <button
                onClick={handleShare}
                className="flex-1 flex items-center justify-center gap-2 border border-gray-200 text-gray-600 rounded-xl py-2.5 text-sm font-medium hover:bg-gray-50 transition-colors"
              >
                <Share2 className="w-4 h-4" />
                Bagikan
              </button>
            </div>

            {/* Owner actions */}
            {isOwner && (
              <div className="mt-4 pt-4 border-t space-y-2">
                <Link
                  href={`/listings/${listing.listingId}/edit`}
                  className="w-full flex items-center justify-center gap-2 btn-primary text-sm py-2.5"
                >
                  <Edit2 className="w-4 h-4" />
                  Edit Iklan
                </Link>
                <button
                  onClick={onDelete}
                  className="w-full flex items-center justify-center gap-2 border border-red-200 text-red-500 rounded-xl py-2.5 text-sm font-medium hover:bg-red-50 transition-colors"
                >
                  <Trash2 className="w-4 h-4" />
                  Hapus Iklan
                </button>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

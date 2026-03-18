import type {
  Listing,
  ListingFormImage,
  ListingImageValue,
  UploadPrepareRequest,
  UploadPrepareResponse,
} from '@/types';

type ListingGalleryImage = {
  id: string;
  url: string;
  thumbnailUrl?: string;
  isFeatured: boolean;
};

function isListingImageView(image: ListingImageValue): image is Exclude<ListingImageValue, string> {
  return typeof image === 'object' && image !== null;
}

export function getListingCardImage(listing: Pick<Listing, 'images' | 'featuredThumbnailUrl'>): string | undefined {
  if (listing.featuredThumbnailUrl) {
    return listing.featuredThumbnailUrl;
  }

  const firstImage = listing.images?.[0];
  if (!firstImage) {
    return undefined;
  }

  if (typeof firstImage === 'string') {
    return firstImage;
  }

  return firstImage.thumbnailUrl || firstImage.url;
}

export function getListingGalleryImages(
  listing: Pick<Listing, 'images' | 'featuredThumbnailUrl'>
): ListingGalleryImage[] {
  const gallery = (listing.images || [])
    .map((image, index): ListingGalleryImage | null => {
      if (typeof image === 'string') {
        return {
          id: `legacy-${index}`,
          url: image,
          thumbnailUrl: image,
          isFeatured: index === 0,
        };
      }

      const url = image.url || image.thumbnailUrl;
      if (!url) {
        return null;
      }

      return {
        id: image.imageId || `image-${index}`,
        url,
        thumbnailUrl: image.thumbnailUrl || url,
        isFeatured: Boolean(image.isFeatured),
      };
    })
    .filter((image): image is ListingGalleryImage => image !== null);

  if (gallery.length === 0 && listing.featuredThumbnailUrl) {
    return [
      {
        id: 'featured-thumbnail',
        url: listing.featuredThumbnailUrl,
        thumbnailUrl: listing.featuredThumbnailUrl,
        isFeatured: true,
      },
    ];
  }

  if (gallery.some((image) => image.isFeatured)) {
    return gallery;
  }

  return gallery.map((image, index) => ({
    ...image,
    isFeatured: index === 0,
  }));
}

export function createNewListingFormImage(file: File, previewUrl: string): ListingFormImage {
  return {
    id: `${file.name}-${file.size}-${crypto.randomUUID()}`,
    kind: 'new',
    file,
    previewUrl,
    isFeatured: false,
  };
}

export function normalizeListingFormImages(images: Listing['images'] = []): ListingFormImage[] {
  return images
    .map((image, index): ListingFormImage | null => {
      if (typeof image === 'string') {
        return {
          id: `legacy-${index}`,
          kind: 'existing',
          previewUrl: image,
          remoteUrl: image,
          isFeatured: index === 0,
        };
      }

      const previewUrl = image.thumbnailUrl || image.url;
      if (!previewUrl) {
        return null;
      }

      return {
        id: image.imageId || `image-${index}`,
        kind: 'existing',
        imageId: image.imageId,
        previewUrl,
        remoteUrl: image.url || previewUrl,
        isFeatured: Boolean(image.isFeatured),
      };
    })
    .filter((image): image is ListingFormImage => image !== null)
    .map((image, index, items) => ({
      ...image,
      isFeatured: items.some((item) => item.isFeatured) ? image.isFeatured : index === 0,
    }));
}

export function markFeaturedImage(images: ListingFormImage[], imageId: string): ListingFormImage[] {
  return images.map((image) => ({
    ...image,
    isFeatured: image.id === imageId,
  }));
}

export function removeListingFormImage(images: ListingFormImage[], imageId: string): ListingFormImage[] {
  const remaining = images.filter((image) => image.id !== imageId);
  if (remaining.length === 0 || remaining.some((image) => image.isFeatured)) {
    return remaining;
  }

  return remaining.map((image, index) => ({
    ...image,
    isFeatured: index === 0,
  }));
}

type UploadImageDependencies = {
  prepareUpload: (request: UploadPrepareRequest) => Promise<UploadPrepareResponse>;
  uploadObject: (presignedUrl: string, file: File) => Promise<void>;
};

export async function uploadPendingListingImages(
  images: ListingFormImage[],
  dependencies: UploadImageDependencies,
  listingId?: string
): Promise<{
  retainedImageIds: string[];
  legacyImages: string[];
  newImageUploadSessionIds: string[];
  featuredImageId?: string;
  featuredUploadSessionId?: string;
}> {
  const retainedImages = images.filter((image) => image.kind === 'existing' && image.imageId);
  const retainedLegacyImages = images
    .filter((image) => image.kind === 'existing' && !image.imageId && image.remoteUrl)
    .map((image) => image.remoteUrl as string);
  const newImages = images.filter((image) => image.kind === 'new' && image.file);
  const featuredImage = images.find((image) => image.isFeatured);
  const orderedLegacyImages =
    featuredImage?.kind === 'existing' && !featuredImage.imageId && featuredImage.remoteUrl
      ? [featuredImage.remoteUrl, ...retainedLegacyImages.filter((url) => url !== featuredImage.remoteUrl)]
      : retainedLegacyImages;

  if (newImages.length === 0) {
    return {
      retainedImageIds: retainedImages
        .map((image) => image.imageId)
        .filter((imageId): imageId is string => Boolean(imageId)),
      legacyImages: orderedLegacyImages,
      newImageUploadSessionIds: [],
      featuredImageId: featuredImage?.kind === 'existing' ? featuredImage.imageId : undefined,
    };
  }

  const prepareRequest: UploadPrepareRequest = {
    listingId,
    retainedImageCount: retainedImages.length,
    finalImageCount: images.length,
    newImages: newImages.map((image) => ({
      contentType: image.file?.type || 'image/jpeg',
      sizeBytes: image.file?.size || 0,
    })),
  };

  const response = await dependencies.prepareUpload(prepareRequest);

  if (response.slots.length !== newImages.length) {
    throw new Error('Jumlah slot upload tidak sesuai dengan jumlah foto baru.');
  }

  await Promise.all(
    response.slots.map((slot, index) => dependencies.uploadObject(slot.presignedUrl, newImages[index].file as File))
  );

  const featuredNewImageIndex =
    featuredImage?.kind === 'new' ? newImages.findIndex((image) => image.id === featuredImage.id) : -1;

  return {
    retainedImageIds: retainedImages
      .map((image) => image.imageId)
      .filter((imageId): imageId is string => Boolean(imageId)),
    legacyImages: orderedLegacyImages,
    newImageUploadSessionIds: response.slots.map((slot) => slot.sessionId),
    featuredImageId: featuredImage?.kind === 'existing' ? featuredImage.imageId : undefined,
    featuredUploadSessionId:
      featuredNewImageIndex >= 0 ? response.slots[featuredNewImageIndex]?.sessionId : undefined,
  };
}

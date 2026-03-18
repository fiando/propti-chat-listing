'use client';

import { useRef, useState } from 'react';
import { Upload, X, Image as ImageIcon, Crown, Plus, Star } from 'lucide-react';
import { cn } from '@/lib/utils';
import type { ListingFormImage } from '@/types';
import { ImageLimits } from '@/types';
import {
  createNewListingFormImage,
  markFeaturedImage,
  removeListingFormImage,
} from '@/lib/listing-images';

interface ImageUploadProps {
  images: ListingFormImage[];
  onChange: (images: ListingFormImage[]) => void;
  maxImages?: number;
}

const MAX_IMAGE_PX = 1280;
const IMAGE_QUALITY = 0.82;

function loadImage(url: string): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const image = new Image();
    image.onload = () => resolve(image);
    image.onerror = () => reject(new Error('Gagal memuat foto untuk dikompres.'));
    image.src = url;
  });
}

async function compressImage(file: File, maxPx = MAX_IMAGE_PX, quality = IMAGE_QUALITY): Promise<File> {
  const sourceUrl = URL.createObjectURL(file);

  try {
    const image = await loadImage(sourceUrl);
    let { width, height } = image;

    if (width > maxPx || height > maxPx) {
      if (width >= height) {
        height = Math.round((height * maxPx) / width);
        width = maxPx;
      } else {
        width = Math.round((width * maxPx) / height);
        height = maxPx;
      }
    }

    const canvas = document.createElement('canvas');
    canvas.width = width;
    canvas.height = height;

    const context = canvas.getContext('2d');
    if (!context) {
      return file;
    }

    context.drawImage(image, 0, 0, width, height);

    const blob = await new Promise<Blob | null>((resolve) => {
      canvas.toBlob(resolve, 'image/jpeg', quality);
    });

    if (!blob) {
      return file;
    }

    const sanitizedName = file.name.replace(/\.[^.]+$/, '') || 'listing-image';
    return new File([blob], `${sanitizedName}.jpg`, {
      type: 'image/jpeg',
      lastModified: file.lastModified,
    });
  } finally {
    URL.revokeObjectURL(sourceUrl);
  }
}

export function ImageUpload({ images, onChange, maxImages = 3 }: ImageUploadProps) {
  const [dragging, setDragging] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  const handleFiles = async (files: FileList | null) => {
    if (!files) return;

    const remaining = maxImages - images.length;
    const selectedFiles = Array.from(files).slice(0, remaining);
    const compressedFiles = await Promise.all(selectedFiles.map((file) => compressImage(file)));
    const nextImages = compressedFiles.map((file) =>
      createNewListingFormImage(file, URL.createObjectURL(file))
    );

    const mergedImages = [...images, ...nextImages];
    const hasFeatured = mergedImages.some((image) => image.isFeatured);

    onChange(
      hasFeatured
        ? mergedImages
        : mergedImages.map((image, index) => ({
            ...image,
            isFeatured: index === 0,
          }))
    );
  };

  const removeImage = (imageId: string) => {
    const imageToRemove = images.find((image) => image.id === imageId);
    if (imageToRemove?.kind === 'new') {
      URL.revokeObjectURL(imageToRemove.previewUrl);
    }
    onChange(removeListingFormImage(images, imageId));
  };

  const setFeaturedImage = (imageId: string) => {
    onChange(markFeaturedImage(images, imageId));
  };

  return (
    <div className="space-y-4">
      {images.length < maxImages && (
        <div
          onDragOver={(event) => {
            event.preventDefault();
            setDragging(true);
          }}
          onDragLeave={() => setDragging(false)}
          onDrop={(event) => {
            event.preventDefault();
            setDragging(false);
            void handleFiles(event.dataTransfer.files);
          }}
          onClick={() => inputRef.current?.click()}
          className={cn(
            'border-2 border-dashed rounded-2xl p-8 text-center cursor-pointer transition-all',
            dragging
              ? 'border-brand-primary bg-brand-light/30'
              : 'border-gray-300 hover:border-brand-accent hover:bg-gray-50'
          )}
        >
          <div className="w-14 h-14 bg-gray-100 rounded-2xl flex items-center justify-center mx-auto mb-3">
            <Upload className="w-7 h-7 text-gray-400" />
          </div>
          <p className="font-medium text-gray-700 mb-1">
            Drag & drop foto atau <span className="text-brand-primary">klik untuk pilih</span>
          </p>
          <p className="text-xs text-gray-400">
            JPG, PNG, WebP — Maks {maxImages} foto ({images.length}/{maxImages})
          </p>
          <input
            ref={inputRef}
            type="file"
            accept="image/*"
            multiple
            className="hidden"
            onChange={(event) => void handleFiles(event.target.files)}
          />
        </div>
      )}

      {images.length > 0 && (
        <div className="grid grid-cols-2 gap-3 md:grid-cols-3">
          {images.map((image) => (
            <div
              key={image.id}
              className="relative overflow-hidden rounded-xl border border-gray-200 bg-gray-100"
            >
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img src={image.previewUrl} alt="" className="aspect-video w-full object-cover" />

              <div className="absolute inset-x-0 bottom-0 flex items-center justify-between gap-2 bg-gradient-to-t from-black/70 to-transparent p-3">
                <button
                  type="button"
                  onClick={() => setFeaturedImage(image.id)}
                  className={cn(
                    'inline-flex items-center gap-1 rounded-full px-2.5 py-1 text-xs font-semibold transition-colors',
                    image.isFeatured
                      ? 'bg-amber-400 text-white'
                      : 'bg-white/85 text-gray-700 hover:bg-white'
                  )}
                >
                  <Star className={cn('h-3.5 w-3.5', image.isFeatured && 'fill-current')} />
                  {image.isFeatured ? 'Foto utama' : 'Jadikan foto utama'}
                </button>

                <button
                  type="button"
                  onClick={() => removeImage(image.id)}
                  className="flex h-8 w-8 items-center justify-center rounded-full bg-red-500 text-white transition-colors hover:bg-red-600"
                  aria-label="Hapus foto"
                >
                  <X className="h-4 w-4" />
                </button>
              </div>
            </div>
          ))}

          {images.length < maxImages && (
            <button
              type="button"
              onClick={() => inputRef.current?.click()}
              className="aspect-video rounded-xl border-2 border-dashed border-gray-300 hover:border-brand-accent flex items-center justify-center text-gray-400 hover:text-brand-primary transition-all"
            >
              <div className="text-center">
                <Plus className="mx-auto mb-1 h-6 w-6" />
                <span className="text-xs">Tambah Foto</span>
              </div>
            </button>
          )}
        </div>
      )}

      {maxImages <= 3 && (
        <div className="flex items-center gap-3 rounded-xl border border-amber-200 bg-amber-50 p-3">
          <Crown className="h-5 w-5 flex-shrink-0 text-amber-500" />
          <div>
            <p className="text-xs font-semibold text-amber-700">
              Upgrade ke Premium untuk upload hingga {ImageLimits.premium} foto
            </p>
            <p className="text-xs text-amber-600">
              Iklan dengan banyak foto mendapat 3x lebih banyak penonton
            </p>
          </div>
        </div>
      )}

      {images.length === 0 && (
        <div className="flex items-center gap-3 rounded-xl border border-blue-100 bg-blue-50 p-3">
          <ImageIcon className="h-5 w-5 flex-shrink-0 text-blue-400" />
          <p className="text-xs text-blue-600">
            Iklan dengan foto mendapat <strong>10x lebih banyak</strong> penonton. Tambahkan minimal
            1 foto!
          </p>
        </div>
      )}
    </div>
  );
}

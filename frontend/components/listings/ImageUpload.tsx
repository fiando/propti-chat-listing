'use client';

import { useState, useRef } from 'react';
import { Upload, X, Image as ImageIcon, Crown, Plus } from 'lucide-react';
import { cn } from '@/lib/utils';

interface ImageUploadProps {
  listingId?: string;
  images: string[];
  onChange: (images: string[]) => void;
  maxImages?: number;
}

// Compress an image file using the canvas API.
// Images are scaled to at most MAX_IMAGE_PX on the longest side and encoded
// as JPEG at IMAGE_QUALITY. This runs entirely in-browser with no network
// round-trip, so there is no user-perceived delay — the compressed preview
// appears as fast as the original would have.
const MAX_IMAGE_PX = 1280;
const IMAGE_QUALITY = 0.82;

function compressImage(file: File, maxPx = MAX_IMAGE_PX, quality = IMAGE_QUALITY): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onerror = reject;
    reader.onload = (e) => {
      const dataUrl = e.target?.result as string;
      const img = new Image();
      img.onerror = reject;
      img.onload = () => {
        let { width, height } = img;
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
        const ctx = canvas.getContext('2d');
        if (!ctx) {
          // Fallback: use original if canvas is unavailable.
          resolve(dataUrl);
          return;
        }
        ctx.drawImage(img, 0, 0, width, height);
        resolve(canvas.toDataURL('image/jpeg', quality));
      };
      img.src = dataUrl;
    };
    reader.readAsDataURL(file);
  });
}

export function ImageUpload({
  images,
  onChange,
  maxImages = 3,
}: ImageUploadProps) {
  const [dragging, setDragging] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  const handleFiles = (files: FileList | null) => {
    if (!files) return;
    const remaining = maxImages - images.length;
    const newFiles = Array.from(files).slice(0, remaining);

    Promise.all(newFiles.map((file) => compressImage(file))).then((urls) => {
      onChange([...images, ...urls]);
    });
  };

  const removeImage = (index: number) => {
    onChange(images.filter((_, i) => i !== index));
  };

  return (
    <div className="space-y-4">
      {/* Upload area */}
      {images.length < maxImages && (
        <div
          onDragOver={(e) => {
            e.preventDefault();
            setDragging(true);
          }}
          onDragLeave={() => setDragging(false)}
          onDrop={(e) => {
            e.preventDefault();
            setDragging(false);
            handleFiles(e.dataTransfer.files);
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
            Drag & drop foto atau{' '}
            <span className="text-brand-primary">klik untuk pilih</span>
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
            onChange={(e) => handleFiles(e.target.files)}
          />
        </div>
      )}

      {/* Image preview grid */}
      {images.length > 0 && (
        <div className="grid grid-cols-3 gap-3">
          {images.map((img, i) => (
            <div key={i} className="relative group aspect-video rounded-xl overflow-hidden bg-gray-100">
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img src={img} alt="" className="w-full h-full object-cover" />
              <div className="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
                <button
                  type="button"
                  onClick={(e) => {
                    e.stopPropagation();
                    removeImage(i);
                  }}
                  className="w-8 h-8 bg-red-500 rounded-full flex items-center justify-center text-white hover:bg-red-600 transition-colors"
                >
                  <X className="w-4 h-4" />
                </button>
              </div>
              {i === 0 && (
                <div className="absolute bottom-1.5 left-1.5 bg-black/60 text-white text-xs px-1.5 py-0.5 rounded">
                  Utama
                </div>
              )}
            </div>
          ))}

          {/* Add more slot */}
          {images.length < maxImages && (
            <button
              type="button"
              onClick={() => inputRef.current?.click()}
              className="aspect-video rounded-xl border-2 border-dashed border-gray-300 hover:border-brand-accent flex items-center justify-center text-gray-400 hover:text-brand-primary transition-all"
            >
              <div className="text-center">
                <Plus className="w-6 h-6 mx-auto mb-1" />
                <span className="text-xs">Tambah Foto</span>
              </div>
            </button>
          )}
        </div>
      )}

      {/* Premium upsell */}
      {maxImages <= 3 && (
        <div className="flex items-center gap-3 bg-amber-50 border border-amber-200 rounded-xl p-3">
          <Crown className="w-5 h-5 text-amber-500 flex-shrink-0" />
          <div>
            <p className="text-xs font-semibold text-amber-700">
              Upgrade ke Premium untuk upload hingga 15 foto
            </p>
            <p className="text-xs text-amber-600">Iklan dengan banyak foto mendapat 3x lebih banyak penonton</p>
          </div>
        </div>
      )}

      {/* Empty state */}
      {images.length === 0 && (
        <div className="flex items-center gap-3 bg-blue-50 border border-blue-100 rounded-xl p-3">
          <ImageIcon className="w-5 h-5 text-blue-400 flex-shrink-0" />
          <p className="text-xs text-blue-600">
            Iklan dengan foto mendapat <strong>10x lebih banyak</strong> penonton. Tambahkan minimal 1 foto!
          </p>
        </div>
      )}
    </div>
  );
}

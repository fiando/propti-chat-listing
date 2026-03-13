import { useId } from 'react';
import { cn } from '@/lib/utils';

interface ProptiLogoProps {
  size?: number;
  className?: string;
  /** Show the wordmark "Propti" next to the icon */
  showWordmark?: boolean;
  /** Color variant for the wordmark text */
  wordmarkColor?: 'dark' | 'white';
}

/**
 * Propti brand logo — standalone SVG icon with optional wordmark.
 * Uses the brand green gradient (#1B4332 → #52B788), a modern house icon,
 * and a 4-point AI sparkle in the upper-right to reflect the AI-powered purpose.
 */
export function ProptiLogo({
  size = 36,
  className,
  showWordmark = false,
  wordmarkColor = 'dark',
}: ProptiLogoProps) {
  const uid = useId();
  const gradientId = `propti-grad-${uid.replace(/:/g, '')}`;

  return (
    <span className={cn('inline-flex items-center gap-2 flex-shrink-0', className)}>
      <svg
        width={size}
        height={size}
        viewBox="0 0 40 40"
        xmlns="http://www.w3.org/2000/svg"
        aria-label="Propti"
        role="img"
      >
        <defs>
          <linearGradient id={gradientId} x1="0" y1="0" x2="1" y2="1">
            <stop offset="0%" stopColor="#1B4332" />
            <stop offset="100%" stopColor="#52B788" />
          </linearGradient>
        </defs>
        {/* Background */}
        <rect width="40" height="40" rx="8" fill={`url(#${gradientId})`} />
        {/* Modern house (no chimney) */}
        <path d="M8 20 L20 8 L32 20 L32 34 L24 34 L24 26 L16 26 L16 34 L8 34 Z" fill="white" />
        {/* Door */}
        <rect x="17" y="26" width="6" height="8" rx="1" fill="#2D6A4F" />
        {/* 4-point AI sparkle (upper-right) */}
        <path d="M33 3.5 L34.1 5.9 L36.5 7 L34.1 8.1 L33 10.5 L31.9 8.1 L29.5 7 L31.9 5.9 Z" fill="white" opacity="0.9" />
      </svg>

      {showWordmark && (
        <span
          className={cn(
            'text-xl font-black leading-none tracking-tight',
            wordmarkColor === 'white' ? 'text-white' : 'text-brand-primary'
          )}
        >
          Propti
        </span>
      )}
    </span>
  );
}

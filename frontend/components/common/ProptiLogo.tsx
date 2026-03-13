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

const iconMarkPath =
  'M20 8.5L31 17.25V25.5C31 29.642 27.642 33 23.5 33H22L28 36.75V33H16.5C12.358 33 9 29.642 9 25.5V17.25L20 8.5Z';

const iconDoorPath =
  'M17 23C17 22.448 17.448 22 18 22H22C22.552 22 23 22.448 23 23V33H17V23Z';

/**
 * Propti brand logo — standalone SVG icon with optional wordmark.
 * The mark combines a home roof, a chat bubble tail, and an open doorway
 * to reflect Propti's core flow: turning chat-based property info into listings.
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
            <stop offset="55%" stopColor="#2D6A4F" />
            <stop offset="100%" stopColor="#52B788" />
          </linearGradient>
        </defs>
        <rect width="40" height="40" rx="10" fill={`url(#${gradientId})`} />
        {/* Home roof + chat bubble tail */}
        <path d={iconMarkPath} fill="white" />
        {/* Open doorway */}
        <path d={iconDoorPath} fill="#2D6A4F" />
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

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
 * Uses the brand green gradient (#1B4332 → #40916C) and white house icon.
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
            <stop offset="100%" stopColor="#40916C" />
          </linearGradient>
        </defs>
        {/* Background */}
        <rect width="40" height="40" rx="8" fill={`url(#${gradientId})`} />
        {/* House roof */}
        <polygon points="20,6 7,18 33,18" fill="white" />
        {/* Chimney */}
        <rect x="25" y="8" width="3.5" height="7" fill="white" />
        {/* House body */}
        <rect x="10" y="18" width="20" height="14" fill="white" />
        {/* Door */}
        <rect x="16.5" y="23" width="7" height="9" rx="1" fill="#40916C" />
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

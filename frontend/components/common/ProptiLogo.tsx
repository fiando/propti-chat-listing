import Image from 'next/image';
import { cn } from '@/lib/utils';

interface ProptiLogoProps {
  size?: number;
  className?: string;
  /** Show the wordmark "Propti" next to the icon */
  showWordmark?: boolean;
  /** Color variant for the wordmark text */
  wordmarkColor?: 'dark' | 'white';
}

export function ProptiLogo({
  size = 36,
  className,
  showWordmark = false,
  wordmarkColor = 'dark',
}: ProptiLogoProps) {
  const src = showWordmark ? '/propti-logo.svg' : '/propti-mark.svg';
  const width = showWordmark ? Math.round(size * 2.45) : Math.round(size * 0.87);

  return (
    <span
      className={cn(
        'inline-flex items-center flex-shrink-0',
        showWordmark && wordmarkColor === 'white' && 'rounded-xl bg-white px-3 py-2',
        className
      )}
    >
      <Image src={src} alt="Propti" width={width} height={size} priority={showWordmark} />
    </span>
  );
}

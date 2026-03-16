import Image from 'next/image';
import { cn } from '@/lib/utils';

interface ProptiLogoProps {
  size?: number;
  className?: string;
  showWordmark?: boolean;
}

const FULL_LOGO_ASPECT_RATIO = 337 / 138;
const ICON_MARK_ASPECT_RATIO = 120 / 138;
const FULL_LOGO_SRC = '/propti-logo.svg';
const ICON_MARK_SRC = '/propti-mark.svg';
const MIN_IMAGE_DIMENSION_PX = 1;

export function ProptiLogo({
  size = 36,
  className,
  showWordmark = false,
}: ProptiLogoProps) {
  const src = showWordmark ? FULL_LOGO_SRC : ICON_MARK_SRC;
  const alt = showWordmark ? 'Propti logo' : 'Propti icon';
  const width = Math.max(
    MIN_IMAGE_DIMENSION_PX,
    Math.round(size * (showWordmark ? FULL_LOGO_ASPECT_RATIO : ICON_MARK_ASPECT_RATIO))
  );

  return (
    <span className={cn('inline-flex items-center flex-shrink-0', className)}>
      <Image src={src} alt={alt} width={width} height={size} priority={showWordmark} />
    </span>
  );
}

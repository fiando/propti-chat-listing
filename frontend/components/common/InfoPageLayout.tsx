import Link from 'next/link';

type InfoSection = {
  title: string;
  paragraphs: string[];
  bullets?: string[];
  supportLink?: {
    label: string;
    href: string;
  };
};

type InfoPageLayoutProps = {
  eyebrow: string;
  title: string;
  description: string;
  sections: InfoSection[];
  primaryCta?: {
    label: string;
    href: string;
  };
  secondaryCta?: {
    label: string;
    href: string;
  };
  updatedLabel?: string;
};

export function InfoPageLayout({
  eyebrow,
  title,
  description,
  sections,
  primaryCta,
  secondaryCta,
  updatedLabel = 'Terakhir diperbarui: 13 Maret 2026',
}: InfoPageLayoutProps) {
  return (
    <div className="bg-gradient-to-b from-brand-light/40 to-transparent">
      <div className="mx-auto max-w-5xl px-4 py-12 md:py-16">
        <div className="mb-8">
          <Link href="/" className="text-sm text-brand-primary hover:underline">
            ← Kembali ke beranda
          </Link>
        </div>

        <div className="rounded-3xl bg-white shadow-sm ring-1 ring-black/5 overflow-hidden">
          <div className="bg-brand-primary px-6 py-10 text-white md:px-10">
            <p className="text-sm font-semibold uppercase tracking-[0.2em] text-white/70">{eyebrow}</p>
            <h1 className="mt-3 text-3xl font-bold md:text-4xl">{title}</h1>
            <p className="mt-4 max-w-3xl text-sm leading-7 text-white/80 md:text-base">{description}</p>

            {(primaryCta || secondaryCta) && (
              <div className="mt-6 flex flex-wrap gap-3">
                {primaryCta && (
                  <Link
                    href={primaryCta.href}
                    className="rounded-full bg-white px-5 py-2.5 text-sm font-semibold text-brand-primary transition hover:bg-white/90"
                  >
                    {primaryCta.label}
                  </Link>
                )}
                {secondaryCta && (
                  <Link
                    href={secondaryCta.href}
                    className="rounded-full border border-white/25 px-5 py-2.5 text-sm font-semibold text-white transition hover:bg-white/10"
                  >
                    {secondaryCta.label}
                  </Link>
                )}
              </div>
            )}
          </div>

          <div className="space-y-6 px-6 py-8 md:px-10 md:py-10">
            <p className="text-sm text-gray-500">{updatedLabel}</p>

            {sections.map((section) => (
              <section key={section.title} className="rounded-2xl bg-gray-50 p-6">
                <h2 className="text-xl font-semibold text-gray-900">{section.title}</h2>
                <div className="mt-3 space-y-3 text-sm leading-7 text-gray-600 md:text-base">
                  {section.paragraphs.map((paragraph) => (
                    <p key={paragraph}>{paragraph}</p>
                  ))}
                </div>
                {section.supportLink && (
                  <a
                    href={section.supportLink.href}
                    className="mt-4 inline-flex text-sm font-semibold text-brand-primary transition hover:underline md:text-base"
                  >
                    {section.supportLink.label}
                  </a>
                )}
                {section.bullets && section.bullets.length > 0 && (
                  <ul className="mt-4 space-y-2 text-sm leading-7 text-gray-600 md:text-base">
                    {section.bullets.map((bullet) => (
                      <li key={bullet} className="flex gap-3">
                        <span className="mt-2 h-2 w-2 flex-shrink-0 rounded-full bg-brand-secondary" />
                        <span>{bullet}</span>
                      </li>
                    ))}
                  </ul>
                )}
              </section>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

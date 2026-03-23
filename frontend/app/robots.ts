import type { MetadataRoute } from 'next';

export default function robots(): MetadataRoute.Robots {
  return {
    rules: {
      userAgent: '*',
      allow: ['/listings/', '/'],
      disallow: ['/api/', '/login', '/callback', '/settings', '/profile'],
    },
    sitemap: 'https://propti.id/sitemap.xml',
  };
}

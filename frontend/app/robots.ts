import type { MetadataRoute } from 'next';

export default function robots(): MetadataRoute.Robots {
  return {
    rules: {
      userAgent: '*',
      allow: ['/listings/', '/'],
      disallow: ['/api/', '/login', '/callback', '/profile'],
    },
    sitemap: 'https://propti.id/sitemap.xml',
  };
}

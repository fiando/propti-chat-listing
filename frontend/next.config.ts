import type { NextConfig } from 'next';

const nextConfig: NextConfig = {
  images: {
    domains: [
      'lh3.googleusercontent.com',
      's3.amazonaws.com',
      'propti-media.s3.amazonaws.com',
    ],
  },
};

export default nextConfig;

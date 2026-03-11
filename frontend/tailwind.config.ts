import type { Config } from 'tailwindcss';

const config: Config = {
  content: [
    './pages/**/*.{js,ts,jsx,tsx,mdx}',
    './components/**/*.{js,ts,jsx,tsx,mdx}',
    './app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        brand: {
          primary: '#1B4332',
          secondary: '#40916C',
          accent: '#74C69D',
          gold: '#D4A853',
          light: '#D8F3DC',
        },
        whatsapp: '#25D366',
      },
      fontFamily: {
        sans: ['var(--font-inter)', 'system-ui', 'sans-serif'],
      },
      backgroundImage: {
        'gradient-brand': 'linear-gradient(135deg, #1B4332 0%, #40916C 100%)',
        'gradient-hero': 'linear-gradient(135deg, #1B4332 0%, #2D6A4F 50%, #40916C 100%)',
      },
      boxShadow: {
        card: '0 2px 16px rgba(27, 67, 50, 0.08)',
        'card-hover': '0 8px 32px rgba(27, 67, 50, 0.16)',
      },
    },
  },
  plugins: [],
};

export default config;

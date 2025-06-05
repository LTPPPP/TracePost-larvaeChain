import type { NextConfig } from 'next';
import path from 'path';

const nextConfig: NextConfig = {
  // Enable standalone output for Docker
  output: 'standalone',
  
  webpack(config) {
    config.resolve.alias = {
      ...config.resolve.alias,
      '@': path.resolve(__dirname, 'src'),
      '@components': path.resolve(__dirname, 'src/components'),
      '@layouts': path.resolve(__dirname, 'src/components/layouts'),
      '@ui': path.resolve(__dirname, 'src/components/ui'),
      '@hooks': path.resolve(__dirname, 'src/hooks'),
      '@middlewares': path.resolve(__dirname, 'src/middlewares'),
      '@styles': path.resolve(__dirname, 'src/styles')
    };
    // SVGR config: chỉ áp dụng khi import SVG từ TS/JS
    config.module.rules.push({
      test: /\.svg$/,
      use: ['@svgr/webpack']
    });
    return config;
  }
};

export default nextConfig;

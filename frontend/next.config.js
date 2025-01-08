/** @type {import('next').NextConfig} */
const nextConfig = {
  experimental: {
    appDir: true,
  },
  poweredByHeader: false,
  reactStrictMode: true,
  pageExtensions: ['js', 'jsx', 'ts', 'tsx'],
  // Remove the webpack configuration as Next.js handles CSS modules automatically
  optimizeFonts: true,
  swcMinify: true,
  compiler: {
    removeConsole: process.env.NODE_ENV === "production",
  }
}

module.exports = nextConfig

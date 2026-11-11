/** @type {import('next').NextConfig} */
const nextConfig = {
  // Standalone output makes a small Docker image for the frontend.
  output: "standalone",
  reactStrictMode: true,
};

export default nextConfig;

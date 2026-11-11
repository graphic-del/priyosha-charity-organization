/** @type {import('next').NextConfig} */
const nextConfig = {
  // Standalone output makes a small Docker image for the frontend.
  output: "standalone",
  reactStrictMode: true,
  // Do not leak the framework version in a response header.
  poweredByHeader: false,
  async headers() {
    // Security headers for every page. The Content-Security-Policy allows
    // the site itself, the API at https://api.priyoshacharity.com, images
    // (including the logo and data: images), and the inline styles/scripts
    // Next.js needs to render. Nothing may be framed from here.
    // See SECURITY.md — no site is 100% unhackable, but these headers
    // close off whole classes of attacks (clickjacking, MIME sniffing,
    // and loading the site inside another page).
    const securityHeaders = [
      { key: "Strict-Transport-Security", value: "max-age=31536000; includeSubDomains" },
      { key: "X-Content-Type-Options", value: "nosniff" },
      { key: "X-Frame-Options", value: "DENY" },
      { key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
      { key: "Permissions-Policy", value: "camera=(), microphone=(), geolocation=()" },
      {
        key: "Content-Security-Policy",
        value: [
          "default-src 'self'",
          "script-src 'self' 'unsafe-inline'",
          "style-src 'self' 'unsafe-inline'",
          "img-src 'self' data: https:",
          "font-src 'self'",
          "connect-src 'self' https://api.priyoshacharity.com",
          "frame-ancestors 'none'",
          "base-uri 'self'",
          "form-action 'self'",
        ].join("; "),
      },
    ];
    return [{ source: "/:path*", headers: securityHeaders }];
  },
};

export default nextConfig;

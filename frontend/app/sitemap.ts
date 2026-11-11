import type { MetadataRoute } from "next";

// Canonical site URL — same source as layout.tsx. Live domain:
// https://priyoshacharity.com
const siteUrl =
  process.env.NEXT_PUBLIC_SITE_URL ||
  process.env.FRONTEND_URL ||
  "https://priyoshacharity.com";

export default function sitemap(): MetadataRoute.Sitemap {
  return [
    {
      url: siteUrl,
      lastModified: new Date(),
      changeFrequency: "weekly",
      priority: 1,
    },
  ];
}

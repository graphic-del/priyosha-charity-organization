import type { MetadataRoute } from "next";

// Canonical site URL — same source as layout.tsx. Live domain:
// https://priyoshacharity.com
const siteUrl =
  process.env.NEXT_PUBLIC_SITE_URL ||
  process.env.FRONTEND_URL ||
  "https://priyoshacharity.com";

export default function sitemap(): MetadataRoute.Sitemap {
  const now = new Date();
  return ["", "/need", "/programs", "/give", "/contact", "/about"].map((path, i) => ({
    url: `${siteUrl}${path}`,
    lastModified: now,
    changeFrequency: "weekly" as const,
    priority: i === 0 ? 1 : 0.8,
  }));
}

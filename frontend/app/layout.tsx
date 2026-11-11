import type { Metadata } from "next";
import "./globals.css";

// Canonical site URL for metadata / sitemap / robots. The live domain is
// https://priyoshacharity.com; override with NEXT_PUBLIC_SITE_URL or
// FRONTEND_URL if it ever changes.
const siteUrl =
  process.env.NEXT_PUBLIC_SITE_URL ||
  process.env.FRONTEND_URL ||
  "https://priyoshacharity.com";

export const metadata: Metadata = {
  metadataBase: new URL(siteUrl),
  title: "Priyosha Charity Organization — Better futures for people and animals in Bangladesh",
  description:
    "Priyosha Charity Organization is a US-based nonprofit raising charity from around the world to reduce problems in Bangladesh — for people facing hunger and homelessness, and for cats and dogs who suffer the most.",
  openGraph: {
    title: "Priyosha Charity Organization",
    description:
      "A US-based nonprofit raising charity from around the world for people and animals in Bangladesh.",
    url: siteUrl,
    siteName: "Priyosha Charity Organization",
    type: "website",
  },
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}

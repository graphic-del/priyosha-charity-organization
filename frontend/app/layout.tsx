import type { Metadata } from "next";
import "./globals.css";
import Reveal from "./reveal";

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
      <body>
        <header className="site">
          <div className="nav">
            <a className="brand" href="/">
              <img src="/logo.png" alt="Priyosha Charity logo — a person embracing a dog, forming a heart" />
              <span>
                <strong>Priyosha Charity Organization</strong>
                <small>US-based nonprofit · Helping Bangladesh</small>
              </span>
            </a>
            <nav>
              <a href="/">Home</a>
              <a href="/need">The Need</a>
              <a href="/programs">Programs</a>
              <a href="/about">About</a>
              <a href="/contact">Contact</a>
              <a className="give-link" href="/give">Give</a>
            </nav>
          </div>
        </header>

        {children}
        <Reveal />

        <footer className="site">
          <div className="wrap">
            <div className="foot-grid">
              <div>
                <img src="/logo.png" alt="Priyosha Charity logo" />
                <p>
                  <strong>Priyosha Charity Organization</strong> — a US-based nonprofit.
                  <br />
                  Better futures for humans and animals in Bangladesh.
                </p>
              </div>
              <div>
                <h3>Explore</h3>
                <p><a href="/">Home</a></p>
                <p><a href="/need">The Need</a></p>
                <p><a href="/programs">Programs</a></p>
                <p><a href="/about">About</a></p>
              </div>
              <div>
                <h3>Get involved</h3>
                <p><a href="/give">Give — make a pledge</a></p>
                <p><a href="/contact">Contact &amp; volunteer</a></p>
                <p>Food for people</p>
                <p>Food and care for cats and dogs</p>
              </div>
            </div>
            <p className="foot-bottom">
              Figures on this site are shown with their sources, and with the warning
              that they cannot be added together.
            </p>
          </div>
        </footer>
      </body>
    </html>
  );
}

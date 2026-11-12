"use client";

import { useEffect, useState } from "react";
import { API_URL, FALLBACK_PROGRAMS, FALLBACK_STATS, type Program, type Stat } from "./data";

export default function Home() {
  const [stats, setStats] = useState<Stat[]>(FALLBACK_STATS);
  const [programs, setPrograms] = useState<Program[]>(FALLBACK_PROGRAMS);

  useEffect(() => {
    fetch(`${API_URL}/api/stats`)
      .then((r) => (r.ok ? r.json() : Promise.reject(new Error("stats failed"))))
      .then((d) => { if (Array.isArray(d.stats)) setStats(d.stats); })
      .catch(() => {});
    fetch(`${API_URL}/api/programs`)
      .then((r) => (r.ok ? r.json() : Promise.reject(new Error("programs failed"))))
      .then((d) => { if (Array.isArray(d.programs)) setPrograms(d.programs); })
      .catch(() => {});
  }, []);

  return (
    <main className="page-anim">
      <section className="hero-band band-maroon">
        <div className="wrap">
          <div className="hero hero-full">
            <div className="hero-copy">
              <p className="eyebrow">Priyosha Charity Organization</p>
              <h1>Priyosha backs Bangladesh.</h1>
              <p className="hero-sub">
                Better futures for people <span>and animals</span> in Bangladesh.
              </p>
              <p>
                Priyosha Charity Organization is a US-based nonprofit. We raise charity
                from around the world to reduce problems in Bangladesh — for people
                facing hunger and homelessness, and for cats and dogs who suffer the most.
              </p>
              <p className="cta-line">Join. Give. Spark Change.</p>
              <div className="actions">
                <a className="btn light" href="/give">Give now</a>
                <a className="btn ghost-light" href="/programs">See our programs</a>
                <a className="btn ghost-light" href="/contact">Volunteer</a>
              </div>
            </div>
            <div className="logo-card float">
              <img
                className="logo"
                src="/logo.png"
                alt="Priyosha Charity logo — a person embracing a dog, forming a heart"
              />
            </div>
          </div>
        </div>
      </section>

      <section className="band-white" style={{ paddingTop: 0, paddingBottom: 8 }}>
        <div className="wrap">
          <div className="impact reveal">
            <div><strong>18.1 million</strong><span>people facing crisis-level food insecurity</span></div>
            <div><strong>22,185</strong><span>people counted sleeping in public places</span></div>
            <div><strong>16.7 lakh</strong><span>dogs estimated in Bangladesh</span></div>
          </div>
        </div>
      </section>

      <section className="band">
        <div className="wrap reveal">
          <h2>Our mission</h2>
          <p className="lede">
            To bring support from around the world to Bangladesh, and turn it into
            food, shelter and care — for humans and for animals, especially the cats
            and dogs who suffer the most.
          </p>
          <div className="mission-grid">
            <div><h3>Food</h3><p>Meals and food packs for families who do not have enough to eat.</p></div>
            <div><h3>Shelter</h3><p>Support for people sleeping rough, and a path to more secure housing.</p></div>
            <div><h3>Care for animals</h3><p>Food, shelter and basic veterinary care for stray cats and dogs.</p></div>
          </div>
          <p style={{ marginTop: 28 }}>
            <a className="btn ghost" href="/about">More about us →</a>
          </p>
        </div>
      </section>

      <section className="band-white">
        <div className="wrap reveal">
          <h2>Explore the issues &amp; our impact</h2>
          <p className="lede">The need is large — click a tile to open The Need page and read the full story, with sources.</p>
          <div className="issue-tiles">
            {stats.map((s, i) => (
              <a className={`issue-tile tone-${i % 4} tile-link`} key={s.id} href="/need">
                <p className="figure">{s.figure}</p>
                <h3>{s.title}</h3>
                <p>{s.detail}</p>
                <p className="source">Read more on The Need page →</p>
              </a>
            ))}
          </div>
        </div>
      </section>

      <section className="band">
        <div className="wrap reveal">
          <h2>Get involved</h2>
          <p className="lede">Three programmes, one goal: a better future for humans and animals.</p>
          <div className="involve-grid">
            {programs.map((p) => (
              <article className="involve-card" key={p.id}>
                <div className="involve-top"><span className="tag">{p.focus}</span></div>
                <h3>{p.title}</h3>
                <p>{p.description}</p>
                <a className="btn" href="/programs">See this program →</a>
              </article>
            ))}
          </div>
          <div className="involve-more">
            <a className="btn" href="/give">Make a pledge</a>
            <a className="btn ghost" href="/contact">Volunteer with us</a>
          </div>
        </div>
      </section>
    </main>
  );
}

"use client";

import { useEffect, useState } from "react";
import { API_URL, FALLBACK_STATS, FALLBACK_WARNING, type Stat } from "../data";

export default function NeedPage() {
  const [stats, setStats] = useState<Stat[]>(FALLBACK_STATS);
  const [warning, setWarning] = useState(FALLBACK_WARNING);

  useEffect(() => {
    fetch(`${API_URL}/api/stats`)
      .then((r) => (r.ok ? r.json() : Promise.reject(new Error("stats failed"))))
      .then((d) => {
        if (Array.isArray(d.stats)) setStats(d.stats);
        if (typeof d.warning === "string") setWarning(d.warning);
      })
      .catch(() => {});
  }, []);

  return (
    <main className="page-anim">
      <section className="band-maroon page-head">
        <div className="wrap">
          <p className="eyebrow">The Need</p>
          <h1>The Need in Bangladesh</h1>
          <p className="lede">
            The need is large — and the numbers we have are incomplete. We show what
            is known, what is not known, and where each figure comes from.
          </p>
        </div>
      </section>

      <section className="band-white">
        <div className="wrap reveal">
          <div className="issue-tiles">
            {stats.map((s, i) => (
              <article className={`issue-tile tone-${i % 4}`} key={s.id}>
                <p className="figure">{s.figure}</p>
                <h3>{s.title}</h3>
                <p>{s.detail}</p>
                <p className="source">Source: {s.source}</p>
              </article>
            ))}
          </div>
          <p className="warning"><strong>Please read:</strong> {warning}</p>
          <div className="involve-more">
            <a className="btn" href="/give">Give to help →</a>
            <a className="btn ghost" href="/programs">See our programs</a>
          </div>
        </div>
      </section>
    </main>
  );
}

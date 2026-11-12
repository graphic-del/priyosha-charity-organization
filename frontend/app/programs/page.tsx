"use client";

import { useEffect, useState } from "react";
import { API_URL, FALLBACK_PROGRAMS, type Program } from "../data";

export default function ProgramsPage() {
  const [programs, setPrograms] = useState<Program[]>(FALLBACK_PROGRAMS);

  useEffect(() => {
    fetch(`${API_URL}/api/programs`)
      .then((r) => (r.ok ? r.json() : Promise.reject(new Error("programs failed"))))
      .then((d) => { if (Array.isArray(d.programs)) setPrograms(d.programs); })
      .catch(() => {});
  }, []);

  return (
    <main className="page-anim">
      <section className="band-maroon page-head">
        <div className="wrap">
          <p className="eyebrow">Programs</p>
          <h1>Get involved</h1>
          <p className="lede">Three programmes, one goal: a better future for humans and animals. There is a part for you to play.</p>
        </div>
      </section>

      <section className="band">
        <div className="wrap reveal">
          <div className="involve-grid">
            {programs.map((p) => (
              <article className="involve-card" key={p.id}>
                <div className="involve-top"><span className="tag">{p.focus}</span></div>
                <h3>{p.title}</h3>
                <p>{p.description}</p>
                <a className="btn" href="/give">Give to this work →</a>
              </article>
            ))}
          </div>
          <div className="involve-more">
            <a className="btn ghost" href="/contact">Volunteer with us</a>
            <a className="btn ghost" href="/need">See the need</a>
          </div>
        </div>
      </section>
    </main>
  );
}

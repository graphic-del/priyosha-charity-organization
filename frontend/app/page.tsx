"use client";

import { useEffect, useState } from "react";

// The Go api-gateway. Set NEXT_PUBLIC_API_URL in production to your
// deployed gateway URL, e.g. https://api.priyoshacharity.com
const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

type Stat = {
  id: string;
  title: string;
  figure: string;
  detail: string;
  source: string;
};

type Program = {
  id: string;
  title: string;
  description: string;
  focus: string;
};

// Fallback copy of the content-service data, used only if the API cannot be
// reached (so the page still shows the correct figures). The API is the
// source of truth when it is up — edit figures in backend/content-service.
const FALLBACK_WARNING =
  "These figures cannot be added together: hunger and homelessness can affect the same people, and animal population estimates do not measure suffering.";

const FALLBACK_STATS: Stat[] = [
  {
    id: "food-insecurity",
    title: "People lacking enough food",
    figure: "18.1 million — 1 crore 81 lakh",
    detail:
      "Projected to face crisis-level food insecurity or worse during September–December 2026, within the areas assessed.",
    source: "Integrated Food Security Phase Classification (IPC)",
  },
  {
    id: "sleeping-in-public",
    title: "People sleeping in public places",
    figure: "22,185",
    detail:
      "Counted in the 2022 census. This narrow category covers people sleeping on streets, at stations and similar places. It does not capture everyone without secure housing, and it is an older count.",
    source: "Bangladesh Bureau of Statistics, 2022 census (nsds.bbs.gov.bd)",
  },
  {
    id: "dogs",
    title: "Dogs",
    figure: "About 16.7 lakh dogs",
    detail:
      "Estimated in a 2024 study. This estimates the dog population; it does not establish how many are homeless or hungry.",
    source: "2024 study (PMC)",
  },
  {
    id: "animals-overall",
    title: "Hungry or homeless animals overall, including cats",
    figure: "No reliable nationwide total found",
    detail:
      "Research identifies gaps in Bangladesh's stray-animal population data. Giving an exact suffering count would be guesswork.",
    source: "pmc.ncbi.nlm.nih.gov",
  },
];

const FALLBACK_PROGRAMS: Program[] = [
  {
    id: "food-for-people",
    title: "Food for people",
    description:
      "Food support for people facing crisis-level food insecurity in Bangladesh — meals and food packs for families who do not have enough to eat.",
    focus: "People",
  },
  {
    id: "support-sleeping-rough",
    title: "Support for people sleeping rough",
    description:
      "Help for people sleeping in public places — on streets, at stations and similar places — including shelter support, essentials and a path to more secure housing.",
    focus: "People",
  },
  {
    id: "care-for-animals",
    title: "Food and care for cats and dogs",
    description:
      "Food, shelter and care for stray cats and dogs who suffer the most — feeding, basic veterinary care and rescue where we can reach them.",
    focus: "Animals — cats and dogs",
  },
];

const AMOUNTS = [10, 25, 50, 100, 250];

export default function Home() {
  const [stats, setStats] = useState<Stat[]>(FALLBACK_STATS);
  const [warning, setWarning] = useState(FALLBACK_WARNING);
  const [programs, setPrograms] = useState<Program[]>(FALLBACK_PROGRAMS);

  // Pledge calculator state
  const [amount, setAmount] = useState(50);
  const [customAmount, setCustomAmount] = useState("");
  const [frequency, setFrequency] = useState<"one-time" | "monthly">("one-time");
  const [donorName, setDonorName] = useState("");
  const [donorEmail, setDonorEmail] = useState("");
  const [pledgeStatus, setPledgeStatus] = useState<{ ok: boolean; text: string } | null>(null);
  const [pledgeBusy, setPledgeBusy] = useState(false);

  // Contact form state
  const [contactName, setContactName] = useState("");
  const [contactEmail, setContactEmail] = useState("");
  const [interest, setInterest] = useState("donate");
  const [contactMessage, setContactMessage] = useState("");
  const [contactStatus, setContactStatus] = useState<{ ok: boolean; text: string } | null>(null);
  const [contactBusy, setContactBusy] = useState(false);

  // Load the need and programmes from the Go content-service via the gateway.
  useEffect(() => {
    fetch(`${API_URL}/api/stats`)
      .then((r) => (r.ok ? r.json() : Promise.reject(new Error("stats failed"))))
      .then((d) => {
        if (Array.isArray(d.stats)) setStats(d.stats);
        if (typeof d.warning === "string") setWarning(d.warning);
      })
      .catch(() => {
        // Keep the fallback data above; the page still shows correct figures.
      });
    fetch(`${API_URL}/api/programs`)
      .then((r) => (r.ok ? r.json() : Promise.reject(new Error("programs failed"))))
      .then((d) => {
        if (Array.isArray(d.programs)) setPrograms(d.programs);
      })
      .catch(() => {});
  }, []);

  const effectiveAmount = customAmount ? Number(customAmount) || 0 : amount;

  async function submitPledge(e: React.FormEvent) {
    e.preventDefault();
    setPledgeStatus(null);
    if (!effectiveAmount || effectiveAmount <= 0) {
      setPledgeStatus({ ok: false, text: "Please choose or enter an amount greater than $0." });
      return;
    }
    setPledgeBusy(true);
    try {
      const res = await fetch(`${API_URL}/api/pledges`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          amount: effectiveAmount,
          frequency,
          name: donorName,
          email: donorEmail,
        }),
      });
      if (!res.ok) throw new Error(await res.text());
      const data = await res.json();
      setPledgeStatus({
        ok: true,
        text:
          data.message ||
          `Thank you, ${donorName} — your pledge of $${effectiveAmount} ${frequency} has been recorded.`,
      });
    } catch (err) {
      setPledgeStatus({
        ok: false,
        text: `Sorry — we could not record your pledge. Please check that the backend is running (${API_URL}) and try again.`,
      });
    } finally {
      setPledgeBusy(false);
    }
  }

  async function submitContact(e: React.FormEvent) {
    e.preventDefault();
    setContactStatus(null);
    setContactBusy(true);
    try {
      const res = await fetch(`${API_URL}/api/messages`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: contactName,
          email: contactEmail,
          interest,
          message: contactMessage,
        }),
      });
      if (!res.ok) throw new Error(await res.text());
      setContactStatus({ ok: true, text: "Thank you — your message has been recorded. We will get back to you." });
      setContactMessage("");
    } catch (err) {
      setContactStatus({
        ok: false,
        text: `Sorry — we could not send your message. Please check that the backend is running (${API_URL}) and try again.`,
      });
    } finally {
      setContactBusy(false);
    }
  }

  return (
    <>
      <header className="site">
        <div className="nav">
          <a className="brand" href="#">
            <img src="/logo.png" alt="Priyosha Charity logo — a person embracing a dog, forming a heart" />
            <span>
              <strong>Priyosha Charity Organization</strong>
              <small>US-based nonprofit · Helping Bangladesh</small>
            </span>
          </a>
          <nav>
            <a href="#need">The Need</a>
            <a href="#programs">Programs</a>
            <a href="#give">Give</a>
            <a href="#contact">Contact</a>
          </nav>
        </div>
      </header>

      <main>
        <section className="hero">
          <div>
            <h1>
              Better futures for people <span>and animals</span> in Bangladesh.
            </h1>
            <p>
              Priyosha Charity Organization is a US-based nonprofit. We raise charity
              from around the world to reduce problems in Bangladesh — for people
              facing hunger and homelessness, and for cats and dogs who suffer the most.
            </p>
            <div className="actions">
              <a className="btn" href="#give">Give</a>
              <a className="btn ghost" href="#programs">See our programs</a>
            </div>
          </div>
          <img
            className="logo"
            src="/logo.png"
            alt="Priyosha Charity logo — a person embracing a dog, forming a heart"
          />
        </section>

        <section id="mission">
          <h2>Our mission</h2>
          <p className="lede">
            To bring support from around the world to Bangladesh, and turn it into
            food, shelter and care — for humans and for animals, especially the cats
            and dogs who suffer the most.
          </p>
        </section>

        <section id="need">
          <h2>The Need in Bangladesh</h2>
          <p className="lede">
            The need is large — and the numbers we have are incomplete. We show what
            is known, what is not known, and where each figure comes from.
          </p>
          <div className="cards">
            {stats.map((s) => (
              <article className="card" key={s.id}>
                <h3>{s.title}</h3>
                <p className="figure">{s.figure}</p>
                <p>{s.detail}</p>
                <p className="source">Source: {s.source}</p>
              </article>
            ))}
          </div>
          <p className="warning">
            <strong>Please read:</strong> {warning}
          </p>
        </section>

        <section id="programs">
          <h2>Programs</h2>
          <p className="lede">Three programmes, one goal: a better future for humans and animals.</p>
          <div className="cards">
            {programs.map((p) => (
              <article className="card" key={p.id}>
                <h3>{p.title}</h3>
                <p>{p.description}</p>
                <p className="source">Focus: {p.focus}</p>
              </article>
            ))}
          </div>
        </section>

        <section id="give">
          <h2>Give — make a pledge</h2>
          <p className="lede">
            Choose an amount and how often you would like to give. Your pledge is
            recorded by our donation service. No payment is taken on this site yet —
            see the note below.
          </p>
          <div className="calculator">
            <form onSubmit={submitPledge}>
              <div className="amounts" role="group" aria-label="Pledge amount">
                {AMOUNTS.map((a) => (
                  <button
                    type="button"
                    key={a}
                    className={!customAmount && amount === a ? "selected" : ""}
                    onClick={() => {
                      setAmount(a);
                      setCustomAmount("");
                    }}
                  >
                    ${a}
                  </button>
                ))}
              </div>
              <label htmlFor="custom-amount">Or enter your own amount (USD)</label>
              <input
                id="custom-amount"
                type="number"
                min="1"
                step="1"
                value={customAmount}
                placeholder="e.g. 75"
                onChange={(e) => setCustomAmount(e.target.value)}
              />
              <div className="freq">
                <label>
                  <input
                    type="radio"
                    name="frequency"
                    checked={frequency === "one-time"}
                    onChange={() => setFrequency("one-time")}
                  />{" "}
                  One-time
                </label>
                <label>
                  <input
                    type="radio"
                    name="frequency"
                    checked={frequency === "monthly"}
                    onChange={() => setFrequency("monthly")}
                  />{" "}
                  Monthly
                </label>
              </div>
              <p className="pledge-summary">
                Your pledge: <strong>${effectiveAmount || 0} {frequency}</strong>
                {frequency === "monthly" ? " — that's $" + (effectiveAmount || 0) * 12 + " over a year." : ""}
              </p>
              <label htmlFor="donor-name">Your name</label>
              <input
                id="donor-name"
                type="text"
                required
                value={donorName}
                onChange={(e) => setDonorName(e.target.value)}
                placeholder="Your name"
              />
              <label htmlFor="donor-email">Your email</label>
              <input
                id="donor-email"
                type="email"
                required
                value={donorEmail}
                onChange={(e) => setDonorEmail(e.target.value)}
                placeholder="you@example.org"
              />
              <p style={{ marginTop: 16 }}>
                <button className="btn" type="submit" disabled={pledgeBusy}>
                  {pledgeBusy ? "Recording…" : "Record my pledge"}
                </button>
              </p>
              {pledgeStatus && (
                <p className={`notice${pledgeStatus.ok ? "" : " error"}`}>{pledgeStatus.text}</p>
              )}
              <p className="source">
                Note: this records a pledge of intent. Real payment processing
                (for example Stripe) is not wired up yet — see the README.
              </p>
            </form>
          </div>
        </section>

        <section id="contact">
          <h2>Contact &amp; volunteer</h2>
          <p className="lede">
            Want to help, volunteer, or ask a question? Send us a message — it is
            recorded by our contact service and we will get back to you.
          </p>
          <div className="form-box">
            <form onSubmit={submitContact}>
              <label htmlFor="contact-name">Your name</label>
              <input
                id="contact-name"
                type="text"
                required
                value={contactName}
                onChange={(e) => setContactName(e.target.value)}
                placeholder="Your name"
              />
              <label htmlFor="contact-email">Your email</label>
              <input
                id="contact-email"
                type="email"
                required
                value={contactEmail}
                onChange={(e) => setContactEmail(e.target.value)}
                placeholder="you@example.org"
              />
              <label htmlFor="interest">I am interested in</label>
              <select id="interest" value={interest} onChange={(e) => setInterest(e.target.value)}>
                <option value="donate">Donating</option>
                <option value="volunteer">Volunteering</option>
                <option value="other">Something else</option>
              </select>
              <label htmlFor="contact-message">Your message</label>
              <textarea
                id="contact-message"
                required
                value={contactMessage}
                onChange={(e) => setContactMessage(e.target.value)}
                placeholder="How would you like to help?"
              />
              <p style={{ marginTop: 16 }}>
                <button className="btn" type="submit" disabled={contactBusy}>
                  {contactBusy ? "Sending…" : "Send message"}
                </button>
              </p>
              {contactStatus && (
                <p className={`notice${contactStatus.ok ? "" : " error"}`}>{contactStatus.text}</p>
              )}
            </form>
            <p className="source">
              Contact details: email and phone are not published yet — replace the
              placeholder in the README with the organization&apos;s real contact
              details before launch. Placeholder: donate@priyosha-charity.example
            </p>
          </div>
        </section>

        <section id="founder">
          <h2>Our founder</h2>
          <div className="founder">
            <div className="initials" aria-hidden="true">SM</div>
            <div>
              <h3 style={{ margin: "0 0 6px" }}>Md Shadman Mahmood</h3>
              <p>
                Md Shadman Mahmood founded Priyosha Charity Organization to help the
                people and animals of Bangladesh who suffer the most — people facing
                hunger and sleeping in public places, and the cats and dogs with no
                one to care for them. The organization is based in the United States
                and raises charity from around the world for that work.
              </p>
            </div>
          </div>
        </section>
      </main>

      <footer className="site">
        <img src="/logo.png" alt="Priyosha Charity logo" />
        <p>
          <strong>Priyosha Charity Organization</strong> — a US-based nonprofit.
          <br />
          Better futures for humans and animals in Bangladesh.
        </p>
        <p className="source">
          Figures on this page are shown with their sources, and with the warning
          that they cannot be added together.
        </p>
      </footer>
    </>
  );
}

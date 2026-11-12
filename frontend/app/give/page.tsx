"use client";

import { useState } from "react";
import { API_URL } from "../data";

const AMOUNTS = [10, 25, 50, 100, 250];

export default function GivePage() {
  const [amount, setAmount] = useState(50);
  const [customAmount, setCustomAmount] = useState("");
  const [frequency, setFrequency] = useState<"one-time" | "monthly">("one-time");
  const [donorName, setDonorName] = useState("");
  const [donorEmail, setDonorEmail] = useState("");
  const [pledgeStatus, setPledgeStatus] = useState<{ ok: boolean; text: string } | null>(null);
  const [pledgeBusy, setPledgeBusy] = useState(false);

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
        body: JSON.stringify({ amount: effectiveAmount, frequency, name: donorName, email: donorEmail }),
      });
      if (!res.ok) throw new Error(await res.text());
      const data = await res.json();
      setPledgeStatus({
        ok: true,
        text: data.message || `Thank you, ${donorName} — your pledge of $${effectiveAmount} ${frequency} has been recorded.`,
      });
    } catch {
      setPledgeStatus({
        ok: false,
        text: `Sorry — we could not record your pledge. Please check that the backend is running (${API_URL}) and try again.`,
      });
    } finally {
      setPledgeBusy(false);
    }
  }

  return (
    <main className="page-anim">
      <section className="band-maroon page-head">
        <div className="wrap">
          <p className="eyebrow">Give</p>
          <h1>Give — make a pledge</h1>
          <p className="lede">
            Choose an amount and how often you would like to give. Your pledge is
            recorded by our donation service. No payment is taken on this site yet —
            see the note below.
          </p>
        </div>
      </section>

      <section className="band-white">
        <div className="wrap reveal" style={{ maxWidth: 760 }}>
          <div className="calculator" style={{ marginTop: 0 }}>
            <form onSubmit={submitPledge}>
              <div className="amounts" role="group" aria-label="Pledge amount">
                {AMOUNTS.map((a) => (
                  <button
                    type="button"
                    key={a}
                    className={!customAmount && amount === a ? "selected" : ""}
                    onClick={() => { setAmount(a); setCustomAmount(""); }}
                  >
                    ${a}
                  </button>
                ))}
              </div>
              <label htmlFor="custom-amount">Or enter your own amount (USD)</label>
              <input id="custom-amount" type="number" min="1" step="1" value={customAmount} placeholder="e.g. 75" onChange={(e) => setCustomAmount(e.target.value)} />
              <div className="freq">
                <label><input type="radio" name="frequency" checked={frequency === "one-time"} onChange={() => setFrequency("one-time")} /> One-time</label>
                <label><input type="radio" name="frequency" checked={frequency === "monthly"} onChange={() => setFrequency("monthly")} /> Monthly</label>
              </div>
              <p className="pledge-summary">
                Your pledge: <strong>${effectiveAmount || 0} {frequency}</strong>
                {frequency === "monthly" ? " — that's $" + (effectiveAmount || 0) * 12 + " over a year." : ""}
              </p>
              <label htmlFor="donor-name">Your name</label>
              <input id="donor-name" type="text" required value={donorName} onChange={(e) => setDonorName(e.target.value)} placeholder="Your name" />
              <label htmlFor="donor-email">Your email</label>
              <input id="donor-email" type="email" required value={donorEmail} onChange={(e) => setDonorEmail(e.target.value)} placeholder="you@example.org" />
              <p style={{ marginTop: 16 }}>
                <button className="btn" type="submit" disabled={pledgeBusy}>{pledgeBusy ? "Recording…" : "Record my pledge"}</button>
              </p>
              {pledgeStatus && <p className={`notice${pledgeStatus.ok ? "" : " error"}`}>{pledgeStatus.text}</p>}
              <p className="source">
                Note: this records a pledge of intent. Real payment processing
                (for example Stripe) is not wired up yet — see the README.
              </p>
            </form>
          </div>
        </div>
      </section>
    </main>
  );
}

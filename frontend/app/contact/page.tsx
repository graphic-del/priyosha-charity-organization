"use client";

import { useState } from "react";
import { API_URL } from "../data";

export default function ContactPage() {
  const [contactName, setContactName] = useState("");
  const [contactEmail, setContactEmail] = useState("");
  const [interest, setInterest] = useState("donate");
  const [contactMessage, setContactMessage] = useState("");
  const [contactStatus, setContactStatus] = useState<{ ok: boolean; text: string } | null>(null);
  const [contactBusy, setContactBusy] = useState(false);

  async function submitContact(e: React.FormEvent) {
    e.preventDefault();
    setContactStatus(null);
    setContactBusy(true);
    try {
      const res = await fetch(`${API_URL}/api/messages`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name: contactName, email: contactEmail, interest, message: contactMessage }),
      });
      if (!res.ok) throw new Error(await res.text());
      setContactStatus({ ok: true, text: "Thank you — your message has been recorded. We will get back to you." });
      setContactMessage("");
    } catch {
      setContactStatus({
        ok: false,
        text: `Sorry — we could not send your message. Please check that the backend is running (${API_URL}) and try again.`,
      });
    } finally {
      setContactBusy(false);
    }
  }

  return (
    <main className="page-anim">
      <section className="band-maroon page-head">
        <div className="wrap">
          <p className="eyebrow">Contact</p>
          <h1>Contact &amp; volunteer</h1>
          <p className="lede">
            Want to help, volunteer, or ask a question? Send us a message — it is
            recorded by our contact service and we will get back to you.
          </p>
        </div>
      </section>

      <section className="band-white">
        <div className="wrap reveal" style={{ maxWidth: 760 }}>
          <div className="form-box" style={{ marginTop: 0 }}>
            <form onSubmit={submitContact}>
              <label htmlFor="contact-name">Your name</label>
              <input id="contact-name" type="text" required value={contactName} onChange={(e) => setContactName(e.target.value)} placeholder="Your name" />
              <label htmlFor="contact-email">Your email</label>
              <input id="contact-email" type="email" required value={contactEmail} onChange={(e) => setContactEmail(e.target.value)} placeholder="you@example.org" />
              <label htmlFor="interest">I am interested in</label>
              <select id="interest" value={interest} onChange={(e) => setInterest(e.target.value)}>
                <option value="donate">Donating</option>
                <option value="volunteer">Volunteering</option>
                <option value="other">Something else</option>
              </select>
              <label htmlFor="contact-message">Your message</label>
              <textarea id="contact-message" required value={contactMessage} onChange={(e) => setContactMessage(e.target.value)} placeholder="How would you like to help?" />
              <p style={{ marginTop: 16 }}>
                <button className="btn" type="submit" disabled={contactBusy}>{contactBusy ? "Sending…" : "Send message"}</button>
              </p>
              {contactStatus && <p className={`notice${contactStatus.ok ? "" : " error"}`}>{contactStatus.text}</p>}
            </form>
            <p className="source">
              Contact details: email and phone are not published yet — replace the
              placeholder in the README with the organization&apos;s real contact
              details before launch. Placeholder: donate@priyosha-charity.example
            </p>
          </div>
        </div>
      </section>
    </main>
  );
}

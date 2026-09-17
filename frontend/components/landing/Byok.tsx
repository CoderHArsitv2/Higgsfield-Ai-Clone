"use client";

import { useEffect, useRef } from "react";
import { gsap, prefersReducedMotion } from "@/lib/gsap";

const POINTS = [
  {
    title: "Your key, your quota",
    body: "Paste a provider key and every model behind it unlocks for your account only. You pay the provider directly at their price.",
  },
  {
    title: "Encrypted at rest",
    body: "Keys are AES-256-GCM encrypted before they touch the database and are never sent back to the browser — not even to you.",
  },
  {
    title: "No credits spent",
    body: "Jobs run on your own key skip our credit system entirely. Remove the key and those models simply lock again.",
  },
];

/** A deliberate breather: no motion tricks, just space. */
export function Byok() {
  const root = useRef<HTMLElement>(null);

  useEffect(() => {
    if (prefersReducedMotion()) {
      gsap.set(".reveal-target", { opacity: 1 });
      return;
    }
    const ctx = gsap.context(() => {
      gsap.set(".reveal-target", { opacity: 1 });
      gsap.from(".byok-item", {
        opacity: 0,
        y: 26,
        duration: 0.9,
        stagger: 0.1,
        ease: "expo.out",
        scrollTrigger: { trigger: root.current, start: "top 75%" },
      });
    }, root);
    return () => ctx.revert();
  }, []);

  return (
    <section id="keys" ref={root} className="border-t border-line px-6 py-28">
      <div className="mx-auto grid max-w-7xl gap-16 lg:grid-cols-[minmax(0,0.9fr)_minmax(0,1.1fr)]">
        <div className="reveal-target">
          <p className="mb-5 font-mono text-xs uppercase tracking-[0.2em] text-dim">
            Bring your own keys
          </p>
          <h2 className="max-w-[13ch] font-display text-[clamp(2.25rem,5vw,4rem)] leading-[0.98] tracking-tight">
            Or skip our billing entirely
          </h2>
        </div>

        <dl className="reveal-target divide-y divide-line border-t border-line">
          {POINTS.map((p) => (
            <div key={p.title} className="byok-item py-7">
              <dt className="text-[15px] font-medium tracking-tight">{p.title}</dt>
              <dd className="mt-2 max-w-lg text-sm leading-relaxed text-muted">
                {p.body}
              </dd>
            </div>
          ))}
        </dl>
      </div>
    </section>
  );
}

"use client";

import { useEffect, useRef } from "react";
import { gsap, prefersReducedMotion } from "@/lib/gsap";

const SHOTS = [
  { title: "Rain alley", model: "Kling 3.0", hue: 8, ratio: "aspect-[3/4]" },
  { title: "Coastal drone", model: "Veo 3.1", hue: 190, ratio: "aspect-[16/9]" },
  { title: "Studio portrait", model: "Nano Banana Pro", hue: 32, ratio: "aspect-[3/4]" },
  { title: "Product macro", model: "FLUX 1.1 Ultra", hue: 268, ratio: "aspect-[1/1]" },
  { title: "Night market", model: "Seedance 2.0", hue: 340, ratio: "aspect-[16/9]" },
  { title: "Desert crane", model: "Sora 2", hue: 42, ratio: "aspect-[3/4]" },
];

/**
 * Vertical scroll drives horizontal movement. It is the one place on the page
 * that breaks the reading axis, which is why it gets a whole section to itself
 * instead of competing with anything.
 */
export function Showcase() {
  const root = useRef<HTMLElement>(null);
  const track = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (prefersReducedMotion()) return;

    const ctx = gsap.context(() => {
      const el = track.current;
      if (!el) return;

      const distance = () => el.scrollWidth - window.innerWidth + 96;

      gsap.to(el, {
        x: () => -distance(),
        ease: "none",
        scrollTrigger: {
          trigger: root.current,
          start: "top top",
          end: () => `+=${distance()}`,
          pin: true,
          scrub: 0.7,
          invalidateOnRefresh: true,
          anticipatePin: 1,
        },
      });
    }, root);

    return () => ctx.revert();
  }, []);

  return (
    <section
      ref={root}
      className="relative overflow-hidden border-t border-line py-24"
    >
      <div className="mx-auto mb-12 max-w-7xl px-6">
        <p className="mb-5 font-mono text-xs uppercase tracking-[0.2em] text-dim">
          Made with Aperture
        </p>
        <h2 className="max-w-[18ch] font-display text-[clamp(2.25rem,5vw,4rem)] leading-[0.98] tracking-tight">
          One gallery, whichever model made it
        </h2>
      </div>

      <div ref={track} className="flex gap-5 px-6 will-change-transform">
        {SHOTS.map((s) => (
          <figure key={s.title} className="shrink-0">
            <div
              className={`${s.ratio} w-[clamp(240px,34vw,420px)] overflow-hidden rounded-2xl border border-line`}
              style={{
                background: `linear-gradient(150deg, hsl(${s.hue} 75% 52% / 0.35), hsl(${s.hue + 45} 70% 45% / 0.12) 60%, #101014 100%)`,
              }}
            />
            <figcaption className="mt-3 flex items-center justify-between font-mono text-[11px] uppercase tracking-wider">
              <span className="text-muted">{s.title}</span>
              <span className="text-dim">{s.model}</span>
            </figcaption>
          </figure>
        ))}
      </div>
    </section>
  );
}

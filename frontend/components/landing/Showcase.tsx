"use client";

import { useEffect, useRef, useState } from "react";
import { gsap, prefersReducedMotion } from "@/lib/gsap";

const SHOTS = [
  {
    clip: "rain-city",
    title: "Rain alley",
    model: "Kling 3.0",
    ratio: "aspect-[9/16]",
  },
  {
    clip: "coastline",
    title: "Coastal drone",
    model: "Veo 3.1",
    ratio: "aspect-video",
  },
  {
    clip: "portrait",
    title: "Studio portrait",
    model: "Nano Banana Pro",
    ratio: "aspect-[9/16]",
  },
  {
    clip: "studio",
    title: "Product macro",
    model: "FLUX 1.1 Ultra",
    ratio: "aspect-video",
  },
  {
    clip: "neon-street",
    title: "Night market",
    model: "Seedance 2.0",
    ratio: "aspect-[9/16]",
  },
  {
    clip: "desert",
    title: "Desert crane",
    model: "Sora 2",
    ratio: "aspect-video",
  },
];

/**
 * Vertical scroll drives horizontal movement. It is the one place on the page
 * that breaks the reading axis, which is why it gets a whole section to itself
 * instead of competing with anything.
 */
export function Showcase() {
  const root = useRef<HTMLElement>(null);
  const track = useRef<HTMLDivElement>(null);
  // Scroll drives the track horizontally. With motion reduced that never runs,
  // so the row has to be scrollable by hand or everything past the fold is
  // simply unreachable.
  const [reduced, setReduced] = useState(false);

  useEffect(() => {
    if (prefersReducedMotion()) {
      setReduced(true);
      return;
    }

    const ctx = gsap.context(() => {
      const el = track.current;
      if (!el) return;

      const distance = () =>
        Math.max(0, el.scrollWidth - window.innerWidth + 96);

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
      className={`relative border-t border-line py-24 ${
        reduced ? "" : "overflow-hidden"
      }`}
    >
      <div className="mx-auto mb-12 max-w-7xl px-6">
        <p className="mb-5 font-mono text-xs uppercase tracking-[0.2em] text-dim">
          Made with Aperture
        </p>
        <h2 className="max-w-[18ch] font-display text-[clamp(2.25rem,5vw,4rem)] leading-[0.98] tracking-tight">
          One gallery, whichever model made it
        </h2>
      </div>

      <div
        ref={track}
        className={`flex gap-5 px-6 ${
          reduced
            ? "snap-x snap-mandatory overflow-x-auto pb-4"
            : "will-change-transform"
        }`}
      >
        {SHOTS.map((s) => (
          <figure key={s.clip} className="shrink-0 snap-start">
            <div
              className={`${s.ratio} w-[clamp(240px,30vw,380px)] overflow-hidden rounded-2xl border border-line bg-panel`}
            >
              <video
                src={`/showcase/${s.clip}.mp4`}
                poster={`/showcase/${s.clip}.jpg`}
                className="h-full w-full object-cover"
                autoPlay
                muted
                loop
                playsInline
                preload="none"
              />
            </div>
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

"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { gsap, prefersReducedMotion } from "@/lib/gsap";
import { LazyVideo } from "@/components/ui/LazyVideo";

const CLIPS = [
  { clip: "rain-city", label: "Rain alley", model: "Kling 3.0" },
  { clip: "coastline", label: "Coastal drone", model: "Veo 3.1" },
  { clip: "portrait", label: "Studio portrait", model: "Nano Banana Pro" },
  { clip: "desert", label: "Desert crane", model: "Sora 2" },
  { clip: "neon-street", label: "Night market", model: "Seedance 2.0" },
];

/** Long enough to read the caption and see the clip move, short enough to hold. */
const SLIDE_MS = 3600;

/**
 * The right half of the hero: five clips on a carousel, each labelled with the
 * model that would have made it.
 *
 * Only the active clip plays. Five simultaneous decodes above the fold is the
 * kind of thing that makes a page feel heavy before the viewer has scrolled at
 * all, and four of them would be invisible anyway.
 */
export function HeroMedia() {
  const root = useRef<HTMLDivElement>(null);
  const [index, setIndex] = useState(0);
  const [paused, setPaused] = useState(false);
  const [reduced, setReduced] = useState(false);

  const go = useCallback((next: number) => {
    setIndex(((next % CLIPS.length) + CLIPS.length) % CLIPS.length);
  }, []);

  useEffect(() => {
    if (prefersReducedMotion()) {
      setReduced(true);
      gsap.set(".hero-slide-0", { opacity: 1 });
      gsap.set(".hero-frame", { opacity: 1, y: 0, scale: 1 });
      return;
    }

    const ctx = gsap.context(() => {
      gsap.fromTo(
        ".hero-frame",
        { opacity: 0, y: 40, scale: 0.94 },
        {
          opacity: 1,
          y: 0,
          scale: 1,
          duration: 1.3,
          ease: "expo.out",
          delay: 0.35,
          stagger: 0.1,
        },
      );

      // The stack breathes slightly so the composition is never quite static.
      gsap.to(".hero-frame-main", {
        yPercent: -2.2,
        duration: 6,
        ease: "sine.inOut",
        repeat: -1,
        yoyo: true,
        delay: 1.6,
      });

      gsap.to(root.current, {
        yPercent: -8,
        ease: "none",
        scrollTrigger: {
          trigger: root.current,
          start: "top 80%",
          end: "bottom top",
          scrub: 0.8,
        },
      });
    }, root);

    return () => ctx.revert();
  }, []);

  // Auto-advance. Paused on hover and while the OS asks for reduced motion,
  // so the carousel never moves under someone who is reading it.
  useEffect(() => {
    if (reduced || paused) return;
    const t = window.setTimeout(() => go(index + 1), SLIDE_MS);
    return () => window.clearTimeout(t);
  }, [index, paused, reduced, go]);

  // Cross-fade to the new slide and refill the progress rail.
  useEffect(() => {
    if (reduced) return;
    const ctx = gsap.context(() => {
      CLIPS.forEach((_, i) => {
        gsap.to(`.hero-slide-${i}`, {
          opacity: i === index ? 1 : 0,
          scale: i === index ? 1 : 1.04,
          duration: 0.8,
          ease: "power2.out",
          overwrite: "auto",
        });
      });

      gsap.fromTo(
        ".hero-caption",
        { opacity: 0, y: 10 },
        {
          opacity: 1,
          y: 0,
          duration: 0.5,
          ease: "power2.out",
          overwrite: "auto",
        },
      );

      gsap.set(".hero-tick-fill", { scaleX: 0 });
      if (!paused) {
        gsap.to(`.hero-tick-fill-${index}`, {
          scaleX: 1,
          duration: SLIDE_MS / 1000,
          ease: "none",
          overwrite: "auto",
        });
      }
    }, root);
    return () => ctx.revert();
  }, [index, paused, reduced]);

  const current = CLIPS[index];

  return (
    <div
      ref={root}
      className="relative mx-auto w-full max-w-[420px] lg:ml-auto lg:mr-0 lg:max-w-[460px]"
      onMouseEnter={() => setPaused(true)}
      onMouseLeave={() => setPaused(false)}
    >
      {/* Two offset frames behind the stack, for depth. Purely decorative. */}
      <div
        aria-hidden
        className="hero-frame absolute inset-0 -translate-x-5 translate-y-5 rotate-[-4deg] rounded-2xl border border-line bg-panel/60"
      />
      <div
        aria-hidden
        className="hero-frame absolute inset-0 translate-x-4 translate-y-2 rotate-[3deg] rounded-2xl border border-line bg-panel/40"
      />

      <figure className="hero-frame hero-frame-main relative overflow-hidden rounded-2xl border border-line bg-panel shadow-2xl shadow-black/60">
        <div className="relative aspect-[4/5] w-full">
          {CLIPS.map((c, i) => (
            <div
              key={c.clip}
              className={`hero-slide-${i} absolute inset-0`}
              style={{ opacity: i === 0 ? 1 : 0 }}
            >
              <LazyVideo
                src={`/showcase/${c.clip}.mp4`}
                poster={`/showcase/${c.clip}.jpg`}
                className="h-full w-full object-cover"
                eager={i === 0}
                active={i === index}
              />
            </div>
          ))}

          <div className="pointer-events-none absolute inset-x-0 bottom-0 h-2/5 bg-gradient-to-t from-void via-void/60 to-transparent" />

          <figcaption className="hero-caption absolute inset-x-0 bottom-0 flex items-end justify-between gap-3 p-4">
            <span className="font-mono text-[10px] uppercase tracking-[0.14em] text-fg/85">
              {current.label}
            </span>
            <span className="rounded-full border border-line bg-void/70 px-2.5 py-1 font-mono text-[10px] text-accent">
              {current.model}
            </span>
          </figcaption>
        </div>

        {/* One tick per clip; the active one fills over the slide duration. */}
        <div className="flex gap-1.5 px-4 pb-4 pt-3">
          {CLIPS.map((c, i) => (
            <button
              key={c.clip}
              onClick={() => go(i)}
              aria-label={`Show ${c.label}`}
              aria-current={i === index}
              className="group h-4 flex-1"
            >
              <span className="block h-0.5 w-full overflow-hidden rounded-full bg-line">
                <span
                  className={`hero-tick-fill hero-tick-fill-${i} block h-full w-full origin-left rounded-full bg-accent`}
                  style={{ transform: i === 0 ? undefined : "scaleX(0)" }}
                />
              </span>
            </button>
          ))}
        </div>
      </figure>
    </div>
  );
}

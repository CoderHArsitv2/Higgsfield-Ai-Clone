"use client";

import { useEffect, useRef } from "react";
import { gsap, prefersReducedMotion } from "@/lib/gsap";

const CARDS = [
  {
    clip: "rain-city",
    model: "Kling 3.0",
    label: "Rain alley",
    className: "left-0 top-[10%] z-20 aspect-[9/16] w-[46%] rotate-[-4deg]",
    float: { y: 16, dur: 5.5 },
  },
  {
    clip: "coastline",
    model: "Veo 3.1",
    label: "Coastal drone",
    className: "right-0 top-0 z-10 aspect-video w-[48%] rotate-[3deg]",
    float: { y: -14, dur: 6.5 },
  },
  {
    clip: "portrait",
    model: "Nano Banana Pro",
    label: "Studio portrait",
    className:
      "bottom-[2%] right-[4%] z-30 aspect-[9/16] w-[38%] rotate-[-2deg]",
    float: { y: 12, dur: 7 },
  },
];

/**
 * The right half of the hero. Real generated-looking footage rather than an
 * empty column: three cards at slight angles, each labelled with the model that
 * would have made it, drifting independently so the composition never sits
 * still.
 */
export function HeroMedia() {
  const root = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (prefersReducedMotion()) {
      gsap.set(".hero-card", { opacity: 1, y: 0, scale: 1 });
      return;
    }

    const ctx = gsap.context(() => {
      gsap.set(".hero-card", { opacity: 0, y: 40, scale: 0.94 });

      gsap.to(".hero-card", {
        opacity: 1,
        y: 0,
        scale: 1,
        duration: 1.3,
        ease: "expo.out",
        stagger: 0.12,
        delay: 0.35,
      });

      // Each card drifts on its own cycle, so they never line up and the group
      // reads as alive rather than as one block moving.
      CARDS.forEach((c, i) => {
        gsap.to(`.hero-card-${i}`, {
          yPercent: c.float.y > 0 ? 3 : -3,
          duration: c.float.dur,
          ease: "sine.inOut",
          repeat: -1,
          yoyo: true,
          delay: 1.4 + i * 0.3,
        });
      });

      // Parallax: the stack drifts up a little slower than the page.
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

  return (
    <div
      ref={root}
      className="relative mx-auto aspect-square w-full max-w-[460px] lg:ml-auto lg:mr-0 lg:max-w-[560px]"
    >
      {CARDS.map((c, i) => (
        <figure
          key={c.clip}
          className={`hero-card hero-card-${i} absolute overflow-hidden rounded-2xl border border-line bg-panel shadow-2xl shadow-black/60 ${c.className}`}
        >
          <video
            src={`/showcase/${c.clip}.mp4`}
            poster={`/showcase/${c.clip}.jpg`}
            className="h-full w-full object-cover"
            autoPlay
            muted
            loop
            playsInline
            preload="metadata"
          />
          <figcaption className="absolute inset-x-0 bottom-0 flex items-center justify-between bg-gradient-to-t from-void/95 to-transparent px-3 pb-2.5 pt-8">
            <span className="font-mono text-[9px] uppercase tracking-wider text-fg/80">
              {c.label}
            </span>
            <span className="rounded-full bg-void/70 px-2 py-0.5 font-mono text-[9px] text-accent">
              {c.model}
            </span>
          </figcaption>
        </figure>
      ))}
    </div>
  );
}

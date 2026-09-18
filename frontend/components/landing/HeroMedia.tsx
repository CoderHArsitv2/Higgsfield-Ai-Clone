"use client";

import { useEffect, useRef, useState } from "react";
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
 * Depth 0 is the card on top of the deck; each step back is smaller, further
 * right and rotated a little more. Anything past the third card sits at BACK,
 * which is that same pose at zero opacity -- so the card leaving the front
 * animates *into* the stack and fades there, rather than flying across the
 * hero. That one detail is what makes the move read as a shuffle.
 */
const DEPTH = [
  { x: 0, y: 0, scale: 1, rotate: 0, opacity: 1 },
  { x: 30, y: -16, scale: 0.955, rotate: 2.6, opacity: 1 },
  { x: 54, y: -29, scale: 0.915, rotate: 5, opacity: 1 },
];
const BACK = { x: 72, y: -39, scale: 0.885, rotate: 6.8, opacity: 0 };

const poseOf = (depth: number) => DEPTH[depth] ?? BACK;

/**
 * The right half of the hero: a deck of clips that shuffles itself, each card
 * labelled with the model that would have made it.
 *
 * Only the front clip plays. Five simultaneous decodes above the fold is the
 * kind of thing that makes a page feel heavy before the viewer has scrolled at
 * all, and four of them are behind another card anyway.
 */
export function HeroMedia() {
  const root = useRef<HTMLDivElement>(null);
  const [front, setFront] = useState(0);
  const [reduced, setReduced] = useState(false);
  const n = CLIPS.length;

  useEffect(() => {
    if (prefersReducedMotion()) {
      setReduced(true);
      gsap.set(".hero-deck", { opacity: 1, y: 0, scale: 1 });
      return;
    }

    // Entrance and drift live on a wrapper, never on the cards. The cards carry
    // their own inline transform for the deck pose, and GSAP writing transform
    // to the same element would fight it every time the deck advances.
    const ctx = gsap.context(() => {
      gsap.fromTo(
        ".hero-deck",
        { opacity: 0, y: 40, scale: 0.94 },
        {
          opacity: 1,
          y: 0,
          scale: 1,
          duration: 1.3,
          ease: "expo.out",
          delay: 0.35,
        },
      );

      gsap.to(".hero-deck", {
        yPercent: -2.2,
        duration: 6,
        ease: "sine.inOut",
        repeat: -1,
        yoyo: true,
        delay: 1.8,
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

  // Shuffles on its own and only on its own -- there is nothing to click. A
  // hero is looked at, not operated.
  useEffect(() => {
    if (reduced) return;
    const t = window.setInterval(() => setFront((f) => (f + 1) % n), SLIDE_MS);
    return () => window.clearInterval(t);
  }, [reduced, n]);

  const current = CLIPS[front];

  return (
    <div
      ref={root}
      className="relative mx-auto w-full max-w-[420px] lg:ml-auto lg:mr-0 lg:max-w-[460px]"
    >
      <div className="hero-deck relative pr-14 pt-10" style={{ opacity: 0 }}>
        <Backdrop clip={current.clip} />

        <div className="relative aspect-[4/5] w-full">
          {CLIPS.map((c, i) => {
            const depth = (i - front + n) % n;
            const pose = reduced ? poseOf(Math.min(i, 3)) : poseOf(depth);
            const isFront = reduced ? i === 0 : depth === 0;

            return (
              <figure
                key={c.clip}
                aria-hidden={!isFront}
                style={{
                  transform: `translate3d(${pose.x}px, ${pose.y}px, 0) scale(${pose.scale}) rotate(${pose.rotate}deg)`,
                  opacity: pose.opacity,
                  zIndex: n - depth,
                }}
                className={`absolute inset-0 overflow-hidden rounded-2xl border border-line bg-panel transition-[transform,opacity] duration-[900ms] ease-[cubic-bezier(0.16,1,0.3,1)] ${
                  isFront ? "shadow-2xl shadow-black/60" : ""
                }`}
              >
                <LazyVideo
                  src={`/showcase/${c.clip}.mp4`}
                  poster={`/showcase/${c.clip}.jpg`}
                  className="h-full w-full object-cover"
                  eager={i === 0}
                  active={isFront}
                />

                {/* Cards behind are dimmed with an overlay rather than lower
                    opacity, so the deck stays solid instead of letting the
                    backdrop show through the whole stack. */}
                {!isFront && <span className="absolute inset-0 bg-void/55" />}

                {isFront && (
                  <>
                    <div className="pointer-events-none absolute inset-x-0 bottom-0 h-2/5 bg-gradient-to-t from-void via-void/60 to-transparent" />
                    <figcaption className="absolute inset-x-0 bottom-0 flex items-end justify-between gap-3 p-4">
                      <span className="font-mono text-[10px] uppercase tracking-[0.14em] text-fg/85">
                        {c.label}
                      </span>
                      <span className="rounded-full border border-line bg-void/70 px-2.5 py-1 font-mono text-[10px] text-accent">
                        {c.model}
                      </span>
                    </figcaption>
                  </>
                )}
              </figure>
            );
          })}
        </div>
      </div>
    </div>
  );
}

/**
 * A blurred blow-up of the front clip's own poster, sitting behind the deck and
 * cross-fading as it turns, so the corner of the hero takes its colour from
 * whatever card is on top. Kept local to the deck and well under the page-wide
 * hero glow, which is already doing the heavy lifting on this screen.
 */
function Backdrop({ clip }: { clip: string }) {
  return (
    <div aria-hidden className="pointer-events-none absolute -inset-10 -z-10">
      {CLIPS.map((c) => (
        <div
          key={c.clip}
          className={`deck-glow absolute inset-0 rounded-[3rem] bg-cover bg-center transition-opacity duration-[1400ms] ease-out ${
            c.clip === clip ? "opacity-30" : "opacity-0"
          }`}
          style={{
            backgroundImage: `url(/showcase/${c.clip}.jpg)`,
            filter: "blur(56px) saturate(1.6)",
          }}
        />
      ))}
    </div>
  );
}

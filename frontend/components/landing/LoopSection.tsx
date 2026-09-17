"use client";

import { useEffect, useRef, useState } from "react";
import { gsap, ScrollTrigger, prefersReducedMotion } from "@/lib/gsap";

const BEATS = [
  {
    n: "01",
    title: "Describe it",
    body: "One prompt bar for every modality. No mode switching, no separate tools for video and stills.",
  },
  {
    n: "02",
    title: "Pick a model",
    body: "Kling, Veo, Sora, Seedance, FLUX, Nano Banana. Each one exposes its own controls, so nothing is flattened to a lowest common denominator.",
  },
  {
    n: "03",
    title: "Get it back",
    body: "Jobs queue server-side and survive a refresh. Results land in one gallery regardless of which provider made them.",
  },
];

const DIM = 0.32;

/**
 * The centrepiece scroll moment: the section pins and the three beats advance
 * with scroll position, so the viewer controls the pace.
 *
 * Driven by an explicit active index rather than a scrubbed `fromTo` timeline.
 * A scrub leaves every element in its "from" state whenever progress is 0 --
 * which is true before the trigger starts and after scrolling back past it --
 * so all three beats sat dimmed at once and the wrong panel showed through.
 * An index always has exactly one valid resting state, at any scroll position
 * and on a deep link straight into the section.
 */
export function LoopSection() {
  const root = useRef<HTMLElement>(null);
  const [reduced, setReduced] = useState(false);

  useEffect(() => {
    if (prefersReducedMotion()) {
      setReduced(true);
      gsap.set(".reveal-target", { opacity: 1 });
      return;
    }

    const ctx = gsap.context(() => {
      gsap.set(".reveal-target", { opacity: 1 });

      // Resting state: beat one is already active before any scrolling.
      gsap.set(".beat-0", { opacity: 1 });
      gsap.set([".beat-1", ".beat-2"], { opacity: DIM });
      gsap.set(".panel-0", { opacity: 1, scale: 1 });
      gsap.set([".panel-1", ".panel-2"], { opacity: 0, scale: 0.97 });
      gsap.set(".loop-progress", { scaleX: 1 / BEATS.length });

      let active = 0;
      const show = (i: number) => {
        if (i === active) return;
        active = i;
        BEATS.forEach((_, j) => {
          gsap.to(`.beat-${j}`, {
            opacity: j === i ? 1 : DIM,
            duration: 0.45,
            ease: "power2.out",
            overwrite: "auto",
          });
          gsap.to(`.panel-${j}`, {
            opacity: j === i ? 1 : 0,
            scale: j === i ? 1 : 0.97,
            duration: 0.5,
            ease: "power2.out",
            overwrite: "auto",
          });
        });
        gsap.to(".loop-progress", {
          scaleX: (i + 1) / BEATS.length,
          duration: 0.4,
          ease: "power2.out",
          overwrite: "auto",
        });
      };

      ScrollTrigger.create({
        trigger: root.current,
        start: "top top",
        end: "+=2100",
        pin: true,
        anticipatePin: 1,
        invalidateOnRefresh: true,
        onUpdate: (self) => {
          const i = Math.min(
            BEATS.length - 1,
            Math.floor(self.progress * BEATS.length),
          );
          show(i);
        },
        // Scrolling back out of the section must restore the resting state,
        // not leave whichever beat happened to be last.
        onLeaveBack: () => show(0),
      });
    }, root);

    return () => ctx.revert();
  }, []);

  return (
    <section
      id="loop"
      ref={root}
      className="relative flex min-h-[100svh] items-center border-t border-line px-6 py-24"
    >
      <div className="mx-auto grid w-full max-w-6xl gap-16 lg:grid-cols-[minmax(0,0.85fr)_minmax(0,1fr)] lg:items-center">
        <div>
          <p className="mb-10 font-mono text-xs uppercase tracking-[0.2em] text-dim">
            The whole product
          </p>

          <div className="space-y-9">
            {BEATS.map((b, i) => (
              <div key={b.n} className={`beat-${i} reveal-target`}>
                <div className="flex items-baseline gap-4">
                  <span className="font-mono text-xs text-accent">{b.n}</span>
                  <h3 className="font-display text-3xl tracking-tight md:text-4xl">
                    {b.title}
                  </h3>
                </div>
                <p className="mt-3 max-w-md pl-10 text-sm leading-relaxed text-muted">
                  {b.body}
                </p>
              </div>
            ))}
          </div>

          <div className="mt-12 h-px w-full max-w-md bg-line">
            <div className="loop-progress h-px origin-left bg-accent" />
          </div>
        </div>

        <div
          className={
            reduced ? "w-full space-y-4" : "relative aspect-[4/3] w-full"
          }
        >
          <LoopPanel index={0} reduced={reduced}>
            <div className="flex h-full flex-col justify-end gap-3 p-8">
              <div className="rounded-xl border border-line bg-void/60 p-4">
                <p className="font-mono text-[11px] text-dim">prompt</p>
                <p className="mt-2 text-sm leading-relaxed text-fg">
                  Slow dolly through a rain-soaked Tokyo alley at night, neon
                  reflections, shallow depth of field
                </p>
              </div>
            </div>
          </LoopPanel>

          <LoopPanel index={1} reduced={reduced}>
            <div className="grid h-full grid-cols-2 content-center gap-2.5 p-8">
              {["Kling 3.0", "Veo 3.1", "Sora 2", "Seedance 2.0"].map(
                (m, i) => (
                  <div
                    key={m}
                    className={`rounded-xl border p-4 text-sm ${
                      i === 0
                        ? "border-accent/60 bg-accent/10 text-fg"
                        : "border-line bg-void/40 text-dim"
                    }`}
                  >
                    <div className="flex items-center justify-between">
                      {m}
                      {i === 0 && (
                        <span className="h-1.5 w-1.5 rounded-full bg-accent" />
                      )}
                    </div>
                  </div>
                ),
              )}
            </div>
          </LoopPanel>

          <LoopPanel index={2} reduced={reduced}>
            <div className="grid h-full grid-cols-2 gap-2 p-8">
              {["rain-city", "neon-street", "coastline", "desert"].map(
                (clip) => (
                  <video
                    key={clip}
                    src={`/showcase/${clip}.mp4`}
                    className="h-full w-full rounded-xl border border-line object-cover"
                    muted
                    loop
                    playsInline
                    autoPlay
                    preload="none"
                  />
                ),
              )}
            </div>
          </LoopPanel>
        </div>
      </div>
    </section>
  );
}

function LoopPanel({
  index,
  reduced,
  children,
}: {
  index: number;
  reduced: boolean;
  children: React.ReactNode;
}) {
  return (
    <div
      className={`panel-${index} reveal-target overflow-hidden rounded-2xl border border-line bg-panel ${
        reduced ? "aspect-[4/3]" : "absolute inset-0"
      }`}
    >
      {children}
    </div>
  );
}

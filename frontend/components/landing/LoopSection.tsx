"use client";

import { useEffect, useRef, useState } from "react";
import { gsap, prefersReducedMotion } from "@/lib/gsap";

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

/**
 * The centrepiece scroll moment: the section pins and the three beats are
 * scrubbed by scroll position rather than autoplaying, so the viewer controls
 * the pace. This replaces five competing announcement blocks on the original
 * site with one idea told in sequence.
 */
export function LoopSection() {
  const root = useRef<HTMLElement>(null);
  // The three panels are stacked absolutely and revealed by the scrub. With
  // motion reduced that scrub never runs, so they would sit on top of each
  // other and only the last one would be visible. In that mode they lay out in
  // normal flow instead.
  const [reduced, setReduced] = useState(false);

  useEffect(() => {
    if (prefersReducedMotion()) {
      setReduced(true);
      gsap.set(".reveal-target", { opacity: 1 });
      return;
    }

    const ctx = gsap.context(() => {
      gsap.set(".reveal-target", { opacity: 1 });

      const tl = gsap.timeline({
        scrollTrigger: {
          trigger: root.current,
          start: "top top",
          end: "+=2600",
          pin: true,
          scrub: 0.8,
          anticipatePin: 1,
        },
      });

      BEATS.forEach((_, i) => {
        const at = i * 1;
        if (i > 0) {
          tl.to(
            `.beat-${i - 1}`,
            { opacity: 0.18, y: -14, duration: 0.4 },
            at,
          ).to(
            `.panel-${i - 1}`,
            { opacity: 0, scale: 0.97, duration: 0.4 },
            at,
          );
        }
        tl.fromTo(
          `.beat-${i}`,
          { opacity: 0.18, y: 14 },
          { opacity: 1, y: 0, duration: 0.4 },
          at,
        )
          .fromTo(
            `.panel-${i}`,
            { opacity: 0, scale: 0.97 },
            { opacity: 1, scale: 1, duration: 0.5 },
            at,
          )
          .to(
            ".loop-progress",
            { scaleX: (i + 1) / BEATS.length, duration: 0.4 },
            at,
          );
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
            <div className="loop-progress h-px origin-left scale-x-0 bg-accent" />
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
              {[0, 1, 2, 3].map((i) => (
                <div
                  key={i}
                  className="rounded-xl border border-line"
                  style={{
                    background: `linear-gradient(${135 + i * 40}deg, rgba(255,90,54,${0.28 - i * 0.05}), rgba(138,124,255,${0.2 + i * 0.04}))`,
                  }}
                />
              ))}
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

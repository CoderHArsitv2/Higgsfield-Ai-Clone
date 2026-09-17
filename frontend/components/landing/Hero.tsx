"use client";

import { useEffect, useRef } from "react";
import {
  gsap,
  SplitText,
  prefersReducedMotion,
  scrollToSection,
} from "@/lib/gsap";
import { Magnetic } from "@/components/ui/Magnetic";
import { HeroMedia } from "./HeroMedia";

export function Hero({ modelCount }: { modelCount: number }) {
  const root = useRef<HTMLElement>(null);
  const headline = useRef<HTMLHeadingElement>(null);

  useEffect(() => {
    const revealEverything = () => {
      gsap.set(".reveal-target", { opacity: 1, y: 0 });
      if (headline.current) gsap.set(headline.current, { opacity: 1 });
    };

    if (prefersReducedMotion()) {
      revealEverything();
      return;
    }

    let ctx: gsap.Context | undefined;
    let cancelled = false;

    const build = () => {
      ctx = gsap.context(() => {
        // Split into lines then characters, and mask each line, so characters
        // rise out from behind a clean edge rather than fading in place.
        const split = new SplitText(headline.current, {
          type: "lines,chars",
          linesClass: "overflow-hidden pb-[0.12em]",
        });

        const tl = gsap.timeline({ defaults: { ease: "expo.out" } });

        tl.set(".reveal-target", { opacity: 1 })
          .from(split.chars, {
            yPercent: 115,
            duration: 1.1,
            stagger: { each: 0.016, from: "start" },
          })
          .from(".hero-sub", { opacity: 0, y: 18, duration: 0.9 }, "-=0.72")
          .from(
            ".hero-cta",
            { opacity: 0, y: 16, duration: 0.8, stagger: 0.08 },
            "-=0.66",
          )
          .from(
            ".hero-meta",
            { opacity: 0, duration: 0.8, stagger: 0.06 },
            "-=0.6",
          );

        // Slow ambient drift on the backdrop so the page is never quite static.
        //
        // Translate only. Animating `scale` on an element carrying a 120px blur
        // forces the browser to re-rasterise a viewport-sized blurred layer on
        // every frame -- one of the most expensive things a page can do, and it
        // was running continuously underneath every scroll. A pure translate
        // stays on the compositor.
        gsap.to(".hero-glow", {
          xPercent: 10,
          yPercent: -6,
          duration: 20,
          ease: "sine.inOut",
          repeat: -1,
          yoyo: true,
        });

        // Headline drifts up as you scroll away; the glow lags behind it.
        gsap.to(".hero-parallax", {
          yPercent: -9,
          ease: "none",
          scrollTrigger: {
            trigger: root.current,
            start: "top top",
            end: "bottom top",
            scrub: 0.6,
          },
        });

        return () => split.revert();
      }, root);
    };

    // Splitting before the display webfont is applied measures the fallback
    // font's line boxes, so the headline breaks in the wrong places and then
    // reflows visibly once the real font arrives.
    // A font that never resolves must not leave the headline hidden.
    const fonts = Promise.race([
      document.fonts?.ready ?? Promise.resolve(),
      new Promise((r) => window.setTimeout(r, 600)),
    ]);
    fonts
      .catch(() => undefined)
      .then(() => {
        if (cancelled) return;
        try {
          build();
        } catch {
          revealEverything();
        }
      });

    return () => {
      cancelled = true;
      ctx?.revert();
    };
  }, []);

  return (
    <section
      ref={root}
      className="relative flex min-h-[100svh] items-center overflow-hidden px-6 pt-16"
    >
      <div
        aria-hidden
        className="hero-glow pointer-events-none absolute left-1/2 top-1/3 h-[70vh] w-[70vw] -translate-x-1/2 -translate-y-1/2 rounded-full opacity-45 blur-[120px]"
        style={{
          willChange: "transform",
          background:
            "radial-gradient(circle at 30% 30%, #ff5a36 0%, transparent 55%), radial-gradient(circle at 70% 60%, #8a7cff 0%, transparent 55%)",
        }}
      />

      <div className="relative mx-auto grid w-full max-w-7xl items-center gap-12 lg:grid-cols-[minmax(0,1.05fr)_minmax(0,0.95fr)] lg:gap-16">
        <div className="hero-parallax">
          <p className="hero-meta reveal-target mb-8 flex items-center gap-3 font-mono text-xs uppercase tracking-[0.2em] text-dim">
            <span className="inline-block h-px w-8 bg-dim" />
            {modelCount} models · one workspace
          </p>

          <h1
            ref={headline}
            className="reveal-target max-w-[16ch] font-display text-[clamp(3rem,10vw,8.5rem)] font-normal leading-[0.92] tracking-[-0.02em]"
          >
            Every model. <em className="italic text-accent">One</em> studio.
          </h1>

          <p className="hero-sub reveal-target mt-8 max-w-xl text-lg leading-relaxed text-muted">
            Video, stills and voice from a single prompt bar. Use our keys, or
            bring your own and pay the model providers directly.
          </p>

          <div className="mt-11 flex flex-wrap items-center gap-4">
            <Magnetic className="hero-cta reveal-target">
              <a
                href="/login"
                className="inline-flex items-center gap-2 rounded-full bg-accent px-7 py-3.5 text-sm font-medium text-void transition-colors hover:bg-accent-soft"
              >
                Start creating
                <svg
                  width="14"
                  height="14"
                  viewBox="0 0 14 14"
                  fill="none"
                  aria-hidden
                >
                  <path
                    d="M1 7h11M7.5 2.5 12 7l-4.5 4.5"
                    stroke="currentColor"
                    strokeWidth="1.5"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  />
                </svg>
              </a>
            </Magnetic>

            <a
              href="#loop"
              onClick={(e) => {
                e.preventDefault();
                scrollToSection("#loop");
              }}
              className="hero-cta reveal-target rounded-full border border-line px-7 py-3.5 text-sm text-muted transition-colors hover:border-dim hover:text-fg"
            >
              See how it works
            </a>
          </div>

          <p className="hero-meta reveal-target mt-14 font-mono text-xs text-dim">
            No card required · Sandbox models run free
          </p>
        </div>

        <HeroMedia />
      </div>
    </section>
  );
}

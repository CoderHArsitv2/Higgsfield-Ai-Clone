"use client";

import { Magnetic } from "@/components/ui/Magnetic";

const WORDS = ["Video", "Stills", "Voice", "Motion", "Edits", "Upscales"];

export function FooterCta() {
  return (
    <footer className="relative overflow-hidden border-t border-line">
      <div
        aria-hidden
        className="pointer-events-none absolute -bottom-40 left-1/2 h-[60vh] w-[90vw] -translate-x-1/2 rounded-full opacity-25 blur-[130px]"
        style={{
          background: "radial-gradient(circle, #ff5a36 0%, transparent 65%)",
        }}
      />

      <div className="relative flex select-none overflow-hidden py-10">
        <div className="marquee-track flex shrink-0 items-center gap-8 whitespace-nowrap pr-8">
          {[...WORDS, ...WORDS, ...WORDS, ...WORDS].map((w, i) => (
            <span
              key={i}
              className="font-display text-4xl italic text-dim md:text-6xl"
            >
              {w}
              <span className="not-italic text-accent"> · </span>
            </span>
          ))}
        </div>
      </div>

      <div className="relative mx-auto max-w-7xl px-6 pb-16 pt-10 text-center">
        <h2 className="mx-auto max-w-[16ch] font-display text-[clamp(2.5rem,7vw,5.5rem)] leading-[0.95] tracking-tight">
          Start with the sandbox. No key needed.
        </h2>

        <div className="mt-10 flex justify-center">
          <Magnetic>
            <a
              href="/auth/login"
              className="inline-flex items-center gap-2 rounded-full bg-accent px-8 py-4 text-sm font-medium text-void transition-colors hover:bg-accent-soft"
            >
              Sign in with Auth0
            </a>
          </Magnetic>
        </div>

        <div className="mt-20 flex flex-col items-center justify-between gap-4 border-t border-line pt-8 font-mono text-[11px] uppercase tracking-wider text-dim sm:flex-row">
          <span>Aperture — a Higgsfield study</span>
          <span>Built with Next.js and Go</span>
        </div>
      </div>
    </footer>
  );
}

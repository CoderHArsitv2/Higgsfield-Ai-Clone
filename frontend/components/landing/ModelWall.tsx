"use client";

import { useEffect, useRef, useState } from "react";
import { gsap, prefersReducedMotion } from "@/lib/gsap";
import type { PublicModel } from "@/lib/types";

const FILTERS = [
  { key: "all", label: "Everything" },
  { key: "video", label: "Video" },
  { key: "image", label: "Image" },
  { key: "audio", label: "Voice" },
] as const;

/**
 * The model wall replaces the five separate product-announcement sections the
 * original site uses. Locked models are shown rather than hidden: the breadth
 * is the pitch, and a lock is a prompt to add a key, not a dead end.
 */
export function ModelWall({ models }: { models: PublicModel[] }) {
  const root = useRef<HTMLElement>(null);
  const [filter, setFilter] = useState<string>("all");

  const shown =
    filter === "all" ? models : models.filter((m) => m.modality === filter);

  useEffect(() => {
    if (prefersReducedMotion()) {
      gsap.set(".reveal-target", { opacity: 1 });
      return;
    }
    const ctx = gsap.context(() => {
      gsap.set(".reveal-target", { opacity: 1 });
      gsap.from(".wall-head > *", {
        opacity: 0,
        y: 24,
        duration: 0.9,
        stagger: 0.08,
        ease: "expo.out",
        scrollTrigger: { trigger: ".wall-head", start: "top 82%" },
      });
    }, root);
    return () => ctx.revert();
  }, []);

  // Re-stagger on filter change so the grid reshuffles with intent rather than
  // snapping.
  useEffect(() => {
    if (prefersReducedMotion()) return;
    const cards = gsap.utils.toArray<HTMLElement>(".model-card");
    gsap.fromTo(
      cards,
      { opacity: 0, y: 18 },
      {
        opacity: 1,
        y: 0,
        duration: 0.55,
        ease: "expo.out",
        stagger: { each: 0.022, grid: "auto", from: "start" },
        overwrite: true,
      },
    );
  }, [filter]);

  const enabled = models.filter((m) => m.enabled).length;

  return (
    <section id="models" ref={root} className="border-t border-line px-6 py-28">
      <div className="mx-auto max-w-7xl">
        <div className="wall-head mb-14 flex flex-wrap items-end justify-between gap-8">
          <div className="reveal-target">
            <p className="mb-5 font-mono text-xs uppercase tracking-[0.2em] text-dim">
              The catalogue
            </p>
            <h2 className="max-w-[14ch] font-display text-[clamp(2.5rem,6vw,4.75rem)] leading-[0.95] tracking-tight">
              {models.length} models, nothing hidden
            </h2>
            <p className="mt-5 max-w-lg text-sm leading-relaxed text-muted">
              <span className="text-mint">{enabled} ready to run now.</span> The
              rest unlock the moment a key exists — ours or yours. We show you
              the whole shelf either way.
            </p>
          </div>

          <div className="reveal-target flex flex-wrap gap-2">
            {FILTERS.map((f) => (
              <button
                key={f.key}
                onClick={() => setFilter(f.key)}
                className={`rounded-full border px-4 py-1.5 text-sm transition-colors ${
                  filter === f.key
                    ? "border-fg bg-fg text-void"
                    : "border-line text-muted hover:border-dim hover:text-fg"
                }`}
              >
                {f.label}
              </button>
            ))}
          </div>
        </div>

        {/* Separate tiles rather than one bordered block. The catalogue length
            is not a multiple of the column count at any breakpoint, so a single
            block always ends on a partial row -- which reads as a broken edge
            however the separators are drawn. Tiles make a partial row look
            deliberate, and the count can change freely as filters are applied. */}
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {shown.map((m) => (
            <ModelCard key={m.id} model={m} />
          ))}
        </div>
      </div>
    </section>
  );
}

function ModelCard({ model }: { model: PublicModel }) {
  return (
    <article
      className={`model-card group relative flex min-h-[172px] flex-col justify-between overflow-hidden rounded-xl border border-line bg-void p-6 transition-colors hover:border-dim hover:bg-panel ${
        model.enabled ? "" : "opacity-60"
      }`}
    >
      <div>
        <div className="flex items-start justify-between gap-4">
          <h3 className="text-[15px] font-medium tracking-tight">
            {model.name}
          </h3>
          {model.enabled ? (
            <span
              className="mt-1 h-1.5 w-1.5 shrink-0 rounded-full bg-mint"
              title="Ready"
            />
          ) : (
            <svg
              className="mt-0.5 shrink-0 text-dim"
              width="13"
              height="13"
              viewBox="0 0 14 14"
              fill="none"
              aria-label="Locked"
            >
              <rect
                x="2.5"
                y="6"
                width="9"
                height="6.5"
                rx="1.5"
                stroke="currentColor"
                strokeWidth="1.2"
              />
              <path
                d="M4.75 6V4.25a2.25 2.25 0 0 1 4.5 0V6"
                stroke="currentColor"
                strokeWidth="1.2"
              />
            </svg>
          )}
        </div>
        <p className="mt-2.5 line-clamp-2 text-[13px] leading-relaxed text-muted">
          {model.description}
        </p>
      </div>

      <div className="mt-5 flex items-center justify-between font-mono text-[11px] uppercase tracking-wider">
        <span className="text-dim">{model.provider_name}</span>
        <span className={model.enabled ? "text-mint" : "text-dim"}>
          {model.enabled ? "Ready" : "Needs key"}
        </span>
      </div>

      {model.featured && (
        <span className="absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-accent to-transparent opacity-70" />
      )}
    </article>
  );
}

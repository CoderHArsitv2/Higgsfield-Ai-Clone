"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { api, ApiError } from "@/lib/api";
import type { Generation, Model, ProviderInfo } from "@/lib/types";
import { ModelPicker } from "./ModelPicker";
import { ParamControls } from "./ParamControls";
import { JobCard } from "./JobCard";
import {
  CardGridSkeleton,
  ModelPickerSkeleton,
} from "@/components/ui/Skeleton";

type Value = string | number | boolean;

const POLL_MS = 2500;

// The studio shows only the newest few: the composer is the point of this page
// and a full history would push it off the screen. Everything else is one click
// away at /generations, which pages and filters server-side.
const RECENT = 3;

const isActive = (j: Generation) =>
  j.status === "queued" || j.status === "running";

export function Studio({
  credits,
  onCreditsChange,
}: {
  credits: number;
  onCreditsChange: (n: number) => void;
}) {
  const [models, setModels] = useState<Model[]>([]);
  const [providers, setProviders] = useState<ProviderInfo[]>([]);
  const [selected, setSelected] = useState<Model>();
  const [params, setParams] = useState<Record<string, Value>>({});
  const [prompt, setPrompt] = useState("");
  const [jobs, setJobs] = useState<Generation[]>([]);
  const [error, setError] = useState<string>();
  const [busy, setBusy] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    (async () => {
      try {
        const [catalog, history] = await Promise.all([
          api.models(),
          api.generations({ limit: "24" }),
        ]);
        setModels(catalog.models);
        setProviders(catalog.providers ?? []);
        setJobs(history.generations);
        const first = catalog.models.find((m) => m.enabled);
        if (first) selectModel(first);
      } catch (e) {
        setError(
          e instanceof ApiError ? e.message : "Could not load the studio",
        );
      } finally {
        setLoading(false);
      }
    })();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const hasActive = jobs.some(isActive);

  // Poll only while something is actually in flight, so an idle tab is silent.
  useEffect(() => {
    if (!hasActive) return;
    const t = setInterval(async () => {
      try {
        const { generations } = await api.generations({ limit: "24" });
        setJobs(generations);
      } catch {
        // A dropped poll is not worth showing; the next tick retries.
      }
    }, POLL_MS);
    return () => clearInterval(t);
  }, [hasActive]);

  const selectModel = useCallback((m: Model) => {
    setSelected(m);
    const next: Record<string, Value> = {};
    for (const p of m.params ?? []) {
      if (p.default !== undefined) next[p.key] = p.default as Value;
    }
    setParams(next);
  }, []);

  const submit = async () => {
    if (!selected || !prompt.trim() || busy) return;
    setBusy(true);
    setError(undefined);
    try {
      const res = await api.create({
        model_id: selected.id,
        prompt: prompt.trim(),
        params,
      });
      setJobs((prev) => [res.generation, ...prev]);
      onCreditsChange(res.credits_remaining);
      setPrompt("");
    } catch (e) {
      setError(
        e instanceof ApiError ? e.message : "Could not start the generation",
      );
    } finally {
      setBusy(false);
    }
  };

  const cancel = async (id: string) => {
    try {
      const { generation } = await api.cancel(id);
      setJobs((prev) => prev.map((j) => (j.id === id ? generation : j)));
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Could not cancel");
    }
  };

  const lockedCount = useMemo(
    () => models.filter((m) => !m.enabled).length,
    [models],
  );

  const activeCount = useMemo(() => jobs.filter(isActive).length, [jobs]);
  const recent = useMemo(() => jobs.slice(0, RECENT), [jobs]);

  // A flex row, not a grid: in a two-row grid the sidebar set the height of the
  // first row, so a short composer left a column of dead space beside a tall
  // model list. Here the sidebar is its own sticky column and the right column
  // stacks composer over generations with nothing between them.
  return (
    <div className="mx-auto flex max-w-[1600px] flex-col gap-6 px-5 py-6 lg:flex-row lg:items-start">
      {/* Capped and scrollable on phones too: stacked, the full catalogue is
            taller than the screen and pushed the prompt box well below the
            fold. On desktop it becomes the sticky full-height column. */}
      <aside className="max-h-[45svh] w-full shrink-0 overflow-y-auto pr-2 lg:sticky lg:top-20 lg:max-h-[calc(100svh-6rem)] lg:w-[280px]">
        {loading ? (
          <ModelPickerSkeleton />
        ) : (
          <>
            <ModelPicker
              models={models}
              providers={providers}
              selected={selected}
              onSelect={selectModel}
            />
            {lockedCount > 0 && (
              <Link
                href="/settings/keys"
                className="mt-6 block rounded-lg border border-line p-3 text-xs leading-relaxed text-muted transition-colors hover:border-dim hover:bg-panel"
              >
                <span className="text-fg">{lockedCount} models locked.</span>{" "}
                Add your own provider keys to unlock them →
              </Link>
            )}
          </>
        )}
      </aside>

      <div className="flex min-w-0 flex-1 flex-col gap-6">
        <main className="min-w-0">
          <section className="relative overflow-hidden rounded-2xl border border-line bg-panel p-5">
            {/* A single soft wash off the accent, so the composer reads as the
              live surface without adding another competing bright colour. */}
            <div
              aria-hidden
              className="pointer-events-none absolute -top-24 left-1/2 h-48 w-[36rem] -translate-x-1/2 rounded-full bg-accent/10 blur-3xl"
            />

            <div className="relative">
              <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
                <div className="flex min-w-0 items-center gap-2">
                  <h2 className="truncate text-sm font-medium">
                    {selected ? selected.name : "Pick a model"}
                  </h2>
                  {selected && (
                    <span className="shrink-0 rounded-full bg-panel-2 px-2 py-0.5 font-mono text-[10px] uppercase tracking-wider text-dim">
                      {selected.provider_name}
                    </span>
                  )}
                </div>
                {selected && (
                  <span className="font-mono text-[11px] text-dim">
                    {selected.access === "byok"
                      ? "your key · no credits"
                      : `${selected.credit_cost} credits · ${credits} left`}
                  </span>
                )}
              </div>

              <textarea
                value={prompt}
                onChange={(e) => setPrompt(e.target.value)}
                onKeyDown={(e) => {
                  if ((e.metaKey || e.ctrlKey) && e.key === "Enter") submit();
                }}
                rows={3}
                placeholder={
                  selected?.modality === "audio"
                    ? "The script to speak…"
                    : "Describe the shot. Camera, subject, light, mood…"
                }
                className="w-full resize-y rounded-xl border border-line bg-void px-4 py-3 text-sm leading-relaxed outline-none transition-colors placeholder:text-dim focus:border-accent/50"
              />

              {selected?.params?.length ? (
                <div className="mt-4">
                  <ParamControls
                    params={selected.params}
                    values={params}
                    onChange={(k, v) => setParams((p) => ({ ...p, [k]: v }))}
                  />
                </div>
              ) : null}

              {error && (
                <p className="mt-4 rounded-lg border border-accent/40 bg-accent/10 px-3 py-2 text-xs leading-relaxed text-accent-soft">
                  {error}
                </p>
              )}

              <div className="mt-4 flex items-center justify-between gap-3">
                <span className="font-mono text-[10px] uppercase tracking-wider text-dim">
                  ⌘ + enter to run
                </span>
                <button
                  onClick={submit}
                  disabled={!selected || !prompt.trim() || busy}
                  className="rounded-full bg-accent px-6 py-2.5 text-sm font-medium text-void transition-all duration-150 hover:bg-accent-soft disabled:cursor-not-allowed disabled:bg-panel-2 disabled:text-dim enabled:hover:shadow-[0_0_24px_-6px_var(--color-accent)]"
                >
                  {busy ? "Starting…" : "Generate"}
                </button>
              </div>
            </div>
          </section>
        </main>

        {/* The newest few only, with the full history one click away. Loading
            shows the same card geometry rather than a spinner, so the section
            does not resize under the composer when the data lands. */}
        <section className="min-w-0">
          <div className="mb-3 flex flex-wrap items-center justify-between gap-3 border-t border-line pt-5">
            <div className="flex items-center gap-2.5">
              <h2 className="font-mono text-[11px] uppercase tracking-[0.18em] text-muted">
                Latest generations
              </h2>
              {activeCount > 0 && (
                <span className="inline-flex items-center gap-1.5 font-mono text-[10px] uppercase tracking-wider text-accent">
                  <span className="h-1.5 w-1.5 animate-pulse rounded-full bg-accent" />
                  {activeCount} running
                </span>
              )}
            </div>

            {jobs.length > 0 && (
              <Link
                href="/generations"
                className="font-mono text-[10px] uppercase tracking-wider text-dim transition-colors hover:text-fg"
              >
                View all →
              </Link>
            )}
          </div>

          {loading ? (
            <CardGridSkeleton count={RECENT} />
          ) : recent.length === 0 ? (
            <EmptyState />
          ) : (
            <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
              {recent.map((j) => (
                <JobCard key={j.id} job={j} onCancel={cancel} />
              ))}
            </div>
          )}
        </section>
      </div>
    </div>
  );
}

function EmptyState() {
  return (
    <div className="rounded-2xl border border-dashed border-line p-14 text-center">
      <svg
        width="26"
        height="26"
        viewBox="0 0 24 24"
        fill="none"
        aria-hidden
        className="mx-auto text-dim"
      >
        <rect
          x="2.75"
          y="4.75"
          width="18.5"
          height="14.5"
          rx="2.5"
          stroke="currentColor"
          strokeWidth="1.3"
        />
        <path
          d="m9.75 9.5 5 2.5-5 2.5z"
          stroke="currentColor"
          strokeWidth="1.3"
          strokeLinejoin="round"
        />
      </svg>
      <p className="mt-4 text-sm text-muted">Nothing here yet.</p>
      <p className="mt-1 text-xs text-dim">
        Sandbox models run free — try one to see the whole loop.
      </p>
    </div>
  );
}

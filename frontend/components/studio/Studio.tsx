"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import Link from "next/link";
import { api, ApiError } from "@/lib/api";
import type { Generation, Model } from "@/lib/types";
import { ModelPicker } from "./ModelPicker";
import { ParamControls } from "./ParamControls";
import { JobCard } from "./JobCard";

type Value = string | number | boolean;

const POLL_MS = 2500;

export function Studio({
  credits,
  onCreditsChange,
}: {
  credits: number;
  onCreditsChange: (n: number) => void;
}) {
  const [models, setModels] = useState<Model[]>([]);
  const [selected, setSelected] = useState<Model>();
  const [params, setParams] = useState<Record<string, Value>>({});
  const [prompt, setPrompt] = useState("");
  const [jobs, setJobs] = useState<Generation[]>([]);
  const [error, setError] = useState<string>();
  const [busy, setBusy] = useState(false);
  const [loading, setLoading] = useState(true);

  // Held in a ref so the polling effect does not restart on every job update.
  const jobsRef = useRef<Generation[]>([]);
  jobsRef.current = jobs;

  useEffect(() => {
    (async () => {
      try {
        const [catalog, history] = await Promise.all([
          api.models(),
          api.generations({ limit: "24" }),
        ]);
        setModels(catalog.models);
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

  const hasActive = jobs.some(
    (j) => j.status === "queued" || j.status === "running",
  );

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

  return (
    <div className="mx-auto grid max-w-[1600px] gap-6 px-5 py-6 lg:grid-cols-[260px_minmax(0,1fr)]">
      <aside className="lg:sticky lg:top-20 lg:h-[calc(100svh-6rem)] lg:overflow-y-auto lg:pr-2">
        {loading ? (
          <p className="font-mono text-xs text-dim">Loading models…</p>
        ) : (
          <>
            <ModelPicker
              models={models}
              selected={selected}
              onSelect={selectModel}
            />
            {lockedCount > 0 && (
              <Link
                href="/settings/keys"
                className="mt-6 block rounded-lg border border-line p-3 text-xs leading-relaxed text-muted transition-colors hover:border-dim"
              >
                <span className="text-fg">{lockedCount} models locked.</span>{" "}
                Add your own provider keys to unlock them →
              </Link>
            )}
          </>
        )}
      </aside>

      <main className="min-w-0 space-y-6">
        <section className="rounded-2xl border border-line bg-panel p-5">
          <div className="mb-3 flex items-center justify-between">
            <h2 className="text-sm font-medium">
              {selected ? selected.name : "Pick a model"}
            </h2>
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
            className="w-full resize-y rounded-xl border border-line bg-void px-4 py-3 text-sm leading-relaxed outline-none transition-colors placeholder:text-dim focus:border-dim"
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

          <div className="mt-4 flex items-center justify-between">
            <span className="font-mono text-[10px] uppercase tracking-wider text-dim">
              ⌘ + enter to run
            </span>
            <button
              onClick={submit}
              disabled={!selected || !prompt.trim() || busy}
              className="rounded-full bg-accent px-6 py-2.5 text-sm font-medium text-void transition-colors hover:bg-accent-soft disabled:cursor-not-allowed disabled:bg-panel-2 disabled:text-dim"
            >
              {busy ? "Starting…" : "Generate"}
            </button>
          </div>
        </section>

        <section>
          <h2 className="mb-3 font-mono text-[11px] uppercase tracking-[0.18em] text-dim">
            Your generations
          </h2>
          {jobs.length === 0 ? (
            <div className="rounded-2xl border border-dashed border-line p-14 text-center">
              <p className="text-sm text-muted">Nothing here yet.</p>
              <p className="mt-1 text-xs text-dim">
                Sandbox models run free — try one to see the whole loop.
              </p>
            </div>
          ) : (
            <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
              {jobs.map((j) => (
                <JobCard key={j.id} job={j} onCancel={cancel} />
              ))}
            </div>
          )}
        </section>
      </main>
    </div>
  );
}

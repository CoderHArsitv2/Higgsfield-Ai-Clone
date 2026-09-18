"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import Link from "next/link";
import { api, ApiError } from "@/lib/api";
import type { Generation, GenerationStatus } from "@/lib/types";
import { JobCard } from "./JobCard";
import { CardGridSkeleton } from "@/components/ui/Skeleton";

const PAGE = 12;
const POLL_MS = 2500;

type Filter = "all" | "active" | GenerationStatus;
const FILTERS: { id: Filter; label: string }[] = [
  { id: "all", label: "All" },
  { id: "active", label: "In flight" },
  { id: "succeeded", label: "Done" },
  { id: "failed", label: "Failed" },
  { id: "canceled", label: "Canceled" },
];

const isActive = (j: Generation) =>
  j.status === "queued" || j.status === "running";

/**
 * The full history, paged from the server.
 *
 * Studio keeps only the newest few so the composer stays above the fold; this
 * is where the rest lives. Filtering goes through the API rather than the
 * loaded page, so "Failed" means every failure on record, not just the failures
 * that happen to be in the first twelve rows.
 */
export function GenerationsBrowser() {
  const [jobs, setJobs] = useState<Generation[]>([]);
  const [total, setTotal] = useState(0);
  const [filter, setFilter] = useState<Filter>("all");
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState<string>();

  // Identifies the newest request so a slow response for an abandoned filter
  // cannot overwrite the results of the one the user is now looking at.
  const reqId = useRef(0);

  const query = useCallback((f: Filter, offset: number) => {
    const p: Record<string, string> = {
      limit: String(PAGE),
      offset: String(offset),
    };
    // "In flight" is two statuses, which the API filter cannot express, so it
    // is filtered client-side over an unfiltered page.
    if (f !== "all" && f !== "active") p.status = f;
    return p;
  }, []);

  const load = useCallback(
    async (f: Filter) => {
      const id = ++reqId.current;
      setLoading(true);
      setError(undefined);
      try {
        const res = await api.generations(query(f, 0));
        if (id !== reqId.current) return;
        setJobs(res.generations);
        setTotal(res.total);
      } catch (e) {
        if (id !== reqId.current) return;
        setError(
          e instanceof ApiError ? e.message : "Could not load your generations",
        );
      } finally {
        if (id === reqId.current) setLoading(false);
      }
    },
    [query],
  );

  useEffect(() => {
    load(filter);
  }, [filter, load]);

  const loadMore = async () => {
    setLoadingMore(true);
    try {
      const res = await api.generations(query(filter, jobs.length));
      setJobs((prev) => [...prev, ...res.generations]);
      setTotal(res.total);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Could not load more");
    } finally {
      setLoadingMore(false);
    }
  };

  // Refresh in place while anything on this page is still running.
  const hasActive = jobs.some(isActive);
  useEffect(() => {
    if (!hasActive) return;
    const t = setInterval(async () => {
      try {
        const res = await api.generations(query(filter, 0));
        setJobs((prev) => {
          const fresh = new Map(res.generations.map((g) => [g.id, g]));
          return prev.map((j) => fresh.get(j.id) ?? j);
        });
      } catch {
        // A dropped poll is not worth showing; the next tick retries.
      }
    }, POLL_MS);
    return () => clearInterval(t);
  }, [hasActive, filter, query]);

  const cancel = async (id: string) => {
    try {
      const { generation } = await api.cancel(id);
      setJobs((prev) => prev.map((j) => (j.id === id ? generation : j)));
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Could not cancel");
    }
  };

  const shown = filter === "active" ? jobs.filter(isActive) : jobs;
  const canLoadMore = filter !== "active" && jobs.length < total;

  return (
    <div>
      <div className="mb-5 flex flex-wrap items-center justify-between gap-3">
        <div className="flex flex-wrap gap-1.5">
          {FILTERS.map((f) => (
            <button
              key={f.id}
              onClick={() => setFilter(f.id)}
              aria-pressed={filter === f.id}
              className={`rounded-full border px-3 py-1.5 font-mono text-[10px] uppercase tracking-wider transition-colors ${
                filter === f.id
                  ? "border-accent/60 bg-accent/15 text-accent-soft"
                  : "border-line text-dim hover:border-dim hover:text-muted"
              }`}
            >
              {f.label}
            </button>
          ))}
        </div>
        {!loading && (
          <span className="font-mono text-[11px] text-dim">
            {shown.length} of {total}
          </span>
        )}
      </div>

      {error && (
        <p className="mb-4 rounded-lg border border-accent/40 bg-accent/10 px-3 py-2 text-xs leading-relaxed text-accent-soft">
          {error}
        </p>
      )}

      {loading ? (
        <CardGridSkeleton count={6} />
      ) : shown.length === 0 ? (
        <div className="rounded-2xl border border-dashed border-line p-16 text-center">
          <p className="text-sm text-muted">
            {filter === "all"
              ? "You have not generated anything yet."
              : "Nothing matches that filter."}
          </p>
          {filter === "all" && (
            <Link
              href="/studio"
              className="mt-3 inline-block text-xs text-accent transition-colors hover:text-accent-soft"
            >
              Go to the studio →
            </Link>
          )}
        </div>
      ) : (
        <>
          <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
            {shown.map((j) => (
              <JobCard key={j.id} job={j} onCancel={cancel} />
            ))}
          </div>

          {canLoadMore && (
            <div className="mt-7 flex justify-center">
              <button
                onClick={loadMore}
                disabled={loadingMore}
                className="inline-flex items-center gap-2 rounded-full border border-line px-5 py-2.5 text-sm text-muted transition-colors hover:border-dim hover:text-fg disabled:cursor-not-allowed disabled:opacity-60"
              >
                {loadingMore && (
                  <span className="h-3.5 w-3.5 animate-spin rounded-full border border-current border-t-transparent" />
                )}
                {loadingMore
                  ? "Loading…"
                  : `Load ${Math.min(PAGE, total - jobs.length)} more`}
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
}

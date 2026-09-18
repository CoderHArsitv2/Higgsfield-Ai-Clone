"use client";

import { useEffect, useState } from "react";
import { api, ApiError } from "@/lib/api";
import type { ProviderInfo, StoredKey } from "@/lib/types";

export function KeyManager() {
  const [providers, setProviders] = useState<ProviderInfo[]>([]);
  const [keys, setKeys] = useState<StoredKey[]>([]);
  const [drafts, setDrafts] = useState<Record<string, string>>({});
  const [busy, setBusy] = useState<string>();
  const [error, setError] = useState<string>();
  const [loading, setLoading] = useState(true);

  const load = async () => {
    try {
      const [catalog, stored] = await Promise.all([api.models(), api.keys()]);
      setProviders(catalog.providers.filter((p) => p.accepts_byok));
      setKeys(stored.keys);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Could not load your keys");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const save = async (providerId: string) => {
    const value = (drafts[providerId] ?? "").trim();
    if (!value) return;
    setBusy(providerId);
    setError(undefined);
    try {
      await api.saveKey(providerId, value);
      setDrafts((d) => ({ ...d, [providerId]: "" }));
      await load();
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Could not save that key");
    } finally {
      setBusy(undefined);
    }
  };

  const remove = async (providerId: string) => {
    setBusy(providerId);
    try {
      await api.deleteKey(providerId);
      await load();
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Could not remove that key");
    } finally {
      setBusy(undefined);
    }
  };

  if (loading) {
    // Same geometry as the real rows, so the list does not jump on arrival.
    return (
      <div
        className="space-y-3"
        aria-busy="true"
        aria-label="Loading providers"
      >
        {[0, 1, 2, 3].map((i) => (
          <div key={i} className="skeleton h-20 rounded-xl" />
        ))}
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {error && (
        <p className="rounded-lg border border-accent/40 bg-accent/10 px-3 py-2 text-xs text-accent-soft">
          {error}
        </p>
      )}

      {providers.map((p) => {
        const stored = keys.find((k) => k.provider_id === p.id);
        return (
          <section
            key={p.id}
            className="rounded-xl border border-line bg-panel p-5"
          >
            <div className="flex flex-wrap items-start justify-between gap-4">
              <div>
                <div className="flex items-center gap-2.5">
                  <h3 className="text-sm font-medium">{p.name}</h3>
                  {p.has_server_key && (
                    <span className="rounded-full border border-mint/40 px-2 py-0.5 font-mono text-[10px] uppercase tracking-wider text-mint">
                      platform key active
                    </span>
                  )}
                  {stored && (
                    <span className="rounded-full border border-iris/40 px-2 py-0.5 font-mono text-[10px] uppercase tracking-wider text-iris">
                      your key
                    </span>
                  )}
                </div>
                <p className="mt-1 font-mono text-[11px] text-dim">
                  {p.model_count} models · env {p.env_key}
                </p>
              </div>

              {p.docs_url && (
                <a
                  href={p.docs_url}
                  target="_blank"
                  rel="noreferrer"
                  className="font-mono text-[11px] text-dim underline-offset-4 transition-colors hover:text-fg hover:underline"
                >
                  get a key ↗
                </a>
              )}
            </div>

            <div className="mt-4 flex flex-wrap items-center gap-2">
              {stored ? (
                <>
                  <code className="rounded-lg border border-line bg-void px-3 py-2 font-mono text-xs text-muted">
                    {stored.preview}
                  </code>
                  <button
                    onClick={() => remove(p.id)}
                    disabled={busy === p.id}
                    className="rounded-lg border border-line px-3 py-2 text-xs text-muted transition-colors hover:border-accent/50 hover:text-accent disabled:opacity-50"
                  >
                    Remove
                  </button>
                </>
              ) : (
                <>
                  <input
                    type="password"
                    autoComplete="off"
                    value={drafts[p.id] ?? ""}
                    onChange={(e) =>
                      setDrafts((d) => ({ ...d, [p.id]: e.target.value }))
                    }
                    onKeyDown={(e) => e.key === "Enter" && save(p.id)}
                    placeholder={`Paste your ${p.name} key`}
                    className="min-w-0 flex-1 rounded-lg border border-line bg-void px-3 py-2 font-mono text-xs outline-none transition-colors placeholder:text-dim focus:border-dim"
                  />
                  <button
                    onClick={() => save(p.id)}
                    disabled={busy === p.id || !(drafts[p.id] ?? "").trim()}
                    className="rounded-lg bg-fg px-4 py-2 text-xs font-medium text-void transition-opacity disabled:opacity-40"
                  >
                    {busy === p.id ? "Saving…" : "Save"}
                  </button>
                </>
              )}
            </div>
          </section>
        );
      })}
    </div>
  );
}

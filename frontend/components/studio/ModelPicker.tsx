"use client";

import { useMemo, useState } from "react";
import type { Model, Modality, ProviderInfo } from "@/lib/types";

const ORDER: Modality[] = ["video", "image", "audio"];
const LABEL: Record<Modality, string> = {
  video: "Video",
  image: "Image",
  audio: "Voice",
};

/**
 * Models are grouped by provider rather than by modality.
 *
 * Modality is the cheaper question -- most people arrive knowing whether they
 * want video or stills -- so it is a filter across the top. Provider is the
 * expensive one, because it decides whether a model runs on our key, on yours,
 * or not at all, and those are the rows that share a fix: add one key and a
 * whole group unlocks together. Grouping by provider puts that shared fix next
 * to the models it unlocks.
 */
export function ModelPicker({
  models,
  providers,
  selected,
  onSelect,
}: {
  models: Model[];
  providers: ProviderInfo[];
  selected?: Model;
  onSelect: (m: Model) => void;
}) {
  const [modality, setModality] = useState<Modality | "all">("all");
  const [query, setQuery] = useState("");
  const [collapsed, setCollapsed] = useState<Record<string, boolean>>({});

  const counts = useMemo(() => {
    const c: Record<string, number> = { all: models.length };
    for (const m of models) c[m.modality] = (c[m.modality] ?? 0) + 1;
    return c;
  }, [models]);

  const visible = useMemo(() => {
    const q = query.trim().toLowerCase();
    return models.filter((m) => {
      if (modality !== "all" && m.modality !== modality) return false;
      if (!q) return true;
      return (
        m.name.toLowerCase().includes(q) ||
        m.provider_name.toLowerCase().includes(q) ||
        m.description?.toLowerCase().includes(q)
      );
    });
  }, [models, modality, query]);

  // Providers the API told us about, ordered so the ones you can actually run
  // float up. Any provider with no visible model after filtering is dropped.
  const groups = useMemo(() => {
    const byProvider = new Map<string, Model[]>();
    for (const m of visible) {
      const list = byProvider.get(m.provider_id);
      if (list) list.push(m);
      else byProvider.set(m.provider_id, [m]);
    }

    const name = new Map(providers.map((p) => [p.id, p.name]));
    return [...byProvider.entries()]
      .map(([id, list]) => ({
        id,
        name: name.get(id) ?? list[0]?.provider_name ?? id,
        models: list,
        ready: list.filter((m) => m.enabled).length,
      }))
      .sort((a, b) => {
        if (!!a.ready !== !!b.ready) return a.ready ? -1 : 1;
        return a.name.localeCompare(b.name);
      });
  }, [visible, providers]);

  return (
    <div className="space-y-4">
      <div className="relative">
        <svg
          width="13"
          height="13"
          viewBox="0 0 14 14"
          fill="none"
          aria-hidden
          className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-dim"
        >
          <circle
            cx="6"
            cy="6"
            r="4.25"
            stroke="currentColor"
            strokeWidth="1.3"
          />
          <path
            d="m9.5 9.5 3 3"
            stroke="currentColor"
            strokeWidth="1.3"
            strokeLinecap="round"
          />
        </svg>
        <input
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Search models"
          aria-label="Search models"
          className="w-full rounded-lg border border-line bg-void py-2 pl-8 pr-8 text-[13px] outline-none transition-colors placeholder:text-dim focus:border-dim"
        />
        {query && (
          <button
            onClick={() => setQuery("")}
            aria-label="Clear search"
            className="absolute right-2 top-1/2 -translate-y-1/2 rounded px-1.5 py-0.5 text-dim transition-colors hover:text-fg"
          >
            ×
          </button>
        )}
      </div>

      <div className="flex flex-wrap gap-1.5">
        {(["all", ...ORDER] as const).map((m) => {
          const on = modality === m;
          const n = counts[m] ?? 0;
          if (m !== "all" && n === 0) return null;
          return (
            <button
              key={m}
              onClick={() => setModality(m)}
              aria-pressed={on}
              className={`rounded-full border px-2.5 py-1 font-mono text-[10px] uppercase tracking-wider transition-colors ${
                on
                  ? "border-accent/60 bg-accent/15 text-accent-soft"
                  : "border-line text-dim hover:border-dim hover:text-muted"
              }`}
            >
              {m === "all" ? "All" : LABEL[m]} {n}
            </button>
          );
        })}
      </div>

      {groups.length === 0 ? (
        <p className="rounded-lg border border-dashed border-line px-3 py-6 text-center text-xs text-dim">
          No model matches “{query}”.
        </p>
      ) : (
        <div className="space-y-5">
          {groups.map((g) => {
            const isCollapsed = collapsed[g.id] ?? false;
            return (
              <section key={g.id}>
                <button
                  onClick={() =>
                    setCollapsed((c) => ({ ...c, [g.id]: !isCollapsed }))
                  }
                  aria-expanded={!isCollapsed}
                  className="group mb-2 flex w-full items-center gap-2 text-left"
                >
                  <svg
                    width="9"
                    height="9"
                    viewBox="0 0 10 10"
                    fill="none"
                    aria-hidden
                    className={`shrink-0 text-dim transition-transform duration-200 ${isCollapsed ? "" : "rotate-90"}`}
                  >
                    <path
                      d="m3.5 2 3.5 3-3.5 3"
                      stroke="currentColor"
                      strokeWidth="1.4"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                    />
                  </svg>
                  <h3 className="truncate font-mono text-[11px] uppercase tracking-[0.18em] text-muted transition-colors group-hover:text-fg">
                    {g.name}
                  </h3>
                  <span className="h-px flex-1 bg-line" />
                  <span
                    className={`shrink-0 font-mono text-[10px] ${g.ready ? "text-dim" : "text-dim/70"}`}
                    title={`${g.ready} of ${g.models.length} ready to run`}
                  >
                    {g.ready}/{g.models.length}
                  </span>
                </button>

                {!isCollapsed && (
                  <div className="space-y-1">
                    {g.models.map((m) => (
                      <ModelRow
                        key={m.id}
                        model={m}
                        active={selected?.id === m.id}
                        onSelect={onSelect}
                      />
                    ))}
                  </div>
                )}
              </section>
            );
          })}
        </div>
      )}
    </div>
  );
}

function ModelRow({
  model: m,
  active,
  onSelect,
}: {
  model: Model;
  active: boolean;
  onSelect: (m: Model) => void;
}) {
  return (
    <button
      onClick={() => m.enabled && onSelect(m)}
      disabled={!m.enabled}
      title={m.enabled ? m.description : m.lock_reason}
      aria-current={active}
      className={`relative w-full rounded-lg border px-3 py-2 text-left transition-all duration-150 ${
        active
          ? "border-accent/60 bg-accent/10"
          : m.enabled
            ? "border-transparent hover:border-line hover:bg-panel"
            : "cursor-not-allowed border-transparent opacity-45"
      }`}
    >
      {/* A selected model gets a bar rather than only a tint, so the current
          choice survives being scrolled to the edge of the list. */}
      {active && (
        <span className="absolute inset-y-1.5 left-0 w-0.5 rounded-full bg-accent" />
      )}

      <div className="flex items-center justify-between gap-2">
        <span className="truncate text-[13px] font-medium">{m.name}</span>
        {m.enabled ? (
          <span
            className={`shrink-0 rounded-full px-1.5 py-0.5 font-mono text-[10px] ${
              m.access === "byok"
                ? "bg-iris/15 text-iris"
                : "bg-panel-2 text-dim"
            }`}
          >
            {m.credit_cost > 0 && m.access !== "byok"
              ? `${m.credit_cost}c`
              : "own key"}
          </span>
        ) : (
          <svg
            width="11"
            height="11"
            viewBox="0 0 14 14"
            fill="none"
            aria-hidden
            className="shrink-0 text-dim"
          >
            <rect
              x="2.5"
              y="6"
              width="9"
              height="6.5"
              rx="1.5"
              stroke="currentColor"
              strokeWidth="1.3"
            />
            <path
              d="M4.75 6V4.25a2.25 2.25 0 0 1 4.5 0V6"
              stroke="currentColor"
              strokeWidth="1.3"
            />
          </svg>
        )}
      </div>

      {/* Provider used to live on this line, but the group heading already says
          it. The description is the thing you cannot get anywhere else. */}
      {m.description && (
        <p className="mt-0.5 truncate text-[11px] leading-relaxed text-dim">
          {m.enabled ? m.description : (m.lock_reason ?? m.description)}
        </p>
      )}
    </button>
  );
}

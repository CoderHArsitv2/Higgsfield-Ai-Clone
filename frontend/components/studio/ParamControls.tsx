"use client";

import type { ParamSpec } from "@/lib/types";

type Value = string | number | boolean;

/**
 * Controls are rendered from the model's own param spec rather than hardcoded
 * per model, so a new model with new options needs no frontend change.
 */
export function ParamControls({
  params,
  values,
  onChange,
}: {
  params: ParamSpec[];
  values: Record<string, Value>;
  onChange: (key: string, value: Value) => void;
}) {
  if (!params.length) return null;

  return (
    <div className="grid gap-3 sm:grid-cols-2">
      {params.map((p) => {
        const value = values[p.key] ?? (p.default as Value) ?? "";

        return (
          <label key={p.key} className="block">
            <span className="mb-1.5 block font-mono text-[10px] uppercase tracking-[0.15em] text-dim">
              {p.label}
              {p.optional && (
                <span className="ml-1 normal-case tracking-normal">
                  (optional)
                </span>
              )}
            </span>

            {p.kind === "select" && (
              <select
                value={String(value)}
                onChange={(e) => onChange(p.key, e.target.value)}
                className="w-full rounded-lg border border-line bg-panel px-3 py-2 text-sm outline-none transition-colors focus:border-dim"
              >
                {p.options?.map((o) => (
                  <option key={o.value} value={o.value}>
                    {o.label}
                  </option>
                ))}
              </select>
            )}

            {p.kind === "number" && (
              <input
                type="number"
                value={String(value)}
                min={p.min}
                max={p.max}
                step={p.step ?? 1}
                onChange={(e) =>
                  onChange(
                    p.key,
                    e.target.value === "" ? "" : Number(e.target.value),
                  )
                }
                className="w-full rounded-lg border border-line bg-panel px-3 py-2 text-sm outline-none transition-colors focus:border-dim"
              />
            )}

            {p.kind === "text" && (
              <input
                type="text"
                value={String(value)}
                placeholder={p.help}
                onChange={(e) => onChange(p.key, e.target.value)}
                className="w-full rounded-lg border border-line bg-panel px-3 py-2 text-sm outline-none transition-colors focus:border-dim"
              />
            )}

            {p.kind === "bool" && (
              <button
                type="button"
                onClick={() => onChange(p.key, !value)}
                className={`h-9 w-full rounded-lg border text-sm transition-colors ${
                  value
                    ? "border-accent/60 bg-accent/10"
                    : "border-line bg-panel text-muted"
                }`}
              >
                {value ? "On" : "Off"}
              </button>
            )}
          </label>
        );
      })}
    </div>
  );
}

"use client";

import type { Model } from "@/lib/types";

const ORDER: Model["modality"][] = ["video", "image", "audio"];
const LABEL: Record<Model["modality"], string> = {
  video: "Video",
  image: "Image",
  audio: "Voice",
};

export function ModelPicker({
  models,
  selected,
  onSelect,
}: {
  models: Model[];
  selected?: Model;
  onSelect: (m: Model) => void;
}) {
  return (
    <div className="space-y-7">
      {ORDER.map((modality) => {
        const group = models.filter((m) => m.modality === modality);
        if (!group.length) return null;
        const ready = group.filter((m) => m.enabled).length;

        return (
          <section key={modality}>
            <div className="mb-2.5 flex items-baseline justify-between">
              <h3 className="font-mono text-[11px] uppercase tracking-[0.18em] text-dim">
                {LABEL[modality]}
              </h3>
              <span className="font-mono text-[11px] text-dim">
                {ready}/{group.length}
              </span>
            </div>

            <div className="space-y-1">
              {group.map((m) => (
                <button
                  key={m.id}
                  onClick={() => m.enabled && onSelect(m)}
                  disabled={!m.enabled}
                  title={m.enabled ? m.description : m.lock_reason}
                  className={`w-full rounded-lg border px-3 py-2.5 text-left transition-colors ${
                    selected?.id === m.id
                      ? "border-accent/60 bg-accent/10"
                      : m.enabled
                        ? "border-transparent hover:border-line hover:bg-panel"
                        : "cursor-not-allowed border-transparent opacity-45"
                  }`}
                >
                  <div className="flex items-center justify-between gap-2">
                    <span className="truncate text-[13px] font-medium">{m.name}</span>
                    {m.enabled ? (
                      <span className="shrink-0 font-mono text-[10px] text-dim">
                        {m.credit_cost > 0 && m.access !== "byok" ? `${m.credit_cost}c` : "own key"}
                      </span>
                    ) : (
                      <svg width="11" height="11" viewBox="0 0 14 14" fill="none" className="shrink-0 text-dim">
                        <rect x="2.5" y="6" width="9" height="6.5" rx="1.5" stroke="currentColor" strokeWidth="1.3" />
                        <path d="M4.75 6V4.25a2.25 2.25 0 0 1 4.5 0V6" stroke="currentColor" strokeWidth="1.3" />
                      </svg>
                    )}
                  </div>
                  <p className="mt-0.5 truncate font-mono text-[10px] uppercase tracking-wider text-dim">
                    {m.provider_name}
                  </p>
                </button>
              ))}
            </div>
          </section>
        );
      })}
    </div>
  );
}

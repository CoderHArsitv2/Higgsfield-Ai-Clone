"use client";

import type { Generation } from "@/lib/types";

const STATUS: Record<string, { label: string; dot: string; text: string }> = {
  queued: { label: "Queued", dot: "bg-muted", text: "text-muted" },
  running: {
    label: "Running",
    dot: "bg-accent animate-pulse",
    text: "text-accent",
  },
  succeeded: { label: "Done", dot: "bg-mint", text: "text-mint" },
  failed: { label: "Failed", dot: "bg-accent", text: "text-accent" },
  canceled: { label: "Canceled", dot: "bg-dim", text: "text-dim" },
};

/** Short relative age -- "4m", "2h" -- so a full grid stays scannable. */
function age(iso: string): string {
  const s = Math.max(0, (Date.now() - new Date(iso).getTime()) / 1000);
  if (s < 60) return "just now";
  const m = Math.floor(s / 60);
  if (m < 60) return `${m}m ago`;
  const h = Math.floor(m / 60);
  if (h < 24) return `${h}h ago`;
  return `${Math.floor(h / 24)}d ago`;
}

export function JobCard({
  job,
  onCancel,
}: {
  job: Generation;
  onCancel: (id: string) => void;
}) {
  const active = job.status === "queued" || job.status === "running";
  const assets = job.assets ?? [];
  const status = STATUS[job.status] ?? STATUS.queued;
  const failed = job.status === "failed";

  return (
    <article
      className={`group relative flex flex-col overflow-hidden rounded-xl border bg-panel transition-colors duration-200 ${
        failed ? "border-accent/30" : "border-line hover:border-dim"
      }`}
    >
      {/* The asset is taken out of flow so it contributes no intrinsic height.
          Otherwise a portrait image sets this flex item's automatic minimum
          size, beats aspect-video, and pushes the caption out of the card. */}
      <div className="relative aspect-video min-h-0 overflow-hidden bg-void">
        {assets.length > 0 ? (
          <div className="absolute inset-0">
            <AssetView job={job} />
          </div>
        ) : (
          <div className="flex h-full flex-col items-center justify-center gap-3 px-6 text-center">
            {active ? (
              <>
                <div className="h-0.5 w-28 overflow-hidden rounded-full bg-line">
                  <div className="h-full w-1/3 animate-[marquee_1.2s_linear_infinite] bg-accent" />
                </div>
                <p className="font-mono text-[11px] uppercase tracking-wider text-muted">
                  {job.status === "queued" ? "Queued" : "Generating"}
                </p>
              </>
            ) : (
              <>
                {failed && (
                  <svg
                    width="18"
                    height="18"
                    viewBox="0 0 20 20"
                    fill="none"
                    aria-hidden
                    className="text-accent/70"
                  >
                    <circle
                      cx="10"
                      cy="10"
                      r="7.25"
                      stroke="currentColor"
                      strokeWidth="1.4"
                    />
                    <path
                      d="M10 6.25v4.5"
                      stroke="currentColor"
                      strokeWidth="1.4"
                      strokeLinecap="round"
                    />
                    <circle cx="10" cy="13.4" r="0.85" fill="currentColor" />
                  </svg>
                )}
                <p className="max-w-xs text-xs leading-relaxed text-dim">
                  {job.error || "No output"}
                </p>
              </>
            )}
          </div>
        )}

        {/* Status rides on the media so it is readable before you reach the
            caption, and stays out of the way until the card is hovered. */}
        <span className="pointer-events-none absolute left-2.5 top-2.5 inline-flex items-center gap-1.5 rounded-full border border-line/80 bg-void/80 px-2 py-0.5 font-mono text-[10px] uppercase tracking-wider backdrop-blur-sm">
          <span className={`h-1.5 w-1.5 rounded-full ${status.dot}`} />
          <span className={status.text}>{status.label}</span>
        </span>
      </div>

      <div className="flex flex-1 flex-col gap-2.5 p-4">
        <p className="line-clamp-2 text-[13px] leading-relaxed text-muted">
          {job.prompt}
        </p>

        <div className="mt-auto flex items-center justify-between gap-2 font-mono text-[10px] uppercase tracking-wider">
          <span className="truncate text-dim" title={job.model_id}>
            {job.model_id.split("/")[1] ?? job.model_id}
          </span>
          <div className="flex shrink-0 items-center gap-3">
            {job.used_own_key && <span className="text-iris">own key</span>}
            <span
              className="text-dim"
              title={new Date(job.created_at).toLocaleString()}
            >
              {age(job.created_at)}
            </span>
            {active && (
              <button
                onClick={() => onCancel(job.id)}
                className="text-dim transition-colors hover:text-fg"
              >
                cancel
              </button>
            )}
            {!active && assets[0] && (
              <a
                href={assets[0].url}
                target="_blank"
                rel="noreferrer"
                className="text-dim transition-colors hover:text-fg"
              >
                open
              </a>
            )}
          </div>
        </div>
      </div>
    </article>
  );
}

function AssetView({ job }: { job: Generation }) {
  const assets = job.assets ?? [];
  const first = assets[0];

  if (first.kind === "video") {
    return (
      <video
        src={first.url}
        poster={first.thumbnail_url}
        controls
        loop
        muted
        playsInline
        className="h-full w-full object-cover"
      />
    );
  }

  if (first.kind === "audio") {
    return (
      <div className="flex h-full items-center justify-center px-6">
        <audio src={first.url} controls className="w-full" />
      </div>
    );
  }

  return (
    <div
      className={`grid h-full auto-rows-fr ${assets.length > 1 ? "grid-cols-2" : "grid-cols-1"} gap-px`}
    >
      {assets.map((a) => (
        // eslint-disable-next-line @next/next/no-img-element
        <img
          key={a.id}
          src={a.thumbnail_url || a.url}
          alt={job.prompt}
          loading="lazy"
          className="h-full w-full object-cover transition-transform duration-500 ease-out group-hover:scale-[1.03]"
        />
      ))}
    </div>
  );
}

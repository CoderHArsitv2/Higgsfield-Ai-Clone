"use client";

import type { Generation } from "@/lib/types";

const STATUS_COLOR: Record<string, string> = {
  queued: "text-muted",
  running: "text-accent",
  succeeded: "text-mint",
  failed: "text-accent",
  canceled: "text-dim",
};

export function JobCard({
  job,
  onCancel,
}: {
  job: Generation;
  onCancel: (id: string) => void;
}) {
  const active = job.status === "queued" || job.status === "running";
  const assets = job.assets ?? [];

  return (
    <article className="overflow-hidden rounded-xl border border-line bg-panel">
      <div className="aspect-video bg-void">
        {assets.length > 0 ? (
          <AssetView job={job} />
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
              <p className="max-w-xs text-xs leading-relaxed text-dim">
                {job.error || "No output"}
              </p>
            )}
          </div>
        )}
      </div>

      <div className="space-y-2.5 p-4">
        <p className="line-clamp-2 text-[13px] leading-relaxed text-muted">{job.prompt}</p>

        <div className="flex items-center justify-between font-mono text-[10px] uppercase tracking-wider">
          <span className="text-dim">{job.model_id.split("/")[1] ?? job.model_id}</span>
          <div className="flex items-center gap-3">
            {job.used_own_key && <span className="text-iris">own key</span>}
            <span className={STATUS_COLOR[job.status]}>{job.status}</span>
            {active && (
              <button
                onClick={() => onCancel(job.id)}
                className="text-dim transition-colors hover:text-fg"
              >
                cancel
              </button>
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
    <div className={`grid h-full ${assets.length > 1 ? "grid-cols-2" : "grid-cols-1"} gap-px`}>
      {assets.map((a) => (
        // eslint-disable-next-line @next/next/no-img-element
        <img
          key={a.id}
          src={a.thumbnail_url || a.url}
          alt={job.prompt}
          loading="lazy"
          className="h-full w-full object-cover"
        />
      ))}
    </div>
  );
}

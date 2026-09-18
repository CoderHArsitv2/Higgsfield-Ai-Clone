import { ProviderIcon } from "./ProviderIcon";
import { LazyVideo } from "@/components/ui/LazyVideo";
import type { Connection } from "@/lib/connections";

/**
 * Our own sign-in screen, in place of Auth0's Universal Login.
 *
 * Every button is a plain link to `/auth/login?connection=…`, so this whole
 * screen needs no client JavaScript: the SDK forwards the connection to Auth0's
 * /authorize, which sends the user straight to that provider rather than to an
 * Auth0-hosted chooser.
 *
 * These are `<a>`, not `next/link`. /auth/login is served by middleware, not by
 * a route in this app, so there is nothing for a client-side navigation to do.
 * More importantly, `next/link` prefetches: every button visible on this screen
 * would silently start a login, and each one banks an encrypted `__txn_<state>`
 * cookie that is never completed. The SDK caps those cookies and evicts the
 * oldest, so the transaction from the button the user actually clicked can be
 * thrown away before Auth0 redirects back -- surfacing as "The state parameter
 * is invalid." on /auth/callback.
 */
export function AuthPanel({
  connections,
  emailConnection,
  returnTo,
  error,
}: {
  connections: Connection[];
  emailConnection: string;
  returnTo: string;
  error?: string;
}) {
  const href = (connection?: string) => {
    const q = new URLSearchParams({ returnTo });
    if (connection) q.set("connection", connection);
    return `/auth/login?${q}`;
  };

  return (
    <div className="grid w-full max-w-4xl overflow-hidden rounded-3xl border border-line bg-panel shadow-2xl shadow-black/70 md:grid-cols-2">
      {/* Left: the product doing the thing you are signing in to do. */}
      <aside className="relative hidden min-h-[520px] md:block">
        <LazyVideo
          src="/showcase/rain-city.mp4"
          poster="/showcase/rain-city.jpg"
          className="absolute inset-0 h-full w-full object-cover"
          eager
        />
        <div className="absolute inset-0 bg-gradient-to-t from-void via-void/20 to-transparent" />
        <div className="absolute inset-x-0 bottom-0 p-7">
          <span className="inline-flex items-center gap-1.5 rounded-full border border-line bg-void/70 px-2.5 py-1 font-mono text-[10px] uppercase tracking-wider text-muted">
            Kling 3.0 · 1080p
          </span>
          <h2 className="mt-4 font-display text-3xl leading-tight tracking-tight">
            Rain alley, one prompt
          </h2>
          <p className="mt-2 max-w-xs text-sm leading-relaxed text-muted">
            34 models behind a single prompt bar. Start on the sandbox — no key,
            no card.
          </p>
        </div>
      </aside>

      {/* Right: the actual choice. */}
      <div className="flex flex-col justify-center px-7 py-10 sm:px-10">
        <div className="mb-8 text-center">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img
            src="/logo-mark.svg"
            alt=""
            width={40}
            height={40}
            className="mx-auto h-10 w-10"
          />
          <h1 className="mt-5 font-display text-3xl tracking-tight">
            Welcome to Aperture
          </h1>
          <p className="mt-2 text-sm text-muted">
            Sign in and generate for free
          </p>
        </div>

        {error && (
          <p className="mb-5 rounded-xl border border-accent/40 bg-accent/10 px-4 py-3 text-xs leading-relaxed text-accent-soft">
            {error}
          </p>
        )}

        <div className="space-y-2.5">
          {connections.map((c) => (
            <a
              key={c.id}
              href={href(c.id)}
              className="flex h-12 items-center justify-center gap-3 rounded-xl border border-line bg-void/40 text-sm font-medium transition-colors hover:border-dim hover:bg-void"
            >
              <ProviderIcon name={c.icon} />
              Continue with {c.label}
            </a>
          ))}
        </div>

        <div className="my-6 flex items-center gap-4">
          <span className="h-px flex-1 bg-line" />
          <span className="font-mono text-[10px] uppercase tracking-[0.18em] text-dim">
            or
          </span>
          <span className="h-px flex-1 bg-line" />
        </div>

        <a
          href={href(emailConnection || undefined)}
          className="flex h-12 items-center justify-center gap-3 rounded-xl border border-line bg-void/40 text-sm font-medium transition-colors hover:border-dim hover:bg-void"
        >
          <svg
            width="18"
            height="18"
            viewBox="0 0 18 18"
            fill="none"
            stroke="currentColor"
            strokeWidth="1.5"
            aria-hidden
          >
            <rect x="1.75" y="3.75" width="14.5" height="10.5" rx="2" />
            <path d="m2.5 5 6.5 4.5L15.5 5" />
          </svg>
          Continue with Email
        </a>

        <p className="mt-8 text-center text-xs leading-relaxed text-dim">
          Authentication is handled by Auth0. We never see or store a password.
        </p>
      </div>
    </div>
  );
}

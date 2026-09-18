"use client";

import Link, { useLinkStatus } from "next/link";
import { usePathname } from "next/navigation";

const TABS = [
  { href: "/studio", label: "Studio" },
  { href: "/generations", label: "Generations" },
  { href: "/settings/keys", label: "Keys" },
];

/**
 * Nav tabs report their own pending state.
 *
 * Moving between tabs renders on the server, so there is a real wait before
 * anything changes. Without a signal the app looks like it ignored the click
 * and people click again. `useLinkStatus` is scoped to the enclosing Link, so
 * the spinner appears on the tab actually being navigated to.
 */
function TabPending() {
  const { pending } = useLinkStatus();
  if (!pending) return null;
  return (
    <span
      role="status"
      aria-label="Loading"
      className="ml-1.5 inline-block h-3 w-3 animate-spin rounded-full border border-current border-t-transparent align-[-1px] opacity-70"
    />
  );
}

function Tab({ href, label }: { href: string; label: string }) {
  const pathname = usePathname();
  const active = pathname === href || pathname.startsWith(`${href}/`);

  return (
    <Link
      href={href}
      aria-current={active ? "page" : undefined}
      className={`relative inline-flex items-center py-4 transition-colors ${
        active ? "text-fg" : "text-muted hover:text-fg"
      }`}
    >
      {label}
      <TabPending />
      {active && (
        <span className="absolute inset-x-0 bottom-0 h-px bg-accent" />
      )}
    </Link>
  );
}

export function AppNav({
  credits,
  spent,
  refunded,
  name,
  picture,
}: {
  credits: number;
  spent?: number;
  refunded?: number;
  name?: string;
  picture?: string;
}) {
  return (
    <header className="sticky top-0 z-40 border-b border-line bg-void/85 backdrop-blur-xl">
      <div className="mx-auto flex h-14 max-w-[1600px] items-center justify-between px-5">
        <div className="flex items-center gap-7">
          <Link href="/" className="flex items-center gap-2.5 text-sm">
            {/* A 600-byte SVG: next/image would add a request and an optimizer
                pass for no benefit, and it does not optimise SVG anyway.
                Served straight from public/ so the mark has one source of truth. */}
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src="/logo-mark.svg"
              alt=""
              width={24}
              height={24}
              className="h-6 w-6"
            />
            <span className="font-medium">Aperture</span>
          </Link>
          <nav className="flex items-center gap-5 text-sm">
            {TABS.map((t) => (
              <Tab key={t.href} {...t} />
            ))}
          </nav>
        </div>

        <div className="flex items-center gap-4">
          <span
            className="rounded-full border border-line px-3 py-1 font-mono text-[11px] text-muted"
            title={[
              "Credits are only spent on models running on platform keys.",
              spent !== undefined ? `${spent} spent in total` : null,
              refunded ? `${refunded} refunded from failed jobs` : null,
            ]
              .filter(Boolean)
              .join("\n")}
          >
            {credits} credits
          </span>
          {picture ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img
              src={picture}
              alt={name ?? "You"}
              className="h-7 w-7 rounded-full"
            />
          ) : (
            <span className="grid h-7 w-7 place-items-center rounded-full bg-panel-2 text-xs">
              {(name ?? "?").charAt(0).toUpperCase()}
            </span>
          )}
          <a
            href="/auth/logout"
            className="text-sm text-dim transition-colors hover:text-fg"
          >
            Sign out
          </a>
        </div>
      </div>
    </header>
  );
}

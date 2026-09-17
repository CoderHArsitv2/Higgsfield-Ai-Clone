import Link from "next/link";

export function AppNav({
  credits,
  name,
  picture,
}: {
  credits: number;
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
          <nav className="flex items-center gap-5 text-sm text-muted">
            <Link href="/studio" className="transition-colors hover:text-fg">
              Studio
            </Link>
            <Link
              href="/settings/keys"
              className="transition-colors hover:text-fg"
            >
              Keys
            </Link>
          </nav>
        </div>

        <div className="flex items-center gap-4">
          <span
            className="rounded-full border border-line px-3 py-1 font-mono text-[11px] text-muted"
            title="Credits are only spent on models running on platform keys"
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

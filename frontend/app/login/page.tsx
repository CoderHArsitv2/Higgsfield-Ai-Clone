import Link from "next/link";
import { redirect } from "next/navigation";
import { auth0 } from "@/lib/auth0";
import { socialConnections, emailConnection } from "@/lib/connections";
import { AuthPanel } from "@/components/auth/AuthPanel";

export const metadata = { title: "Sign in — Aperture" };

/** Only allow relative paths, so ?returnTo cannot bounce a user off-site. */
function safeReturnTo(value?: string) {
  if (!value || !value.startsWith("/") || value.startsWith("//"))
    return "/studio";
  return value;
}

export default async function LoginPage({
  searchParams,
}: {
  searchParams: Promise<{ returnTo?: string; error?: string }>;
}) {
  const params = await searchParams;
  const returnTo = safeReturnTo(params.returnTo);

  // Already signed in: skip the screen entirely.
  const session = await auth0.getSession();
  if (session) redirect(returnTo);

  return (
    <main className="relative flex min-h-[100svh] items-center justify-center px-5 py-10">
      <div
        aria-hidden
        className="pointer-events-none absolute left-1/2 top-1/3 h-[60vh] w-[70vw] -translate-x-1/2 -translate-y-1/2 rounded-full opacity-30 blur-[130px]"
        style={{
          background:
            "radial-gradient(circle at 35% 35%, #ff5a36 0%, transparent 55%), radial-gradient(circle at 65% 65%, #8a7cff 0%, transparent 55%)",
        }}
      />

      <Link
        href="/"
        className="absolute left-6 top-6 z-10 flex items-center gap-2.5 text-sm text-muted transition-colors hover:text-fg"
      >
        <svg
          width="14"
          height="14"
          viewBox="0 0 14 14"
          fill="none"
          stroke="currentColor"
          strokeWidth="1.5"
          strokeLinecap="round"
          aria-hidden
        >
          <path d="M13 7H2M6.5 2.5 2 7l4.5 4.5" />
        </svg>
        Back
      </Link>

      <div className="relative w-full max-w-4xl">
        <AuthPanel
          connections={socialConnections()}
          emailConnection={emailConnection()}
          returnTo={returnTo}
          error={params.error}
        />
      </div>
    </main>
  );
}

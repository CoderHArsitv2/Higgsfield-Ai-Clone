import { redirect } from "next/navigation";
import { auth0 } from "@/lib/auth0";
import { getMe } from "@/lib/server-api";
import { AppNav } from "@/components/studio/AppNav";
import { KeyManager } from "@/components/studio/KeyManager";

export const metadata = { title: "Provider keys — Aperture" };

export default async function KeysPage() {
  const session = await auth0.getSession();
  if (!session) redirect("/login?returnTo=/settings/keys");

  const me = await getMe();

  return (
    <>
      <AppNav
        credits={me?.user.credits ?? 0}
        name={me?.user.name || session.user.name || session.user.email}
        picture={me?.user.picture || session.user.picture}
      />

      <main className="mx-auto max-w-3xl px-5 py-10">
        <h1 className="font-display text-4xl tracking-tight">Provider keys</h1>
        <p className="mt-3 max-w-xl text-sm leading-relaxed text-muted">
          Add a key and every model behind that provider unlocks for your
          account only. Jobs run on your own key are billed by the provider and
          spend no credits here.
        </p>
        <p className="mt-3 max-w-xl text-xs leading-relaxed text-dim">
          Keys are encrypted with AES-256-GCM before storage and are never
          returned to the browser — this page only ever shows a masked preview.
        </p>

        <div className="mt-8">
          <KeyManager />
        </div>
      </main>
    </>
  );
}

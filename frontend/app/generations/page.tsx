import { redirect } from "next/navigation";
import { auth0 } from "@/lib/auth0";
import { getMe } from "@/lib/server-api";
import { AppNav } from "@/components/studio/AppNav";
import { GenerationsBrowser } from "@/components/studio/GenerationsBrowser";

export const metadata = { title: "Generations — Aperture" };

export default async function GenerationsPage() {
  const session = await auth0.getSession();
  if (!session) redirect("/login?returnTo=/generations");

  const me = await getMe();

  return (
    <>
      <AppNav
        credits={me?.user.credits ?? 0}
        spent={me?.user.credits_spent}
        refunded={me?.user.credits_refunded}
        name={me?.user.name || session.user.name || session.user.email}
        picture={me?.user.picture || session.user.picture}
      />

      <main className="mx-auto max-w-[1600px] px-5 py-8">
        <h1 className="font-display text-4xl tracking-tight">Generations</h1>
        <p className="mt-2.5 max-w-xl text-sm leading-relaxed text-muted">
          Everything you have run, newest first. The studio keeps the latest few
          next to the composer; the full history lives here.
        </p>

        <div className="mt-8">
          <GenerationsBrowser />
        </div>
      </main>
    </>
  );
}

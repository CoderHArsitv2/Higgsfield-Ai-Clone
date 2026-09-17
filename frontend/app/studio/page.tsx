import { redirect } from "next/navigation";
import { auth0 } from "@/lib/auth0";
import { getMe } from "@/lib/server-api";
import { StudioShell } from "@/components/studio/StudioShell";

export const metadata = { title: "Studio — Aperture" };

export default async function StudioPage() {
  const session = await auth0.getSession();
  if (!session) redirect("/login?returnTo=/studio");

  // The Auth0 session gives us a name and picture immediately; credits come
  // from our own API and are the one value worth waiting for.
  const me = await getMe();

  return (
    <StudioShell
      initialCredits={me?.user.credits ?? 0}
      spent={me?.user.credits_spent}
      refunded={me?.user.credits_refunded}
      name={me?.user.name || session.user.name || session.user.email}
      picture={me?.user.picture || session.user.picture}
    />
  );
}

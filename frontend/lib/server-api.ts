import "server-only";
import { auth0 } from "./auth0";
import type { User } from "./types";

const API = process.env.BACKEND_URL ?? "http://localhost:8080";

/**
 * Server-side call to the Go API using the session's access token. Used for the
 * first paint of authenticated pages so the shell renders with real data
 * instead of flashing placeholder values.
 */
export async function getMe(): Promise<{ user: User; byok: string[] } | null> {
  try {
    const { token } = await auth0.getAccessToken();
    const res = await fetch(`${API}/api/v1/me`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: "no-store",
      signal: AbortSignal.timeout(6000),
    });
    if (!res.ok) return null;
    const data = (await res.json()) as { user: User; byok_providers: string[] };
    return { user: data.user, byok: data.byok_providers ?? [] };
  } catch {
    return null;
  }
}

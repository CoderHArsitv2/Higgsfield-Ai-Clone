import "server-only";

/**
 * Identity connections offered on our own sign-in screen.
 *
 * Each button links to `/auth/login?connection=<id>`, which the SDK forwards to
 * Auth0's /authorize. Naming the connection skips Auth0's Universal Login
 * chooser entirely, so the user lands directly on Google's account picker
 * rather than an intermediate page.
 *
 * A connection only works once it is enabled in the Auth0 dashboard
 * (Authentication → Social) AND switched on for this application. Clicking a
 * button for a connection that is not enabled returns an Auth0 error, so the
 * list is env-driven and defaults to what a fresh tenant actually has on.
 */
export interface Connection {
  id: string;
  label: string;
  /** Inline brand mark. Kept as data so the button component stays dumb. */
  icon: "google" | "microsoft" | "apple" | "github" | "generic";
}

const KNOWN: Record<string, Connection> = {
  "google-oauth2": { id: "google-oauth2", label: "Google", icon: "google" },
  windowslive: { id: "windowslive", label: "Microsoft", icon: "microsoft" },
  apple: { id: "apple", label: "Apple", icon: "apple" },
  github: { id: "github", label: "GitHub", icon: "github" },
};

/**
 * Defaults to Google alone: it is the one social connection a new Auth0 tenant
 * enables out of the box. Add more only once they are configured, or the screen
 * ships with buttons that fail.
 *
 *   AUTH0_CONNECTIONS=google-oauth2,windowslive,apple
 */
export function socialConnections(): Connection[] {
  const raw = process.env.AUTH0_CONNECTIONS || "google-oauth2";
  return raw
    .split(",")
    .map((id) => id.trim())
    .filter(Boolean)
    .map((id) => KNOWN[id] ?? { id, label: id, icon: "generic" as const });
}

/**
 * The database connection used for the email option. Empty string means fall
 * back to Universal Login with no connection named, which lets Auth0 decide.
 */
export function emailConnection(): string {
  return (
    process.env.AUTH0_EMAIL_CONNECTION ?? "Username-Password-Authentication"
  );
}

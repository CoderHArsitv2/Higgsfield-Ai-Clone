import { Auth0Client } from "@auth0/nextjs-auth0/server";

/**
 * Auth0 is the only identity provider in this app -- there is no local
 * password flow to fall back to, by design.
 */

/**
 * Where Auth0 sends the user back to after login.
 *
 * Three levels, most specific first:
 *
 *   AUTH0_REDIRECT_URI   full URL, wins over everything. Use this when the
 *                        origin the browser sees is not the origin this
 *                        process knows about -- behind ngrok, a tunnel, or a
 *                        proxy.
 *   APP_BASE_URL         origin(s) the app is served from. The SDK appends the
 *                        callback path. Accepts a comma-separated list, so
 *                        localhost and a tunnel can both be valid without a
 *                        code change.
 *   AUTH0_CALLBACK_PATH  the path itself, if /auth/callback is taken.
 *
 * Whatever this resolves to must appear *verbatim* in the Auth0 application's
 * "Allowed Callback URLs". A mismatch is rejected by Auth0 before login, with
 * a callback-url-mismatch error naming both the attempted and allowed values.
 */
const callbackPath = process.env.AUTH0_CALLBACK_PATH || "/auth/callback";
const redirectUri = process.env.AUTH0_REDIRECT_URI?.trim() || undefined;

// The SDK documents APP_BASE_URL as accepting a comma-separated list, but its
// constructor runs `new URL()` over the raw string and throws "Invalid URL" on
// the comma. Split it here and pass an array, which the option type accepts.
const origins = (process.env.APP_BASE_URL || "")
  .split(",")
  .map((o) => o.trim().replace(/\/$/, ""))
  .filter(Boolean);
const appBaseUrl =
  origins.length > 1 ? origins : origins.length === 1 ? origins[0] : undefined;

export const auth0 = new Auth0Client({
  appBaseUrl,
  routes: { callback: callbackPath },
  authorizationParameters: {
    scope: process.env.AUTH0_SCOPE || "openid profile email offline_access",
    // Without an audience Auth0 issues an opaque token the Go API cannot
    // verify. This must match AUTH0_AUDIENCE on the backend exactly, and must
    // be a custom API you created -- not Auth0's own Management API.
    audience: process.env.AUTH0_AUDIENCE,
    ...(redirectUri ? { redirect_uri: redirectUri } : {}),
  },
});

/**
 * The callback URL this process will actually send, for logging and for the
 * setup check. Exported so a misconfiguration is visible without having to
 * read an Auth0 tenant log to discover it.
 */
export function resolvedCallbackUrl(): string {
  if (redirectUri) return redirectUri;
  const first = origins[0];
  return first
    ? `${first}${callbackPath}`
    : `<APP_BASE_URL unset>${callbackPath}`;
}

// Print it once in development. A callback mismatch otherwise only shows up as
// a generic failure in the browser and a log entry inside the Auth0 tenant.
if (process.env.NODE_ENV !== "production") {
  console.info(
    `[auth0] callback URL: ${resolvedCallbackUrl()}  ` +
      `(this exact string must be in Auth0 → Allowed Callback URLs)`,
  );
}

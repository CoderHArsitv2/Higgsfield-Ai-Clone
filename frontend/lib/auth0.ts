import { Auth0Client } from "@auth0/nextjs-auth0/server";

/**
 * Auth0 is the only identity provider in this app -- there is no local
 * password flow to fall back to, by design.
 *
 * The `audience` matters: without it Auth0 issues an opaque token that the Go
 * API cannot verify. Asking for the API audience is what makes the returned
 * access token a JWT the backend can validate against the tenant JWKS.
 */
export const auth0 = new Auth0Client({
  authorizationParameters: {
    scope: "openid profile email offline_access",
    audience: process.env.AUTH0_AUDIENCE,
  },
});
